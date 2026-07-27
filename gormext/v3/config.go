package gormext

import (
	"time"
)

// defaultConfig 在未提供任何 option 时应用。
//
// 安全第一:Traffic span 默认开启 (仅记 cmd/cost/code/sql,不含参数),
// 但 SQL 绑定参数 (vars) 默认**不**记入 traffic/errorLog/slowLog ——
// 参数中常含密码、token、手机号等敏感信息。需要参数时显式
// WithEnableVars(true);如需对参数脱敏后再记,用 WithRedactor。
//
// metrics 与 errorLog 默认关闭:metrics 要求在 context 中注入 Exporter
// (见 monitor.Init);errorLog 在业务有显式需要时再开。SlowLogFloor
// 默认 5s,超过即记 Warn。
var defaultConfig = Config{
	EnableTraffic:  true,
	EnableMetrics:  false,
	EnableErrorLog: false,
	EnableVars:     false,
	SlowLogFloor:   5 * time.Second,
}

// Config 控制 Tracker.Apply 接入哪些回调层。
type Config struct {
	// EnableTraffic 通过 logger/v3 的 traffic logger 开启每条 SQL 的流量日志
	// (记 cmd/cost/code/sql)。当 context 携带 tracer.FlagDebug 时,无论
	// 此 flag 取值如何也会自动记录 (per-request debug)。
	//
	// 注意:traffic 记录的 SQL 绑定参数受 EnableVars 控制,默认关闭以防
	// 敏感信息泄露。
	EnableTraffic bool `json:"enable_traffic" yaml:"enable_traffic"`
	// EnableMetrics 通过 monitor/v3 开启每条 SQL 的 latency/counter 记录。
	// 默认关闭:它要求在 context 中注入 Exporter (见 monitor.Init)。
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics"`
	// EnableErrorLog 开启数据库错误日志:非 gorm.ErrRecordNotFound 的错误
	// 输出 Error (带 WithError(db.Error)),记录未找到降级为 Debug,避免
	// 正常的"查无此记录"淹没真正的错误。
	EnableErrorLog bool `json:"enable_error_log" yaml:"enable_error_log"`
	// EnableVars 开启后,traffic/errorLog/slowLog 会附带 SQL 绑定参数 (vars)。
	// 默认**关闭**:参数中常含密码、token、手机号等敏感信息。开启前评估
	// 是否需要配合 Redactor 脱敏。SQL 语句本身始终记录。
	EnableVars bool `json:"enable_vars" yaml:"enable_vars"`
	// Redactor 对 SQL 绑定参数做脱敏后再记入 traffic/errorLog/slowLog。
	// 为 nil 且 EnableVars=true 时,参数按原值记录。Redactor 同样作用于
	// EnableVars=false 的情形:若你想在默认关闭的情况下仍记脱敏后的参数,
	// 设 EnableVars=true 并提供 Redactor。
	Redactor Redactor `json:"-" yaml:"-"`
	// SlowLogFloor 是慢查询阈值:当查询耗时超过此值时输出 Warn 日志 (带
	// SQL 与实际耗时;参数受 EnableVars/Redactor 控制)。为 0 时不检测
	// 慢查询。负值在 Apply 时会被规范化为 0 (即关闭慢查询)。
	SlowLogFloor time.Duration `json:"slow_log_floor" yaml:"slow_log_floor"`
}

// ConfigOption 变更 Config。NewTrackerWithOpts 将其叠加在 defaultConfig 之上。
// nil option 被安全跳过,不会 panic。
type ConfigOption func(cfg *Config)

// WithEnableTraffic 启用或禁用流量日志。
func WithEnableTraffic(enable bool) ConfigOption {
	return func(cfg *Config) { cfg.EnableTraffic = enable }
}

// WithEnableMetrics 启用或禁用 Prometheus 指标记录。
func WithEnableMetrics(enable bool) ConfigOption {
	return func(cfg *Config) { cfg.EnableMetrics = enable }
}

// WithEnableErrorLog 启用或禁用数据库错误日志。
func WithEnableErrorLog(enable bool) ConfigOption {
	return func(cfg *Config) { cfg.EnableErrorLog = enable }
}

// WithEnableVars 启用或禁用 SQL 绑定参数的记录 (作用于 traffic/errorLog/
// slowLog)。默认关闭以防敏感信息泄露。
func WithEnableVars(enable bool) ConfigOption {
	return func(cfg *Config) { cfg.EnableVars = enable }
}

// WithRedactor 设置参数脱敏器。它对 SQL 绑定参数做原地替换后再记入日志。
// 典型用法:按字段名或 SQL 占位位置把 password/token/phone 等替换为 ***。
func WithRedactor(r Redactor) ConfigOption {
	return func(cfg *Config) { cfg.Redactor = r }
}

// WithSlowLogFloor 设置慢查询日志阈值。0 或负值表示不检测慢查询。
func WithSlowLogFloor(floor time.Duration) ConfigOption {
	return func(cfg *Config) { cfg.SlowLogFloor = floor }
}

// Redactor 对一条 SQL 的命令 (sql) 与绑定参数 (vars) 做脱敏,返回脱敏后
// 的值供日志记录。返回值会被原样写入 traffic/errorLog/slowLog;不应修改
// 传入的切片 (调用方仍可能继续使用原 vars)。
//
// 实现示例:把 vars 中所有 string 值按 key 猜测脱敏:
//
//	func redact(sql string, vars []any) (string, []any) {
//	    out := make([]any, len(vars))
//	    for i, v := range vars {
//	        if s, ok := v.(string); ok && len(s) > 0 {
//	            out[i] = "***"
//	        } else {
//	            out[i] = v
//	        }
//	    }
//	    return sql, out
//	}
type Redactor func(sql string, vars []any) (string, []any)
