package logger

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestOutputTrimmer_Json(t *testing.T) {
	type fields struct {
		arrLimit   int
		strLimit   int
		deepLimit  int
		wholeLimit int
		ignores    map[string]bool
	}
	type args struct {
		obj any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "when obj is nil then return empty string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: nil,
			},
			want: `null`,
		},
		{
			name: "when obj is empty string then return empty string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: "",
			},
			want: `""`,
		},
		{
			name: "when obj is string then return string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: "abc",
			},
			want: `"abc"`,
		},
		{
			name: "when obj is int then return int string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: 123,
			},
			want: `123`,
		},
		{
			name: "when obj is struct then return json string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					type user struct {
						Name string `json:"name"`
						Age  int    `json:"age"`
					}
					return user{
						Name: "Alice",
						Age:  18,
					}
				}(),
			},
			want: `{"age":18,"name":"Alice"}`,
		},
		{
			name: "when obj is struct ptr then return json string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					type user struct {
						Name string `json:"name"`
						Age  int    `json:"age"`
					}
					return &user{
						Name: "Alice",
						Age:  18,
					}
				}(),
			},
			want: `{"age":18,"name":"Alice"}`,
		},
		{
			name: "when obj is unsupported type then return error string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					return func() {}
				}(),
			},
			want: `null`,
		},
		{
			name: "when obj is interface type then return string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() myinterface {
					return &mystruct{}
				}(),
			},
			want: `{"v":0}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ot := &OutputTrimmer{
				arrLimit:   tt.fields.arrLimit,
				strLimit:   tt.fields.strLimit,
				deepLimit:  tt.fields.deepLimit,
				wholeLimit: tt.fields.wholeLimit,
				ignores:    tt.fields.ignores,
			}
			if got := ot.Json(tt.args.obj); got != tt.want {
				t.Errorf("Json() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_valuableType(t *testing.T) {
	type args struct {
		v []reflect.Value
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "when v is nil type then return false",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(nil),
				},
			},
			want: false,
		},
		{
			name: "when v is int type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(int8(123)),
					reflect.ValueOf(int16(123)),
					reflect.ValueOf(int32(123)),
					reflect.ValueOf(int64(123)),
					reflect.ValueOf(int32(123)),
					reflect.ValueOf(int(123)),
				},
			},
			want: true,
		},
		{
			name: "when v is uint type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(uint8(123)),
					reflect.ValueOf(uint16(123)),
					reflect.ValueOf(uint32(123)),
					reflect.ValueOf(uint64(123)),
					reflect.ValueOf(uint32(123)),
					reflect.ValueOf(uint(123)),
				},
			},
			want: true,
		},
		{
			name: "when v is bool type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(true),
					reflect.ValueOf(false),
				},
			},
			want: true,
		},
		{
			name: "when v is time type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(time.Now()),
					reflect.ValueOf(2 * time.Millisecond),
				},
			},
			want: true,
		},
		{
			name: "when v is bool ptr type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(func() *bool {
						var b = true
						return &b
					}()),
				},
			},
			want: true,
		},
		{
			name: "when v is string type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf("abc"),
					reflect.ValueOf(""),
					reflect.ValueOf(func() *string {
						s := "abc"
						return &s
					}()),
					reflect.ValueOf(func() *string {
						s := ""
						return &s
					}()),
				},
			},
			want: true,
		},
		{
			name: "when v is float type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(float32(1.23)),
					reflect.ValueOf(float64(1.23)),
				},
			},
			want: true,
		},
		{
			name: "when v is complex type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(complex(1, 2)),
					reflect.ValueOf(complex64(1.23)),
					reflect.ValueOf(complex128(1.23)),
				},
			},
			want: true,
		},
		{
			name: "when v is map type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(map[string]any{
						"a": 1,
					}),
				},
			},
			want: true,
		},
		{
			name: "when v is slice or array type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf([3]int{1, 2, 3}),
					reflect.ValueOf([3]byte{1, 2, 3}),
				},
			},
			want: true,
		},
		{
			name: "when v is struct type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(func() any {
						type user struct {
							name string
							age  int
						}
						return user{
							name: "gopher",
							age:  12,
						}
					}()),
				},
			},
			want: true,
		},
		{
			name: "when v is struct ptr type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(func() any {
						type user struct {
							name string
							age  int
						}
						return &user{
							name: "gopher",
							age:  12,
						}
					}()),
				},
			},
			want: true,
		},
		{
			name: "when v is struct ptr nil type then return false",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(nil),
					reflect.ValueOf(func() any {
						type user struct {
							name string
							age  int
						}
						var u *user
						return u
					}()),
				},
			},
			want: false,
		},
		{
			name: "when v is error type then return true",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(errors.New("oops")),
				},
			},
			want: true,
		},
		{
			name: "when v is unsupported type then return false",
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(func() {}),
					reflect.ValueOf(make(chan int)),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, vv := range tt.args.v {
				if got := valuableType(vv); got != tt.want {
					t.Errorf("%v: valuableType() = %v, want %v", vv, got, tt.want)
				}
			}
		})
	}
}

func TestOutputTrimmer_valOfPrimaryType(t *testing.T) {
	type fields struct {
		arrLimit   int
		strLimit   int
		deepLimit  int
		wholeLimit int
		ignores    map[string]bool
	}
	type args struct {
		v []reflect.Value
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantVal []any
		wantOk  bool
	}{
		{
			name: "when nil then not ok",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(nil),
				},
			},
			wantVal: []any{
				nil,
			},
			wantOk: false,
		},
		{
			name: "when bool values then ok",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(true),
					reflect.ValueOf(false),
				},
			},
			wantVal: []any{
				true,
				false,
			},
			wantOk: true,
		},
		{
			name: "when int values then ok",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(int(123)),
					reflect.ValueOf(int8(123)),
					reflect.ValueOf(int16(123)),
					reflect.ValueOf(int32(123)),
					reflect.ValueOf(int64(123)),
				},
			},
			wantVal: []any{
				int64(123),
				int64(123),
				int64(123),
				int64(123),
				int64(123),
			},
			wantOk: true,
		},
		{
			name: "when uint values then ok",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(uint(123)),
					reflect.ValueOf(uint8(123)),
					reflect.ValueOf(uint16(123)),
					reflect.ValueOf(uint32(123)),
					reflect.ValueOf(uint64(123)),
				},
			},
			wantVal: []any{
				uint64(123),
				uint64(123),
				uint64(123),
				uint64(123),
				uint64(123),
			},
			wantOk: true,
		},
		{
			name: "when float values then ok",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(float32(1)),
					reflect.ValueOf(float64(1)),
				},
			},
			wantVal: []any{
				float64(1),
				float64(1),
			},
			wantOk: true,
		},
		{
			name: "when string values then ok",
			fields: fields{
				arrLimit:   3,
				strLimit:   3,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: []reflect.Value{
					reflect.ValueOf("abc"),
					reflect.ValueOf("abcdef"),
				},
			},
			wantVal: []any{
				"abc",
				"abc...",
			},
			wantOk: true,
		},
		{
			name: "when ptr values then ok",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(func() *bool {
						var b = true
						return &b
					}()),
					reflect.ValueOf(func() any {
						var b = false
						return &b
					}()),
				},
			},
			wantVal: []any{
				true,
				false,
			},
			wantOk: true,
		},
		{
			name: "when struct values then not ok",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: []reflect.Value{
					reflect.ValueOf(func() any {
						var b = struct {
							flag string
						}{}
						return b
					}()),
					reflect.ValueOf(func() any {
						var b = struct {
							flag string
						}{}
						return &b
					}()),
				},
			},
			wantVal: []any{
				nil,
				nil,
			},
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ot := &OutputTrimmer{
				arrLimit:   tt.fields.arrLimit,
				strLimit:   tt.fields.strLimit,
				deepLimit:  tt.fields.deepLimit,
				wholeLimit: tt.fields.wholeLimit,
				ignores:    tt.fields.ignores,
			}

			if len(tt.args.v) != len(tt.wantVal) {
				t.Errorf("args size not equals want val size")
				return
			}

			for i, vv := range tt.args.v {
				gotVal, gotOk := ot.valOfPrimaryType(vv)
				if !reflect.DeepEqual(gotVal, tt.wantVal[i]) {
					t.Logf("got type: %T, want type: %T", gotVal, tt.wantVal[i])
					t.Errorf("valOfPrimaryType() gotVal = %v, want %v", gotVal, tt.wantVal[i])
				}
				if gotOk != tt.wantOk {
					t.Errorf("valOfPrimaryType() gotOk = %v, want %v", gotOk, tt.wantOk)
				}
			}
		})
	}
}

func Test_isBytes(t *testing.T) {
	type args struct {
		v      reflect.Value
		chkLen int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "when v is byte slice type then return true",
			args: args{
				v: reflect.ValueOf([]byte{1, 2, 3, 4}),
			},
			want: true,
		},
		{
			name: "when v is byte array type then return true",
			args: args{
				v: reflect.ValueOf([4]byte{1, 2, 3, 4}),
			},
			want: true,
		},
		{
			name: "when v is empty byte slice type then return true",
			args: args{
				v: reflect.ValueOf([]byte{}),
			},
			want: true,
		},
		{
			name: "when v is empty byte array type then return true",
			args: args{
				v: reflect.ValueOf([0]byte{}),
			},
			want: true,
		},
		{
			name: "when v is int slice then return false",
			args: args{
				v:      reflect.ValueOf([]int{1, 2, 3, 4}),
				chkLen: 3,
			},
			want: false,
		},
		{
			name: "when v is empty int slice then return false",
			args: args{
				v:      reflect.ValueOf([]int{}),
				chkLen: 3,
			},
		},
		{
			name: "when v is byte slice then return false",
			args: args{
				v: reflect.ValueOf([]any{
					byte(1),
					byte(2),
					byte(3),
					byte(4),
				}),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBytes(tt.args.v); got != tt.want {
				t.Errorf("isBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputTrimmer_valOfSpecialType(t *testing.T) {
	type fields struct {
		arrLimit   int
		strLimit   int
		deepLimit  int
		wholeLimit int
		ignores    map[string]bool
	}
	type args struct {
		v reflect.Value
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantVal any
		wantOk  bool
	}{
		{
			name: "when v is time type then return time string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			wantVal: "2021-01-01T00:00:00.000",
			wantOk:  true,
		},
		{
			name: "when v is time ptr type then return time string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(func() *time.Time {
					t := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
					return &t
				}()),
			},
			wantVal: "2021-01-01T00:00:00.000",
			wantOk:  true,
		},
		{
			name: "when v is error type then return error string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(fmt.Errorf("oops")),
			},
			wantVal: "oops",
			wantOk:  true,
		},
		{
			name: "when v is string type then return time string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(func() string {
					return "abc"
				}()),
			},
			wantVal: "abc",
			wantOk:  true,
		},
		{
			name: "when v is string type and larger than limit then return time string",
			fields: fields{
				arrLimit:   3,
				strLimit:   3,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(func() string {
					return "abcd"
				}()),
			},
			wantVal: "abc...",
			wantOk:  true,
		},
		{
			name: "when v is bytes type then return bytes string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf([]byte{1, 2, 3, 4}),
			},
			wantVal: "AQIDBA==",
			wantOk:  true,
		},
		{
			name: "when v is byte array type then return bytes string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf([4]byte{1, 2, 3, 4}),
			},
			wantVal: "AQIDBA==",
			wantOk:  true,
		},
		{
			name: "when v is byte array type len larger than limit then return nil",
			fields: fields{
				arrLimit:   3,
				strLimit:   3,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf([4]byte{1, 2, 3, 4}),
			},
			wantVal: nil,
			wantOk:  false,
		},
		{
			name: "when v is byte array ptr type then return bytes string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(func() *[4]byte {
					bs := [4]byte{1, 2, 3, 4}
					return &bs
				}()),
			},
			wantVal: "AQIDBA==",
			wantOk:  true,
		},
		{
			name: "when v is byte slice ptr type then return bytes string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(func() *[]byte {
					bs := []byte{1, 2, 3, 4}
					return &bs
				}()),
			},
			wantVal: "AQIDBA==",
			wantOk:  true,
		},
		{
			name: "when v is func type then return bytes string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(func() {}),
			},
			wantVal: nil,
			wantOk:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ot := &OutputTrimmer{
				arrLimit:   tt.fields.arrLimit,
				strLimit:   tt.fields.strLimit,
				deepLimit:  tt.fields.deepLimit,
				wholeLimit: tt.fields.wholeLimit,
				ignores:    tt.fields.ignores,
			}
			gotVal, gotOk := ot.valOfSpecialType(tt.args.v)
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("valOfSpecialType() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			if gotOk != tt.wantOk {
				t.Errorf("valOfSpecialType() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestOutputTrimmer_bytesString(t *testing.T) {
	type fields struct {
		arrLimit   int
		strLimit   int
		deepLimit  int
		wholeLimit int
		ignores    map[string]bool
	}
	type args struct {
		v reflect.Value
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		wantOk bool
	}{
		{
			name: "when v is byte slice then return string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf([]byte{1, 2, 3}),
			},
			want:   "AQID",
			wantOk: true,
		},
		{
			name: "when v is byte array then return string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf([3]byte{1, 2, 3}),
			},
			want:   "AQID",
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ot := &OutputTrimmer{
				arrLimit:   tt.fields.arrLimit,
				strLimit:   tt.fields.strLimit,
				deepLimit:  tt.fields.deepLimit,
				wholeLimit: tt.fields.wholeLimit,
				ignores:    tt.fields.ignores,
			}
			got, got1 := ot.bytesString(tt.args.v)
			if got != tt.want {
				t.Errorf("bytesString() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantOk {
				t.Errorf("bytesString() got1 = %v, want %v", got1, tt.wantOk)
			}
		})
	}
}

func TestOutputTrimmer_trimStruct(t *testing.T) {
	type fields struct {
		arrLimit   int
		strLimit   int
		deepLimit  int
		wholeLimit int
		ignores    map[string]bool
	}
	type args struct {
		v       reflect.Value
		deepLmt int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]any
	}{
		{
			name: "when v is nil with deep 0 then return nil",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v:       reflect.ValueOf(nil),
				deepLmt: 0,
			},
			want: map[string]any{},
		},
		{
			name: "when v is nil with deep 10 then return nil",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v:       reflect.ValueOf(nil),
				deepLmt: 10,
			},
			want: map[string]any{},
		},
		{
			name: "when v is struct then return map[string]any",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(func() any {
					type user struct {
						Name string `json:"name"`
						Age  int    `json:"age"`
					}
					return user{
						Name: "Alice",
						Age:  18,
					}
				}()),
				deepLmt: 10,
			},
			want: map[string]any{
				"name": "Alice",
				"age":  int64(18),
			},
		},
		{
			name: "when v is struct with slice then return map[string]any",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(func() any {
					type user struct {
						Name           string                   `json:"name"`
						Age            int                      `json:"age"`
						NickNames      []string                 `json:"nick_names"`
						Ext            []byte                   `json:"ext"`
						TraveledCities map[string]time.Duration `json:"traveled_cities"`
						Profile        *string                  `json:"profile"`
						Err            error                    `json:"err"`
					}
					return user{
						Name: "Alice",
						Age:  18,
						NickNames: []string{
							"Alice",
							"Ali",
							"Al",
							"A",
						},
						Ext: bytes.Repeat([]byte{1, 2, 3}, 100),
						TraveledCities: map[string]time.Duration{
							"Beijing":  4 * 24 * time.Hour,
							"Shanghai": 20 * 24 * time.Hour,
						},
						Profile: func() *string {
							var s = "abc123"
							return &s
						}(),
						Err: fmt.Errorf("oops"),
					}
				}()),
				deepLmt: 10,
			},
			want: map[string]any{
				"name": "Alice",
				"age":  int64(18),
				"nick_names": []any{
					"Alice",
					"Ali",
					"Al",
				},
				"ext": []any{uint64(1), uint64(2), uint64(3)},
				"traveled_cities": map[string]any{
					"Beijing":  "96h0m0s",
					"Shanghai": "480h0m0s",
				},
				"profile":           "abc123",
				"err":               "oops",
				"_size__nick_names": 4,
				"_size__ext":        300,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ot := &OutputTrimmer{
				arrLimit:   tt.fields.arrLimit,
				strLimit:   tt.fields.strLimit,
				deepLimit:  tt.fields.deepLimit,
				wholeLimit: tt.fields.wholeLimit,
				ignores:    tt.fields.ignores,
			}
			if got := ot.trimStruct(tt.args.v, tt.args.deepLmt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trimStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputTrimmer_trimSlice(t *testing.T) {
	type fields struct {
		arrLimit   int
		strLimit   int
		deepLimit  int
		wholeLimit int
		ignores    map[string]bool
	}
	type args struct {
		v       reflect.Value
		deepLmt int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []any
	}{
		{
			name: "when v is int slice with deep 10 then return trimmed slice",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v:       reflect.ValueOf([]int{1, 2, 3, 4}),
				deepLmt: 10,
			},
			want: []any{
				int64(1),
				int64(2),
				int64(3),
			},
		},
		{
			name: "when v is int ptr slice with deep 10 then return trimmed slice",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(func() []*int {
					return []*int{
						intPtr(1),
						nil,
						intPtr(3),
						intPtr(4),
					}
				}()),
				deepLmt: 10,
			},
			want: []any{
				int64(1),
				nil,
				int64(3),
			},
		},
		{
			name: "when v is struct slice with deep 10 then return trimmed slice",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(func() []any {
					type user struct {
						Name string `json:"name"`
						Age  int    `json:"age"`
					}
					return []any{
						user{
							Name: "Alice",
							Age:  18,
						},
						user{
							Name: "Bob",
							Age:  20,
						},
					}

				}()),
				deepLmt: 10,
			},
			want: []any{
				map[string]any{
					"name": "Alice",
					"age":  int64(18),
				},
				map[string]any{
					"name": "Bob",
					"age":  int64(20),
				},
			},
		},
		{
			name: "when v is slice slice with deep 10 then return trimmed slice",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				v: reflect.ValueOf(func() []any {
					return []any{
						[]string{
							"Alice",
							"Ali",
						},
						[]string{
							"Bob",
							"Bo",
						},
					}

				}()),
				deepLmt: 10,
			},
			want: []any{
				[]any{
					"Alice",
					"Ali",
				},
				[]any{
					"Bob",
					"Bo",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ot := &OutputTrimmer{
				arrLimit:   tt.fields.arrLimit,
				strLimit:   tt.fields.strLimit,
				deepLimit:  tt.fields.deepLimit,
				wholeLimit: tt.fields.wholeLimit,
				ignores:    tt.fields.ignores,
			}
			if got := ot.trimSlice(tt.args.v, tt.args.deepLmt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trimSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

type myinterface interface {
	SomeMethod()
}

type mystruct struct {
	v int
}

func (m *mystruct) SomeMethod() {
	fmt.Println("some method")
}

func TestOutputTrimmer_TrimObject(t *testing.T) {
	type fields struct {
		arrLimit   int
		strLimit   int
		deepLimit  int
		wholeLimit int
		ignores    map[string]bool
	}
	type args struct {
		obj any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet any
	}{
		{
			name: "when obj is nil then return nil",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: nil,
			},
			wantRet: nil,
		},
		{
			name: "when obj is int then return int64",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: 123,
			},
			wantRet: int64(123),
		},
		{
			name: "when obj is int ptr then return int64",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: intPtr(123),
			},
			wantRet: int64(123),
		},
		{
			name: "when obj is string then return string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: "abc",
			},
			wantRet: "abc",
		},
		{
			name: "when obj is string ptr then return string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() *string {
					s := "abc"
					return &s
				}(),
			},
			wantRet: "abc",
		},
		{
			name: "when obj is string larger than limit then return trimmed string",
			fields: fields{
				arrLimit:   3,
				strLimit:   3,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: "abcdef",
			},
			wantRet: "abc...",
		},
		{
			name: "when obj is string ptr larger than limit then return trimmed string",
			fields: fields{
				arrLimit:   3,
				strLimit:   3,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() *string {
					s := "abcdef"
					return &s
				}(),
			},
			wantRet: "abc...",
		},
		{
			name: "when obj is struct then return map[string]any",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					type user struct {
						Name       string        `json:"name"`
						Age        int           `json:"age"`
						GoSchoolAt time.Time     `json:"go_school_at"`
						DailyView  time.Duration `json:"daily_view"`
					}
					return user{
						Name:       "Alice",
						Age:        18,
						GoSchoolAt: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
						DailyView:  40 * time.Minute,
					}
				}(),
			},
			wantRet: map[string]any{
				"name":         "Alice",
				"age":          int64(18),
				"go_school_at": "2001-01-01T00:00:00.000",
				"daily_view":   "40m0s",
			},
		},
		{
			name: "when obj is struct larger than deep limit then return trimmed map[string]any",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  3,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					type deep struct {
						Flag string `json:"flag"`
					}
					type ext struct {
						Status string `json:"status"`
						Deep   deep   `json:"deep"`
					}
					type user struct {
						Name string `json:"name"`
						Age  int    `json:"age"`
						Ext  ext    `json:"ext"`
					}
					return user{
						Name: "Alice",
						Age:  18,
						Ext: ext{
							Status: "OK",
							Deep: deep{
								Flag: "flag",
							},
						},
					}
				}(),
			},
			wantRet: map[string]any{
				"name": "Alice",
				"age":  int64(18),
				"ext": map[string]any{
					"status": "OK",
				},
			},
		},
		{
			name: "when obj is struct ptr then return map[string]any",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					type user struct {
						Name string `json:"name"`
						Age  int    `json:"age"`
					}
					return &user{
						Name: "Alice",
						Age:  18,
					}
				}(),
			},
			wantRet: map[string]any{
				"name": "Alice",
				"age":  int64(18),
			},
		},
		{
			name: "when obj is slice then return []any",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					return []string{
						"Alice",
						"Bob",
						"Charlie",
					}
				}(),
			},
			wantRet: []any{
				"Alice",
				"Bob",
				"Charlie",
			},
		},
		{
			name: "when obj is bytes then return base64 bytes string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					return []byte{1, 2, 3, 4}
				}(),
			},
			wantRet: "AQIDBA==",
		},
		{
			name: "when obj is bytes less than limit then return base64 bytes string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					return bytes.Repeat([]byte{1, 2, 3, 4}, 20)
				}(),
			},
			wantRet: `AQIDBAECAwQBAgMEAQIDBAECAwQBAgMEAQIDBAECAwQBAgMEAQIDBAECAwQBAgMEAQIDBAECAwQBAgMEAQIDBAECAwQBAgMEAQIDBAECAwQ=`,
		},
		{
			name: "when obj is bytes larger than limit then return trimmed bytes",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					return bytes.Repeat([]byte{1, 2, 3, 4}, 100)
				}(),
			},
			wantRet: []any{
				uint64(1),
				uint64(2),
				uint64(3),
			},
		},
		{
			name: "when obj is array then return []any",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					return [3]string{
						"Alice",
						"Bob",
						"Charlie",
					}
				}(),
			},
			wantRet: []any{
				"Alice",
				"Bob",
				"Charlie",
			},
		},
		{
			name: "when obj is slice ptr then return []any",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					return &[]string{
						"Alice",
						"Bob",
						"Charlie",
					}
				}(),
			},
			wantRet: []any{
				"Alice",
				"Bob",
				"Charlie",
			},
		},
		{
			name: "when obj is slice larger than limit then return trimmed []any",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					return []string{
						"Alice",
						"Bob",
						"Charlie",
						"Danny",
					}
				}(),
			},
			wantRet: []any{
				"Alice",
				"Bob",
				"Charlie",
			},
		},
		{
			name: "when obj is map then return map[string]any",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					return map[string]string{
						"name": "Alice",
						"age":  "18",
					}
				}(),
			},
			wantRet: map[string]any{
				"name": "Alice",
				"age":  "18",
			},
		},
		{
			name: "when obj is error then return error string",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					return fmt.Errorf("oops")
				}(),
			},
			wantRet: "oops",
		},
		{
			name: "when obj is chan then return nil",
			fields: fields{
				arrLimit:   3,
				strLimit:   128,
				deepLimit:  10,
				wholeLimit: 1000,
				ignores:    make(map[string]bool),
			},
			args: args{
				obj: func() any {
					return make(chan int)
				}(),
			},
			wantRet: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ot := &OutputTrimmer{
				arrLimit:   tt.fields.arrLimit,
				strLimit:   tt.fields.strLimit,
				deepLimit:  tt.fields.deepLimit,
				wholeLimit: tt.fields.wholeLimit,
				ignores:    tt.fields.ignores,
			}
			if gotRet := ot.TrimObject(tt.args.obj); !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("TrimObject() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}
