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

	tests := []struct {
		name      string
		fieldIdl  *idl.IntField
		val       reflect.Value
		expectErr bool
	}{
		{
			name: "Valid value",
			fieldIdl: &idl.IntField{
				Gt: &gt,
				Lt: &lt,
				In: in,
			},
			val: func() reflect.Value {
				var id int64 = 12
				return reflect.ValueOf(id)
			}(),
			expectErr: false,
		},
		{
			name: "Valid value",
			fieldIdl: &idl.IntField{
				Gt: &gt,
				Lt: &lt,
				In: in,
			},
			val: func() reflect.Value {
				var id int64 = 12
				return reflect.ValueOf(&id)
			}(),
			expectErr: false,
		},
		{
			name: "Nil value with default",
			fieldIdl: &idl.IntField{
				Default:  &defaultValue,
				Required: &required,
			},
			val: func() reflect.Value {
				var id *int64
				return reflect.ValueOf(id)
			}(),
			expectErr: false,
		},
		{
			name: "Required field is nil without default",
			fieldIdl: &idl.IntField{
				Required: &required,
			},
			val: func() reflect.Value {
				var id *int64
				return reflect.ValueOf(id)
			}(),
			expectErr: true,
		},
		{
			name: "Value less than GT",
			fieldIdl: &idl.IntField{
				Gt: &gt,
			},
			val: func() reflect.Value {
				id := proto.Int32(4)
				return reflect.ValueOf(id)
			}(),
			expectErr: true,
		},
		{
			name: "Value equal to GTE",
			fieldIdl: &idl.IntField{
				Gte: &gte,
			},
			val: func() reflect.Value {
				var id int64 = 6
				return reflect.ValueOf(id)
			}(),
			expectErr: false,
		},
		{
			name: "Value greater than or equal to LTE",
			fieldIdl: &idl.IntField{
				Lte: &lte,
			},
			val: func() reflect.Value {
				var id int64 = 14
				return reflect.ValueOf(id)
			}(),
			expectErr: false,
		},
		{
			name: "Value Equal",
			fieldIdl: &idl.IntField{
				Eq: &eq,
			},
			val: func() reflect.Value {
				var id int64 = 14
				return reflect.ValueOf(id)
			}(),
			expectErr: true,
		},
		{
			name: "Value Equal",
			fieldIdl: &idl.IntField{
				Eq: &eq,
			},
			val: func() reflect.Value {
				var id int64 = 14
				return reflect.ValueOf(&id)
			}(),
			expectErr: true,
		},
		{
			name: "Value Not Equal",
			fieldIdl: &idl.IntField{
				Eq: &eq,
			},
			val: func() reflect.Value {
				var id int64 = 12
				return reflect.ValueOf(id)
			}(),
			expectErr: false,
		},
		{
			name: "Value Not Equal to NE",
			fieldIdl: &idl.IntField{
				Ne: &ne,
			},
			val: func() reflect.Value {
				var id int64 = 12
				return reflect.ValueOf(id)
			}(),
			expectErr: false,
		},
		{
			name: "Value Equal to NE",
			fieldIdl: &idl.IntField{
				Ne: &ne,
			},
			val: func() reflect.Value {
				var id int64 = 13
				return reflect.ValueOf(id)
			}(),
			expectErr: true,
		},
		{
			name: "Value not in IN list",
			fieldIdl: &idl.IntField{
				In: in,
			},
			val: func() reflect.Value {
				var id int64 = 16
				return reflect.ValueOf(id)
			}(),
			expectErr: true},
		{
			name: "Value in NOT IN list",
			fieldIdl: &idl.IntField{
				NotIn: notIn,
			},
			val: func() reflect.Value {
				var id int64 = 1
				return reflect.ValueOf(id)
			}(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntField(tt.fieldIdl, "Id", tt.val)
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
