package annotation

import (
	"fmt"
	"reflect"
	"strconv"
)

// ParseDefault parses the default tag value of a struct and sets the field value.
func ParseDefault(structPtr any) error {
	v := reflect.ValueOf(structPtr)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		if err := updateStructField(v, field); err != nil {
			return err
		}
	}
	return nil
}

// updateStructField updates the field value based on the default tag value.
func updateStructField(v reflect.Value, field reflect.StructField) error {
	defaultVal := field.Tag.Get(string(Default))
	fieldVal := v.FieldByName(field.Name)

	if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
		// Initialize pointer field.
		fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
	}

	if fieldVal.Kind() == reflect.Struct || (fieldVal.Kind() == reflect.Ptr && fieldVal.Elem().Kind() == reflect.Struct) {
		// Recursively process nested struct fields.
		if fieldVal.Kind() == reflect.Ptr {
			return ParseDefault(fieldVal.Interface())
		} else {
			return ParseDefault(fieldVal.Addr().Interface())
		}
	}

	if defaultVal == "" {
		return nil
	}

	if err := setStringValue(fieldVal, defaultVal); err != nil {
		return err
	}
	return nil
}

// setStringValue sets the value of a field based on the field type.
func setStringValue(fieldVal reflect.Value, val string) error {
	if !fieldVal.CanSet() {
		return fmt.Errorf("cannot set value for field: %v", fieldVal.Type())
	}

	if fieldVal.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
		}
		fieldVal = fieldVal.Elem()
		if !fieldVal.CanSet() {
			return fmt.Errorf("cannot set value for field: %v", fieldVal.Type())
		}
	}

	switch fieldVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		fieldVal.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		fieldVal.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		fieldVal.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		fieldVal.SetBool(v)
	case reflect.Slice:
		if fieldVal.Type().Elem().Kind() == reflect.Uint8 {
			fieldVal.SetBytes([]byte(val))
		}
	case reflect.String:
		fieldVal.SetString(val)
	default:
		// ignore unsupported types
	}
	return nil
}
