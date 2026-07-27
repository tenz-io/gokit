package annotation

import (
	"reflect"
	"sync"
)

// Rule 是在 plan-build 时绑定到某个 field 的已编译检查。它已捕获任何解析过的
// 参数(编译后的 regex、float64 阈值、...),因此运行时检查无需分配和解析。
// 它返回值是否满足规则,失败时返回一条自然消息。
type Rule func(rv reflect.Value) (ok bool, msg string)

// namedRule 将 Rule 与其标识 name/param(用于 FieldError)以及由尾部 msg=...
// 修饰符设置的可选消息覆盖配对。
type namedRule struct {
	name  string
	param string
	run   Rule
	msg   string // 非空时覆盖规则的自然失败消息
}

// Validator 为某个 field 编译规则。给定原始参数字符串和 field 的
// StructField,返回运行时检查或一个 config 错误(参数错误、不支持的类型)。
// config 错误以永久失败的规则呈现,而非 panic,因此拼写错误不会使请求崩溃。
type Validator func(param string, ft reflect.StructField) (Rule, error)

var (
	registryMu sync.RWMutex
	registry   = map[string]Validator{}
)

// Register 在给定规则名下添加或替换 validator。可安全并发调用;自定义规则
// 应在 init() 时注册,以便在任何 Plan 构建前可用。
func Register(name string, v Validator) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[name] = v
}

func lookupValidator(name string) (Validator, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	v, ok := registry[name]
	return v, ok && v != nil
}
