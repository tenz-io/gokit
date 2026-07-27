package errcode

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	e := &Error{Code: 404, Message: "not found", Status: http.StatusNotFound}
	assert.Equal(t,
		"error: code = 404 message = not found status = 404",
		e.Error())
}

func TestNew(t *testing.T) {
	t.Run("default status is 200 when none given", func(t *testing.T) {
		e := New(0, "ok").(*Error)
		assert.Equal(t, http.StatusOK, e.Status)
		assert.Equal(t, 0, e.Code)
		assert.Equal(t, "ok", e.Message)
	})
	t.Run("explicit status overrides default", func(t *testing.T) {
		e := New(7, "boom", http.StatusBadGateway).(*Error)
		assert.Equal(t, http.StatusBadGateway, e.Status)
	})
	t.Run("returned value is a non-pointer error", func(t *testing.T) {
		err := New(1, "x")
		var target *Error
		assert.True(t, errors.As(err, &target))
	})
}

// TestConstructors 验证每个便捷构造函数都把它映射到正确的 HTTP 状态码,
// 且 Code/Message 透传无误。
func TestConstructors(t *testing.T) {
	cases := []struct {
		name string
		fn   func(int, string) error
		want int
	}{
		{"BadRequest", BadRequest, http.StatusBadRequest},
		{"Unauthorized", Unauthorized, http.StatusUnauthorized},
		{"Forbidden", Forbidden, http.StatusForbidden},
		{"NotFound", NotFound, http.StatusNotFound},
		{"MethodNotAllowed", MethodNotAllowed, http.StatusMethodNotAllowed},
		{"Timeout", Timeout, http.StatusRequestTimeout},
		{"Conflict", Conflict, http.StatusConflict},
		{"TooManyRequests", TooManyRequests, http.StatusTooManyRequests},
		{"InternalServer", InternalServer, http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn(tc.want, tc.name)
			e, ok := err.(*Error)
			assert.True(t, ok, "expected *errcode.Error")
			assert.Equal(t, tc.want, e.Code)
			assert.Equal(t, tc.name, e.Message)
			assert.Equal(t, tc.want, e.Status)
		})
	}
}

func TestFromError(t *testing.T) {
	t.Run("unwraps wrapped errcode", func(t *testing.T) {
		wrapped := errors.New("ctx")
		err := errors.Join(NotFound(http.StatusNotFound, "nf"), wrapped)
		e, ok := FromError(err)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNotFound, e.Status)
		assert.Equal(t, "nf", e.Message)
	})
	t.Run("non-errcode returns nil false", func(t *testing.T) {
		e, ok := FromError(errors.New("plain"))
		assert.False(t, ok)
		assert.Nil(t, e)
	})
	t.Run("nil error returns nil false", func(t *testing.T) {
		e, ok := FromError(nil)
		assert.False(t, ok)
		assert.Nil(t, e)
	})
}
