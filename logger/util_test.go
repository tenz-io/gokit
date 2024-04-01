package logger

import "testing"

func Test_isSliceOrArray(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "when v is a slice",
			args: args{v: []int{1, 2, 3}},
			want: true,
		},
		{
			name: "when v is an array",
			args: args{v: [3]int{1, 2, 3}},
			want: true,
		},
		{
			name: "when v is a nil slice",
			args: args{v: []int(nil)},
			want: true,
		},
		{
			name: "when v is a nil",
			args: args{v: nil},
			want: false,
		},
		{
			name: "when v is a map",
			args: args{v: map[string]int{"a": 1}},
			want: false,
		},
		{
			name: "when v is a string",
			args: args{v: "hello"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isArray(tt.args.v); got != tt.want {
				t.Errorf("isSliceOrArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lenIfArrayType(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name       string
		args       args
		wantLength int
		wantOk     bool
	}{
		{
			name:       "when v is a slice",
			args:       args{v: []int{1, 2, 3}},
			wantLength: 3,
			wantOk:     true,
		},
		{
			name:       "when v is an array",
			args:       args{v: [3]int{1, 2, 3}},
			wantLength: 3,
			wantOk:     true,
		},
		{
			name:       "when v is an empty array",
			args:       args{v: [0]int{}},
			wantLength: 0,
			wantOk:     true,
		},
		{
			name:       "when v is a nil slice",
			args:       args{v: []int(nil)},
			wantLength: 0,
			wantOk:     true,
		},
		{
			name:       "when v is a nil",
			args:       args{v: nil},
			wantLength: 0,
			wantOk:     false,
		},
		{
			name:       "when v is a map",
			args:       args{v: map[string]int{"a": 1}},
			wantLength: 0,
			wantOk:     false,
		},
		{
			name:       "when v is a string",
			args:       args{v: "hello"},
			wantLength: 0,
			wantOk:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLength, gotOk := lenIfArrayType(tt.args.v)
			if gotLength != tt.wantLength {
				t.Errorf("lenIfArrayType() gotLength = %v, want %v", gotLength, tt.wantLength)
			}
			if gotOk != tt.wantOk {
				t.Errorf("lenIfArrayType() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
