package ginext

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

// getStructFieldNames returns the field names of a struct
// only for public fields
// e.g:
//
//	struct MultipartRequest{ {
//		File []byte `protobuf:"bytes,1,opt,name=file,proto3"`
//		Filename string `protobuf:"bytes,2,opt,name=filename,proto3"`
//		Label string `protobuf:"bytes,3,opt,name=label,proto3"`
//	    status string
//	}
//
// status will not be returned
func getStructFieldNames(v any) ([]string, error) {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr {
		return nil, errors.New("v must be a struct pointer")
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return nil, errors.New("v must be a struct pointer")
	}

	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		fields = append(fields, field.Name)
	}
	return fields, nil
}

// getProtoFieldNames returns the field names of a proto message
// e.g:
//
//	struct MultipartRequest{ {
//		File []byte `protobuf:"bytes,1,opt,name=file,proto3"`
//		Filename string `protobuf:"bytes,2,opt,name=filename,proto3"`
//		Label string `protobuf:"bytes,3,opt,name=label,proto3"`
//	    status string
//	}
//
// status will not be ignored
// returns map[protobufName]fieldName
func getFieldNames(ptr any) (map[string]string, error) {
	typ, err := getPtrElem(ptr)
	if err != nil {
		return nil, err
	}

	var fields = make(map[string]string)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		if protoName, ok := getProtoName(field); ok {
			fields[protoName] = field.Name
			continue
		}
		if jsonName, ok := getJSONName(field); ok {
			fields[jsonName] = field.Name
			continue
		}
		fields[field.Name] = field.Name
	}
	return fields, nil
}

// getProtoName returns the proto name of a field
// e.g:
// Filename string `protobuf:"bytes,2,opt,name=filename,proto3"`
// returns filename
func getProtoName(field reflect.StructField) (string, bool) {
	protoTag := field.Tag.Get("protobuf")
	if protoTag == "" {
		return "", false
	}

	elems := strings.Split(protoTag, ",")
	if len(elems) == 1 {
		return elems[0], true
	}

	for _, tag := range elems {
		if strings.HasPrefix(tag, "name=") {
			return strings.TrimPrefix(tag, "name="), true
		}
	}
	return "", false
}

// getJSONName returns the json name of a field
// e.g:
// Filename string `json:"filename,omitempty"`
// returns filename
func getJSONName(field reflect.StructField) (string, bool) {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return "", false
	}

	elems := strings.Split(jsonTag, ",")
	if len(elems) > 1 {
		return elems[0], true
	}

	return jsonTag, true
}

// getFormName returns the form name of a field
// e.g:
// Page int32 `protobuf:"varint,2,opt,name=page,proto3" json:"page,omitempty" form:"page"`
// returns page
func getFormName(field reflect.StructField) (string, bool) {
	formTag := field.Tag.Get("form")
	if formTag == "" {
		return "", false
	}

	elems := strings.Split(formTag, ",")
	if len(elems) > 1 {
		return elems[0], true
	}

	return formTag, true
}

// getUriName returns the uri name of a field
// e.g:
//
//	AuthorId int32 `protobuf:"varint,4,opt,name=author_id,json=authorId,proto3" json:"author_id,omitempty" form:"author_id" uri:"author_id"`
//
// returns author_id
func getUriName(field reflect.StructField) (string, bool) {
	uriTag := field.Tag.Get("uri")
	if uriTag == "" {
		return "", false
	}

	elems := strings.Split(uriTag, ",")
	if len(elems) > 1 {
		return elems[0], true
	}

	return uriTag, true

}

func setField(val reflect.Value, fieldName string, strVal string) error {
	field := val.Elem().FieldByName(fieldName)
	if !field.IsValid() {
		return errors.New("field not found")
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(strVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := strconv.ParseInt(strVal, 10, 64); err == nil {
			field.SetInt(i)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if i, err := strconv.ParseUint(strVal, 10, 64); err == nil {
			field.SetUint(i)
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(strVal, 64); err == nil {
			field.SetFloat(f)
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(strVal); err == nil {
			field.SetBool(b)
		}
	default:
		return errors.New("unsupported field type")
	}
	return nil
}

// validatePtr validates if v is a struct pointer
func getPtrElem(v any) (reflect.Type, error) {
	if v == nil {
		return nil, errors.New("is nil")
	}
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Ptr {
		return nil, errors.New("need a pointer")
	}

	typ = typ.Elem()
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("need a struct pointer")
	}

	return typ, nil
}
