package ginext

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/annotation/v3"
	"github.com/tenz-io/gokit/ginext/v2/errcode"
)

// warpError translates bind/validation errors into an HTTP errcode.Error so the
// response layer can render a consistent 400 with the field details.
func warpError(_ *gin.Context, err error) error {
	if err == nil {
		return nil
	}
	// Collected validation failures: surface them all.
	if verrs, ok := annotation.AsErrors(err); ok && verrs.Has() {
		return errcode.New(http.StatusBadRequest, verrs.Error())
	}
	// Malformed JSON payload.
	var jsonErr *json.UnmarshalTypeError
	if errors.As(err, &jsonErr) {
		return errcode.New(http.StatusBadRequest, jsonErr.Error())
	}
	return errcode.New(http.StatusBadRequest, fmt.Sprintf("invalid request: %s", err.Error()))
}
