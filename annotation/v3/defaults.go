package annotation

import (
	"fmt"
	"reflect"
)

// ApplyDefaults 为 ptr 所指 struct 中当前为零值的叶子设置 default-tag 值。
// 与 v2 的 ParseDefault 不同,它不会分配无关的 nil pointer field(例如无
// default tag 的 *int),因此 config struct 的可选 pointer 保持 nil,除非显式
// 设置了 default。ptr 必须是指向 struct 的非空 pointer。
func ApplyDefaults(ptr any) error {
	p, err := PlanFor(ptr)
	if err != nil {
		return err
	}
	root := reflect.ValueOf(ptr).Elem()
	return applyDefaults(p.fields, root)
}

func applyDefaults(fields []*Field, parent reflect.Value) error {
	for _, f := range fields {
		fv := parent.FieldByIndex(f.Index)

		// 递归进入嵌套 struct;仅当嵌套 struct 确实在某处带有 default 时才实例化
		// pointer(避免无谓的分配)。
		if len(f.children) > 0 {
			if fv.Kind() == reflect.Ptr {
				if fv.IsNil() {
					if !f.hasDefaultDeep() {
						continue // 可选,保持不变
					}
					fv.Set(reflect.New(fv.Type().Elem()))
				}
				if err := applyDefaults(f.children, fv.Elem()); err != nil {
					return err
				}
				continue
			}
			if fv.Kind() == reflect.Struct {
				if err := applyDefaults(f.children, fv); err != nil {
					return err
				}
				continue
			}
		}

		if f.Default == "" {
			continue
		}
		if !isZeroValue(fv) {
			continue // 已设置,保留调用方的值
		}
		if err := setFromString(fv, f.Default); err != nil {
			return fmt.Errorf("field %s: %w", f.Path, err)
		}
	}
	return nil
}

// hasDefaultDeep 报告 f 或任意后代是否带有 default tag。用于决定是否实例化
// 可选的嵌套 *Struct。
func (f *Field) hasDefaultDeep() bool {
	if f.Default != "" {
		return true
	}
	for _, c := range f.children {
		if c.hasDefaultDeep() {
			return true
		}
	}
	return false
}

func setFromString(rv reflect.Value, s string) error {
	rv = peelOnce(rv)
	if !rv.CanSet() {
		return fmt.Errorf("field not settable: %s", rv.Type())
	}
	return SetString(rv, s)
}

// isZeroValue 报告 rv 是否为其类型的零值,仅针对 ApplyDefaults 关心的少数集合
// (scalar、slice、pointer)。
func isZeroValue(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		return rv.IsNil()
	case reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	}
	return false
}
