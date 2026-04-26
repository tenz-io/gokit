package annotation

import (
	"reflect"
	"testing"
)

func TestGetAnnotations(t *testing.T) {
	type args struct {
		field reflect.StructField
	}
	tests := []struct {
		name string
		args args
		want []Annotation
	}{
		{
			name: "when field has no annotations",
			args: args{
				field: reflect.StructField{},
			},
			want: []Annotation{},
		},
		{
			name: "when field has annotations",
			args: args{
				field: reflect.StructField{
					Tag: `json:"name" bind:"form,name=name" validate:"required"`,
				},
			},
			want: []Annotation{Bind, JSON, Validate},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAnnotations(tt.args.field); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}
