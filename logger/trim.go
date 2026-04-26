package logger

import (
	"encoding/base64"
	"reflect"
	"strings"
	"time"
)

const (
	defaultArrLimit  = 3
	defaultStrLimit  = 128
	defaultDeepLimit = 10
)

// OutputTrimmer controls how deeply structured values are reflected into log output.
type OutputTrimmer struct {
	arrLimit  int
	strLimit  int
	deepLimit int
	ignores   map[string]bool
}

func newTrimmer(cfg *TrimConfig) *OutputTrimmer {
	if cfg == nil {
		return &OutputTrimmer{
			arrLimit:  defaultArrLimit,
			strLimit:  defaultStrLimit,
			deepLimit: defaultDeepLimit,
		}
	}
	arr := cfg.ArrLimit
	if arr <= 0 {
		arr = defaultArrLimit
	}
	str := cfg.StrLimit
	if str <= 0 {
		str = defaultStrLimit
	}
	deep := cfg.DeepLimit
	if deep <= 0 {
		deep = defaultDeepLimit
	}
	ig := make(map[string]bool, len(cfg.Ignores))
	for _, k := range cfg.Ignores {
		ig[k] = true
	}
	return &OutputTrimmer{arrLimit: arr, strLimit: str, deepLimit: deep, ignores: ig}
}

// TrimFields applies output trimming to a key-value field list.
func (ot *OutputTrimmer) TrimFields(args []any) []any {
	if len(args) < 2 {
		return args
	}
	out := make([]any, 0, len(args))
	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			out = append(out, args[i], args[i+1])
			continue
		}
		if ot.ignores != nil && ot.ignores[key] {
			continue
		}
		out = append(out, key, ot.trimAny(args[i+1], ot.deepLimit))
	}
	return out
}

// TrimArgs applies output trimming to alternating key-value arguments.
func (ot *OutputTrimmer) TrimArgs(args []any) []any {
	return ot.TrimFields(args)
}

func (ot *OutputTrimmer) trimAny(v any, depth int) any {
	if v == nil || depth <= 0 {
		return nil
	}
	rv := reflect.ValueOf(v)

	// Dereference pointers/interfaces.
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint()
	case reflect.Float32, reflect.Float64:
		return rv.Float()
	case reflect.String:
		return ot.trimString(rv.String())
	case reflect.Struct:
		if rv.Type() == timeType {
			return rv.Interface().(time.Time).Format("2006-01-02T15:04:05.000")
		}
		if rv.Type() == durationType {
			return rv.Interface().(time.Duration).String()
		}
		return ot.trimStruct(rv, depth-1)
	case reflect.Map:
		return ot.trimMap(rv, depth-1)
	case reflect.Slice, reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return ot.trimBytes(rv)
		}
		return ot.trimSlice(rv, depth)
	default:
		return fmtValue(rv)
	}
}

func (ot *OutputTrimmer) trimStruct(v reflect.Value, depth int) map[string]any {
	if depth <= 0 {
		return nil
	}
	m := make(map[string]any)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := f.Name
		if tag, ok := f.Tag.Lookup("json"); ok {
			if tag == "-" {
				continue
			}
			if idx := strings.IndexByte(tag, ','); idx >= 0 {
				name = tag[:idx]
			} else {
				name = tag
			}
		}
		if ot.ignores != nil && ot.ignores[name] {
			continue
		}
		if !strings.HasPrefix(name, "XXX_") {
			m[name] = ot.trimAny(v.Field(i).Interface(), depth)
		}
	}
	return m
}

func (ot *OutputTrimmer) trimMap(v reflect.Value, depth int) map[string]any {
	if depth <= 0 {
		return nil
	}
	m := make(map[string]any, v.Len())
	for _, k := range v.MapKeys() {
		name := fmtValue(k)
		if ot.ignores != nil && ot.ignores[name] {
			continue
		}
		m[name] = ot.trimAny(v.MapIndex(k).Interface(), depth)
	}
	return m
}

func (ot *OutputTrimmer) trimSlice(v reflect.Value, depth int) []any {
	l := v.Len()
	if l == 0 {
		return nil
	}
	limit := l
	if limit > ot.arrLimit {
		limit = ot.arrLimit
	}
	arr := make([]any, limit)
	for i := 0; i < limit; i++ {
		arr[i] = ot.trimAny(v.Index(i).Interface(), depth)
	}
	return arr
}

func (ot *OutputTrimmer) trimBytes(v reflect.Value) string {
	l := v.Len()
	if l == 0 {
		return "[]"
	}
	limit := ot.arrLimit
	if ot.strLimit > limit {
		limit = ot.strLimit
	}
	if l <= limit {
		return base64.StdEncoding.EncodeToString(v.Bytes())
	}
	return ""
}

func (ot *OutputTrimmer) trimString(s string) string {
	if ot.strLimit > 0 && len(s) > ot.strLimit {
		return s[:ot.strLimit] + "..."
	}
	return s
}

var (
	timeType     = reflect.TypeOf(time.Now())
	durationType = reflect.TypeOf(time.Second)
)

func fmtValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return itoa(int(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return itoa(int(v.Uint()))
	default:
		return ""
	}
}

func itoa(n int) string {
	// Quick path for small integers.
	if n < 10 && n > -1 {
		return string(rune('0' + n))
	}
	return string([]byte{
		byte('0' + n/100),
		byte('0' + (n/10)%10),
		byte('0' + n%10),
	})
}
