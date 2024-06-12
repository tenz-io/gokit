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

const (
	arrFieldPrefix = "_size__"
)

var (
	defaultOutputTrimmer = OutputTrimmer{
		arrLimit:  defaultArrLimit,
		strLimit:  defaultStrLimit,
		deepLimit: defaultDeepLimit,
		ignores:   make(map[string]bool),
	}
)

type OutputTrimmer struct {
	arrLimit  int
	strLimit  int
	deepLimit int
	ignores   map[string]bool
}

type TrimOption func(*OutputTrimmer)

func SetupDefaultTrimmer(opts ...TrimOption) {
	trimmer := NewOutputTrimmer(opts...)
	defaultOutputTrimmer = *trimmer
}

func WithArrLimit(limit int) TrimOption {
	return func(t *OutputTrimmer) {
		t.arrLimit = limit
	}
}

func WithStrLimit(limit int) TrimOption {
	return func(t *OutputTrimmer) {
		t.strLimit = limit
	}
}

func WithDeepLimit(limit int) TrimOption {
	return func(t *OutputTrimmer) {
		t.deepLimit = limit
	}
}

func WithIgnores(ignores ...string) TrimOption {
	return func(t *OutputTrimmer) {
		if t.ignores == nil {
			t.ignores = make(map[string]bool)
		}
		for _, ignore := range ignores {
			t.ignores[ignore] = true
		}
	}
}

func NewOutputTrimmer(opts ...TrimOption) *OutputTrimmer {
	ot := defaultOutputTrimmer
	for _, opt := range opts {
		opt(&ot)
	}
	return &ot
}

func ObjectTrimWithOpts(obj any, opts ...TrimOption) any {
	return NewOutputTrimmer(opts...).TrimObject(obj)
}

func (ot *OutputTrimmer) TrimObject(obj any) (ret any) {
	return ot.trimObject(obj, ot.deepLimit)
}

func (ot *OutputTrimmer) trimObject(obj any, deepLmt int) any {
	if obj == nil || deepLmt <= 0 {
		return nil
	}

	v := reflect.ValueOf(obj)
	if val, ok := ot.valOfSupportType(v); ok {
		return val
	}

	for ; v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface; v = v.Elem() {
		if v.IsNil() {
			return nil
		}
	}

	switch v.Kind() {
	case reflect.Struct:
		return ot.trimStruct(v, deepLmt-1)
	case reflect.Map:
		return ot.trimMap(v, deepLmt-1)
	case reflect.Array, reflect.Slice:
		return ot.trimSlice(v, deepLmt)
	default:
		//ignore
	}

	return nil
}

func (ot *OutputTrimmer) trimStruct(v reflect.Value, deepLmt int) map[string]any {
	var (
		m = map[string]any{}
	)

	if deepLmt <= 0 {
		return m
	}

	if v.Kind() != reflect.Struct {
		return m
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name

		// ignore unexported field
		if fieldName[0] < 'A' || fieldName[0] > 'Z' {
			continue
		}

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

		if !ot.visibleName(fieldName) {
			continue
		}

		fv := v.Field(i)
		if val, ok := ot.valOfSupportType(fv); ok {
			m[fieldName] = val
			continue
		}

		for ; fv.Kind() == reflect.Ptr || fv.Kind() == reflect.Interface; fv = fv.Elem() {
			if fv.IsNil() {
				break
			}
		}

		switch fv.Kind() {
		case reflect.Ptr, reflect.Interface:
			if fv.IsNil() {
				continue
			}
		case reflect.Struct:
			if ret := ot.trimStruct(fv, deepLmt-1); len(ret) > 0 {
				m[fieldName] = ret
			}
		case reflect.Map:
			if ret := ot.trimMap(fv, deepLmt-1); len(ret) > 0 {
				m[fieldName] = ret
			}
		case reflect.Array, reflect.Slice:
			if ret := ot.trimSlice(fv, deepLmt); len(ret) > 0 {
				m[fieldName] = ret
				m[arrFieldPrefix+fieldName] = fv.Len()
			}
		default:
			// ignore
		}
	}

	return m
}

func (ot *OutputTrimmer) trimMap(v reflect.Value, deepLmt int) map[string]any {
	m := make(map[string]any)
	if deepLmt <= 0 {
		return m
	}

	if v.Kind() != reflect.Map {
		return m
	}

	for _, k := range v.MapKeys() {
		if !ot.visibleName(k.String()) {
			continue
		}

		fv := v.MapIndex(k)
		if val, ok := ot.valOfSupportType(fv); ok {
			m[k.String()] = val
			continue
		}

		for ; fv.Kind() == reflect.Ptr || fv.Kind() == reflect.Interface; fv = fv.Elem() {
			if fv.IsNil() {
				break
			}
		}

		switch fv.Kind() {
		case reflect.Ptr, reflect.Interface:
			if fv.IsNil() {
				continue
			}
		case reflect.Map:
			if ret := ot.trimMap(fv, deepLmt-1); len(ret) > 0 {
				m[k.String()] = ret
			}
		case reflect.Struct:
			if ret := ot.trimStruct(fv, deepLmt-1); len(ret) > 0 {
				m[k.String()] = ret
			}
		case reflect.Array, reflect.Slice:
			if ret := ot.trimSlice(fv, deepLmt); len(ret) > 0 {
				m[k.String()] = ret
			}
		default:
			//ignore
		}
	}

	return m
}

func (ot *OutputTrimmer) trimSlice(v reflect.Value, deepLmt int) []any {
	var (
		arr []any
		l   = v.Len()
	)

	if l == 0 {
		return arr
	}

	if l > ot.arrLimit {
		l = ot.arrLimit
	}

	for i := 0; i < l; i++ {
		fv := v.Index(i)

		if val, ok := ot.valOfSupportType(fv); ok {
			arr = append(arr, val)
			continue
		}

		for ; fv.Kind() == reflect.Ptr || fv.Kind() == reflect.Interface; fv = fv.Elem() {
			if fv.IsNil() {
				break
			}
		}

		switch fv.Kind() {
		case reflect.Ptr, reflect.Interface:
			if fv.IsNil() {
				arr = append(arr, nil)
			}
		case reflect.Struct:
			arr = append(arr, ot.trimStruct(fv, deepLmt-1))
		case reflect.Map:
			arr = append(arr, ot.trimMap(fv, deepLmt-1))
		case reflect.Array, reflect.Slice:
			arr = append(arr, ot.trimSlice(fv, deepLmt-1))
		default:
			//ignore
		}
	}

	return arr
}

var (
	errType      = reflect.TypeOf((*error)(nil)).Elem()
	timeType     = reflect.TypeOf(time.Now())
	durationType = reflect.TypeOf(time.Second)
	bytesType    = reflect.TypeOf([]byte{})
	strType      = reflect.TypeOf("")

	timeFormat = "2006-01-02T15:04:05.000"
)

// valOfSpecialType returns the value of a special type
func (ot *OutputTrimmer) valOfSpecialType(v reflect.Value) (val any, ok bool) {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil, false
		}

		// if v is type of error
		if v.CanInterface() && v.Type().Implements(errType) {
			return v.Interface().(error).Error(), true
		}

		v = v.Elem()
	}

	if v.CanInterface() {
		interfacedValue := v.Interface()

		// Check if the value implements the error interface
		if err, isError := interfacedValue.(error); isError {
			return err.Error(), true
		}

		switch v.Type() {
		case timeType:
			return interfacedValue.(time.Time).Format(timeFormat), true
		case durationType:
			return interfacedValue.(time.Duration).String(), true
		case strType:
			return ot.stringLimit(v.String()), true
		default:
			if isBytes(v) {
				if val, ok := ot.bytesString(v); ok {
					return val, true
				}
			}
		}

	}

	return nil, false
}

func isBytes(v reflect.Value) bool {
	return (v.Kind() == reflect.Slice || v.Kind() == reflect.Array) && v.Type().Elem().Kind() == reflect.Uint8
}

func (ot *OutputTrimmer) bytesString(v reflect.Value) (string, bool) {
	vlen := v.Len()
	if vlen == 0 {
		return "[]", true
	}

	maxLen := ifThen(ot.arrLimit > ot.strLimit, ot.arrLimit, ot.strLimit)
	if vlen <= maxLen {
		// if v is a byte slice
		if v.Type().AssignableTo(bytesType) {
			return base64.StdEncoding.EncodeToString(v.Bytes()), true
		}

		// if v is a byte array
		bs := make([]byte, vlen)
		reflect.Copy(reflect.ValueOf(bs), v)
		return base64.StdEncoding.EncodeToString(bs), true
	}

	return "", false
}

// valOfSupportType returns the value of a support type
func (ot *OutputTrimmer) valOfSupportType(v reflect.Value) (val any, ok bool) {
	if !valuableType(v) {
		return nil, false
	}

	if val, ok = ot.valOfSpecialType(v); ok {
		return val, true
	}

	if val, ok = ot.valOfPrimaryType(v); ok {
		return val, true
	}

	return nil, false
}

// valOfPrimaryType returns the value of a primary type or pointer to a primary type
func (ot *OutputTrimmer) valOfPrimaryType(v reflect.Value) (val any, ok bool) {
	if !valuableType(v) {
		return nil, false
	}

	for ; v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface; v = v.Elem() {
		if v.IsNil() {
			return nil, false
		}
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
		return ot.stringLimit(v.String()), true
	default:
		//ignore
	}

	return nil, false
}

// valuableType return ture if the value is valuable
func valuableType(v reflect.Value) bool {
	for ; v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface; v = v.Elem() {
		if v.IsNil() {
			return false
		}
	}

	switch v.Kind() {
	case reflect.Bool,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint,
		reflect.Float32, reflect.Float64,
		reflect.String,
		reflect.Complex64, reflect.Complex128,
		reflect.Struct:
		return true
	case reflect.Map, reflect.Slice, reflect.Array:
		return v.Len() > 0
	default:
		return false
	}
}

// stringLimit returns a string with limited length at most
func (ot *OutputTrimmer) stringLimit(s string) string {
	if ot.strLimit <= 0 {
		return s
	}
	if len(s) > ot.strLimit {
		return s[:ot.strLimit] + "..."
	}
	return s
}

func (ot *OutputTrimmer) visibleName(filedName string) bool {
	if len(ot.ignores) > 0 {
		if _, ok := ot.ignores[filedName]; ok {
			return false
		}
	}

	if strings.HasPrefix(filedName, "XXX_") {
		//skip proto unknown fields
		return false
	}
	return true
}
