package ginext

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext/errcode"
)

type FileResponse interface {
	// GetFile returns the file content
	GetFile() []byte
}

type ResponseFrame struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Response(c *gin.Context, data any) {
	if data == nil {
		data = gin.H{}
	}

	// if data is a FileResponse, return file
	if f, ok := data.(FileResponse); ok {
		fileContent := f.GetFile()
		contentType := http.DetectContentType(fileContent)
		c.Data(http.StatusOK, contentType, fileContent)
		return
	}

	c.JSON(http.StatusOK, ResponseFrame{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, err error, data ...any) {
	var d any
	if len(data) > 0 {
		d = data[0]
	}
	if d == nil {
		d = gin.H{}
	}

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
	c.JSON(http.StatusInternalServerError, ResponseFrame{
		Code:    http.StatusInternalServerError,
		Message: err.Error(),
		Data:    d,
	})
	c.Abort()
}
