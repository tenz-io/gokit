package annotation

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestTagStruct struct {
	URIField    string `bind:"uri,name=uri_field"`
	HeaderField string `bind:"header,name=header_field"`
	QueryField  int    `bind:"query,name=query_field"`
	FormField   string `bind:"form,name=form_field"`
	FileField   []byte `bind:"file,name=file_field"`
	JSONField   string `json:"json_field"`
	ProtoField  string `protobuf:"bytes,1,opt,name=proto_field,proto3"`
	NoTagField  string
}

func TestGetRequestFields(t *testing.T) {
	testStruct := &TestTagStruct{
		URIField:    "uri_value",
		HeaderField: "header_value",
		QueryField:  123,
		FormField:   "form_value",
		FileField:   []byte("file_value"),
		JSONField:   "json_value",
		ProtoField:  "proto_value",
		NoTagField:  "notag_value",
	}

	fields := GetRequestFields(testStruct)
	t.Logf("fields: %s", fields)

	assert.Len(t, fields, 8)

	tests := []struct {
		name     string
		expected RequestField
	}{
		{
			name: "uri_field",
			expected: RequestField{
				FieldName: "URIField",
				TagName:   "uri_field",
				Field:     reflect.TypeOf(testStruct).Elem().Field(0),
				FieldVal:  reflect.ValueOf(testStruct).Elem().Field(0),
				IsUri:     true,
			},
		},
		{
			name: "header_field",
			expected: RequestField{
				FieldName: "HeaderField",
				TagName:   "header_field",
				Field:     reflect.TypeOf(testStruct).Elem().Field(1),
				FieldVal:  reflect.ValueOf(testStruct).Elem().Field(1),
				IsHeader:  true,
			},
		},
		{
			name: "query_field",
			expected: RequestField{
				FieldName: "QueryField",
				TagName:   "query_field",
				Field:     reflect.TypeOf(testStruct).Elem().Field(2),
				FieldVal:  reflect.ValueOf(testStruct).Elem().Field(2),
				IsQuery:   true,
			},
		},
		{
			name: "form_field",
			expected: RequestField{
				FieldName: "FormField",
				TagName:   "form_field",
				Field:     reflect.TypeOf(testStruct).Elem().Field(3),
				FieldVal:  reflect.ValueOf(testStruct).Elem().Field(3),
				IsForm:    true,
			},
		},
		{
			name: "file_field",
			expected: RequestField{
				FieldName: "FileField",
				TagName:   "file_field",
				Field:     reflect.TypeOf(testStruct).Elem().Field(4),
				FieldVal:  reflect.ValueOf(testStruct).Elem().Field(4),
				IsFile:    true,
			},
		},
		{
			name: "json_field",
			expected: RequestField{
				FieldName: "JSONField",
				TagName:   "json_field",
				Field:     reflect.TypeOf(testStruct).Elem().Field(5),
				FieldVal:  reflect.ValueOf(testStruct).Elem().Field(5),
			},
		},
		{
			name: "proto_field",
			expected: RequestField{
				FieldName: "ProtoField",
				TagName:   "proto_field",
				Field:     reflect.TypeOf(testStruct).Elem().Field(6),
				FieldVal:  reflect.ValueOf(testStruct).Elem().Field(6),
			},
		},
		{
			name: "NoTagField",
			expected: RequestField{
				FieldName: "NoTagField",
				TagName:   "NoTagField",
				Field:     reflect.TypeOf(testStruct).Elem().Field(7),
				FieldVal:  reflect.ValueOf(testStruct).Elem().Field(7),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, ok := fields[tt.name]
			t.Logf("field: %+v", field)
			assert.True(t, ok)
			assert.Equal(t, tt.expected.FieldName, field.FieldName)
			assert.Equal(t, tt.expected.TagName, field.TagName)
			assert.Equal(t, tt.expected.Field, field.Field)
			assert.Equal(t, tt.expected.FieldVal.Interface(), field.FieldVal.Interface())
			assert.Equal(t, tt.expected.IsUri, field.IsUri)
			assert.Equal(t, tt.expected.IsHeader, field.IsHeader)
			assert.Equal(t, tt.expected.IsQuery, field.IsQuery)
			assert.Equal(t, tt.expected.IsForm, field.IsForm)
			assert.Equal(t, tt.expected.IsFile, field.IsFile)
		})
	}
}

func TestGetTagMap(t *testing.T) {
	tests := []struct {
		name     string
		tag      reflect.StructTag
		tagName  string
		expected map[string]string
	}{
		{
			name:     "Empty tag",
			tag:      reflect.StructTag(``),
			tagName:  "bind",
			expected: map[string]string{},
		},
		{
			name:     "Single key",
			tag:      reflect.StructTag(`bind:"uri"`),
			tagName:  "bind",
			expected: map[string]string{"uri": ""},
		},
		{
			name:     "Single key-value",
			tag:      reflect.StructTag(`bind:"name=id"`),
			tagName:  "bind",
			expected: map[string]string{"name": "id"},
		},
		{
			name:     "Multiple keys and key-values",
			tag:      reflect.StructTag(`bind:"uri,name=id,type=int"`),
			tagName:  "bind",
			expected: map[string]string{"uri": "", "name": "id", "type": "int"},
		},
		{
			name:     "Irrelevant tag",
			tag:      reflect.StructTag(`json:"name" bind:"uri"`),
			tagName:  "bind",
			expected: map[string]string{"uri": ""},
		},
		{
			name:     "Empty value in tag",
			tag:      reflect.StructTag(`bind:""`),
			tagName:  "bind",
			expected: map[string]string{},
		},
		{
			name:     "Key with empty value",
			tag:      reflect.StructTag(`bind:"uri,name="`),
			tagName:  "bind",
			expected: map[string]string{"uri": "", "name": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTagMap(tt.tag, tt.tagName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
