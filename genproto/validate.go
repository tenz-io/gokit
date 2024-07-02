package genproto

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/tenz-io/gokit/genproto/go/custom/idl"
)

var (
	_ error = (*ValidationError)(nil)
	_ error = (ValidationsError)(nil)
	_ error = (*ProtoError)(nil)
)

type FieldRules map[string]*idl.Field

type ValidationError struct {
	Key     string
	Message string
}

type ProtoError struct {
	Key     string
	Message string
}

type ValidationsError []*ValidationError

func (p *ProtoError) Error() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", p.Key, p.Message)
}

func (v *ValidationError) Error() string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", v.Key, v.Message)
}

func (v ValidationsError) Error() string {
	if v == nil {
		return ""
	}
	errs := make([]string, 0, len(v))
	for _, e := range v {
		errs = append(errs, e.Error())
	}
	return strings.Join(errs, ", ")
}

func (v ValidationsError) HasError() bool {
	return len(v) > 0
}

type FieldData struct {
	Name        string
	IntField    *idl.IntField
	StringField *idl.StringField
	BytesField  *idl.BytesField
	ArrayField  *idl.ArrayField
	FloatField  *idl.FloatField
}

func ValidateIntField(fieldIdl *idl.IntField, fieldName string, v reflect.Value) error {
	if fieldIdl == nil || !v.IsValid() {
		return nil
	}

	var (
		hasDefault  = fieldIdl.Default != nil
		defaultVal  = fieldIdl.GetDefault()
		validations = ValidationsError{}
	)

	if isNilValue(v) && hasDefault {
		setIntValue(v, defaultVal)
	}

	var (
		isNil = isNilValue(v)
	)

	if v.Kind() == reflect.Ptr {
		// get the actual value if v is a pointer
		v = v.Elem()
	}

	var actualVal int64
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		actualVal = v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		actualVal = int64(v.Uint())
	case reflect.Invalid:
		// ignore, caused by nil value
	default:
		return &ProtoError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be int type"),
		}
	}

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && !hasDefault && isNil {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("is required"),
		})
	}

	if fieldIdl.Gt != nil && actualVal <= fieldIdl.GetGt() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be greater than %d", fieldIdl.GetGt()),
		})
	}

	if fieldIdl.Gte != nil && actualVal < fieldIdl.GetGte() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be greater than or equal to %d", fieldIdl.GetGte()),
		})
	}

	if fieldIdl.Lt != nil && actualVal >= fieldIdl.GetLt() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be less than %d", fieldIdl.GetLt()),
		})
	}

	if fieldIdl.Lte != nil && actualVal > fieldIdl.GetLte() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be less than or equal to %d", fieldIdl.GetLte()),
		})
	}

	if fieldIdl.Eq != nil && actualVal != fieldIdl.GetEq() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be equal to %d", fieldIdl.GetEq()),
		})
	}

	if fieldIdl.Ne != nil && actualVal == fieldIdl.GetNe() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be not equal to %d", fieldIdl.GetNe()),
		})
	}

	if len(fieldIdl.In) > 0 {
		var found bool
		for _, v := range fieldIdl.GetIn() {
			if actualVal == v {
				found = true
				break
			}
		}
		if !found {
			validations = append(validations, &ValidationError{
				Key:     fieldName,
				Message: fmt.Sprintf("should be in %v", fieldIdl.GetIn()),
			})
		}
	}

	if len(fieldIdl.NotIn) > 0 {
		var found bool
		for _, v := range fieldIdl.GetNotIn() {
			if actualVal == v {
				found = true
				break
			}
		}
		if found {
			validations = append(validations, &ValidationError{
				Key:     fieldName,
				Message: fmt.Sprintf("should be not in %v", fieldIdl.GetNotIn()),
			})
		}
	}

	if validations.HasError() {
		return validations
	}

	return nil
}

// ValidateStringField validates a string field
func ValidateStringField(fieldIdl *idl.StringField, fieldName string, v reflect.Value) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		hasDefault  = fieldIdl.Default != nil
		defaultVal  = fieldIdl.GetDefault()
		validations = ValidationsError{}
	)

	if isNilValue(v) && hasDefault {
		setStringValue(v, defaultVal)
	}

	var (
		isNil = isNilValue(v)
	)

	if v.Kind() == reflect.Ptr {
		if isNil {
			v = reflect.New(v.Type().Elem())
		} else {
			v = v.Elem()
		}
	}

	var actualVal string
	switch v.Kind() {
	case reflect.String:
		actualVal = v.String()
	default:
		return &ProtoError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be string type"),
		}
	}

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && isNil && !hasDefault {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("is required"),
		})
	}

	if fieldIdl.MinLen != nil && len(actualVal) < int(fieldIdl.GetMinLen()) {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should have minimum length of %d", fieldIdl.GetMinLen()),
		})
	}

	if fieldIdl.MaxLen != nil && len(actualVal) > int(fieldIdl.GetMaxLen()) {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should have maximum length of %d", fieldIdl.GetMaxLen()),
		})
	}

	if len(fieldIdl.In) > 0 {
		var found bool
		for _, v := range fieldIdl.GetIn() {
			if actualVal == v {
				found = true
				break
			}
		}
		if !found {
			validations = append(validations, &ValidationError{
				Key:     fieldName,
				Message: fmt.Sprintf("should be in %v", fieldIdl.GetIn()),
			})
		}
	}

	if len(fieldIdl.NotIn) > 0 {
		var found bool
		for _, v := range fieldIdl.GetNotIn() {
			if actualVal == v {
				found = true
				break
			}
		}
		if found {
			validations = append(validations, &ValidationError{
				Key:     fieldName,
				Message: fmt.Sprintf("should be not in %v", fieldIdl.GetNotIn()),
			})
		}
	}

	if fieldIdl.Pattern != nil {
		matched, err := regexp.MatchString(fieldIdl.GetPattern(), actualVal)
		if err != nil {
			return &ProtoError{
				Key:     fieldName,
				Message: fmt.Sprintf("invalid pattern %s", fieldIdl.GetPattern()),
			}
		}
		if !matched {
			validations = append(validations, &ValidationError{
				Key:     fieldName,
				Message: fmt.Sprintf("should match pattern %s", fieldIdl.GetPattern()),
			})
		}
	}

	if validations.HasError() {
		return validations
	}

	return nil
}

// ValidateBytesField validates a bytes field
func ValidateBytesField(fieldIdl *idl.BytesField, fieldName string, v reflect.Value) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		isPtr       = v.Kind() == reflect.Ptr
		isNil       = isNilValue(v)
		validations = ValidationsError{}
	)

	if isPtr {
		if isNil {
			v = reflect.New(v.Type().Elem())
		} else {
			v = v.Elem()
		}
	}

	var actualVal []byte
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() != reflect.Uint8 {
			return &ProtoError{
				Key:     fieldName,
				Message: fmt.Sprintf("should be []byte"),
			}
		}
		actualVal = v.Bytes()
	default:
		return &ProtoError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be []byte"),
		}
	}

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && (isNil || len(actualVal) == 0) {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("is required"),
		})
	}

	if fieldIdl.MinLen != nil && len(actualVal) < int(fieldIdl.GetMinLen()) {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should have minimum length of %d", fieldIdl.GetMinLen()),
		})
	}

	if fieldIdl.MaxLen != nil && len(actualVal) > int(fieldIdl.GetMaxLen()) {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should have maximum length of %d", fieldIdl.GetMaxLen()),
		})
	}

	return nil
}

// ValidateArrayField validates an array field
func ValidateArrayField(fieldIdl *idl.ArrayField, fieldName string, fieldVal reflect.Value) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		isPtr       = fieldVal.Kind() == reflect.Ptr
		isNil       = isPtr && fieldVal.IsNil()
		validations = ValidationsError{}
	)

	// fieldVal should be []T or a pointer to []T
	if isPtr {
		if isNil {
			fieldVal = reflect.New(fieldVal.Type().Elem())
		} else {
			fieldVal = fieldVal.Elem()
		}
	}

	switch fieldVal.Kind() {
	case reflect.Slice, reflect.Array:

	default:
		return &ProtoError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be slice/array"),
		}
	}

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && (isNil || fieldVal.Len() == 0) {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should not be empty"),
		})
	}

	if fieldIdl.MinItems != nil && int64(fieldVal.Len()) < fieldIdl.GetMinItems() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should have minimum items of %d", fieldIdl.GetMinItems()),
		})
	}

	if fieldIdl.MaxItems != nil && int64(fieldVal.Len()) > fieldIdl.GetMaxItems() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should have maximum items of %d", fieldIdl.GetMaxItems()),
		})
	}

	if fieldIdl.Len != nil && int64(fieldVal.Len()) != fieldIdl.GetLen() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should have %d items", fieldIdl.GetLen()),
		})
	}

	if fieldIdl.Item != nil {
		//for i := 0; i < fieldVal.Len(); i++ {
		//	itemType := fieldVal.Index(i).Type()
		//	itemVal := fieldVal.Index(i)
		//	if itemVal.Kind() == reflect.Ptr {
		//		if itemVal.IsNil() {
		//			itemVal = reflect.New(itemType.Elem())
		//		} else {
		//			itemVal = itemVal.Elem()
		//		}
		//	}
		//
		//	switch itemVal.Kind() {
		//	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		//		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		//
		//		if err := ValidateIntField(fieldIdl.GetItem().GetInt(), itemType, itemVal); err != nil {
		//			if e := new(ValidationError); errors.As(err, &e) {
		//				validations = append(validations, e)
		//			} else if e := new(ValidationsError); errors.As(err, &e) {
		//				validations = append(validations, *e...)
		//			} else {
		//				return err
		//			}
		//		}
		//	case reflect.String:
		//		if err := ValidateStringField(fieldIdl.GetItem().GetStr(), itemType, itemVal); err != nil {
		//			if e := new(ValidationError); errors.As(err, &e) {
		//				validations = append(validations, e)
		//			} else if e := new(ValidationsError); errors.As(err, &e) {
		//				validations = append(validations, *e...)
		//			} else {
		//				return err
		//			}
		//		}
		//	case reflect.Slice, reflect.Array:
		//		if itemVal.Type().Elem().Kind() == reflect.Uint8 {
		//			if err := ValidateBytesField(fieldIdl.GetItem().GetBytes(), itemType, itemVal); err != nil {
		//				if e := new(ValidationError); errors.As(err, &e) {
		//					validations = append(validations, e)
		//				} else if e := new(ValidationsError); errors.As(err, &e) {
		//					validations = append(validations, *e...)
		//				} else {
		//					return err
		//				}
		//			}
		//		} else {
		//			// ignore other types
		//		}
		//	default:
		//	}
		//
		//}
	}

	return nil
}

// ValidateFloatField validates a float field
func ValidateFloatField(fieldIdl *idl.FloatField, filedName string, fieldVal reflect.Value) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		isPtr       = fieldVal.Kind() == reflect.Ptr
		isNil       = fieldVal.IsNil()
		hasDefault  = fieldIdl.Default != nil
		defaultVal  = fieldIdl.GetDefault()
		validations = ValidationsError{}
	)

	// fieldVal should be float32, float64 or a pointer to one of these types
	if isPtr {
		if isNil {
			fieldVal = reflect.New(fieldVal.Type().Elem())
		} else {
			fieldVal = fieldVal.Elem()
		}
	}

	var actualVal float64
	switch fieldVal.Kind() {
	case reflect.Float32, reflect.Float64:
		if isNil {
			fieldVal.SetFloat(defaultVal)
		}
		actualVal = fieldVal.Float()
	default:
		return &ProtoError{
			Key:     filedName,
			Message: fmt.Sprintf("should be float type"),
		}
	}

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && isNil && !hasDefault {
		validations = append(validations, &ValidationError{
			Key:     filedName,
			Message: fmt.Sprintf("should not be empty"),
		})
	}

	if fieldIdl.Gt != nil && actualVal <= fieldIdl.GetGt() {
		validations = append(validations, &ValidationError{
			Key:     filedName,
			Message: fmt.Sprintf("should be greater than %f", fieldIdl.GetGt()),
		})
	}

	if fieldIdl.Gte != nil && actualVal < fieldIdl.GetGte() {
		validations = append(validations, &ValidationError{
			Key:     filedName,
			Message: fmt.Sprintf("should be greater than or equal to %f", fieldIdl.GetGte()),
		})
	}

	if fieldIdl.Lt != nil && actualVal >= fieldIdl.GetLt() {
		validations = append(validations, &ValidationError{
			Key:     filedName,
			Message: fmt.Sprintf("should be less than %f", fieldIdl.GetLt()),
		})
	}

	if fieldIdl.Lte != nil && actualVal > fieldIdl.GetLte() {
		validations = append(validations, &ValidationError{
			Key:     filedName,
			Message: fmt.Sprintf("should be less than or equal to %f", fieldIdl.GetLte()),
		})
	}

	return nil
}

func ValidateField(fieldIdl *idl.Field, fieldName string, fieldVal reflect.Value) error {
	if fieldIdl == nil {
		return nil
	}

	if fieldVal.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			fieldVal = reflect.New(fieldVal.Type().Elem())
		}
	}

	if fieldIdl.GetInt() != nil {
		return ValidateIntField(fieldIdl.GetInt(), fieldName, fieldVal)
	}

	if fieldIdl.GetStr() != nil {
		return ValidateStringField(fieldIdl.GetStr(), fieldName, fieldVal)
	}

	if fieldIdl.GetBytes() != nil {
		return ValidateBytesField(fieldIdl.GetBytes(), fieldName, fieldVal)
	}

	if fieldIdl.GetArray() != nil {
		return ValidateArrayField(fieldIdl.GetArray(), fieldName, fieldVal)
	}

	if fieldIdl.GetFloat() != nil {
		return ValidateFloatField(fieldIdl.GetFloat(), fieldName, fieldVal)
	}

	return nil
}

func Validate(rules FieldRules, val any) error {
	if len(rules) == 0 {
		return nil
	}

	if val == nil {
		return nil
	}

	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v = reflect.New(v.Type().Elem())
		} else {
			v = v.Elem()
		}
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		filedName := t.Field(i).Name

		fieldIdl, ok := rules[filedName]
		if !ok {
			continue
		}

		if err := ValidateField(fieldIdl, filedName, fieldVal); err != nil {
			return err
		}
	}

	return nil
}

// isNilValue checks if a reflect.Value is a nil value
func isNilValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Ptr,
		reflect.Slice,
		reflect.Array:
		return v.IsNil()
	default:
		return false
	}
}

// setIntValue sets the value of a reflect.Value to an int64 value
// v should be a reflect.Value of int/uint/*int/*uint type
func setIntValue(v reflect.Value, i int64) {
	if !v.CanSet() {
		return
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(uint64(i))
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		setIntValue(v.Elem(), i)
	default:
		// ignore
	}
}

// setStringValue sets the value of a reflect.Value to a string value
// v should be a reflect.Value of string/*string type
func setStringValue(v reflect.Value, s string) {
	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		setStringValue(v.Elem(), s)
	default:
		// ignore
	}
}

// setFloatValue sets the value of a reflect.Value to a float64 value
// v should be a reflect.Value of float32/float64 or a pointer to one of these types
func setFloatValue(v reflect.Value, f float64) {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		v.SetFloat(f)
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		setFloatValue(v.Elem(), f)
	default:
		// ignore
	}
}

func setDefaultValue(msg any, fieldName string, defaultValue any) error {
	v := reflect.ValueOf(msg)

	// Check if the provided interface is a pointer to a struct
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct, got %T", msg)
	}

	v = v.Elem() // Dereference to get the struct value

	// Get the field by name
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("no such field: %s in struct", fieldName)
	}

	// Check if the field is a pointer and is nil
	if field.Kind() == reflect.Ptr && field.IsNil() {
		// Check if the defaultValue type matches the field's element type
		defaultVal := reflect.ValueOf(defaultValue)
		if defaultVal.Type().ConvertibleTo(field.Type().Elem()) {
			newVal := reflect.New(field.Type().Elem())
			newVal.Elem().Set(defaultVal.Convert(field.Type().Elem()))
			field.Set(newVal)
		} else {
			return fmt.Errorf("type mismatch: cannot set field %s with type %s to default value of type %s", fieldName, field.Type().Elem(), defaultVal.Type())
		}
	}

	return nil
}
