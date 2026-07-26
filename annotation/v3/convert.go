package annotation

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// SetString assigns a string-encoded value into rv (the field value resolved
// from a Plan). It supports the same scalar set as the v2 setter plus
// time.Duration and []byte, peeling one pointer level. It is intended for
// transport binders that receive raw strings (uri/query/header/form).
func SetString(rv reflect.Value, s string) error {
	var err error
	rv, err = writableValue(rv, "annotation.SetString")
	if err != nil {
		return err
	}
	if rv.Type() == durationType {
		d, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("invalid duration %q: %w", s, err)
		}
		rv.SetInt(int64(d))
		return nil
	}
	switch rv.Kind() {
	case reflect.String:
		rv.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(s, 10, rv.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid int %q: %w", s, err)
		}
		rv.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(s, 10, rv.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid uint %q: %w", s, err)
		}
		rv.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(s, rv.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid float %q: %w", s, err)
		}
		rv.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("invalid bool %q: %w", s, err)
		}
		rv.SetBool(v)
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			rv.SetBytes([]byte(s))
		} else {
			return fmt.Errorf("annotation.SetString: unsupported slice element %s", rv.Type().Elem())
		}
	default:
		return fmt.Errorf("annotation.SetString: unsupported kind %s", rv.Kind())
	}
	return nil
}

// Set assigns a typed value into rv, converting when the types are assignable or
// convertible. It is intended for transport binders that already have a typed
// value (e.g. a file's []byte).
func Set(rv reflect.Value, v any) error {
	var err error
	rv, err = writableValue(rv, "annotation.Set")
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return nil
	}
	// Direct assign.
	if val.Type() == rv.Type() {
		rv.Set(val)
		return nil
	}
	// Convertible.
	if val.Type().ConvertibleTo(rv.Type()) {
		rv.Set(val.Convert(rv.Type()))
		return nil
	}
	return fmt.Errorf("annotation.Set: cannot convert %s to %s", val.Type(), rv.Type())
}

func writableValue(rv reflect.Value, operation string) (reflect.Value, error) {
	if !rv.IsValid() {
		return reflect.Value{}, fmt.Errorf("%s: invalid field", operation)
	}
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			if !rv.CanSet() {
				return reflect.Value{}, fmt.Errorf("%s: field %s is not settable", operation, rv.Type())
			}
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	if !rv.CanSet() {
		return reflect.Value{}, fmt.Errorf("%s: field %s is not settable", operation, rv.Type())
	}
	return rv, nil
}

// peelOnce dereferences a single pointer level, allocating if nil, so callers
// can write through *T fields uniformly.
func peelOnce(rv reflect.Value) reflect.Value {
	for {
		if rv.Kind() != reflect.Ptr {
			break
		}
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	return rv
}

var durationType = reflect.TypeOf(time.Duration(0))
