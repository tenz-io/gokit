package annotation

import (
	"reflect"
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
