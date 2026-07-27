# gormext

GORM 插件：通过 GORM 回调注入流量日志、Prometheus 指标、数据库错误日志和慢查询检测。v3 是一次无兼容包袱的干净重写，基于 `logger/v3`、`monitor/v3`、`tracer/v3`，与保留不动的 `gormext/v2` 并存。

gormext **只提供 `Tracker`**——把它一次性 `Apply` 到 `*gorm.DB` 上，之后照常 `db.WithContext(ctx).First/...`，所有 `Query`/`Create`/`Update`/`Delete`/`Row`/`Raw` 操作自动经过流量/指标/错误/慢查询四层回调。不封装 gorm：标准库的 `*gorm.DB` 已经够用，再包一层只是多一层间接和约定。

```go
import "github.com/tenz-io/gokit/gormext/v3"
```

## 模块介绍

gormext 解决 GORM 调用的四类横切需求：

- **记录 SQL 流量日志**（`WithEnableTraffic`）：用 `logger/v3` 的 traffic logger 记录每条 SQL 的 cmd、耗时、状态码、SQL 语句；链路处于 `tracer.FlagDebug` 态时自动触发。流量日志**不读、不记结果集**，只记 SQL 元数据，开销小且不打扰 gorm 的结果读取。SQL 绑定参数受 `EnableVars` 控制，默认关闭以防敏感信息泄露。
- **上报数据库指标**（`WithEnableMetrics`）：用 `monitor/v3` 按命令类型（`db_query`/`db_create`/`db_update`/`db_delete`/`db_row`/`db_raw`）记录每条 SQL 的耗时与成功率。
- **区分记录数据库错误日志**（`WithEnableErrorLog`）：非 `gorm.ErrRecordNotFound` 的错误输出 Error（带 `WithError(db.Error)`，含真实错误信息），记录未找到降级为 Debug，避免正常的"查无此记录"淹没真正的错误。
- **检测并告警慢查询**（`WithSlowLogFloor`）：超过阈值的 SQL 会带上 SQL 语句与实际耗时输出 Warn 日志，用于发现慢 SQL。SQL 慢日志会直接打印具体语句，比 monitor 的 latency 直方图更能定位是哪条 SQL 慢。

核心能力：

- `Tracker`：将四类回调一次性 `Apply` 到 `*gorm.DB`，无需在业务代码里逐个埋点。
- 回调名带 `gormext:` 前缀，避免与用户自身注册的回调冲突。

### 安全默认

- **SQL 参数默认不进日志**：`EnableVars` 默认 `false`，traffic/errorLog/slowLog 只记 SQL 语句，不记绑定参数（密码、token、手机号等可能明文泄露）。需要时 `WithEnableVars(true)`，建议配合 `WithRedactor` 脱敏。
- **Redactor 扩展点**：`WithRedactor(func(sql, vars) (string, []any))` 对参数按字段/类型/位置脱敏后再记入日志，不影响 SQL 语句本身。
- **Debug 仍可绕过 `EnableTraffic=false`**：当 context 携带 `tracer.FlagDebug` 时，即使 `EnableTraffic=false` 也会捕获流量（per-request debug）。因此生产环境若不希望临时泄露参数，Redactor 是必要的。

### 统一错误分类

metrics/traffic/errorLog 共用 `classify(err)` → `ok`/`not_found`/`err`，三层口径一致：

| 分类 | 触发条件 | metrics code | traffic code | errorLog 级别 |
|---|---|---|---|---|
| `ok` | `err == nil` | `0` (ok) | `ok` | 不记 |
| `not_found` | `errors.Is(err, gorm.ErrRecordNotFound)` | `0` (ok，**不拉高失败率**) | `not_found` | Debug |
| `err` | 其余非 nil 错误 | `1` (err) | `error` | Error (带 `WithError`) |

关键点：`gorm.ErrRecordNotFound` 在三层都视为"正常"，因此大量正常的"查无此记录"不会污染失败率与告警。

## 能力清单

| 能力 | 含义 |
|---|---|
| 一键接入 GORM 回调 | `Tracker.Apply(db)` 在 `Query`/`Create`/`Update`/`Delete`/`Row`/`Raw` 六类操作上统一注册前后置回调，无需在业务代码里逐个埋点 |
| 记录 SQL 流量日志 | 开启 `EnableTraffic` 后，每条 SQL 及其耗时、结果码、SQL 语句作为一条 traffic 记录落到 `logger`，用于排查具体请求执行了哪些 SQL |
| SQL 参数脱敏 | `EnableVars` 默认关闭；开启后可配 `Redactor` 对密码/token 等参数脱敏再记入 traffic/errorLog/slowLog |
| Debug 模式下自动补记流量日志 | 即使未显式开启 `EnableTraffic`，只要 `tracer.FromContext(ctx).IsDebug()` 为真，也会自动记录流量日志 |
| 上报数据库指标 | 开启 `EnableMetrics` 后，按命令类型用 `monitor.Begin`/`EndWithCode` 统计各类数据库操作的耗时与成功率 |
| 统一错误分类 | metrics/traffic/errorLog 共用 `classify`，not_found 在三层均视为正常，不拉高失败率 |
| 区分记录数据库错误日志 | 开启 `EnableErrorLog` 后，非 `gorm.ErrRecordNotFound` 的错误输出 Error（带真实错误信息），记录未找到降级为 Debug |
| 检测并告警慢查询 | `WithSlowLogFloor` 设置阈值，超过的 SQL 会带上 SQL 语句与耗时输出 Warn 日志 |
| 覆盖 db.Exec / db.Raw | 注册 Raw processor，使 `db.Exec`/`db.Raw`/部分迁移 DDL 也产生指标与流量日志 |
| 灵活的配置方式 | 提供 `NewTracker(Config)` 直接传入完整配置，或 `NewTrackerWithOpts` 基于默认配置叠加 `With*` Option 链式组合（nil option 安全跳过，负 SlowLogFloor 归一为 0） |

## 快速开始

```go
package main

import (
	"context"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/tenz-io/gokit/gormext/v3"
	"github.com/tenz-io/gokit/logger/v3"
	"github.com/tenz-io/gokit/monitor/v3"
)

func init() {
	logger.ConfigureWithOpts(
		logger.WithLevel(logger.DebugLevel),
		logger.WithConsole(true),
		logger.WithFilePath("log"),
		logger.WithTraffic(true),
	)
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	tracker := gormext.NewTrackerWithOpts(
		gormext.WithEnableTraffic(true),
		gormext.WithEnableMetrics(true),
		gormext.WithEnableErrorLog(true),
		// 安全:开启参数记录但配 Redactor 把字符串参数 (含密码) 脱敏为 ***。
		gormext.WithEnableVars(true),
		gormext.WithRedactor(gormext.Redactor(redactSecrets)),
		gormext.WithSlowLogFloor(500*time.Millisecond),
	)
	if err = tracker.Apply(db); err != nil {
		log.Fatal(err)
	}

	// Inject a monitor Exporter so WithEnableMetrics records something.
	ctx := monitor.Init(context.Background(), "myapp")

	var user User
	if err = db.WithContext(ctx).First(&user, 1).Error; err != nil {
		log.Printf("find: %v", err)
	}
}

type User struct {
	ID       int64  `gorm:"primaryKey"`
	Username string `gorm:"column:username;unique"`
	Password string `gorm:"column:password"`
}

// redactSecrets 把 SQL 绑定参数中的字符串值脱敏为 ***,其余类型原样保留。
func redactSecrets(sql string, vars []any) (string, []any) {
	out := make([]any, len(vars))
	for i, v := range vars {
		if _, ok := v.(string); ok {
			out[i] = "***"
		} else {
			out[i] = v
		}
	}
	return sql, out
}
```

## API 速查

| 符号 | 说明 |
|---|---|
| `Tracker` | 接口，`Apply(db *gorm.DB) error` 将回调注册到 `*gorm.DB`（init 阶段调一次） |
| `NewTracker(config Config) Tracker` | 使用完整 `Config` 创建 Tracker（负 `SlowLogFloor` 归一为 0） |
| `NewTrackerWithOpts(opts ...ConfigOption) Tracker` | 基于默认配置叠加 Option 创建 Tracker（nil option 安全跳过） |
| `Config` | 配置项：`EnableTraffic`、`EnableMetrics`、`EnableErrorLog`、`EnableVars`、`Redactor`、`SlowLogFloor` |
| `ConfigOption` | `func(*Config)`，用于函数式配置 |
| `WithEnableTraffic(bool) ConfigOption` | 开关流量日志记录 |
| `WithEnableMetrics(bool) ConfigOption` | 开关 Prometheus 指标上报 |
| `WithEnableErrorLog(bool) ConfigOption` | 开关数据库错误日志 |
| `WithEnableVars(bool) ConfigOption` | 开关 SQL 绑定参数记录（默认关，防泄露） |
| `WithRedactor(Redactor) ConfigOption` | 设置参数脱敏器 |
| `WithSlowLogFloor(time.Duration) ConfigOption` | 设置慢查询阈值，0/负值表示不检测 |
| `type Redactor func(sql string, vars []any) (string, []any)` | 参数脱敏函数签名 |

## 回调行为

`Tracker.Apply(db)` 对 `Query`/`Create`/`Update`/`Delete`/`Row`/`Raw` 六类操作各注册一对 `gormext:start_*` / `gormext:end_*` 回调。每条 SQL 的执行流：

1. **`before(cmd)`**（前置回调）：记起始时间；若 `EnableMetrics` 则 `monitor.Begin(ctx, cmd)`；若 `EnableTraffic` 或 `tracer.FromContext(ctx).IsDebug()` 则 `logger.FromContext(ctx).StartTraffic(cmd).WithTyp(TrafficTypSend)`。meta 经 `db.Statement.Context` 传给 `after`。
2. **执行 SQL**（gorm 自身）。
3. **`after`**（后置回调）：先 `classify(db.Error)` 得到统一分类，再按配置分支结束 metrics、traffic，并按需输出错误/慢查询日志。

| 分支 | 条件 | 行为 |
|---|---|---|
| metrics | `EnableMetrics` | `metricsRec.EndWithCode(monitorCode(class))` —— ok/not_found→`0`，err→`1` |
| traffic | `EnableTraffic` 或 debug | `ok`→`End(nil,"ok",...)`；`not_found`→`End(nil,"not_found",...,"error",err.Error())`；`err`→`EndWithError(err,...)`。sql/vars 走脱敏 |
| errorLog | `EnableErrorLog` 且 `db.Error != nil` | `not_found`→Debug("record not found")；`err`→Error("db error")；均带 `WithError(db.Error)` |
| slowLog | `SlowLogFloor > 0` 且 `time.Since(start) > floor` | `Warn("slow query")`（带 sql/duration；vars 受 EnableVars/Redactor 控制） |

注意点：

- **SQL 参数默认不记**：`EnableVars=false` 时 traffic/errorLog/slowLog 只记 SQL 语句，不记 vars，防敏感泄露。`Redactor` 仅在 `EnableVars=true` 时生效。
- **traffic 在 debug 模式始终激活**：即使 `EnableTraffic=false`，`tracer.FlagDebug` 仍会捕获流量，便于线上临时排障；因此 Redactor 对生产是必要的。
- **metrics 无 Exporter 时 no-op**：`EnableMetrics=true` 但 context 未注入 Exporter 时，`monitor.Begin` 返回 nil-safe 的 Recorder，回调走空路径不 panic。
- **重复 `Apply` 不报错**：gorm 对重名回调只发 warn 并保留两份 handler（非替换），因此对同一 db 多次 `Apply` 会让每个回调被调用两次（指标/日志翻倍）。Apply 设计为 init 阶段调一次。
- **慢查询对 SQL 更实用**：与 `httpext/v3` 移除 slowLog 不同，gormext 保留 `SlowLogFloor`——SQL 慢日志会直接打印 SQL 语句（参数受控），而 monitor 的 latency 直方图只能看到聚合耗时分布。

## 语义边界

使用前务必了解以下边界（GORM 回调机制的本质，非 v3 的 bug）：

- **回调衡量的是 GORM operation 粒度，不一定是"一条物理 SQL"**：`Before("*")`/`After("*")` 会含事务、hook、association、preload；DryRun 等情况下可能根本不执行 SQL。指标语义是 **GORM operation latency**，而非 database execution latency。
- **Row/Rows 的语义是"查询派发耗时/派发是否成功"**：`Row()`/`Rows()` 回调触发时，结果通常尚未 Scan/迭代/Close —— `QueryRowContext` 的错误可能延迟到 `row.Scan()` 才出现，`Rows()` 的迭代错误不会进入当前 `db.Error`。因此 `db_row` 不承诺结果集读取的完整耗时与错误。
- **Apply 中途失败留半注册状态**：实践中 Apply 只在 DB 初始化阶段调用、且失败即 fatal，半注册状态不会到达业务路径。

## 测试

测试在 `tracker_test.go`，用内存 sqlite + 真实 gorm.DB 端到端验证四层回调的实际副作用，而不只是"不 panic"。覆盖率约 90%。

| 测试 | 验证内容 |
|---|---|
| `TestTracker_Apply_NilDB` | nil db 为 no-op，不 panic |
| `TestTracker_Apply_NilOptionIsSafe` | nil option 被安全跳过，不 panic |
| `TestTracker_NewTracker_NormalizesNegativeFloor` | 负 SlowLogFloor 归一为 0 |
| `TestTracker_Apply_RepeatCalls` | 重复 Apply 不报错，后续查询仍成功 |
| `TestTracker_Traffic_WritesSQLSpanNoVarsByDefault` | 默认 traffic 写 SQL span 但**不含**绑定参数（防泄露） |
| `TestTracker_Traffic_VarsRecordedWhenEnabled` | `WithEnableVars(true)` 时参数进 traffic.log |
| `TestTracker_Traffic_RedactorScrubsParams` | `WithRedactor` 把密码脱敏为 `***`，不进 traffic.log |
| `TestTracker_Traffic_ErrorPathCode` | 未命中时 traffic code 为 `not_found`（非 ok/error） |
| `TestTracker_DebugContextCapturesTraffic` | `EnableTraffic=false` 时 `tracer.FlagDebug` 仍触发 traffic |
| `TestTracker_Metrics_NoExporterIsNoOp` | `EnableMetrics=true` 但无 Exporter 时 no-op 不 panic |
| `TestTracker_Metrics_RecordsPerCommand` | create+query 各触发 Incr/Decr/Observe，dsCmd 与 code(`0`) 正确 |
| `TestTracker_Metrics_NotFoundIsOKCode` | not_found 在 metrics 归 `0`(ok) 且 errorLog 降 Debug（三层一致） |
| `TestTracker_ErrorLog_NotFoundIsDebug` | not_found 走 Debug 带 WithError，error.log 不含 |
| `TestTracker_ErrorLog_RealErrorContainsActualError` | 真实错误（唯一键冲突）进 error.log，含 "UNIQUE constraint failed" 实际信息 |
| `TestTracker_SlowLog_FiresWarn` | 阈值足够小时慢查询进 warn.log，带 duration |
| `TestTracker_SlowLog_BelowFloorSilent` | 阈值远大于实际耗时不写慢日志 |
| `TestTracker_SlowLog_NoVarsByDefault` | 慢查询日志默认不含绑定参数（防泄露） |
| `TestTracker_DisablesAllIsBareGorm` | 四层全关时行为同裸 gorm，无 traffic.log |
| `TestTracker_RawCoversExec` | `db.Exec` 走 Raw，产生 `db_raw` 指标与 traffic |
| `TestTracker_UpdateDeleteProduceMetrics` | Update/Delete 各产生 `db_update`/`db_delete` 指标 |
| `TestTracker_RowMetricDispatchSemantics` | `db.Row()` 产生 `db_row` 指标（派发语义） |

## 开发

```bash
make test      # go test ./... -cover -v
make cover     # 生成 coverage.html
make vet       # go vet
make tidy      # go mod tidy
make run-example  # 运行 example/ 下的示例程序
```

## 与 v2 的行为差异

v3 不保证与 v2 兼容，以下是显式的行为差异：

| 差异点 | v2 | v3 |
| --- | --- | --- |
| 依赖 | logger/v2、monitor/v2、tracer/v2 | logger/v3、monitor/v3、tracer/v3 |
| 选项命名 | `WithTraffic`/`WithMetrics`/`WithErrorLog` | `WithEnableTraffic`/`WithEnableMetrics`/`WithEnableErrorLog`/`WithEnableVars`/`WithRedactor`，与 `httpext/v3` 的 `WithEnable*` 一致 |
| SQL 参数记录 | traffic/errorLog/slowLog **默认明文记录** vars（密码会泄露） | **默认关闭**（`EnableVars`）；可配 `Redactor` 脱敏 |
| 错误日志内容 | `le.Error("db error")` 仅带 sql/vars | 补 `WithError(db.Error)`，含真实错误信息（超时/断连/唯一键等） |
| 错误分类 | metrics 把 not_found 记为 err（拉高失败率）；errorLog 当 Debug —— 口径分裂 | 统一 `classify`：not_found 在 metrics/traffic/errorLog 三层均视为正常 |
| 流量日志 API | `logger.StartTrafficRec(ctx, &ReqEntity{...})` + `rec.End(&RespEntity{...}, logger.Fields{...})` | `logger.FromContext(ctx).StartTraffic(cmd).WithTyp(logger.TrafficTypSend)` + `rec.End(nil, code, "sql", ..., "vars", ...)` / `rec.EndWithError(err, ...)` |
| 流量日志内容 | 记 SQL + vars + result（`RespEntity`） | 记 SQL + code，**不记 result body**，对齐 `httpext/v3` |
| 指标 API | `monitor.BeginRecord` + `EndWithError` | `monitor.Begin(ctx, cmd)` + `rec.EndWithCode(code)` |
| 操作覆盖 | Query/Create/Update/Delete/Row（漏 Exec/Raw） | **加 Raw**，覆盖 `db.Exec`/`db.Raw`/迁移 DDL |
| 文件结构 | `config.go` + `tracker.go` | 多文件拆分 `doc`/`config`/`tracker`/`callbacks`/`classify`，与 `httpext/v3` 一致（含 `tracker_test.go`，v2 无测试） |
| 回调命名 | `start_query`/`end_query` 等 | 加 `gormext:` 前缀（`gormext:start_query`），避免冲突 |
| 慢查询日志 | `WithSlowLogFloor` + 慢日志 Warn | 保留（SQL 慢日志打印 SQL，比 monitor 直方图更实用） |
| nil option / 负 floor | 未处理 | nil option 跳过；负 SlowLogFloor 归一为 0 |

调用方迁移速查（**不在本次范围，留待后续逐步改**）：

| 调用方代码（v2） | v3 等价 |
| --- | --- |
| `gormext.WithTraffic(true)` | `gormext.WithEnableTraffic(true)` |
| `gormext.WithMetrics(true)` | `gormext.WithEnableMetrics(true)` |
| `gormext.WithErrorLog(true)` | `gormext.WithEnableErrorLog(true)` |
| `gormext.WithSlowLogFloor(d)` | `gormext.WithSlowLogFloor(d)`（不变；负值归一为 0） |
| （v2 默认记 vars） | `gormext.WithEnableVars(true)` + `gormext.WithRedactor(...)`（默认不记） |
| `logger.StartTrafficRec(ctx, &logger.ReqEntity{Typ:..., Cmd:...})` | `logger.FromContext(ctx).StartTraffic(cmd).WithTyp(logger.TrafficTypSend)` |
| `trafficRec.End(&logger.RespEntity{Code:..., Msg:...}, logger.Fields{"sql":..., "vars":...})` | `trafficRec.End(nil, code, "sql", db.Statement.SQL.String(), "vars", db.Statement.Vars)` 或 `trafficRec.EndWithError(err, ...)` |
| `monitor.BeginRecord(ctx, cmd)` | `monitor.Begin(ctx, cmd)` |

## 引入路径

`import "github.com/tenz-io/gokit/gormext/v3"`
