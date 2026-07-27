// Package annotation 是一个由 struct tag 驱动的 Go struct 工具包:
// 声明式 default、可插拔的 validator,以及一个缓存 field plan,供 HTTP 层
// (或任意 transport)复用以进行绑定。
//
// 对 struct 的一次遍历产生一个缓存的 Plan;Validate、ApplyDefaults 和外部
// binder 都消费同一 plan,而非重复 reflect。
package annotation

import (
	"fmt"
	"strings"
)

// FieldError 是 struct field 上的单个校验失败。
type FieldError struct {
	// Field 是指向该 field 的点分路径,例如 "Config.Addr.Street"。
	Field string
	// Rule 是失败规则的 name,例如 "required"、"gt"。
	// 对于 bind 失败等临时错误,为空。
	Rule string
	// Param 是规则的原始 param,例如 "0"、"^a-z+$"。
	Param string
	// Msg 是人类可读的说明。
	Msg string
}

// Error 实现 error。
func (e FieldError) Error() string {
	switch {
	case e.Msg == "":
		return e.Field + ": invalid"
	case e.Rule == "":
		return e.Field + ": " + e.Msg
	default:
		return e.Field + ": " + e.Msg
	}
}

// ValidationErrors 是校验 struct 时发现的每个失败的有序集合。与在首个错误
// 处返回不同,它累积所有问题,使调用方可一次性上报。
type ValidationErrors []FieldError

// Has 报告是否收集到了任何错误。
func (v ValidationErrors) Has() bool { return len(v) > 0 }

// Error 实现 error。
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, e := range v {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(e.Error())
	}
	return sb.String()
}

// NewFieldError 构建一个 FieldError。
func NewFieldError(field, rule, param, msg string) FieldError {
	return FieldError{Field: field, Rule: rule, Param: param, Msg: msg}
}

// Err 是一个便捷构造函数,返回单错误的 ValidationErrors,适用于规则引擎之外
// 的临时失败(例如格式错误的请求体)。rule 可为空。
func Err(field, rule, msg string) ValidationErrors {
	return ValidationErrors{{Field: field, Rule: rule, Msg: msg}}
}

// Errf 类似于 Err,但使用格式化消息。
func Errf(field, rule, format string, args ...any) ValidationErrors {
	return ValidationErrors{{Field: field, Rule: rule, Msg: fmt.Sprintf(format, args...)}}
}
