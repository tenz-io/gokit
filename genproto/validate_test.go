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

func TestIsNilOrEmpty(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want bool
	}{
		{"nil value", nil, true},
		{"nil pointer", (*int)(nil), true},
		{"empty slice", []int{}, true},
		{"non-empty slice", []int{1, 2, 3}, false},
		{"empty map", map[string]int{}, true},
		{"non-empty map", map[string]int{"key": 1}, false},
		{"empty array", [0]int{}, true},
		{"non-empty array", [1]int{1}, false},
		{"zero value int", 0, false},
		{"non-zero value int", 1, false},
		{"empty string", "", true},
		{"non-empty string", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNilOrEmpty(tt.val); got != tt.want {
				t.Errorf("IsNilOrEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetValue(t *testing.T) {
	var (
		i32 int32  = 42
		i64 int64  = 42
		u32 uint32 = 42
	)

	tests := []struct {
		name       string
		ptrVal     any
		newVal     any
		wantOk     bool
		wantVal    any
		wantNewPtr bool
	}{
		{"set int value", new(int), 42, true, 42, false},
		{"set int32 value", new(int32), i64, true, i32, false},
		{"set uint32 value", new(uint32), i64, true, u32, false},
		{"set float64 value", new(float64), 3.14, true, 3.14, false},
		{"set string value", new(string), "hello", true, "hello", false},
		{"set struct value", new(struct{ X int }), struct{ X int }{X: 10}, true, struct{ X int }{X: 10}, false},
		{"nil pointer", (*int)(nil), 42, true, 42, true},
		{"type mismatch", new(int), "string", false, 0, false},
		{"non-pointer value", 42, 42, false, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newPtrVal, got := SetValue(tt.ptrVal, tt.newVal)
			if got != tt.wantOk {
				t.Errorf("SetValue() = %v, want %v", got, tt.wantOk)
			}

			t.Logf("newPtrVal: %v, %T, ok: %t", newPtrVal, newPtrVal, got)

			if tt.wantOk {
				if tt.wantNewPtr {
					if reflect.ValueOf(newPtrVal).Elem().Interface() != tt.newVal {
						t.Errorf("New pointer value = %v, want %v", reflect.ValueOf(newPtrVal).Elem().Interface(), tt.newVal)
					}
				} else {
					v := reflect.ValueOf(tt.ptrVal).Elem().Interface()
					if !reflect.DeepEqual(v, tt.wantVal) {
						t.Errorf("Set value = %v, want %v", v, tt.wantVal)
					}
				}
			}
		})
	}
}

func TestStringIn(t *testing.T) {
	type args struct {
		s    string
		list []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"string in list", args{"avatar", []string{"avatar", "background", "post"}}, true},
		{"string not in list", args{"invalid", []string{"avatar", "background", "post"}}, false},
		{"empty list", args{"avatar", []string{}}, false},
		{"empty string", args{"", []string{"avatar", "background", "post"}}, false},
		{"empty string and list", args{"", []string{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, StringIn(tt.args.s, tt.args.list), "StringIn(%v, %v)", tt.args.s, tt.args.list)
		})
	}
}

func TestStringMatches(t *testing.T) {
	type args struct {
		s       string
		pattern string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"match pattern", args{"test", "test"}, true},
		{"match pattern with wildcard", args{"test", "te*"}, true},
		{"no match pattern", args{"test", "tes"}, true},
		{"empty string", args{"", "test"}, false},
		{"empty pattern", args{"test", ""}, true},
		{"empty string and pattern", args{"", ""}, true},
		{"empty string and wildcard", args{"", ".*"}, true},
		{"match all", args{"abb", ".*"}, true},
		{"match int", args{"123", "\\d+"}, true},
		{"match email", args{"example.12@gmail.com", "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, StringMatches(tt.args.s, tt.args.pattern), "StringMatches(%v, %v)", tt.args.s, tt.args.pattern)
		})
	}
}

func TestIntIn(t *testing.T) {
	type args[T intTyp] struct {
		i    T
		list []int
	}
	type testCase[T intTyp] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[int32]{
		{"int in list", args[int32]{10, []int{10, 12, 14}}, true},
		{"int not in list", args[int32]{11, []int{10, 12, 14}}, false},
		{"empty list", args[int32]{10, []int{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IntIn(tt.args.i, tt.args.list), "IntIn(%v, %v)", tt.args.i, tt.args.list)
		})
	}
}
