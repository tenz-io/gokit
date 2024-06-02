package annotation

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type RequestField struct {
	FieldName string
	TagName   string
	Field     reflect.StructField
	FieldVal  reflect.Value
	IsUri     bool
	IsHeader  bool
	IsQuery   bool
	IsFile    bool
	IsForm    bool
}

// RequestFields is a map of request fields.
// key is tag name, value is RequestField.
// if tag name is empty, key is field name.
type RequestFields map[string]RequestField

// Values returns the values of the request fields.
func (r RequestFields) Values() []RequestField {
	var values []RequestField
	for _, v := range r {
		values = append(values, v)
	}
	return values
}

// Validate returns true if the request fields contain the name.
func (r RequestField) Validate() error {
	if r.IsFile {
		// file field should have bytes type field
		if r.Field.Type.Kind() != reflect.Slice || r.Field.Type.Elem().Kind() != reflect.Uint8 {
			return fmt.Errorf("field %s should be []byte", r.FieldName)
		}
	}

	return nil
}

// Set sets the value of the request field.
func (r RequestField) Set(value any) error {
	if err := r.Validate(); err != nil {
		return err
	}
	var (
		val = reflect.ValueOf(value)
	)

	if val.Kind() == reflect.Ptr && val.IsNil() {
		return nil
	}

	switch {
	case r.FieldVal.Kind() != reflect.Ptr && val.Kind() == reflect.Ptr:
		// if types match, set the value
		if r.FieldVal.Type() == val.Type().Elem() {
			r.FieldVal.Set(val.Elem())
		} else {
			// try to convert the value
			if val.Type().Elem().ConvertibleTo(r.FieldVal.Type()) {
				r.FieldVal.Set(val.Elem().Convert(r.FieldVal.Type()))
			} else {
				return fmt.Errorf("cannot convert %s to %s", val.Type().Elem(), r.FieldVal.Type())
			}
		}
	case r.FieldVal.Kind() == reflect.Ptr && val.Kind() != reflect.Ptr:
		// if types match, set the value
		if r.FieldVal.Type().Elem() == val.Type() {
			// make a new pointer
			newVal := reflect.New(val.Type())
			newVal.Elem().Set(val)
			r.FieldVal.Set(newVal)
		} else {
			// try to convert the value
			if val.Type().ConvertibleTo(r.FieldVal.Type().Elem()) {
				// make a new pointer match the type
				newVal := reflect.New(r.FieldVal.Type().Elem())
				newVal.Elem().Set(val.Convert(r.FieldVal.Type().Elem()))
				r.FieldVal.Set(newVal)
			} else {
				return fmt.Errorf("cannot convert %s to %s", val.Type(), r.FieldVal.Type().Elem())
			}
		}
	case r.FieldVal.Kind() == reflect.Ptr && val.Kind() == reflect.Ptr:
		// if types match, set the value
		if r.FieldVal.Type().Elem() == val.Type().Elem() {
			r.FieldVal.Set(val)
		} else {
			// try to convert the value
			if val.Type().Elem().ConvertibleTo(r.FieldVal.Type().Elem()) {
				r.FieldVal.Set(val.Elem().Convert(r.FieldVal.Type().Elem()))
			} else {
				return fmt.Errorf("cannot convert %s to %s", val.Type().Elem(), r.FieldVal.Type().Elem())
			}
		}
	case r.FieldVal.Kind() != reflect.Ptr && val.Kind() != reflect.Ptr:
		// if types match, set the value
		if r.FieldVal.Type() == val.Type() {
			r.FieldVal.Set(val)
		} else {
			// try to convert the value
			if val.Type().ConvertibleTo(r.FieldVal.Type()) {
				r.FieldVal.Set(val.Convert(r.FieldVal.Type()))
			} else {
				return fmt.Errorf("cannot convert %s to %s", val.Type(), r.FieldVal.Type())
			}
		}
	}

	return nil
}

// SetString sets the value of the request field.
// converts the string value to the field type.
func (r RequestField) SetString(value string) error {
	if err := r.Validate(); err != nil {
		return err
	}

	switch r.FieldVal.Kind() {
	case reflect.String:
		r.FieldVal.SetString(value)
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		r.FieldVal.SetInt(v)
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		r.FieldVal.SetUint(v)
	case reflect.Float32,
		reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		r.FieldVal.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		r.FieldVal.SetBool(v)
	case reflect.Ptr:
		switch r.FieldVal.Type().Elem().Kind() {
		case reflect.String:
			r.FieldVal.Set(reflect.ValueOf(&value))
		case reflect.Int:
			v, err := convertInt[int](value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		case reflect.Int8:
			v, err := convertInt[int8](value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		case reflect.Int16:
			v, err := convertInt[int64](value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		case reflect.Int32:
			v, err := convertInt[int32](value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		case reflect.Int64:
			v, err := convertInt[int64](value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		case reflect.Uint:
			v, err := convertUint[uint](value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		case reflect.Uint8:
			v, err := convertUint[uint8](value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		case reflect.Uint16:
			v, err := convertUint[uint16](value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		case reflect.Uint32:
			v, err := convertUint[uint32](value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		case reflect.Uint64:
			v, err := convertUint[uint64](value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		case reflect.Float32:
			v, err := convertFloatPtr[float32](value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		case reflect.Float64:
			v, err := convertFloatPtr[float64](value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		case reflect.Bool:
			v, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			r.FieldVal.Set(reflect.ValueOf(&v))
		default:
			return fmt.Errorf("unsupported field type: %s", r.FieldVal.Kind())
		}
	default:
		return fmt.Errorf("unsupported field type: %s", r.FieldVal.Kind())
	}

	return nil
}

// Contains returns true if the request fields contain the name.
func (r RequestFields) Contains(name string) bool {
	_, ok := r[name]
	return ok
}

// String returns the string representation of the request fields.
func (r RequestFields) String() string {
	var sb strings.Builder
	for k, v := range r {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(v.FieldName)
		sb.WriteString(", ")
	}
	return sb.String()
}

// GetRequestFields returns the request fields of a struct.
func GetRequestFields(structPtr any) RequestFields {
	v := reflect.ValueOf(structPtr)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil
	}

	v = v.Elem()
	t := v.Type()
	fields := RequestFields{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)
		tag := field.Tag
		name, bindingType, ok := getTagValue(tag,
			string(Bind),
			string(JSON),
			string(Protobuf),
		)

		if !ok || name == "" {
			name = field.Name
		}

		fields[name] = RequestField{
			FieldName: field.Name,
			TagName:   name,
			Field:     field,
			FieldVal:  fieldVal,
			IsUri:     bindingType == URI,
			IsHeader:  bindingType == Header,
			IsQuery:   bindingType == Query,
			IsForm:    bindingType == Form,
			IsFile:    bindingType == File,
		}
	}
	return fields
}

// getTagValue returns the value of a tag.
func getTagValue(tag reflect.StructTag, keys ...string) (string, BindingType, bool) {
	for _, key := range keys {
		switch key {
		case string(JSON):
			if v, ok := getJSONTagValue(tag); ok {
				return v, None, true
			}
		case string(Protobuf):
			if v, ok := getProtobufTagValue(tag); ok {
				return v, None, true
			}
		case string(Bind):
			if v, bingType, ok := getBindTagValue(tag); ok {
				return v, bingType, true
			}
		default:
			// ignore
		}
	}
	return "", None, false
}

// getJSONTagValue returns the value of a JSON tag.
// e.g. struct { Name string `json:"name,omitempty"` }
func getJSONTagValue(tag reflect.StructTag) (string, bool) {
	v := tag.Get(string(JSON))
	if v == "" {
		return "", false
	}

	elems := strings.Split(v, ",")
	if len(elems) == 0 {
		return "", false
	}
	return elems[0], true
}

// getProtobufTagValue returns the value of a Protobuf tag.
// e.g. struct { Title string `protobuf:"bytes,1,opt,name=title,proto3" }
func getProtobufTagValue(tag reflect.StructTag) (string, bool) {
	tagMap := getTagMap(tag, string(Protobuf))
	if v, ok := tagMap["name"]; ok {
		return v, true
	}

	return "", false
}

// getBindTagValue returns the value of a Bind tag.
// e.g. struct { Name string `bind:"name"` }
func getBindTagValue(tag reflect.StructTag) (string, BindingType, bool) {
	tagMap := getTagMap(tag, string(Bind))
	name, ok := tagMap["name"]
	if !ok {
		return "", "", false
	}
	bingType, ok := getBindingType(tagMap)
	return name, bingType, ok
}

// getTagMap converts a tag to a map.
// e.g: `bind:"uri,name=id"` -> map[string]map[string]string{"uri": "", "name": "id"}
func getTagMap(tag reflect.StructTag, tagName string) map[string]string {
	m := map[string]string{}
	v := tag.Get(tagName)
	if v == "" {
		return m
	}

	elems := strings.Split(v, ",")
	for _, elem := range elems {
		kv := strings.Split(elem, "=")
		if len(kv) == 1 {
			m[kv[0]] = ""
		} else {
			m[kv[0]] = kv[1]
		}
	}
	return m
}

// getBindingType returns the binging type of a tag.
func getBindingType(tagMap map[string]string) (BindingType, bool) {
	for k := range tagMap {
		switch k {
		case string(URI):
			return URI, true
		case string(Query):
			return Query, true
		case string(Header):
			return Header, true
		case string(Form):
			return Form, true
		case string(File):
			return File, true
		default:
			// ignore
		}
	}
	return None, false
}

type (
	intType interface {
		int | int8 | int16 | int32 | int64
	}
	uintType interface {
		uint | uint8 | uint16 | uint32 | uint64
	}
	floatType interface {
		float32 | float64
	}
)

func convertInt[T intType](s string) (T, error) {
	var zero T
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return zero, err
	}
	return T(v), nil
}

func convertUint[T uintType](s string) (T, error) {
	var zero T
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return zero, err
	}
	return T(v), nil
}

func convertFloatPtr[T floatType](s string) (T, error) {
	var zero T
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return zero, err
	}
	return T(v), nil
}
