package ginext

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/tenz-io/gokit/ginext/errcode"
	"github.com/tenz-io/gokit/tracer"
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

// getFieldNames returns the field names of a proto message
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

func getErrCode(err error) string {
	if err == nil {
		return "0"
	}
	if e := new(errcode.Error); errors.As(err, &e) {
		return fmt.Sprintf("%d", e.Code)
	}
	return "1"
}

func getErrMsg(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func getTracerFlag(requestFlag string) tracer.Flag {
	if requestFlag == "" {
		return tracer.FlagNone
	}

	modes := strings.Split(strings.TrimSpace(requestFlag), "|")
	flag := tracer.FlagNone
	for _, mode := range modes {
		switch mode {
		case "normal", "":
		case "debug":
			flag |= tracer.FlagDebug
		case "shadow":
			flag |= tracer.FlagShadow
		case "stress":
			flag |= tracer.FlagStress
		default:
			//ignore
		}
	}
	return flag
}
