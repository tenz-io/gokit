package errcode

import (
	"errors"
	"fmt"
	"net/http"
)

var _ error = (*Error)(nil)

type Error struct {
	Code    int    `json:"code"` // custom error code
	Message string `json:"message"`
	Status  int    `json:"status"` // httpStatus
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: code = %d message = %s status = %d", e.Code, e.Message, e.Status)
}

func New(code int, msg string, status ...int) error {
	err := &Error{
		Code:    code,
		Message: msg,
		Status:  http.StatusOK,
	}
	if len(status) > 0 {
		err.Status = status[0]
	}
	return err
}

// BadRequest generates a 400 error.
func BadRequest(code int, message string) error {
	return New(code, message, http.StatusBadRequest)
}

func Unauthorized(code int, message string) error {
	return New(code, message, http.StatusUnauthorized)
}

func Forbidden(code int, message string) error {
	return New(code, message, http.StatusForbidden)
}

func NotFound(code int, message string) error {
	return New(code, message, http.StatusNotFound)
}

func MethodNotAllowed(code int, message string) error {
	return New(code, message, http.StatusMethodNotAllowed)
}

func Timeout(code int, message string) error {
	return New(code, message, http.StatusRequestTimeout)
}

func Conflict(code int, message string) error {
	return New(code, message, http.StatusConflict)
}

func TooManyRequests(code int, message string) error {
	return New(code, message, http.StatusTooManyRequests)
}

func InternalServer(code int, message string) error {
	return New(code, message, http.StatusInternalServerError)
}

func FromError(err error) (*Error, bool) {
	if e := new(Error); errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
