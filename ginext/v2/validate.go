package ginext

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/annotation"
	"github.com/tenz-io/gokit/ginext/errcode"
)

func warpError(_ *gin.Context, err error) error {
	if e := new(annotation.ValidationErrors); errors.As(err, &e) {
		if e.HasErrors() {
			return errcode.New(http.StatusBadRequest, e.Error())
		}
	}

	if e := new(annotation.ValidationError); errors.As(err, &e) {
		return errcode.New(http.StatusBadRequest, e.Error())
	}

	if e := new(json.UnmarshalTypeError); errors.As(err, &e) {
		return errcode.New(http.StatusBadRequest, e.Error())
	}

	return errcode.New(http.StatusBadRequest, fmt.Sprintf("invalid request: %s", err.Error()))
}
