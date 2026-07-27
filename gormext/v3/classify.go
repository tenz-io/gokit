package gormext

import (
	"errors"

	"gorm.io/gorm"
)

// errClass 是统一的错误分类,被 metrics/traffic/errorLog 共用,保证三层
// 对同一种错误看到一致的语义 (而非 v3 早期 metrics 把 not_found 当 err、
// errorLog 却当 Debug 的口径分裂)。
type errClass string

const (
	classOK       errClass = "ok"        // 无错误
	classNotFound errClass = "not_found" // gorm.ErrRecordNotFound
	classErr      errClass = "err"       // 其余数据库错误
)

// classify 把 db.Error 归入统一分类。metrics/traffic/errorLog 都通过它
// 决定 code/级别,避免口径分裂。
func classify(err error) errClass {
	switch {
	case err == nil:
		return classOK
	case errors.Is(err, gorm.ErrRecordNotFound):
		return classNotFound
	default:
		return classErr
	}
}

// monitorCode 把分类映射为 monitor/v3 的 code label。
//
// 关键设计:not_found 归为 "0" (ok),因此正常的"查无此记录"不会拉高数据库
// 失败率 (monitor 的 normalizeCode 会把任何非 "0" 折叠为 "1"/err,所以
// 这里必须显式传 "0" 才能让 not_found 计为成功)。这与 errorLog 把
// not_found 降级为 Debug 的意图一致:它不是一次"失败"。
func monitorCode(class errClass) string {
	switch class {
	case classOK, classNotFound:
		return "0" // codeOK
	default:
		return "1" // codeErr
	}
}

// trafficCode 把分类映射为 traffic 日志的 code 字段 (自由字符串,不做
// 归一化,因此比 monitor 的 0/1 更可读)。
func trafficCode(class errClass) string { return string(class) }

// redactVars 按 config 决定返回用于日志的 (sql, vars)。
//
//   - EnableVars=false:返回原 sql 与空 vars (不记参数,防泄露)。
//   - EnableVars=true 且 Redactor!=nil:返回脱敏后的 sql/vars。
//   - EnableVars=true 且 Redactor==nil:返回原 sql/vars。
//
// 返回的 vars 是新切片或 nil,不会修改调用方传入的 db.Statement.Vars。
func (t *tracker) redactVars(sql string, vars []any) (string, []any) {
	if !t.config.EnableVars {
		return sql, nil
	}
	if t.config.Redactor == nil {
		// 返回一个副本,避免调用方后续修改影响已记录的字段。
		out := make([]any, len(vars))
		copy(out, vars)
		return sql, out
	}
	_, out := t.config.Redactor(sql, vars)
	return sql, out
}
