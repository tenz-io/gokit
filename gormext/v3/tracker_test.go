package gormext

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/tenz-io/gokit/logger/v3"
	"github.com/tenz-io/gokit/monitor/v3"
	"github.com/tenz-io/gokit/tracer/v3"
)

type testUser struct {
	ID       int64  `gorm:"primaryKey"`
	Username string `gorm:"column:username;unique"`
	Password string `gorm:"column:password"`
}

func (testUser) TableName() string { return "test_user" }

// recordingExporter 是 monitor.Exporter 的录制型测试替身。它把每次调用
// (Incr/Decr/Count/Observe) 记录成一条 call,便于断言回调实际触发了哪些
// 指标,而不是只断言"没 panic"。
type recordingExporter struct {
	mu    sync.Mutex
	calls []metricCall
}

type metricCall struct {
	method string // "Incr"/"Decr"/"Count"/"Observe"
	dsCmd  string
	code   string
	opt    string
	millis float64
}

func (r *recordingExporter) Cmd() string { return "rec" }

func (r *recordingExporter) Set(_ context.Context, _, _ string, _ float64, _ string) {}

func (r *recordingExporter) Incr(_ context.Context, dsCmd, code, opt string) {
	r.record("Incr", dsCmd, code, opt, 0)
}

func (r *recordingExporter) Decr(_ context.Context, dsCmd, code, opt string) {
	r.record("Decr", dsCmd, code, opt, 0)
}

func (r *recordingExporter) Count(_ context.Context, dsCmd, code, opt string) {
	r.record("Count", dsCmd, code, opt, 0)
}

func (r *recordingExporter) CountDelta(_ context.Context, dsCmd, code string, _ uint64, opt string) {
	r.record("CountDelta", dsCmd, code, opt, 0)
}

func (r *recordingExporter) Observe(_ context.Context, dsCmd, code string, millis float64) {
	r.record("Observe", dsCmd, code, "", millis)
}

func (r *recordingExporter) Sample(_ context.Context, _, _ string, _ float64, _ string) {}

func (r *recordingExporter) record(method, dsCmd, code, opt string, millis float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, metricCall{method, dsCmd, code, opt, millis})
}

// countByMethod 返回给定 method 的调用次数 (线程安全)。
func (r *recordingExporter) countByMethod(method string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, c := range r.calls {
		if c.method == method {
			n++
		}
	}
	return n
}

// hasObserve 报告是否发生过一次 (dsCmd, code) 的 Observe 调用。
func (r *recordingExporter) hasObserve(dsCmd, code string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.calls {
		if c.method == "Observe" && c.dsCmd == dsCmd && c.code == code {
			return true
		}
	}
	return false
}

// newDB 打开一个 sqlite 内存库并迁移测试表,返回前清空 test_user。
//
// 用 "file::memory:?cache=shared" 而非裸 ":memory:":后者在 gorm 默认连接
// 池下每次取连接会拿到一个独立的空库,连接池切换后出现 no such table。
// cache=shared 让所有连接共享同一内存库,保证 AutoMigrate 的表对后续
// 查询可见 —— 但代价是测试串行运行时共享同一内存库,因此这里在返回前
// DELETE FROM test_user,确保每个测试拿到一张干净表 (否则唯一约束会因
// 上一个测试残留的行而误报冲突)。
func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	// 兜底:即便 shared cache 出现意外,单连接也保证表可见。
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&testUser{}))
	// 共享内存库:清空表,让每个测试拿到干净数据。
	require.NoError(t, db.Exec("DELETE FROM test_user").Error)
	return db
}

// setupLogger 把全局 logger 指向一个临时目录,并打开 traffic,以便测试可以
// 在结束后读 traffic.log 断言 SQL span 是否落盘。返回 teardown 给
// lumberjack 的落盘一拍时间冲刷 (logger/v3 不导出 Sync,沿用 httpext/v3
// 的 sleep 做法)。注意:logger 是进程级单例,这些测试串行运行,互不并行。
func setupLogger(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	logger.ConfigureWithOpts(
		logger.WithLevel(logger.DebugLevel),
		logger.WithConsole(false),
		logger.WithFilePath(dir),
		logger.WithTraffic(true),
		logger.WithTrafficPath(dir),
	)
	teardown := func() {
		// lumberjack 的 Write 是同步落盘,但 zap 编码器的一拍缓冲需要一点
		// 时间;给 100ms 与 httpext/v3 测试一致。
		time.Sleep(100 * time.Millisecond)
	}
	return dir, teardown
}

// readTraffic 读取 dir 下的 traffic.log,返回全部内容。
func readTraffic(t *testing.T, dir string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(dir, "traffic.log"))
	if err != nil {
		t.Fatalf("read traffic.log: %v", err)
	}
	return string(body)
}

// readLevelLog 读取 dir 下某个级别日志文件的内容。logger/v3 按
// <level>.log 拆分文件,因此可按级别断言 errorLog/slowLog 是否落地。
func readLevelLog(t *testing.T, dir, level string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(dir, level+".log"))
	if err != nil {
		// 文件不存在视作空:该级别从未写过。
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("read %s.log: %v", level, err)
	}
	return string(body)
}

// TestTracker_Apply_NilDB 验证 nil db 为 no-op,不 panic。
func TestTracker_Apply_NilDB(t *testing.T) {
	tr := NewTracker(Config{})
	assert.NoError(t, tr.Apply(nil))
}

// TestTracker_Apply_NilOptionIsSafe 验证 NewTrackerWithOpts 跳过 nil
// option 而非 panic。
func TestTracker_Apply_NilOptionIsSafe(t *testing.T) {
	tr := NewTrackerWithOpts(
		WithEnableTraffic(true),
		nil, // 被安全跳过
		WithEnableMetrics(false),
	)
	assert.NotNil(t, tr)
	assert.NoError(t, tr.Apply(nil))
}

// TestTracker_NewTracker_NormalizesNegativeFloor 验证负 SlowLogFloor 被规范
// 化为 0 (关闭慢查询),而非静默等价关闭。
func TestTracker_NewTracker_NormalizesNegativeFloor(t *testing.T) {
	tr := NewTracker(Config{SlowLogFloor: -1 * time.Second}).(*tracker)
	assert.Equal(t, time.Duration(0), tr.config.SlowLogFloor)
}

// TestTracker_Apply_RepeatCalls 验证对同一 db 重复 Apply 不报错也不会
// 改变可观察行为。gorm 对重名回调只发 warn 并按编译保留两份 handler
// (非替换),因此后续查询仍成功;v3 的 Apply 把"重复调用"当作 init 阶段
// 一次性调用,不返回 error。
func TestTracker_Apply_RepeatCalls(t *testing.T) {
	_, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTracker(Config{EnableTraffic: false, EnableMetrics: false, EnableErrorLog: false})
	require.NoError(t, tr.Apply(db))
	require.NoError(t, tr.Apply(db)) // gorm 去重,不报错

	require.NoError(t, db.WithContext(context.Background()).Create(&testUser{Username: "x", Password: "p"}).Error)
	var u testUser
	require.NoError(t, db.WithContext(context.Background()).First(&u, "username = ?", "x").Error)
}

// TestTracker_Traffic_WritesSQLSpanNoVarsByDefault 验证默认 (EnableVars=
// false) 时 traffic.log 写了 SQL span,但**不含**绑定参数值 (防敏感泄露)。
func TestTracker_Traffic_WritesSQLSpanNoVarsByDefault(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableTraffic(true),
		WithEnableMetrics(false),
		WithEnableErrorLog(false),
		WithSlowLogFloor(0),
		// 注意:不设 WithEnableVars(true),参数默认关闭。
	)
	require.NoError(t, tr.Apply(db))

	ctx := context.Background()
	require.NoError(t, db.WithContext(ctx).Create(&testUser{Username: "sky", Password: "sky123"}).Error)
	var u testUser
	require.NoError(t, db.WithContext(ctx).Where("username = ?", "sky").First(&u).Error)

	body := readTraffic(t, dir)
	assert.Contains(t, body, "db_create")
	assert.Contains(t, body, "db_query")
	assert.Contains(t, body, "INSERT")
	assert.Contains(t, body, "SELECT")
	// SQL 占位 ? 出现在语句中,但绑定变量值 "sky"/"sky123" 不应出现。
	assert.Contains(t, body, "?")
	assert.NotContains(t, body, "sky123", "密码不应明文进 traffic.log")
	// vars 字段为 null (EnableVars=false)。
	assert.Contains(t, body, `"vars": null`)
}

// TestTracker_Traffic_VarsRecordedWhenEnabled 验证 WithEnableVars(true) 时
// 参数会进 traffic.log。
func TestTracker_Traffic_VarsRecordedWhenEnabled(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableTraffic(true),
		WithEnableVars(true),
		WithSlowLogFloor(0),
	)
	require.NoError(t, tr.Apply(db))

	ctx := context.Background()
	require.NoError(t, db.WithContext(ctx).Create(&testUser{Username: "alice", Password: "secret"}).Error)

	body := readTraffic(t, dir)
	assert.Contains(t, body, "db_create")
	assert.Contains(t, body, "alice")
	assert.Contains(t, body, "secret")
}

// TestTracker_Traffic_RedactorScrubsParams 验证 WithRedactor 对参数脱敏:
// 密码值被替换为 ***,不出现在 traffic.log。
func TestTracker_Traffic_RedactorScrubsParams(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	redactor := func(sql string, vars []any) (string, []any) {
		out := make([]any, len(vars))
		for i, v := range vars {
			switch v.(type) {
			case string:
				out[i] = "***"
			default:
				out[i] = v
			}
		}
		return sql, out
	}
	tr := NewTrackerWithOpts(
		WithEnableTraffic(true),
		WithEnableVars(true),
		WithRedactor(Redactor(redactor)),
		WithSlowLogFloor(0),
	)
	require.NoError(t, tr.Apply(db))

	ctx := context.Background()
	require.NoError(t, db.WithContext(ctx).Create(&testUser{Username: "bob", Password: "hunter2"}).Error)

	body := readTraffic(t, dir)
	assert.Contains(t, body, "db_create")
	// SQL 语句仍在,但绑定值被脱敏为 ***。
	assert.NotContains(t, body, "hunter2", "密码应被脱敏")
	assert.NotContains(t, body, "bob", "用户名应被脱敏")
	assert.Contains(t, body, "***")
}

// TestTracker_Traffic_ErrorPathCode 验证查询未命中时 traffic 行的 code 为
// not_found (统一分类),而非 ok 或 err。
func TestTracker_Traffic_ErrorPathCode(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableTraffic(true),
		WithSlowLogFloor(0),
	)
	require.NoError(t, tr.Apply(db))

	err := db.WithContext(context.Background()).
		Where("username = ?", "missing").First(&testUser{}).Error
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	body := readTraffic(t, dir)
	assert.Contains(t, body, `"code": "not_found"`)
	assert.Contains(t, body, "record not found")
}

// TestTracker_DebugContextCapturesTraffic 验证即使 EnableTraffic 为 false,
// 当 context 携带 tracer.FlagDebug 时也会捕获 traffic (per-request opt-in)。
func TestTracker_DebugContextCapturesTraffic(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTracker(Config{EnableTraffic: false, EnableMetrics: false, EnableErrorLog: false})
	require.NoError(t, tr.Apply(db))

	ctx := tracer.WithFlag(context.Background(), tracer.FlagDebug)
	require.NoError(t, db.WithContext(ctx).Create(&testUser{Username: "dbg", Password: "p"}).Error)

	// EnableTraffic=false,但 debug 模式下仍写出 traffic 行。
	body := readTraffic(t, dir)
	assert.Contains(t, body, "db_create")
}

// TestTracker_Metrics_NoExporterIsNoOp 验证 EnableMetrics=true 但 context
// 未注入 Exporter 时,回调走 noop 路径不 panic,且查询正常完成。
func TestTracker_Metrics_NoExporterIsNoOp(t *testing.T) {
	_, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableTraffic(false),
		WithEnableMetrics(true),
		WithSlowLogFloor(0),
	)
	require.NoError(t, tr.Apply(db))

	require.NoError(t, db.WithContext(context.Background()).
		Create(&testUser{Username: "m1", Password: "p"}).Error)
}

// TestTracker_Metrics_RecordsPerCommand 验证注入 recordingExporter 后,
// 一次 create + query 会各自触发 Incr/Decr/Observe,且 Observe 的 dsCmd 为
// db_create / db_query、code 为 "0" (monitor codeOK:nil error 归一为 ok)。
func TestTracker_Metrics_RecordsPerCommand(t *testing.T) {
	_, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableTraffic(false),
		WithEnableMetrics(true),
		WithSlowLogFloor(0),
	)
	require.NoError(t, tr.Apply(db))

	rec := &recordingExporter{}
	ctx := monitor.WithExporter(context.Background(), rec)

	require.NoError(t, db.WithContext(ctx).Create(&testUser{Username: "m1", Password: "p"}).Error)
	var u testUser
	require.NoError(t, db.WithContext(ctx).Where("username = ?", "m1").First(&u).Error)

	// 每条 SQL:一次 Incr (begin) + 一次 Decr (end) + 一次 Count + 一次 Observe。
	assert.GreaterOrEqual(t, rec.countByMethod("Observe"), 2, "Observe 应至少 2 次")
	assert.GreaterOrEqual(t, rec.countByMethod("Incr"), 2, "Incr 应至少 2 次")
	assert.GreaterOrEqual(t, rec.countByMethod("Decr"), 2, "Decr 应至少 2 次")
	// dsCmd + code 正确性:成功路径 code="0"。
	assert.True(t, rec.hasObserve("db_create", "0"), "应有 db_create code=0 的 Observe")
	assert.True(t, rec.hasObserve("db_query", "0"), "应有 db_query code=0 的 Observe")
}

// TestTracker_Metrics_NotFoundIsOKCode 验证 ErrRecordNotFound 在 metrics 层
// 归为 code="0" (ok),因此正常未命中不拉高失败率 (统一分类的核心目的)。
// 同时 errorLog 层仍降级为 Debug —— 三层口径一致:not_found 不是失败。
func TestTracker_Metrics_NotFoundIsOKCode(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableTraffic(false),
		WithEnableMetrics(true),
		WithEnableErrorLog(true),
		WithSlowLogFloor(0),
	)
	require.NoError(t, tr.Apply(db))

	rec := &recordingExporter{}
	ctx := monitor.WithExporter(context.Background(), rec)

	err := db.WithContext(ctx).Where("username = ?", "nope").First(&testUser{}).Error
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// metrics:not_found → code "0" (ok),不拉高失败率。
	assert.True(t, rec.hasObserve("db_query", "0"), "not_found 应归为 code=0 (ok)")
	// errorLog:not_found 降级 Debug,不进 error.log。
	assert.NotContains(t, readLevelLog(t, dir, "error"), "record not found")
	assert.Contains(t, readLevelLog(t, dir, "debug"), "record not found")
}

// TestTracker_ErrorLog_NotFoundIsDebug 验证 ErrRecordNotFound 走 Debug 且
// 带 WithError (实际错误信息可见),error.log 不含它。
func TestTracker_ErrorLog_NotFoundIsDebug(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableErrorLog(true),
		WithSlowLogFloor(0),
	)
	require.NoError(t, tr.Apply(db))

	err := db.WithContext(context.Background()).
		Where("username = ?", "nope").First(&testUser{}).Error
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	debugLog := readLevelLog(t, dir, "debug")
	assert.Contains(t, debugLog, "record not found")
	// WithError(db.Error):实际错误信息 "record not found" 作为 error field 落地。
	assert.Contains(t, debugLog, "record not found")
	assert.NotContains(t, readLevelLog(t, dir, "error"), "record not found")
}

// TestTracker_ErrorLog_RealErrorContainsActualError 验证非 not_found 的真实
// 错误 (唯一约束冲突)在 error.log 含 "db error" 摘要**且**含实际错误信息
// (UNIQUE constraint failed),而非只一句 "db error"。
func TestTracker_ErrorLog_RealErrorContainsActualError(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableErrorLog(true),
		WithSlowLogFloor(0),
	)
	require.NoError(t, tr.Apply(db))

	require.NoError(t, db.WithContext(context.Background()).
		Create(&testUser{Username: "dup", Password: "p"}).Error)
	err := db.WithContext(context.Background()).
		Create(&testUser{Username: "dup", Password: "p"}).Error
	require.Error(t, err)

	errLog := readLevelLog(t, dir, "error")
	assert.Contains(t, errLog, "db error")
	// 实际错误信息必须可见 (P1 回归防护)。
	assert.Contains(t, errLog, "UNIQUE constraint failed")
}

// TestTracker_SlowLog_FiresWarn 验证 SlowLogFloor 足够小时,慢查询在
// warn.log 留 "slow query" 且带 duration。
func TestTracker_SlowLog_FiresWarn(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableTraffic(false),
		WithEnableErrorLog(false),
		WithSlowLogFloor(1*time.Microsecond),
	)
	require.NoError(t, tr.Apply(db))

	require.NoError(t, db.WithContext(context.Background()).
		Create(&testUser{Username: "slow", Password: "p"}).Error)
	// 插入足够多行让一次查询耗时超过 1µs。
	for i := 0; i < 100; i++ {
		require.NoError(t, db.WithContext(context.Background()).
			Create(&testUser{Username: fmt.Sprintf("u%d", i), Password: "p"}).Error)
	}
	var u testUser
	require.NoError(t, db.WithContext(context.Background()).
		Where("username = ?", "slow").First(&u).Error)

	warnLog := readLevelLog(t, dir, "warn")
	assert.Contains(t, warnLog, "slow query")
	assert.Contains(t, warnLog, "duration")
}

// TestTracker_SlowLog_BelowFloorSilent 验证 SlowLogFloor 远大于实际耗时时
// 不写慢日志 (warn.log 为空)。
func TestTracker_SlowLog_BelowFloorSilent(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableTraffic(false),
		WithEnableErrorLog(false),
		WithSlowLogFloor(1*time.Hour),
	)
	require.NoError(t, tr.Apply(db))

	require.NoError(t, db.WithContext(context.Background()).
		Create(&testUser{Username: "fast", Password: "p"}).Error)

	assert.NotContains(t, readLevelLog(t, dir, "warn"), "slow query")
}

// TestTracker_SlowLog_NoVarsByDefault 验证慢查询日志默认不含绑定参数值
// (防敏感泄露),只含 SQL 与 duration。
func TestTracker_SlowLog_NoVarsByDefault(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableTraffic(false),
		WithEnableErrorLog(false),
		WithEnableVars(false),
		WithSlowLogFloor(1*time.Microsecond),
	)
	require.NoError(t, tr.Apply(db))

	require.NoError(t, db.WithContext(context.Background()).
		Create(&testUser{Username: "leak", Password: "leak123"}).Error)

	warnLog := readLevelLog(t, dir, "warn")
	assert.Contains(t, warnLog, "slow query")
	// 密码值不应出现在慢查询日志。
	assert.NotContains(t, warnLog, "leak123")
}

// TestTracker_DisablesAllIsBareGorm 验证四层全关时,Tracker.Apply 后的 db
// 行为与裸 gorm 一致:查询正常、无 traffic 落盘。
func TestTracker_DisablesAllIsBareGorm(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTracker(Config{EnableTraffic: false, EnableMetrics: false, EnableErrorLog: false, SlowLogFloor: 0})
	require.NoError(t, tr.Apply(db))

	require.NoError(t, db.WithContext(context.Background()).
		Create(&testUser{Username: "bare", Password: "p"}).Error)
	var u testUser
	require.NoError(t, db.WithContext(context.Background()).
		Where("username = ?", "bare").First(&u).Error)
	assert.Equal(t, "bare", u.Username)

	_, err := os.Stat(filepath.Join(dir, "traffic.log"))
	assert.True(t, os.IsNotExist(err), "traffic.log 不应在 traffic 关闭时产生")
}

// TestTracker_RawCoversExec 验证 Raw 回调已注册:db.Exec 会产生 db_raw 指标
// 与 traffic span。(对照 P2:Raw 原先完全未覆盖。)
func TestTracker_RawCoversExec(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableTraffic(true),
		WithEnableMetrics(true),
		WithSlowLogFloor(0),
	)
	require.NoError(t, tr.Apply(db))

	rec := &recordingExporter{}
	ctx := monitor.WithExporter(context.Background(), rec)

	// db.Exec 走 Raw processor。
	require.NoError(t, db.WithContext(ctx).Exec("DELETE FROM test_user WHERE 1=1").Error)

	// traffic.log 应有 db_raw 行。
	assert.Contains(t, readTraffic(t, dir), "db_raw")
	// metrics:应有 db_raw 的 Observe (code 0)。
	assert.True(t, rec.hasObserve("db_raw", "0"), "应有 db_raw code=0 的 Observe")
}

// TestTracker_UpdateDeleteProduceMetrics 验证 Update/Delete 操作各自产生指标,
// 补齐操作矩阵 (原先只有 create/query 测试)。
func TestTracker_UpdateDeleteProduceMetrics(t *testing.T) {
	_, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableTraffic(false),
		WithEnableMetrics(true),
		WithSlowLogFloor(0),
	)
	require.NoError(t, tr.Apply(db))

	rec := &recordingExporter{}
	ctx := monitor.WithExporter(context.Background(), rec)

	require.NoError(t, db.WithContext(ctx).Create(&testUser{Username: "u", Password: "p"}).Error)
	// Update。
	require.NoError(t, db.WithContext(ctx).Model(&testUser{}).
		Where("username = ?", "u").Update("password", "p2").Error)
	assert.True(t, rec.hasObserve("db_update", "0"), "应有 db_update code=0")
	// Delete。
	require.NoError(t, db.WithContext(ctx).Where("username = ?", "u").Delete(&testUser{}).Error)
	assert.True(t, rec.hasObserve("db_delete", "0"), "应有 db_delete code=0")
}

// TestTracker_RowMetricDispatchSemantics 验证 Row 回调产出 db_row 指标。
// 它衡量"查询派发"而非结果集读取 (见 README「Row/Rows 的语义」)。
func TestTracker_RowMetricDispatchSemantics(t *testing.T) {
	dir, teardown := setupLogger(t)
	defer teardown()
	db := newDB(t)
	tr := NewTrackerWithOpts(
		WithEnableTraffic(true),
		WithEnableMetrics(true),
		WithSlowLogFloor(0),
	)
	require.NoError(t, tr.Apply(db))

	rec := &recordingExporter{}
	ctx := monitor.WithExporter(context.Background(), rec)

	require.NoError(t, db.WithContext(ctx).Create(&testUser{Username: "r", Password: "p"}).Error)
	row := db.WithContext(ctx).Raw("SELECT password FROM test_user WHERE username = ?", "r").Row()
	var pwd string
	require.NoError(t, row.Scan(&pwd))
	assert.Equal(t, "p", pwd)

	// db_row 指标存在 (派发成功)。
	assert.True(t, rec.hasObserve("db_row", "0"), "应有 db_row code=0")
	// traffic.log 有 db_row 行。
	assert.Contains(t, readTraffic(t, dir), "db_row")
}
