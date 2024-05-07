package ginext

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext/errcode"
)

var _ error = (*ValidateError)(nil)

type ValidateError struct {
	Key     string
	Message string
}

type ValidateErrors []*ValidateError

func (v *ValidateError) Error() string {
	return v.Message
}

func (v ValidateErrors) Error() string {
	return strings.Join(v.Errors(), ",")
}

func (v ValidateErrors) Errors() []string {
	errs := make([]string, len(v))
	for _, err := range v {
		errs = append(errs, err.Error())
	}
	return errs
}

func ShouldBind(c *gin.Context, v any) error {
	if err := c.ShouldBind(v); err != nil {
		return warpError(c, err)
	}
	return nil
}

func ShouldBindUri(c *gin.Context, v any) error {
	if err := c.ShouldBindUri(v); err != nil {
		return warpError(c, err)
	}
	return nil
}

func warpError(c *gin.Context, err error) error {
	var validationErr ValidateErrors
	switch {
	case errors.As(err, &validationErr):
		return errcode.New(400, validationErr.Error())
	case errors.Is(err, new(json.UnmarshalTypeError)):
		return errcode.New(400, err.Error())
	default:
		return errcode.New(400, "invalid request")
	}
}
