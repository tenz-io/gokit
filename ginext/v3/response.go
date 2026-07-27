package ginext

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext/v3/errcode"
)

// FileResponse 由想要以文件流而非 JSON 返回的响应数据实现。
// 当 [Response]/[ResponseStatus] 的 data 实现此接口时,响应层会用
// [http.DetectContentType] 探测内容类型并直接写入文件二进制。
type FileResponse interface {
	// GetFile 返回文件内容。
	GetFile() []byte
}

// ResponseFrame 是统一的成功/错误响应体结构。Code 是业务码(成功为 0),
// Message 是面向人的描述,Data 是业务负载。
type ResponseFrame struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// Response 以 HTTP 200 输出一个统一成功响应。data 为 nil 时写入空对象
// {}。当 data 实现 [FileResponse] 时,直接以文件流返回(此时 HTTP 状态
// 固定为 200,符合文件直传语义)。
func Response(c *gin.Context, data any) {
	writeResponse(c, http.StatusOK, data)
}

// ResponseStatus 以给定 HTTP status 输出一个统一成功响应。用于需要 201
// Created / 204 No Content 等非 200 成功状态时。data 为 nil 时写入空对象。
// 当 data 实现 [FileResponse] 时,直接以文件流返回(此时 HTTP 状态固定
// 为 200,符合文件直传语义,status 参数被忽略)。
func ResponseStatus(c *gin.Context, status int, data any) {
	writeResponse(c, status, data)
}

// writeResponse 是 [Response]/[ResponseStatus] 的共享实现。文件直传路径
// 固定 200;JSON 路径用 status 写 HTTP 状态、200 为 code、写入 ResponseFrame。
func writeResponse(c *gin.Context, status int, data any) {
	if data == nil {
		data = gin.H{}
	}

	// data 实现 FileResponse 时直接返回文件流。
	if f, ok := data.(FileResponse); ok {
		fileContent := f.GetFile()
		contentType := http.DetectContentType(fileContent)
		c.Data(http.StatusOK, contentType, fileContent)
		return
	}

	c.JSON(status, ResponseFrame{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// internalServerMessage 是非 errcode 错误对外暴露的固定消息。不把
// err.Error() 直接写进响应体,以免泄露 SQL/文件路径/下游地址等内部信息。
const internalServerMessage = "internal server error"

// ErrorResponse 输出一个统一错误响应。当 err(或其包装链)是
// [errcode.Error] 时,用其 Status 作为 HTTP 状态码、Code 作为业务码、
// Message 作为面向调用方的描述(errcode 是调用方**有意**暴露的业务码,
// 其 Message 视为可对外)。否则以 500 兜底,响应体只返回固定的
// "internal server error",原始 err 经 [gin.Context.Error] 进入 gin 错误链
// 供日志/中间件记录,但绝不回显给客户端。
//
// 可选 data 会被写入响应体的 Data 字段(nil 时为空对象)。
// 该函数会把 err 附加到 gin 错误链上并中止后续 handler(c.Abort)。
func ErrorResponse(c *gin.Context, err error, data ...any) {
	var d any
	if len(data) > 0 {
		d = data[0]
	}
	if d == nil {
		d = gin.H{}
	}

	// 始终把原始错误挂到 gin 错误链,供 ErrorLogger / 日志中间件取用。
	_ = c.Error(err)
	if e := new(errcode.Error); errors.As(err, &e) {
		c.JSON(e.Status, ResponseFrame{
			Code:    e.Code,
			Message: e.Message,
			Data:    d,
		})
		c.Abort()
		return
	}
	// 非 errcode:走通用 500,只回固定消息,原始错误不外泄。
	c.JSON(http.StatusInternalServerError, ResponseFrame{
		Code:    http.StatusInternalServerError,
		Message: internalServerMessage,
		Data:    d,
	})
	c.Abort()
}
