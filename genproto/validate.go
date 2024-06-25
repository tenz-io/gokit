package genproto

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"google.golang.org/protobuf/proto"

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

type ValidationsError []ValidationError

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
}

// ValidateIntField validates an int field
func ValidateIntField(fieldIdl *idl.IntField, fieldVal reflect.Value) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		fieldType   = fieldVal.Type()
		isPtr       = fieldVal.Kind() == reflect.Ptr
		isNil       = fieldVal.IsNil()
		defaultVal  = fieldIdl.GetDefault()
		validations = ValidationsError{}
	)

	// fieldVal should be int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64
	// or a pointer to one of these types
	if isPtr {
		if isNil {
			fieldVal = reflect.New(fieldType.Elem())
		} else {
			fieldVal = fieldVal.Elem()
		}
	}

	var actualVal int64
	switch fieldVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if isNil {
			fieldVal.SetInt(defaultVal)
		}
		actualVal = fieldVal.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if isNil {
			fieldVal.SetUint(uint64(defaultVal))
		}
		actualVal = int64(fieldVal.Uint())
	default:
		return &ProtoError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should be int/uint, actual %s", fieldType.Name(), fieldVal.Kind().String()),
		}
	}

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && isNil {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s is required", fieldType.Name()),
		})
	}

	if fieldIdl.Gt != nil && actualVal <= fieldIdl.GetGt() {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should be greater than %d, actual %d", fieldType.Name(), fieldIdl.GetGt(), actualVal),
		})
	}

	if fieldIdl.Gte != nil && actualVal < fieldIdl.GetGte() {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should be greater than or equal to %d, actual %d", fieldType.Name(), fieldIdl.GetGte(), actualVal),
		})
	}

	if fieldIdl.Lt != nil && actualVal >= fieldIdl.GetLt() {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should be less than %d, actual %d", fieldType.Name(), fieldIdl.GetLt(), actualVal),
		})
	}

	if fieldIdl.Lte != nil && actualVal > fieldIdl.GetLte() {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should be less than or equal to %d, actual %d", fieldType.Name(), fieldIdl.GetLte(), actualVal),
		})
	}

	if fieldIdl.Eq != nil && actualVal != fieldIdl.GetEq() {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should be equal to %d, actual %d", fieldType.Name(), fieldIdl.GetEq(), actualVal),
		})
	}

	if fieldIdl.Ne != nil && actualVal == fieldIdl.GetNe() {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should not be equal to %d, actual %d", fieldType.Name(), fieldIdl.GetNe(), actualVal),
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
			validations = append(validations, ValidationError{
				Key:     fieldType.Name(),
				Message: fmt.Sprintf("field %s should be in %v, actual %d", fieldType.Name(), fieldIdl.GetIn(), actualVal),
			})
		}
	}

	if validations.HasError() {
		return validations
	}

	return nil
}

// ValidateStringField validates a string field
func ValidateStringField(fieldIdl *idl.StringField, fieldVal reflect.Value) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		fieldType   = fieldVal.Type()
		isPtr       = fieldVal.Kind() == reflect.Ptr
		isNil       = isPtr && fieldVal.IsNil()
		defaultVal  = fieldIdl.GetDefault()
		validations = ValidationsError{}
	)

	// fieldVal should be string or a pointer to string
	if isPtr {
		if isNil {
			fieldVal = reflect.New(fieldType.Elem())
		} else {
			fieldVal = fieldVal.Elem()
		}
	}

	var actualVal string
	switch fieldVal.Kind() {
	case reflect.String:
		if isNil {
			fieldVal.SetString(defaultVal)
		}
		actualVal = fieldVal.String()
	default:
		return &ProtoError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should be string, actual %s", fieldType.Name(), fieldVal.Kind().String()),
		}
	}

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && isNil {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s is required", fieldType.Name()),
		})
	}

	if fieldIdl.MinLen != nil && len(actualVal) < int(fieldIdl.GetMinLen()) {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should have minimum length of %d, actual %d", fieldType.Name(), fieldIdl.GetMinLen(), len(actualVal)),
		})
	}

	if fieldIdl.MaxLen != nil && len(actualVal) > int(fieldIdl.GetMaxLen()) {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should have maximum length of %d, actual %d", fieldType.Name(), fieldIdl.GetMaxLen(), len(actualVal)),
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
			validations = append(validations, ValidationError{
				Key:     fieldType.Name(),
				Message: fmt.Sprintf("field %s should be in %v, actual %s", fieldType.Name(), fieldIdl.GetIn(), actualVal),
			})
		}
	}

	if fieldIdl.Pattern != nil {
		matched, err := regexp.MatchString(fieldIdl.GetPattern(), actualVal)
		if err != nil {
			return &ProtoError{
				Key:     fieldType.Name(),
				Message: fmt.Sprintf("field %s pattern is invalid: %s", fieldType.Name(), err.Error()),
			}
		}
		if !matched {
			validations = append(validations, ValidationError{
				Key:     fieldType.Name(),
				Message: fmt.Sprintf("field %s should match pattern %s, actual %s", fieldType.Name(), fieldIdl.GetPattern(), actualVal),
			})
		}
	}

	if validations.HasError() {
		return validations
	}

	return nil
}

// ValidateBytesField validates a bytes field
func ValidateBytesField(fieldIdl *idl.BytesField, fieldVal reflect.Value) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		fieldType   = fieldVal.Type()
		isPtr       = fieldVal.Kind() == reflect.Ptr
		isNil       = isPtr && fieldVal.IsNil()
		validations = ValidationsError{}
	)

	// fieldVal should be []byte or a pointer to []byte
	if isPtr {
		if isNil {
			fieldVal = reflect.New(fieldType.Elem())
		} else {
			fieldVal = fieldVal.Elem()
		}
	}

	var actualVal []byte
	switch fieldVal.Kind() {
	case reflect.Slice, reflect.Array:
		if fieldVal.Type().Elem().Kind() != reflect.Uint8 {
			return &ProtoError{
				Key:     fieldType.Name(),
				Message: fmt.Sprintf("field %s should be []byte, actual %s", fieldType.Name(), fieldVal.Kind().String()),
			}
		}
		actualVal = fieldVal.Bytes()
	default:
		return &ProtoError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should be []byte, actual %s", fieldType.Name(), fieldVal.Kind().String()),
		}
	}

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && (isNil || len(actualVal) == 0) {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s is required", fieldType.Name()),
		})
	}

	if fieldIdl.MinLen != nil && len(actualVal) < int(fieldIdl.GetMinLen()) {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should have minimum length of %d, actual %d", fieldType.Name(), fieldIdl.GetMinLen(), len(actualVal)),
		})
	}

	if fieldIdl.MaxLen != nil && len(actualVal) > int(fieldIdl.GetMaxLen()) {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should have maximum length of %d, actual %d", fieldType.Name(), fieldIdl.GetMaxLen(), len(actualVal)),
		})
	}

	return nil
}

// ValidateArrayField validates an array field
func ValidateArrayField(fieldIdl *idl.ArrayField, fieldVal reflect.Value) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		fieldType   = fieldVal.Type()
		isPtr       = fieldVal.Kind() == reflect.Ptr
		isNil       = isPtr && fieldVal.IsNil()
		validations = ValidationsError{}
	)

	// fieldVal should be []T or a pointer to []T
	if isPtr {
		if isNil {
			fieldVal = reflect.New(fieldType.Elem())
		} else {
			fieldVal = fieldVal.Elem()
		}
	}

	switch fieldVal.Kind() {
	case reflect.Slice, reflect.Array:

	default:
		return &ProtoError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should be array, actual %s", fieldType.Name(), fieldVal.Kind().String()),
		}
	}

	if fieldIdl.Required != nil && fieldIdl.GetRequired() && (isNil || fieldVal.Len() == 0) {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s is required", fieldType.Name()),
		})
	}

	if fieldIdl.MinItems != nil && int64(fieldVal.Len()) < fieldIdl.GetMinItems() {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should have minimum items of %d, actual %d", fieldType.Name(), fieldIdl.GetMinItems(), fieldVal.Len()),
		})
	}

	if fieldIdl.MaxItems != nil && int64(fieldVal.Len()) > fieldIdl.GetMaxItems() {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should have maximum items of %d, actual %d", fieldType.Name(), fieldIdl.GetMaxItems(), fieldVal.Len()),
		})
	}

	if fieldIdl.Len != nil && int64(fieldVal.Len()) != fieldIdl.GetLen() {
		validations = append(validations, ValidationError{
			Key:     fieldType.Name(),
			Message: fmt.Sprintf("field %s should have length of %d, actual %d", fieldType.Name(), fieldIdl.GetLen(), fieldVal.Len()),
		})
	}

	if fieldIdl.Item != nil {
		for i := 0; i < fieldVal.Len(); i++ {
			itemType := fieldVal.Index(i).Type()
			itemVal := fieldVal.Index(i)
			if itemVal.Kind() == reflect.Ptr {
				if itemVal.IsNil() {
					itemVal = reflect.New(itemType.Elem())
				} else {
					itemVal = itemVal.Elem()
				}
			}

			switch itemVal.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if err := ValidateIntField(fieldIdl.GetItem().GetInt(), itemVal); err != nil {
					validations = append(validations, ValidationError{
						Key:     fieldType.Name(),
						Message: fmt.Sprintf("field %s item %d is invalid: %s", fieldType.Name(), i, err.Error()),
					})
				}
			case reflect.String:
				if err := ValidateStringField(fieldIdl.GetItem().GetStr(), itemVal); err != nil {
					validations = append(validations, ValidationError{
						Key:     fieldType.Name(),
						Message: fmt.Sprintf("field %s item %d is invalid: %s", fieldType.Name(), i, err.Error()),
					})
				}
			case reflect.Slice, reflect.Array:
				if itemVal.Type().Elem().Kind() == reflect.Uint8 {
					if err := ValidateBytesField(fieldIdl.GetItem().GetBytes(), itemVal); err != nil {
						validations = append(validations, ValidationError{
							Key:     fieldType.Name(),
							Message: fmt.Sprintf("field %s item %d is invalid: %s", fieldType.Name(), i, err.Error()),
						})
					}
				} else {
					// ignore other types
				}
			default:
			}

		}
	}

	return nil
}

func ValidateField(fieldIdl *idl.Field, fieldVal reflect.Value) error {
	if fieldIdl == nil {
		return nil
	}

	var (
		fieldType = fieldVal.Type()
	)

	if fieldVal.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			fieldVal = reflect.New(fieldType)
		}
	}

	if fieldIdl.GetInt() != nil {
		return ValidateIntField(fieldIdl.GetInt(), fieldVal)
	}

	if fieldIdl.GetStr() != nil {
		return ValidateStringField(fieldIdl.GetStr(), fieldVal)
	}

	if fieldIdl.GetBytes() != nil {
		return ValidateBytesField(fieldIdl.GetBytes(), fieldVal)
	}

	if fieldIdl.GetArray() != nil {
		return ValidateArrayField(fieldIdl.GetArray(), fieldVal)
	}

	return nil
}

func ValidateMessage(rules FieldRules, message proto.Message) error {
	if len(rules) == 0 {
		return nil
	}

	messageVal := reflect.ValueOf(message)
	messageType := messageVal.Type()
	if messageVal.Kind() == reflect.Ptr {
		if messageVal.IsNil() {
			messageVal = reflect.New(messageType.Elem())
		}
	}

	for i := 0; i < messageVal.NumField(); i++ {
		fieldVal := messageVal.Field(i)
		fieldType := messageType.Field(i)
		fieldIdl, ok := rules[fieldType.Name]
		if !ok {
			continue
		}

		if err := ValidateField(fieldIdl, fieldVal); err != nil {
			return err
		}
	}

	return nil
}
