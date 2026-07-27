package gormext

import (
	"fmt"

	"gorm.io/gorm"
)

// Tracker 通过 Apply 把流量/指标/错误/慢查询回调注册到一个 *gorm.DB 实例。
// 注册一次后,该 db 上的所有 Query/Create/Update/Delete/Row/Raw 操作都会
// 透明地经过回调,无需在业务代码里逐个埋点。
type Tracker interface {
	// Apply 把回调注册到 db。nil db 为 no-op。
	//
	// 契约:Apply 设计为在 DB 初始化阶段调用**恰好一次**。gorm 对重名
	// 回调只发一条 warn 并按编译保留两份 handler,因此对同一 db 重复
	// 调用会让每个回调被调用两次 (指标/日志翻倍),而非报错或替换配置。
	// 如需更换配置,请对新 db 调用 Apply,或先 gorm Remove 旧回调。
	Apply(db *gorm.DB) error
}

// NewTrackerWithOpts 由函数式 option 构建 Tracker,叠加在 defaultConfig
// 之上。nil option 被安全跳过,不会 panic。
func NewTrackerWithOpts(opts ...ConfigOption) Tracker {
	cfg := defaultConfig
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return NewTracker(cfg)
}

// NewTracker 由显式 Config 构建 Tracker。负 SlowLogFloor 被规范化为 0
// (关闭慢查询),避免静默的意外行为。
func NewTracker(config Config) Tracker {
	if config.SlowLogFloor < 0 {
		config.SlowLogFloor = 0
	}
	return &tracker{config: config}
}

type tracker struct {
	config Config
}

// registerFn 把一个 before 回调与一个 after 回调注册到 gorm 的某类操作
// (Query/Create/Update/Delete/Row/Raw)。它以闭包形式持有 gorm 未导出的
// *processor / *callback —— 这两者均无法在签名里命名,因此整个
// "Before/After + Register" 的链式动作被封装进闭包,不把未导出类型
// 暴露到签名或接口中。
type registerFn func(before, after func(db *gorm.DB)) error

// op 描述一类要注入回调的 gorm 操作:回调名后缀、上报给 metrics/traffic
// 的命令名,以及一个取得该操作回调链并完成 Before/After 注册的闭包。
type op struct {
	name string // 回调名后缀,如 "query"
	cmd  string // 上报命令,如 "db_query"
	get  func(db *gorm.DB) registerFn
}

// ops 是 gormext 注入回调的操作集合。覆盖 GORM 六类操作:
// Query/Create/Update/Delete/Row/Raw。Raw 覆盖 db.Exec / db.Raw / 部分
// 迁移 DDL,否则它们不会产生任何指标或日志。顺序无关 (每类操作的
// Before/After 独立注册),但保持稳定的顺序便于排查。每个 get 直接把 v2
// 风格的链式调用写进闭包:gorm 未导出类型在闭包内可见,因此无需命名类型
// 或做接口断言。共享的回调名由 apply 时以 o.name 组装。
var ops = []op{
	{name: "query", cmd: "db_query", get: func(db *gorm.DB) registerFn {
		proc := db.Callback().Query()
		return func(before, after func(db *gorm.DB)) error {
			if err := proc.Before("*").Register("gormext:start_query", before); err != nil {
				return err
			}
			return proc.After("*").Register("gormext:end_query", after)
		}
	}},
	{name: "create", cmd: "db_create", get: func(db *gorm.DB) registerFn {
		proc := db.Callback().Create()
		return func(before, after func(db *gorm.DB)) error {
			if err := proc.Before("*").Register("gormext:start_create", before); err != nil {
				return err
			}
			return proc.After("*").Register("gormext:end_create", after)
		}
	}},
	{name: "update", cmd: "db_update", get: func(db *gorm.DB) registerFn {
		proc := db.Callback().Update()
		return func(before, after func(db *gorm.DB)) error {
			if err := proc.Before("*").Register("gormext:start_update", before); err != nil {
				return err
			}
			return proc.After("*").Register("gormext:end_update", after)
		}
	}},
	{name: "delete", cmd: "db_delete", get: func(db *gorm.DB) registerFn {
		proc := db.Callback().Delete()
		return func(before, after func(db *gorm.DB)) error {
			if err := proc.Before("*").Register("gormext:start_delete", before); err != nil {
				return err
			}
			return proc.After("*").Register("gormext:end_delete", after)
		}
	}},
	{name: "row", cmd: "db_row", get: func(db *gorm.DB) registerFn {
		// 注意:Row/Rows 回调触发时,结果通常尚未 Scan/迭代/Close ——
		// QueryRowContext 的错误可能延迟到 row.Scan() 才出现。因此 db_row
		// 的语义是"查询派发耗时/派发是否成功",不承诺结果集读取的完整耗时
		// 与错误。详见 README「Row/Rows 的语义」。
		proc := db.Callback().Row()
		return func(before, after func(db *gorm.DB)) error {
			if err := proc.Before("*").Register("gormext:start_row", before); err != nil {
				return err
			}
			return proc.After("*").Register("gormext:end_row", after)
		}
	}},
	{name: "raw", cmd: "db_raw", get: func(db *gorm.DB) registerFn {
		// Raw 覆盖 db.Exec / db.Raw 及部分迁移 DDL,否则它们不产生指标。
		proc := db.Callback().Raw()
		return func(before, after func(db *gorm.DB)) error {
			if err := proc.Before("*").Register("gormext:start_raw", before); err != nil {
				return err
			}
			return proc.After("*").Register("gormext:end_raw", after)
		}
	}},
}

// Apply 把 ops 中每类操作的 before/after 回调注册到 db。回调名以
// "gormext:" 为前缀,避免与用户自身注册的回调冲突。
//
// 注意:Apply 在中途注册失败时会留下已注册的回调 (半注册状态)。实践中
// Apply 只在 DB 初始化阶段调用、且失败即 fatal,半注册状态不会到达业务
// 路径;若需严格回滚,请先校验再注册或重建 db。
func (t *tracker) Apply(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	for _, o := range ops {
		if err := o.get(db)(t.before(o.cmd), t.after); err != nil {
			return fmt.Errorf("gormext: register %s callbacks error: %w", o.name, err)
		}
	}
	return nil
}
