package annotation

import (
	"reflect"
	"sync"
)

// Rule is a compiled check bound to one field at plan-build time. It has
// already captured any parsed parameter (a compiled regex, a float64
// threshold, ...), so the runtime check is allocation- and parse-free. It
// returns whether the value satisfied the rule and, on failure, a natural
// message.
type Rule func(rv reflect.Value) (ok bool, msg string)

// namedRule pairs a Rule with its identifying name/param (for FieldError) and
// an optional message override set by a trailing msg=... modifier.
type namedRule struct {
	name  string
	param string
	run   Rule
	msg   string // non-empty overrides the rule's natural failure message
}

// Validator compiles a rule for a field. Given the raw parameter string and
// the field's StructField, it returns the runtime check or a config error
// (bad parameter, unsupported type). Config errors are surfaced as a
// permanently-failing rule rather than a panic, so a typo never crashes a
// request.
type Validator func(param string, ft reflect.StructField) (Rule, error)

var (
	registryMu sync.RWMutex
	registry   = map[string]Validator{}
)

// Register adds or replaces a validator under the given rule name. It is safe
// to call concurrently; custom rules should be registered at init() time so
// they are available before any Plan is built.
func Register(name string, v Validator) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[name] = v
}

func lookupValidator(name string) (Validator, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	v, ok := registry[name]
	return v, ok
}
