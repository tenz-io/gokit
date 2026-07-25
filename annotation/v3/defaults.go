package annotation

import (
	"fmt"
	"reflect"
)

// ApplyDefaults sets default-tag values on the struct behind ptr for any leaf
// that is currently the zero value. Unlike v2's ParseDefault it does NOT
// allocate unrelated nil pointer fields (e.g. a *int with no default tag),
// so a config struct's optional pointers stay nil unless explicitly defaulted.
// ptr must be a non-nil pointer to a struct.
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

		// Recurse into nested structs; instantiate the pointer only if the
		// nested struct actually carries a default somewhere (avoids spurious
		// allocations).
		if len(f.children) > 0 {
			if fv.Kind() == reflect.Ptr {
				if fv.IsNil() {
					if !f.hasDefaultDeep() {
						continue // optional, untouched
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
			continue // already set, keep caller's value
		}
		if err := setFromString(fv, f.Default); err != nil {
			return fmt.Errorf("field %s: %w", f.Path, err)
		}
	}
	return nil
}

// hasDefaultDeep reports whether f or any descendant carries a default tag.
// Used to decide whether to instantiate an optional nested *Struct.
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

// isZeroValue reports whether rv is its type's zero value, for the small set
// ApplyDefaults cares about (scalars, slices, pointers).
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
