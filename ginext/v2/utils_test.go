package ginext

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getStructFieldNames(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "normal",
			args: args{v: &struct {
				File     []byte `protobuf:"bytes,1,opt,name=file,proto3"`
				Filename string `protobuf:"bytes,2,opt,name=filename,proto3"`
				Label    string `protobuf:"bytes,3,opt,name=label,proto3"`
				status   string
			}{}},
			want:    []string{"File", "Filename", "Label"},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getStructFieldNames(tt.args.v)
			t.Logf("got: %v, err: %v", got, err)
			if !tt.wantErr(t, err, fmt.Sprintf("getStructFieldNames(%v)", tt.args.v)) {
				return
			}
			assert.Equalf(t, tt.want, got, "getStructFieldNames(%v)", tt.args.v)
		})
	}
}

func Test_getProtoFieldNames(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "normal",
			args: args{v: &struct {
				File     []byte `protobuf:"bytes,1,opt,name=file,proto3"`
				Filename string `protobuf:"bytes,2,opt,name=filename,proto3"`
				Label    string `protobuf:"label"`
				Offset   int    `json:"offset,omitempty"`
				Limit    int
				status   string
			}{}},
			want: map[string]string{
				"file":     "File",
				"filename": "Filename",
				"label":    "Label",
				"offset":   "Offset",
				"Limit":    "Limit",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFieldNames(tt.args.v)
			t.Logf("got: %v, err: %v", got, err)
			if !tt.wantErr(t, err, fmt.Sprintf("getProtoFieldNames(%v)", tt.args.v)) {
				return
			}
			assert.Equalf(t, tt.want, got, "getProtoFieldNames(%v)", tt.args.v)
		})
	}
}
