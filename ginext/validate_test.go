package ginext

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext/errcode"
)

func Test_warpError(t *testing.T) {
	type args struct {
		c   *gin.Context
		err error
	}
	tests := []struct {
		name   string
		args   args
		assert func(t *testing.T, err error)
	}{
		{
			name: "when error is validation error",
			args: args{
				c: &gin.Context{},
				err: &ValidateError{
					Key:     "test",
					Message: "oops",
				},
			},
			assert: func(t *testing.T, err error) {
				var codeErr *errcode.Error
				if !errors.As(err, &codeErr) {
					t.Errorf("error is not errcode.Error")
					return
				}

				if codeErr.Code != 400 {
					t.Errorf("error code is not 400")
					return
				}

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := warpError(tt.args.c, tt.args.err)
			t.Log(err)
			tt.assert(t, err)
		})
	}
}
