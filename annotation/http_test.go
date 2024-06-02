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

type TestTagStruct2 struct {
	URIField       int64   `bind:"uri,name=uri_field"`
	URIPtrField    *int64  `bind:"uri,name=uri_ptr_field"`
	HeaderField    bool    `bind:"header,name=header_field"`
	HeaderPtrField *bool   `bind:"header,name=header_ptr_field"`
	QueryField     int32   `bind:"query,name=query_field"`
	QueryPtrField  *int32  `bind:"query,name=query_ptr_field"`
	FormField      string  `bind:"form,name=form_field"`
	FormPtrField   *string `bind:"form,name=form_ptr_field"`
	FileField      []byte  `bind:"file,name=file_field"`
	JSONField      string  `json:"json_field"`
	JSONPtrField   *string `json:"json_ptr_field"`
	ProtoField     string  `protobuf:"bytes,1,opt,name=proto_field,proto3"`
	ProtoPtrField  *string `protobuf:"bytes,1,opt,name=proto_ptr_field,proto3"`
	NoTagField     string
	NoTagPtrField  *string
}

func TestRequestField_Set(t *testing.T) {
	testStruct := &TestTagStruct2{}
	requestFields := GetRequestFields(testStruct)
	for key, field := range requestFields {
		switch key {
		case "uri_field":
			err := field.Set(int64(12))
			assert.NoError(t, err)
		case "uri_ptr_field":
			i := int64(34)
			err := field.Set(&i)
			assert.NoError(t, err)
		case "header_field":
			err := field.Set(true)
			assert.NoError(t, err)
		case "header_ptr_field":
			b := true
			err := field.Set(&b)
			assert.NoError(t, err)
		case "query_field":
			err := field.Set(123)
			assert.NoError(t, err)
		case "query_ptr_field":
			i := int32(456)
			err := field.Set(&i)
			assert.NoError(t, err)
		case "form_field":
			err := field.Set("new_form_value")
			assert.NoError(t, err)
		case "form_ptr_field":
			s := "new_form_ptr_value"
			err := field.Set(&s)
			assert.NoError(t, err)
		case "file_field":
			err := field.Set([]byte("new_file_value"))
			assert.NoError(t, err)
		case "json_field":
			err := field.Set("new_json_value")
			assert.NoError(t, err)
		case "json_ptr_field":
			s := "new_json_ptr_value"
			err := field.Set(&s)
			assert.NoError(t, err)
		case "proto_field":
			err := field.Set("new_proto_value")
			assert.NoError(t, err)
		case "proto_ptr_field":
			s := "new_proto_ptr_value"
			err := field.Set(&s)
			assert.NoError(t, err)
		case "NoTagField":
			err := field.Set("new_notag_value")
			assert.NoError(t, err)
		case "NoTagPtrField":
			s := "new_notag_ptr_value"
			err := field.Set(&s)
			assert.NoError(t, err)
		}
	}

	t.Logf("testStruct: %+v", testStruct)
	assert.Equal(t, int64(12), testStruct.URIField)
	assert.Equal(t, int64(34), *testStruct.URIPtrField)
	assert.Equal(t, true, testStruct.HeaderField)
	assert.Equal(t, true, *testStruct.HeaderPtrField)
	assert.Equal(t, int32(123), testStruct.QueryField)
	assert.Equal(t, int32(456), *testStruct.QueryPtrField)
	assert.Equal(t, "new_form_value", testStruct.FormField)
	assert.Equal(t, "new_form_ptr_value", *testStruct.FormPtrField)
	assert.Equal(t, []byte("new_file_value"), testStruct.FileField)
	assert.Equal(t, "new_json_value", testStruct.JSONField)
	assert.Equal(t, "new_json_ptr_value", *testStruct.JSONPtrField)
	assert.Equal(t, "new_proto_value", testStruct.ProtoField)
	assert.Equal(t, "new_proto_ptr_value", *testStruct.ProtoPtrField)
	assert.Equal(t, "new_notag_value", testStruct.NoTagField)
	assert.Equal(t, "new_notag_ptr_value", *testStruct.NoTagPtrField)

}

func TestRequestField_Set2(t *testing.T) {
	testStruct := &TestTagStruct2{}
	requestFields := GetRequestFields(testStruct)
	for key, field := range requestFields {
		switch key {
		case "uri_field":
			i := int64(34)
			err := field.Set(&i)
			assert.NoError(t, err)
		case "uri_ptr_field":
			err := field.Set(int64(12))
			assert.NoError(t, err)
		case "header_field":
			b := true
			err := field.Set(&b)
			assert.NoError(t, err)
		case "header_ptr_field":
			err := field.Set(true)
			assert.NoError(t, err)
		case "query_field":
			i := int32(456)
			err := field.Set(&i)
			assert.NoError(t, err)
		case "query_ptr_field":
			err := field.Set(123)
			assert.NoError(t, err)
		case "form_field":
			s := "new_form_ptr_value"
			err := field.Set(&s)
			assert.NoError(t, err)

		case "form_ptr_field":
			err := field.Set("new_form_value")
			assert.NoError(t, err)
		case "file_field":
			err := field.Set([]byte("new_file_value"))
			assert.NoError(t, err)
		case "json_field":
			s := "new_json_ptr_value"
			err := field.Set(&s)
			assert.NoError(t, err)
		case "json_ptr_field":
			err := field.Set("new_json_value")
			assert.NoError(t, err)
		case "proto_field":
			s := "new_proto_ptr_value"
			err := field.Set(&s)
			assert.NoError(t, err)
		case "proto_ptr_field":
			err := field.Set("new_proto_value")
			assert.NoError(t, err)
		case "NoTagField":
			s := "new_notag_ptr_value"
			err := field.Set(&s)
			assert.NoError(t, err)
		case "NoTagPtrField":
			err := field.Set("new_notag_value")
			assert.NoError(t, err)

		}
	}

	t.Logf("testStruct: %+v", testStruct)
	assert.Equal(t, int64(34), testStruct.URIField)
	assert.Equal(t, int64(12), *testStruct.URIPtrField)
	assert.Equal(t, true, testStruct.HeaderField)
	assert.Equal(t, true, *testStruct.HeaderPtrField)
	assert.Equal(t, int32(456), testStruct.QueryField)
	assert.Equal(t, int32(123), *testStruct.QueryPtrField)
	assert.Equal(t, "new_form_ptr_value", testStruct.FormField)
	assert.Equal(t, "new_form_value", *testStruct.FormPtrField)
	assert.Equal(t, []byte("new_file_value"), testStruct.FileField)
	assert.Equal(t, "new_json_ptr_value", testStruct.JSONField)
	assert.Equal(t, "new_json_value", *testStruct.JSONPtrField)
	assert.Equal(t, "new_proto_ptr_value", testStruct.ProtoField)
	assert.Equal(t, "new_proto_value", *testStruct.ProtoPtrField)
	assert.Equal(t, "new_notag_ptr_value", testStruct.NoTagField)
	assert.Equal(t, "new_notag_value", *testStruct.NoTagPtrField)

}
