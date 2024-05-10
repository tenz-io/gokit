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

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": data})
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
		c.JSON(e.Status, gin.H{"code": e.Code, "message": e.Message, "data": data})
		c.Abort()
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "unknown", "data": gin.H{}})
	c.Abort()
}
