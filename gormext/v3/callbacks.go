package gormext

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/tenz-io/gokit/logger/v3"
	"github.com/tenz-io/gokit/monitor/v3"
	"github.com/tenz-io/gokit/tracer/v3"
)

// trackingKeyType 是在 before/after 回调间承载 meta 的 context key 类型。
// 用一个独立的未导出类型,可避免与任何其他包的 context key 冲突。
type trackingKeyType string

const trackingMetaCtxKey trackingKeyType = "_gormext_tracking_meta"

// meta 保存一次 SQL 执行在 before 回调时采集的上下文,供 after 回调消费。
type meta struct {
	startTime  time.Time
	metricsRec *monitor.Recorder
	trafficRec *logger.TrafficRec
}

// before 返回注册到每类操作的前置回调。它记录起始时间,并按配置启动一个
// metrics Recorder 与一个 traffic span。两者都通过 db.Statement.Context
// 传递给对应的 after 回调。
func (t *tracker) before(cmd string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		var (
			ctx = db.Statement.Context
			m   = &meta{startTime: time.Now()}
		)

		if t.config.EnableMetrics {
			m.metricsRec = monitor.Begin(ctx, cmd)
		}

		// traffic 在配置开启,或 context 处于 debug 模式 (per-request opt-in)
		// 时记录,因此运维方可以在不修改 tracker config 的情况下为单个请求
		// 切换捕获。
		if t.config.EnableTraffic || tracer.FromContext(ctx).IsDebug() {
			m.trafficRec = logger.FromContext(ctx).
				StartTraffic(cmd).
				WithTyp(logger.TrafficTypSend)
		}

		db.Statement.Context = context.WithValue(ctx, trackingMetaCtxKey, m)
	}
}

// after 是注册到每类操作的后置回调。它从 context 取出 before 写入的 meta,
// 然后按配置依次结束 metrics、traffic,并按需输出错误日志与慢查询日志。
//
// 错误分类统一走 classify:metrics/traffic/errorLog 对同一种错误看到一致的
// 语义 (not_found 在三层都被视为"正常"而非"失败")。
func (t *tracker) after(db *gorm.DB) {
	ctx := db.Statement.Context
	class := classify(db.Error)

	// 日志用 entry:始终带 SQL 语句;vars 受 EnableVars/Redactor 控制
	// (redactVars 在关闭时返回 nil,不记参数,防敏感信息泄露)。
	sqlStr := db.Statement.SQL.String()
	logSQL, logVars := t.redactVars(sqlStr, db.Statement.Vars)
	le := logger.FromContext(ctx).With("sql", logSQL)
	if logVars != nil {
		le = le.With("vars", logVars)
	}
	if db.Error != nil {
		le = le.WithError(db.Error)
	}

	m, ok := ctx.Value(trackingMetaCtxKey).(*meta)
	if !ok || m == nil {
		le.Warn("gormext: tracker meta not found")
		return
	}

	// metrics:用分类后的 code 结束 Recorder (not_found → ok code,
	// 不拉高失败率)。
	if t.config.EnableMetrics && m.metricsRec != nil {
		m.metricsRec.EndWithCode(monitorCode(class))
	}

	// traffic:结束 span。code 用可读的 trafficCode (ok/not_found/err),与
	// 分类口径一致;sql/vars 走同样的脱敏路径。
	//   - ok:        End(nil, "ok", ...)
	//   - not_found: End(nil, "not_found", ..., "error", err.Error()) —— 用 End
	//               而非 EndWithError,使 code 显式为 not_found (EndWithError
	//               会强制 code="error");错误文本作为 error field 保留。
	//   - err:       EndWithError(db.Error, ...) —— code="error",msg=err.Error()。
	// traffic 不记录结果 body,保持开销低且不干扰 gorm 的结果读取。
	if (t.config.EnableTraffic || tracer.FromContext(ctx).IsDebug()) && m.trafficRec != nil {
		tcSQL, tcVars := logSQL, logVars // 与结构化日志同一份脱敏结果
		switch class {
		case classOK:
			m.trafficRec.End(nil, trafficCode(class), "sql", tcSQL, "vars", tcVars)
		case classNotFound:
			m.trafficRec.End(nil, trafficCode(class), "sql", tcSQL, "vars", tcVars, "error", db.Error.Error())
		default:
			m.trafficRec.EndWithError(db.Error, "sql", tcSQL, "vars", tcVars)
		}
	}

	// 错误日志:not_found 降级为 Debug (带 WithError 便于排查);其余错误
	// 输出 Error (带 WithError 与具体错误信息)。
	if t.config.EnableErrorLog && db.Error != nil {
		switch class {
		case classNotFound:
			le.Debug("record not found")
		default:
			le.Error("db error")
		}
	}

	// 慢查询:超过阈值则 Warn (带 SQL 与实际耗时;vars 已在 le 上)。
	if t.config.SlowLogFloor > 0 {
		if dur := time.Since(m.startTime); dur > t.config.SlowLogFloor {
			le.With("duration", dur.String()).Warn("slow query")
		}
	}
}
