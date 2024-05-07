package ginext

import (
	"encoding/json"
	"errors"
	"net/http"
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
	if e := new(ValidateErrors); errors.As(err, &e) {
		return errcode.New(http.StatusBadRequest, e.Error())
	}

	if e := new(ValidateError); errors.As(err, &e) {
		return errcode.New(http.StatusBadRequest, e.Error())
	}

	if e := new(json.UnmarshalTypeError); errors.As(err, &e) {
		return errcode.New(http.StatusBadRequest, e.Error())
	}

	return errcode.New(http.StatusBadRequest, "invalid request")
}
