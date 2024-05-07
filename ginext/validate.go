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
	var validationsErr *ValidateErrors
	if errors.As(err, &validationsErr) {
		return errcode.New(400, validationsErr.Error())
	}

	var validationErr *ValidateError
	if errors.As(err, &validationErr) {
		return errcode.New(400, validationErr.Error())
	}

	var unmarshalErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalErr) {
		return errcode.New(400, err.Error())
	}

	return errcode.New(400, "invalid request")
}
