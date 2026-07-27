package logger

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	defaultArrLimit  = 3
	defaultStrLimit  = 128
	defaultDeepLimit = 10
)

// OutputTrimmer 控制结构化值被反射进日志输出的深度:长字符串被截断、
// 大 slice 被封顶、深层 struct/map 被扁平化、具名 field 被丢弃。
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
	return &OutputTrimmer{
		arrLimit:  orDefault(cfg.ArrLimit, defaultArrLimit),
		strLimit:  orDefault(cfg.StrLimit, defaultStrLimit),
		deepLimit: orDefault(cfg.DeepLimit, defaultDeepLimit),
		ignores:   keySet(cfg.Ignores),
	}
}

func orDefault(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}

func keySet(ks []string) map[string]bool {
	m := make(map[string]bool, len(ks))
	for _, k := range ks {
		m[k] = true
	}
	return m
}

// TrimFields 对键值 field 列表应用输出裁剪,完全丢弃被忽略的 key。
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
		if ot.ignores[key] {
			continue
		}
		out = append(out, key, ot.trimAny(args[i+1], ot.deepLimit))
	}
	if len(args)%2 != 0 {
		// 保留畸形输入,以便 zap 能报告这个悬空的 key,而不是
		// 静默丢失调用方数据。
		out = append(out, args[len(args)-1])
	}
	return out
}

// TrimArgs 是 TrimFields 的可变参数孪生版本。
func (ot *OutputTrimmer) TrimArgs(args []any) []any { return ot.TrimFields(args) }

func (ot *OutputTrimmer) trimAny(v any, depth int) any {
	if v == nil {
		return nil
	}
	if d, ok := v.(time.Duration); ok {
		return d.String()
	}
	if tm, ok := v.(time.Time); ok {
		return tm.Format("2006-01-02T15:04:05.000")
	}
	if err, ok := v.(error); ok {
		return ot.trimString(err.Error())
	}
	if s, ok := v.(fmt.Stringer); ok {
		return ot.trimString(s.String())
	}
	rv := reflect.ValueOf(v)

	// 解引用 pointer/interface。
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Type() == timeType {
		return rv.Interface().(time.Time).Format("2006-01-02T15:04:05.000")
	}
	if rv.Type() == durationType {
		return rv.Interface().(time.Duration).String()
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
		if depth <= 0 {
			return nil
		}
		return ot.trimStruct(rv, depth)
	case reflect.Map:
		if depth <= 0 {
			return nil
		}
		return ot.trimMap(rv, depth)
	case reflect.Slice, reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return ot.trimBytes(rv)
		}
		if depth <= 0 {
			return nil
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
		name, ok := fieldName(f)
		if !ok {
			continue
		}
		if ot.ignores[name] {
			continue
		}
		m[name] = ot.trimAny(v.Field(i).Interface(), depth-1)
	}
	return m
}

// fieldName 从 json tag 解析对外名称,回退到 Go field 名。
func fieldName(f reflect.StructField) (string, bool) {
	if tag, ok := f.Tag.Lookup("json"); ok {
		if tag == "-" {
			return "", false
		}
		if idx := strings.IndexByte(tag, ','); idx >= 0 {
			tag = tag[:idx]
		}
		if tag != "" {
			return tag, true
		}
	}
	return f.Name, true
}

func (ot *OutputTrimmer) trimMap(v reflect.Value, depth int) map[string]any {
	if depth <= 0 {
		return nil
	}
	m := make(map[string]any, v.Len())
	for _, k := range v.MapKeys() {
		name := fmtValue(k)
		if ot.ignores[name] {
			continue
		}
		m[name] = ot.trimAny(v.MapIndex(k).Interface(), depth-1)
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
		arr[i] = ot.trimAny(v.Index(i).Interface(), depth-1)
	}
	return arr
}

func (ot *OutputTrimmer) trimBytes(v reflect.Value) string {
	l := v.Len()
	if l == 0 {
		return "[]"
	}
	limit := l
	if limit > ot.arrLimit {
		limit = ot.arrLimit
	}
	b := make([]byte, limit)
	for i := 0; i < limit; i++ {
		b[i] = byte(v.Index(i).Uint())
	}
	encoded := base64.StdEncoding.EncodeToString(b)
	if limit < l {
		encoded += "..."
	}
	return ot.trimString(encoded)
}

func (ot *OutputTrimmer) trimString(s string) string {
	if ot.strLimit > 0 && len(s) > ot.strLimit {
		end := ot.strLimit
		for end > 0 && !utf8.ValidString(s[:end]) {
			end--
		}
		return s[:end] + "..."
	}
	return s
}

var (
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Duration(0))
)

// fmtValue 将 reflect.Value 的基本形态渲染为字符串,用于 map key 及其他
// 裸标量。
func fmtValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	if v.CanInterface() {
		return fmt.Sprint(v.Interface())
	}
	return fmt.Sprint(v)
}
