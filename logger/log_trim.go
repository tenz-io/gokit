package logger

import (
	"encoding/json"
	"fmt"
	syslog "log"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
)

const (
	defaultArrLimit   = 3
	defaultStrLimit   = 128
	defaultDeepLimit  = 10
	defaultWholeLimit = 4096
)

type ObjectTrimmer struct {
	ArrLimit   int
	StrLimit   int
	DeepLimit  int
	WholeLimit int
	Ignores    []string
}

type TrimOption func(*ObjectTrimmer)

func WithArrLimit(limit int) TrimOption {
	return func(t *ObjectTrimmer) {
		t.ArrLimit = limit
	}
}

func WithStrLimit(limit int) TrimOption {
	return func(t *ObjectTrimmer) {
		t.StrLimit = limit
	}
}

func WithDeepLimit(limit int) TrimOption {
	return func(t *ObjectTrimmer) {
		t.DeepLimit = limit
	}
}

func WithWholeLimit(limit int) TrimOption {
	return func(t *ObjectTrimmer) {
		t.WholeLimit = limit
	}
}

func WithIgnores(ignores ...string) TrimOption {
	return func(t *ObjectTrimmer) {
		t.Ignores = ignores
	}
}

func JsonObjectWithOpts(obj any, opts ...TrimOption) string {
	j, err := json.Marshal(TrimObjectWithOpts(obj, opts...))
	if err != nil {
		return ""
	}
	return string(j)
}

func TrimObject(obj any) any {
	return TrimObjectWithOpts(obj)
}

func TrimObjectWithOpts(obj any, opts ...TrimOption) (ret any) {
	// panic recover
	defer func() {
		if r := recover(); r != nil {
			syslog.Printf("panic recovery: %s, stacktrace: %s\n", r, string(debug.Stack()))
			ret = fmt.Sprintf("panic recovery: %s", r)
		}
	}()

	trimmer := &ObjectTrimmer{
		ArrLimit:   defaultArrLimit,
		StrLimit:   defaultStrLimit,
		DeepLimit:  defaultDeepLimit,
		WholeLimit: defaultWholeLimit,
		Ignores:    []string{},
	}

	for _, opt := range opts {
		opt(trimmer)
	}

	return trimObjectWithIgnores(obj, trimmer.ArrLimit, trimmer.StrLimit, trimmer.DeepLimit, trimmer.Ignores...)
}

func trimObjectWithIgnores(obj any, arrLmt, strLmt, deepLmt int, ignores ...string) any {
	ignoreMap := make(map[string]bool)
	if len(ignores) > 0 {
		for _, ignore := range ignores {
			ignoreMap[ignore] = true
		}
	}

	return trimObject(obj, arrLmt, strLmt, deepLmt, ignoreMap)
}

func trimObject(obj any, arrLmt, strLmt, deepLmt int, ignores map[string]bool) any {
	if obj == nil {
		return nil
	}

	v := reflect.ValueOf(obj)

	if isNonValuableType(v) {
		return nil
	}

	if val, ok := valOfSupportType(v, arrLmt, strLmt); ok {
		return val
	}

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Ptr:
		// should not happen
	case reflect.Struct:
		return trimStruct(v, arrLmt, strLmt, deepLmt-1, ignores)
	case reflect.Map:
		return trimMap(v, arrLmt, strLmt, deepLmt-1, ignores)
	case reflect.Array, reflect.Slice:
		return trimSlice(v, arrLmt, strLmt, deepLmt, ignores)
	default:
		//ignore
	}

	return nil
}

func trimStruct(v reflect.Value, arrLmt, strLmt, deepLmt int, ignores map[string]bool) map[string]any {
	m := make(map[string]any)
	if deepLmt <= 0 {
		return m
	}

	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name

		// get json tag
		if tag := t.Field(i).Tag.Get("json"); tag != "" {
			if tag == "-" {
				continue
			}
			if idx := strings.Index(tag, ","); idx >= 0 {
				tag = tag[:idx]
			}
			if tag != "" {
				fieldName = tag
			}
		}

		if !visibleName(fieldName, ignores) {
			continue
		}

		fv := v.Field(i)

		if isNonValuableType(fv) {
			continue
		}

		if val, ok := valOfSupportType(fv, arrLmt, strLmt); ok {
			m[fieldName] = val
			continue
		}

		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		switch fv.Kind() {
		case reflect.Ptr:
			// should never happen
		case reflect.Struct:
			if sv := trimStruct(fv, arrLmt, strLmt, deepLmt-1, ignores); len(sv) > 0 {
				m[fieldName] = sv
			}
		case reflect.Map:
			if mv := trimMap(fv, arrLmt, strLmt, deepLmt-1, ignores); len(mv) > 0 {
				m[fieldName] = trimMap(fv, arrLmt, strLmt, deepLmt-1, ignores)
			}
		case reflect.Array, reflect.Slice:
			if sv := trimSlice(fv, arrLmt, strLmt, deepLmt, ignores); len(sv) > 0 {
				m[fieldName] = trimSlice(fv, arrLmt, strLmt, deepLmt, ignores)
				m["_size__"+fieldName] = fv.Len()
			}
		case reflect.Interface:
			if iv := trimObject(fv.Interface(), arrLmt, strLmt, deepLmt-1, ignores); iv != nil {
				m[fieldName] = iv
			}
		default:
			// ignore
		}
	}

	return m
}

func trimMap(v reflect.Value, arrLmt, strLmt, deepLmt int, ignores map[string]bool) map[string]any {
	m := make(map[string]any)
	if deepLmt <= 0 {
		return m
	}

	if v.Kind() != reflect.Map {
		return m
	}
	for _, k := range v.MapKeys() {
		if !visibleName(k.String(), ignores) {
			continue
		}

		fv := v.MapIndex(k)

		if isNonValuableType(fv) {
			continue
		}

		if val, ok := valOfSupportType(fv, arrLmt, strLmt); ok {
			m[k.String()] = val
			continue
		}

		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		switch fv.Kind() {
		case reflect.Ptr:
		// should never happen
		case reflect.Map:
			m[k.String()] = trimMap(fv, arrLmt, strLmt, deepLmt-1, ignores)
		case reflect.Struct:
			m[k.String()] = trimStruct(fv, arrLmt, strLmt, deepLmt-1, ignores)
		case reflect.Array, reflect.Slice:
			m[k.String()] = trimSlice(fv, arrLmt, strLmt, deepLmt, ignores)
		case reflect.Interface:
			m[k.String()] = trimObject(fv.Interface(), arrLmt, strLmt, deepLmt-1, ignores)
		default:
			//ignore
		}
	}

	return m
}

func trimSlice(v reflect.Value, arrLmt, strLmt, deepLmt int, ignores map[string]bool) []any {
	var arr []any
	l := v.Len()

	if l == 0 {
		return arr
	}

	if l > arrLmt {
		l = arrLmt
	}

	for i := 0; i < l; i++ {
		fv := v.Index(i)

		if isNonValuableType(fv) {
			continue
		}

		if val, ok := valOfSupportType(fv, arrLmt, strLmt); ok {
			arr = append(arr, val)
			continue
		}

		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		switch fv.Kind() {
		case reflect.Ptr:
		// should never happen
		case reflect.Struct:
			arr = append(arr, trimStruct(fv, arrLmt, strLmt, deepLmt-1, ignores))
		case reflect.Map:
			arr = append(arr, trimMap(fv, arrLmt, strLmt, deepLmt-1, ignores))
		case reflect.Array, reflect.Slice:
		// seems like a arr of arr
		// ignore the inner arr
		//arr = append(arr, trimSlice(fv, arrLmt))
		case reflect.Interface:
			arr = append(arr, trimObject(fv.Interface(), arrLmt, strLmt, deepLmt-1, ignores))
		default:
			//ignore
		}
	}

	return arr
}

var (
	errType      = reflect.TypeOf(fmt.Errorf(""))
	timeType     = reflect.TypeOf(time.Now())
	durationType = reflect.TypeOf(time.Second)
	bytesType    = reflect.TypeOf([]byte{})
	stringType   = reflect.TypeOf("")
	timeFormat   = "2006-01-02T15:04:05.000"
)

// valOfSpecialType returns the value of a special type
func valOfSpecialType(v reflect.Value, arrLmt, strLmt int) (val any, ok bool) {
	if isNonValuableType(v) {
		return nil, false
	}

	// if v is kind of error, return the error message
	switch v.Type() {
	case stringType:
		s := v.String()
		return StringLimit(s, strLmt), true
	case errType:
		return v.Interface().(error).Error(), true
	case timeType:
		return v.Interface().(time.Time).Format(timeFormat), true
	case durationType:
		return v.Interface().(time.Duration).String(), true
	default:
		//ignore
	}

	return nil, false
}

// valOfSupportType returns the value of a support type
func valOfSupportType(v reflect.Value, arrLmt, strLmt int) (val any, ok bool) {
	if isNonValuableType(v) {
		return nil, false
	}

	if val, ok = valOfSpecialType(v, arrLmt, strLmt); ok {
		return val, true
	}

	if val, ok = valOfPrimaryType(v, arrLmt, strLmt); ok {
		return val, true
	}

	return nil, false
}

// valOfPrimaryType returns the value of a primary type or pointer to a primary type
func valOfPrimaryType(v reflect.Value, arrLmt, strLmt int) (val any, ok bool) {
	if isNonValuableType(v) {
		return nil, false
	}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Bool:
		return v.Bool(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint(), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.Complex64, reflect.Complex128:
		return v.Complex(), true
	case reflect.String:
		return StringLimit(v.String(), strLmt), true
	default:
		//ignore
	}

	return nil, false
}

// isNonValuableType returns true if the value is not valuable
func isNonValuableType(v reflect.Value) bool {
	if v == reflect.ValueOf(nil) {
		return true
	}
	if !v.CanInterface() {
		return true
	}

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return true
	}
	if v.Kind() == reflect.Interface && v.IsNil() {
		return true
	}

	if v.Kind() == reflect.Invalid {
		return true
	}

	return false
}

// StringLimit returns a string with limited length at most
func StringLimit(s string, limit int) string {
	if limit <= 0 {
		return s
	}
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}

func visibleName(filedName string, ignores map[string]bool) bool {
	if len(ignores) > 0 {
		if _, ok := ignores[filedName]; ok {
			return false
		}
	}

	if strings.HasPrefix(filedName, "XXX_") {
		//skip proto unknown fields
		return false
	}
	return true
}

func visibleVal(val reflect.Value) bool {
	displayable := true
	switch val.Kind() {
	case reflect.Bool:
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	case reflect.Float32, reflect.Float64:
	case reflect.Complex64, reflect.Complex128:
	case reflect.String:
	case reflect.Struct:
	case reflect.Array, reflect.Slice, reflect.Map:
		if val.IsNil() || val.Len() == 0 {
			displayable = false
		}
	case reflect.Pointer, reflect.Interface:
		if val.IsNil() {
			displayable = false
		} else {
			displayable = visibleVal(val.Elem())
		}

	case reflect.Invalid:
		displayable = false
	default:
		displayable = false
	}
	return displayable
}

func ifThen(cond bool, a, b any) any {
	if cond {
		return a
	}
	return b
}
