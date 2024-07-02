package genproto

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"

	"github.com/tenz-io/gokit/genproto/go/custom/idl"
)

func TestValidateIntField(t *testing.T) {
	defaultValue := int64(10)
	required := true
	gt := int64(5)
	gte := int64(6)
	lt := int64(15)
	lte := int64(14)
	eq := int64(12)
	ne := int64(13)
	in := []int64{10, 12, 14}
	notIn := []int64{1, 2, 3}

	type testMsg struct {
		Id  *int64
		Id2 int32
	}

	tests := []struct {
		name      string
		fieldIdl  *idl.IntField
		fieldName string
		val       any
		expectErr bool
	}{
		{
			name: "Valid value",
			fieldIdl: &idl.IntField{
				Gt: &gt,
				Lt: &lt,
				In: in,
			},
			fieldName: "Id",
			val: &testMsg{
				Id: proto.Int64(12),
			},
			expectErr: false,
		},
		{
			name: "Valid value",
			fieldIdl: &idl.IntField{
				Gt: &gt,
				Lt: &lt,
				In: in,
			},
			fieldName: "Id2",
			val: &testMsg{
				Id2: 12,
			},
			expectErr: false,
		},
		{
			name: "Nil value with default",
			fieldIdl: &idl.IntField{
				Default:  &defaultValue,
				Required: &required,
			},
			fieldName: "Id",
			val:       &testMsg{},
			expectErr: false,
		},
		{
			name: "Required field is nil without default",
			fieldIdl: &idl.IntField{
				Required: &required,
			},
			fieldName: "Id",
			val:       &testMsg{},
			expectErr: true,
		},
		{
			name: "Value less than GT",
			fieldIdl: &idl.IntField{
				Gt: &gt,
			},
			fieldName: "Id",
			val: &testMsg{
				Id: proto.Int64(4),
			},
			expectErr: true,
		},
		{
			name: "Value equal to GTE",
			fieldIdl: &idl.IntField{
				Gte: &gte,
			},
			fieldName: "Id",
			val: &testMsg{
				Id: proto.Int64(6),
			},
			expectErr: false,
		},
		{
			name: "Value greater than or equal to LTE",
			fieldIdl: &idl.IntField{
				Lte: &lte,
			},
			fieldName: "Id",
			val: &testMsg{
				Id: proto.Int64(14),
			},
			expectErr: false,
		},
		{
			name: "Value Equal",
			fieldIdl: &idl.IntField{
				Eq: &eq,
			},
			fieldName: "Id",
			val: &testMsg{
				Id: proto.Int64(14),
			},
			expectErr: true,
		},
		{
			name: "Value Equal",
			fieldIdl: &idl.IntField{
				Eq: &eq,
			},
			fieldName: "Id",
			val: &testMsg{
				Id: proto.Int64(14),
			},
			expectErr: true,
		},
		{
			name: "Value Not Equal",
			fieldIdl: &idl.IntField{
				Eq: &eq,
			},
			fieldName: "Id",
			val: &testMsg{
				Id: proto.Int64(12),
			},
			expectErr: false,
		},
		{
			name: "Value Not Equal to NE",
			fieldIdl: &idl.IntField{
				Ne: &ne,
			},
			fieldName: "Id",
			val: &testMsg{
				Id: proto.Int64(12),
			},
			expectErr: false,
		},
		{
			name: "Value Equal to NE",
			fieldIdl: &idl.IntField{
				Ne: &ne,
			},
			fieldName: "Id",
			val: &testMsg{
				Id: proto.Int64(13),
			},
			expectErr: true,
		},
		{
			name: "Value not in IN list",
			fieldIdl: &idl.IntField{
				In: in,
			},
			fieldName: "Id",
			val: &testMsg{
				Id: proto.Int64(16),
			},
			expectErr: true},
		{
			name: "Value in NOT IN list",
			fieldIdl: &idl.IntField{
				NotIn: notIn,
			},
			fieldName: "Id",
			val: &testMsg{
				Id: proto.Int64(1),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntField(tt.fieldIdl, tt.fieldName, tt.val)
			t.Logf("err: %v", err)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateIntField() error = %v, expectErr %v", err, tt.expectErr)
			}

		})
	}
}

type testMsg struct {
	Id   *int64
	Name *string
	Age  *int
}

func TestSetDefaultValue(t *testing.T) {
	defaultInt64 := int64(10)
	defaultString := "default"
	defaultInt := 20

	tests := []struct {
		name         string
		msg          any
		fieldName    string
		defaultValue any
		expectErr    bool
		expectedVal  any
	}{
		{
			name:         "Set default value for *int64 field",
			msg:          &testMsg{},
			fieldName:    "Id",
			defaultValue: defaultInt64,
			expectErr:    false,
			expectedVal:  &defaultInt64,
		},
		{
			name:         "Set default value for *string field",
			msg:          &testMsg{},
			fieldName:    "Name",
			defaultValue: defaultString,
			expectErr:    false,
			expectedVal:  &defaultString,
		},
		{
			name:         "Set default value for *int field",
			msg:          &testMsg{},
			fieldName:    "Age",
			defaultValue: defaultInt,
			expectErr:    false,
			expectedVal:  &defaultInt,
		},
		{
			name:         "Field not found",
			msg:          &testMsg{},
			fieldName:    "NonExistentField",
			defaultValue: defaultInt64,
			expectErr:    true,
			expectedVal:  nil,
		},
		{
			name:         "Type mismatch",
			msg:          &testMsg{},
			fieldName:    "Id",
			defaultValue: defaultString,
			expectErr:    true,
			expectedVal:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setDefaultValue(tt.msg, tt.fieldName, tt.defaultValue)
			if (err != nil) != tt.expectErr {
				t.Errorf("setDefaultValue() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			t.Logf("msg: %+v", tt.msg)

			if !tt.expectErr {
				// get the value of the field
				v := reflect.ValueOf(tt.msg).Elem().FieldByName(tt.fieldName)
				// check if the value is equal to the expected value
				if assert.Equal(t, tt.expectedVal, v.Interface()) {
					t.Logf("Value: %v", v.Interface())
				}
			}
		})
	}
}

type TestStruct struct {
	ValidField   string
	NilField     *string
	AnotherField int
}

func TestCheckNilField(t *testing.T) {
	var nilString *string
	validStruct := TestStruct{
		ValidField:   "value",
		NilField:     nilString,
		AnotherField: 42,
	}

	tests := []struct {
		name      string
		msg       any
		fieldName string
		want      bool
	}{
		{"nil msg", nil, "ValidField", true},
		{"valid field", validStruct, "ValidField", false},
		{"nil field", validStruct, "NilField", true},
		{"invalid field name", validStruct, "NonExistentField", true},
		{"nil struct pointer", (*TestStruct)(nil), "ValidField", true},
		{"non-nil struct pointer", &validStruct, "ValidField", false},
		{"nil field in pointer struct", &validStruct, "NilField", true},
		{"invalid field in pointer struct", &validStruct, "NonExistentField", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNilField(tt.msg, tt.fieldName); got != tt.want {
				t.Errorf("isNilField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getIntFieldVal(t *testing.T) {
	type TestStruct struct {
		IntField         int
		IntPtrField      *int
		UintField        uint
		UintPtrField     *uint
		InvalidField     any
		NilPtrField      *int
		UnsupportedField string
	}

	intVal := 42
	uintVal := uint(42)

	testStruct := TestStruct{
		IntField:         intVal,
		IntPtrField:      &intVal,
		UintField:        uintVal,
		UintPtrField:     &uintVal,
		InvalidField:     nil,
		NilPtrField:      nil,
		UnsupportedField: "test",
	}

	tests := []struct {
		name string
		val  reflect.Value
		want int64
	}{
		{"IntField", reflect.ValueOf(testStruct.IntField), 42},
		{"IntPtrField", reflect.ValueOf(testStruct.IntPtrField), 42},
		{"UintField", reflect.ValueOf(testStruct.UintField), 42},
		{"UintPtrField", reflect.ValueOf(testStruct.UintPtrField), 42},
		{"InvalidField", reflect.ValueOf(testStruct.InvalidField), 0},
		{"NilPtrField", reflect.ValueOf(testStruct.NilPtrField), 0},
		{"UnsupportedField", reflect.ValueOf(testStruct.UnsupportedField), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIntFieldVal(tt.val); got != tt.want {
				t.Errorf("getIntFieldVal() = %v, want %v", got, tt.want)
			}
		})
	}
}
