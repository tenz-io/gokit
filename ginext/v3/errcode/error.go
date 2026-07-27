// Package errcode 定义 ginext 使用的结构化错误码。
//
// Error 同时携带一个面向调用方的逻辑 code(业务错误码)与一个面向传输层的
// HTTP status,使响应层([Response]/[ErrorResponse])既能输出统一的
// {code,message,data} 响应体,又能把 errcode 映射到正确的 HTTP 状态码。
package errcode

import (
	"errors"
	"fmt"
	"net/http"
)

var _ error = (*Error)(nil)

// Error 是一个结构化错误:Code 是业务错误码,Message 是面向人的描述,
// Status 是对应的 HTTP 状态码。它实现 error,并可作为 errors.As 的目标
// 让下游([ErrorResponse])取出 Status 来渲染响应。
type Error struct {
	Code    int    `json:"code"`    // 业务错误码
	Message string `json:"message"` // 面向人的错误描述
	Status  int    `json:"status"`  // HTTP 状态码
}

func (e *Error) Error() string {
	return fmt.Sprintf("error: code = %d message = %s status = %d", e.Code, e.Message, e.Status)
}

// New 构造一个 Error。status 缺省时取 http.StatusOK;调用方通常会传入一个
// 明确的 HTTP 状态码(见 BadRequest/Unauthorized/...)。返回 error 以使
// 调用方无法再意外修改其字段。
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

// BadRequest 生成一个 400 错误。
func BadRequest(code int, message string) error {
	return New(code, message, http.StatusBadRequest)
}

// Unauthorized 生成一个 401 错误。
func Unauthorized(code int, message string) error {
	return New(code, message, http.StatusUnauthorized)
}

// Forbidden 生成一个 403 错误。
func Forbidden(code int, message string) error {
	return New(code, message, http.StatusForbidden)
}

// NotFound 生成一个 404 错误。
func NotFound(code int, message string) error {
	return New(code, message, http.StatusNotFound)
}

// MethodNotAllowed 生成一个 405 错误。
func MethodNotAllowed(code int, message string) error {
	return New(code, message, http.StatusMethodNotAllowed)
}

// Timeout 生成一个 408 错误。
func Timeout(code int, message string) error {
	return New(code, message, http.StatusRequestTimeout)
}

// Conflict 生成一个 409 错误。
func Conflict(code int, message string) error {
	return New(code, message, http.StatusConflict)
}

// TooManyRequests 生成一个 429 错误。
func TooManyRequests(code int, message string) error {
	return New(code, message, http.StatusTooManyRequests)
}

// InternalServer 生成一个 500 错误。
func InternalServer(code int, message string) error {
	return New(code, message, http.StatusInternalServerError)
}

// FromError 从 err 中抽取 *Error。当 err(或其包装链)是 *Error 时返回它
// 与 true;否则返回 nil 与 false。它是对 errors.As 的薄封装。
func FromError(err error) (*Error, bool) {
	if e := new(Error); errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
