package annotation

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetStringValue(t *testing.T) {
	var (
		intValue         int
		intPtrValue      = new(int)
		uintValue        uint
		uintPtrValue     = new(uint)
		floatValue       float64
		floatPtrValue    = new(float64)
		boolValue        bool
		boolPtrValue     = new(bool)
		stringValue      string
		stringPtrValue   = new(string)
		byteSliceValue   []byte
		unsupportedValue complex64
	)

	tests := []struct {
		name      string
		fieldVal  interface{}
		val       string
		expected  interface{}
		expectErr bool
	}{
		{"Set int value", &intValue, "123", 123, false},
		{"Set int pointer value", &intPtrValue, "123", 123, false},
		{"Set uint value", &uintValue, "123", uint(123), false},
		{"Set uint pointer value", &uintPtrValue, "123", uint(123), false},
		{"Set float value", &floatValue, "123.45", 123.45, false},
		{"Set float pointer value", &floatPtrValue, "123.45", 123.45, false},
		{"Set bool value", &boolValue, "true", true, false},
		{"Set bool pointer value", &boolPtrValue, "true", true, false},
		{"Set string value", &stringValue, "hello", "hello", false},
		{"Set string pointer value", &stringPtrValue, "hello", "hello", false},
		{"Set byte slice value", &byteSliceValue, "hello", []byte("hello"), false},
		{"Unsupported value", &unsupportedValue, "hello", complex64(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldVal := reflect.ValueOf(tt.fieldVal).Elem()
			err := setStringValue(fieldVal, tt.val)
			t.Logf("fieldVal: %v, err: %v", fieldVal, err)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.fieldVal != nil && fieldVal.Kind() == reflect.Ptr && !fieldVal.IsNil() {
					assert.Equal(t, tt.expected, fieldVal.Elem().Interface())
				} else {
					assert.Equal(t, tt.expected, fieldVal.Interface())
				}
			}
		})
	}
}

type TestInnerConfig struct {
	InnerField int `default:"42"`
}

type TestConfig struct {
	IntField       int              `default:"123"`
	UintField      uint             `default:"123"`
	FloatField     float64          `default:"123.45"`
	BoolField      bool             `default:"true"`
	StringField    string           `default:"hello"`
	InnerConfig    TestInnerConfig  // Nested struct
	InnerConfigPtr *TestInnerConfig // Pointer to struct
}

type TestStruct struct {
	state int

	IntField          int `default:"123"`
	IntField2         int
	UintField         uint      `default:"123"`
	FloatField        float64   `default:"123.45"`
	BoolField         bool      `default:"true"`
	StringField       string    `default:"hello"`
	ByteSliceField    []byte    `default:"byte_slice"`
	UnsupportedField  complex64 `default:"unsupported"`
	IntPtrField       *int      `default:"123"`
	IntPtrField2      *int
	UintPtrField      *uint    `default:"123"`
	FloatPtrField     *float64 `default:"123.45"`
	BoolPtrField      *bool    `default:"true"`
	StringPtrField    *string  `default:"hello"`
	ByteSlicePtrField *[]byte  `default:"byte_slice"`
	Config            TestConfig
}

func TestParseDefault(t *testing.T) {
	testStruct := &TestStruct{}
	err := ParseDefault(testStruct)
	assert.NoError(t, err)
	assert.Equal(t, 123, testStruct.IntField)
	assert.Equal(t, 0, testStruct.IntField2)
	assert.Equal(t, uint(123), testStruct.UintField)
	assert.Equal(t, 123.45, testStruct.FloatField)
	assert.Equal(t, true, testStruct.BoolField)
	assert.Equal(t, "hello", testStruct.StringField)
	assert.Equal(t, []byte("byte_slice"), testStruct.ByteSliceField)
	assert.NotNil(t, testStruct.IntPtrField)
	assert.Equal(t, 123, *testStruct.IntPtrField)
	assert.NotNil(t, testStruct.IntPtrField2)
	assert.NotNil(t, testStruct.UintPtrField)
	assert.Equal(t, uint(123), *testStruct.UintPtrField)
	assert.NotNil(t, testStruct.FloatPtrField)
	assert.Equal(t, 123.45, *testStruct.FloatPtrField)
	assert.NotNil(t, testStruct.BoolPtrField)
	assert.Equal(t, true, *testStruct.BoolPtrField)
	assert.NotNil(t, testStruct.StringPtrField)
	assert.Equal(t, "hello", *testStruct.StringPtrField)
	assert.NotNil(t, testStruct.ByteSlicePtrField)
	assert.Equal(t, []byte("byte_slice"), *testStruct.ByteSlicePtrField)
	assert.Equal(t, 123, testStruct.Config.IntField)
	assert.Equal(t, uint(123), testStruct.Config.UintField)
	assert.Equal(t, 123.45, testStruct.Config.FloatField)
	assert.Equal(t, true, testStruct.Config.BoolField)
	assert.Equal(t, "hello", testStruct.Config.StringField)
	assert.Equal(t, 42, testStruct.Config.InnerConfig.InnerField)
	assert.NotNil(t, testStruct.Config.InnerConfigPtr)
	assert.Equal(t, 42, testStruct.Config.InnerConfigPtr.InnerField)

}

func TestUpdateStructFieldsUnsupportedField(t *testing.T) {
	type UnsupportedTestStruct struct {
		UnsupportedField complex64 `default:"unsupported"`
	}

	testStruct := &UnsupportedTestStruct{}
	err := ParseDefault(testStruct)
	assert.NoError(t, err)

	assert.Equal(t, complex64(0), testStruct.UnsupportedField)
}

func TestUpdateStructFieldsNonStruct(t *testing.T) {
	var nonStruct int
	err := ParseDefault(&nonStruct)
	assert.Error(t, err)
}

func TestUpdateStructFieldsNilPointer(t *testing.T) {
	var nilPointer *TestStruct
	err := ParseDefault(nilPointer)
	assert.Error(t, err)
}
