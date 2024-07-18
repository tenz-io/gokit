package genproto

import (
	"errors"
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

func mergeValidationsErrors(errs ...ValidationsError) ValidationsError {
	var merged ValidationsError
	for _, err := range errs {
		merged = append(merged, err...)
	}
	return merged
}

type FieldData struct {
	Name        string
	IntField    *idl.IntField
	StringField *idl.StringField
	BytesField  *idl.BytesField
	ArrayField  *idl.ArrayField
	FloatField  *idl.FloatField
}

func ValidateIntField(fieldIdl *idl.IntField, fieldName string, msg any) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		nilField    = isNilField(msg, fieldName)
		validations = ValidationsError{}
	)

	var msgVal = reflect.ValueOf(msg)
	if msgVal.Kind() == reflect.Ptr {
		msgVal = msgVal.Elem()
	}

	var (
		field     = msgVal.FieldByName(fieldName)
		actualVal = getIntFieldVal(field)
	)

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && nilField {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("is required"),
		})
	}

	if err := validateIntField(fieldIdl, fieldName, actualVal); err != nil {
		var vErr = ValidationsError{}
		if errors.As(err, &vErr) && vErr.HasError() {
			validations = mergeValidationsErrors(validations, vErr)
		} else {
			return err
		}
	}

	if validations.HasError() {
		return validations
	}

	return nil
}

func validateIntField(fieldIdl *idl.IntField, fieldName string, val int64) error {
	var validations = ValidationsError{}
	if fieldIdl.Gt != nil && val <= fieldIdl.GetGt() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be greater than %d", fieldIdl.GetGt()),
		})
	}

	if fieldIdl.Gte != nil && val < fieldIdl.GetGte() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be greater than or equal to %d", fieldIdl.GetGte()),
		})
	}

	if fieldIdl.Lt != nil && val >= fieldIdl.GetLt() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be less than %d", fieldIdl.GetLt()),
		})
	}

	if fieldIdl.Lte != nil && val > fieldIdl.GetLte() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be less than or equal to %d", fieldIdl.GetLte()),
		})
	}

	if fieldIdl.Eq != nil && val != fieldIdl.GetEq() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be equal to %d", fieldIdl.GetEq()),
		})
	}

	if fieldIdl.Ne != nil && val == fieldIdl.GetNe() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be not equal to %d", fieldIdl.GetNe()),
		})
	}

	if len(fieldIdl.In) > 0 {
		var found bool
		for _, v := range fieldIdl.GetIn() {
			if val == v {
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
			if val == v {
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
func ValidateStringField(fieldIdl *idl.StringField, fieldName string, msg any) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		nilField    = isNilField(msg, fieldName)
		validations = ValidationsError{}
	)

	var msgVal = reflect.ValueOf(msg)
	if msgVal.Kind() == reflect.Ptr {
		msgVal = msgVal.Elem()
	}

	var (
		field     = msgVal.FieldByName(fieldName)
		actualVal = getStringFieldVal(field)
	)

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && nilField {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("is required"),
		})
	}

	if err := validateStringField(fieldIdl, fieldName, actualVal); err != nil {
		var vErr = ValidationsError{}
		if errors.As(err, &vErr) && vErr.HasError() {
			validations = mergeValidationsErrors(validations, vErr)
		} else {
			return err
		}
	}

	if validations.HasError() {
		return validations
	}

	return nil
}

func validateStringField(fieldIdl *idl.StringField, fieldName string, actualVal string) error {
	var validations = ValidationsError{}
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
func ValidateBytesField(fieldIdl *idl.BytesField, fieldName string, msg any) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		nilField    = isNilField(msg, fieldName)
		validations = ValidationsError{}
	)

	var msgVal = reflect.ValueOf(msg)
	if msgVal.Kind() == reflect.Ptr {
		msgVal = msgVal.Elem()
	}

	var (
		field     = msgVal.FieldByName(fieldName)
		actualVal = getBytesFieldVal(field)
	)

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && (nilField || len(actualVal) == 0) {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("is required"),
		})
	}

	if err := validateBytesField(fieldIdl, fieldName, actualVal); err != nil {
		var vErr = ValidationsError{}
		if errors.As(err, &vErr) && vErr.HasError() {
			validations = mergeValidationsErrors(validations, vErr)
		} else {
			return err
		}
	}

	if validations.HasError() {
		return validations
	}

	return nil
}

func validateBytesField(fieldIdl *idl.BytesField, fieldName string, val []byte) error {
	var validations = ValidationsError{}
	if fieldIdl.MinLen != nil && len(val) < int(fieldIdl.GetMinLen()) {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should have minimum length of %d", fieldIdl.GetMinLen()),
		})
	}

	if fieldIdl.MaxLen != nil && len(val) > int(fieldIdl.GetMaxLen()) {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should have maximum length of %d", fieldIdl.GetMaxLen()),
		})
	}

	if validations.HasError() {
		return validations
	}
	return nil
}

// ValidateArrayField validates an array field
func ValidateArrayField(fieldIdl *idl.ArrayField, fieldName string, msg any) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		nilField    = isNilField(msg, fieldName)
		validations = ValidationsError{}
	)

	var msgVal = reflect.ValueOf(msg)
	if msgVal.Kind() == reflect.Ptr {
		msgVal = msgVal.Elem()
	}

	var (
		field = msgVal.FieldByName(fieldName)
	)

	switch field.Kind() {
	case reflect.Slice, reflect.Array:
		// ignore
	default:
		return &ProtoError{
			Key:     fieldName,
			Message: fmt.Sprintf("should be slice/array"),
		}
	}

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && (nilField || field.Len() == 0) {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should not be empty"),
		})
	}

	if fieldIdl.MinItems != nil && int64(field.Len()) < fieldIdl.GetMinItems() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should have minimum items of %d", fieldIdl.GetMinItems()),
		})
	}

	if fieldIdl.MaxItems != nil && int64(field.Len()) > fieldIdl.GetMaxItems() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should have maximum items of %d", fieldIdl.GetMaxItems()),
		})
	}

	if fieldIdl.Len != nil && int64(field.Len()) != fieldIdl.GetLen() {
		validations = append(validations, &ValidationError{
			Key:     fieldName,
			Message: fmt.Sprintf("should have %d items", fieldIdl.GetLen()),
		})
	}

	if fieldIdl.GetItem() != nil {
		// ignore for now
		for i := 0; i < field.Len(); i++ {
			itemField := field.Index(i)
			switch itemField.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if fieldIdl.Item.GetInt() != nil {
					if err := validateIntField(fieldIdl.Item.GetInt(), fieldName, field.Int()); err != nil {
						var vErr = ValidationsError{}
						if errors.As(err, &vErr) && vErr.HasError() {
							validations = mergeValidationsErrors(validations, vErr)
						} else {
							return err
						}
					}
				}
			case reflect.String:
				if fieldIdl.Item.GetStr() != nil {
					if err := validateStringField(fieldIdl.Item.GetStr(), fieldName, field.String()); err != nil {
						var vErr = ValidationsError{}
						if errors.As(err, &vErr) && vErr.HasError() {
							validations = mergeValidationsErrors(validations, vErr)
						} else {
							return err
						}
					}
				}
			case reflect.Slice, reflect.Array:
				if itemField.Type().Elem().Kind() != reflect.Uint8 {
					continue
				}

				if fieldIdl.Item.GetBytes() != nil {
					if err := validateBytesField(fieldIdl.Item.GetBytes(), fieldName, field.Bytes()); err != nil {
						var vErr = ValidationsError{}
						if errors.As(err, &vErr) && vErr.HasError() {
							validations = mergeValidationsErrors(validations, vErr)
						} else {
							return err
						}
					}
				}
			case reflect.Float32, reflect.Float64:
				if fieldIdl.Item.GetFloat() != nil {
					if err := validateFloatField(fieldIdl.Item.GetFloat(), fieldName, field.Float()); err != nil {
						var vErr = ValidationsError{}
						if errors.As(err, &vErr) && vErr.HasError() {
							validations = mergeValidationsErrors(validations, vErr)
						} else {
							return err
						}
					}
				}
			default:
				// ignore for now
			}
		}
	}

	if validations.HasError() {
		return validations
	}

	return nil
}

// ValidateFloatField validates a float field
func ValidateFloatField(fieldIdl *idl.FloatField, filedName string, msg any) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		nilField    = isNilField(msg, filedName)
		validations = ValidationsError{}
	)

	var msgVal = reflect.ValueOf(msg)
	if msgVal.Kind() == reflect.Ptr {
		msgVal = msgVal.Elem()
	}

	var (
		field     = msgVal.FieldByName(filedName)
		actualVal = getFloatFieldVal(field)
	)

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && nilField {
		validations = append(validations, &ValidationError{
			Key:     filedName,
			Message: fmt.Sprintf("should not be empty"),
		})
	}

	if err := validateFloatField(fieldIdl, filedName, actualVal); err != nil {
		var vErr = ValidationsError{}
		if errors.As(err, &vErr) && vErr.HasError() {
			validations = mergeValidationsErrors(validations, vErr)
		} else {
			return err
		}
	}

	if validations.HasError() {
		return validations
	}

	return nil
}

func validateFloatField(fieldIdl *idl.FloatField, filedName string, val float64) error {
	var validations = ValidationsError{}
	if fieldIdl.Gt != nil && val <= fieldIdl.GetGt() {
		validations = append(validations, &ValidationError{
			Key:     filedName,
			Message: fmt.Sprintf("should be greater than %f", fieldIdl.GetGt()),
		})
	}

	if fieldIdl.Gte != nil && val < fieldIdl.GetGte() {
		validations = append(validations, &ValidationError{
			Key:     filedName,
			Message: fmt.Sprintf("should be greater than or equal to %f", fieldIdl.GetGte()),
		})
	}

	if fieldIdl.Lt != nil && val >= fieldIdl.GetLt() {
		validations = append(validations, &ValidationError{
			Key:     filedName,
			Message: fmt.Sprintf("should be less than %f", fieldIdl.GetLt()),
		})
	}

	if fieldIdl.Lte != nil && val > fieldIdl.GetLte() {
		validations = append(validations, &ValidationError{
			Key:     filedName,
			Message: fmt.Sprintf("should be less than or equal to %f", fieldIdl.GetLte()),
		})
	}

	if validations.HasError() {
		return validations
	}

	return nil
}

func ValidateField(fieldIdl *idl.Field, fieldName string, msg any) error {
	if fieldIdl == nil {
		return nil
	}

	if fieldIdl.GetInt() != nil {
		return ValidateIntField(fieldIdl.GetInt(), fieldName, msg)
	}

	if fieldIdl.GetStr() != nil {
		return ValidateStringField(fieldIdl.GetStr(), fieldName, msg)
	}

	if fieldIdl.GetBytes() != nil {
		return ValidateBytesField(fieldIdl.GetBytes(), fieldName, msg)
	}

	if fieldIdl.GetArray() != nil {
		return ValidateArrayField(fieldIdl.GetArray(), fieldName, msg)
	}

	if fieldIdl.GetFloat() != nil {
		return ValidateFloatField(fieldIdl.GetFloat(), fieldName, msg)
	}

	return nil
}

func Validate(rules FieldRules, msg any) error {
	if len(rules) == 0 {
		return nil
	}

	if msg == nil {
		// should not happen
		return &ProtoError{
			Key:     "",
			Message: "message is nil",
		}
	}

	v := reflect.ValueOf(msg)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return &ProtoError{
				Key:     "",
				Message: "message is nil pointer",
			}
		} else {
			v = v.Elem()
		}
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		filedName := t.Field(i).Name

		fieldIdl, ok := rules[filedName]
		if !ok {
			continue
		}

		if err := ValidateField(fieldIdl, filedName, msg); err != nil {
			return err
		}
	}

	return nil
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

// isNilField checks if a field is nil
func isNilField(msg any, fieldName string) bool {
	if msg == nil {
		return true
	}
	v := reflect.ValueOf(msg)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}

	// Ensure v is a struct
	if v.Kind() != reflect.Struct {
		return true
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return true
	}

	// Check if the field is nil for pointer types
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}

	return false
}

// getIntFieldVal gets the int value of a field
func getIntFieldVal(val reflect.Value) int64 {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(val.Uint())
	case reflect.Invalid:
		return 0
	case reflect.Ptr:
		if val.IsNil() {
			return 0
		}
		return getIntFieldVal(val.Elem())
	default:
		return 0
	}
}

// getStringFieldVal gets the string value of a field
func getStringFieldVal(val reflect.Value) string {
	switch val.Kind() {
	case reflect.String:
		return val.String()
	case reflect.Ptr:
		if val.IsNil() {
			return ""
		}
		return getStringFieldVal(val.Elem())
	default:
		return ""
	}
}

// getBytesFieldVal gets the bytes value of a field
func getBytesFieldVal(val reflect.Value) []byte {
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		if val.Type().Elem().Kind() == reflect.Uint8 {
			return val.Bytes()
		}
	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}
		return getBytesFieldVal(val.Elem())
	default:
		return nil
	}
	return nil
}

// getFloatFieldVal gets the float value of a field
func getFloatFieldVal(val reflect.Value) float64 {
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		return val.Float()
	case reflect.Ptr:
		if val.IsNil() {
			return 0
		}
		return getFloatFieldVal(val.Elem())
	default:
		return 0
	}
}

// IsNilOrEmpty checks if a value is nil or empty
func IsNilOrEmpty(val any) bool {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Ptr:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	case reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	default:
		return false
	}
}

// SetValue sets a new value to a pointer
// if ptrVal is not pointer, it returns false
// If the pointer is nil, it creates a new instance of the type the pointer should point to
// If the new value type does not match the pointer type, it returns false
func SetValue(ptrVal any, newVal any) (newPtrVal any, ok bool) {
	v := reflect.ValueOf(ptrVal)
	if v.Kind() != reflect.Ptr {
		return nil, false
	}

	if v.IsNil() {
		// Handle nil pointers by creating a new instance
		var newInstance reflect.Value
		switch reflect.TypeOf(ptrVal).Elem().Kind() {
		case reflect.Int32:
			newInstance = reflect.New(reflect.TypeOf(int32(0))).Elem()
			newInstance.Set(reflect.ValueOf(int32(newVal.(int64))))
		case reflect.Int64:
			newInstance = reflect.New(reflect.TypeOf(int64(0))).Elem()
			newInstance.Set(reflect.ValueOf(newVal.(int64)))
		case reflect.Uint32:
			newInstance = reflect.New(reflect.TypeOf(uint32(0))).Elem()
			newInstance.Set(reflect.ValueOf(uint32(newVal.(int64))))
		case reflect.Uint64:
			newInstance = reflect.New(reflect.TypeOf(uint64(0))).Elem()
			newInstance.Set(reflect.ValueOf(uint64(newVal.(int64))))
		case reflect.Float32:
			newInstance = reflect.New(reflect.TypeOf(float32(0))).Elem()
			newInstance.Set(reflect.ValueOf(float32(newVal.(float64))))
		case reflect.Float64:
			newInstance = reflect.New(reflect.TypeOf(float64(0))).Elem()
			newInstance.Set(reflect.ValueOf(newVal))
		case reflect.String:
			newInstance = reflect.New(reflect.TypeOf("")).Elem()
			newInstance.Set(reflect.ValueOf(newVal))
		default:
			newInstance = reflect.New(reflect.TypeOf(newVal)).Elem()
			newInstance.Set(reflect.ValueOf(newVal))
		}
		return newInstance.Addr().Interface(), true
	}

	elem := v.Elem()
	newValReflect := reflect.ValueOf(newVal)
	if !newValReflect.Type().AssignableTo(elem.Type()) {
		switch elem.Kind() {
		case reflect.Int32, reflect.Int64:
			elem.SetInt(newValReflect.Int())
			return ptrVal, true
		case reflect.Uint32, reflect.Uint64:
			elem.SetUint(uint64(newValReflect.Int()))
			return ptrVal, true
		case reflect.Float32, reflect.Float64:
			elem.SetFloat(newValReflect.Float())
			return ptrVal, true
		case reflect.String:
			elem.SetString(newValReflect.String())
			return ptrVal, true
		default:
			return nil, false
		}
	}

	elem.Set(newValReflect)
	return ptrVal, true
}

// StringIn checks if a string is in a list of strings
func StringIn(s string, list []string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}

// StringMatches checks if a string matches a pattern
func StringMatches(s string, pattern string) bool {
	p, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	return p.MatchString(s)
}

type intTyp interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

// IntIn checks if an int is in a list of ints
func IntIn[T intTyp](i T, list []int) bool {
	for _, v := range list {
		if int(i) == v {
			return true
		}
	}
	return false
}