package ginterceptor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	syslog "log"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/logger"
)

func (i *interceptor) ApplyTraffic() gin.HandlerFunc {
	if !i.config.EnableTraffic {
		return func(context *gin.Context) {
			context.Next()
		}
	}
	syslog.Println("[gin-interceptor] apply traffic logging")

	return func(c *gin.Context) {
		var (
			ctx     = c.Request.Context()
			reqCopy = captureRequest(c)
		)

		rec := logger.StartTrafficRec(ctx, &logger.ReqEntity{
			Typ: logger.TrafficTypRecv,
			Cmd: c.Request.URL.Path,
			Req: reqCopy,
			Fields: logger.Fields{
				"method":        c.Request.Method,
				"client":        c.ClientIP(),
				"query":         c.Request.URL.Query(),
				"req_header":    c.Request.Header,
				"req_body_size": c.Request.ContentLength,
			},
		})

		// hijack response writer
		rw := &responseWrapper{
			ResponseWriter: c.Writer,
			buffer:         bytes.NewBuffer(make([]byte, 0, c.Writer.Size())),
		}
		c.Writer = rw

		defer func() {
			c.Writer = rw.ResponseWriter

			rec.End(&logger.RespEntity{
				Code: fmt.Sprintf("%d", c.Writer.Status()),
				Msg:  "Success",
				Resp: captureResponse(c, rw.buffer.Bytes()),
			}, logger.Fields{
				"resp_header":    c.Writer.Header(),
				"resp_body_size": c.Writer.Size(),
			})
		}()

		c.Next()
	}
}

// capture http body from gin context request
// input is gin.Context
// output is any
// when context-type is application/json, return map[string]any
// when context-type is application/x-www-form-urlencoded, return map[string]string
// the other case, return nil
func captureRequest(c *gin.Context) (res any) {
	var (
		body        []byte
		err         error
		contentType = strings.ToLower(c.ContentType())
		ctx         = c.Request.Context()
	)

	le := logger.FromContext(ctx).WithFields(logger.Fields{
		"Content-Type": contentType,
	})

	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		return c.Request.PostForm
	}

	if strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "text/xml") {
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			le.WithError(err).Warn("error reading request body")
			return nil
		}

		// clone body for reset body
		bs := bytes.Clone(body)
		defer func() {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bs))
		}()
	} else {
		le.Debug("unsupported dump content-type")
		return "<unsupported content-type>"
	}

	if len(body) == 0 {
		le.Debug("request body is empty")
		return nil
	}

	if strings.HasPrefix(contentType, "application/json") {
		var req map[string]any
		if err = json.Unmarshal(body, &req); err != nil {
			le.WithError(err).Warnf("json unmarshal request failed")
			return "<json unmarshal failed>"
		}

		return req
	}

	// return string for other content-type
	return string(body)
}

// captureResponse capture response from gin context writer
// input is gin.Context
// output is any
// when response writer context-type is application/json, return map[string]any
func captureResponse(c *gin.Context, bs []byte) (res any) {
	var (
		err         error
		contentType string
		ctx         = c.Request.Context()
		le          = logger.FromContext(ctx)
	)

	if len(bs) == 0 {
		le.Debug("response body is empty")
		return nil
	}

	if c.Writer == nil {
		le.Debug("response writer is nil")
		return "<nil writer>"
	}

	contentType = strings.ToLower(c.Writer.Header().Get("Content-Type"))
	le = le.WithFields(logger.Fields{
		"Content-Type": contentType,
	})

	if strings.HasPrefix(contentType, "application/json") {
		var resp map[string]any
		if err = json.Unmarshal(bs, &resp); err != nil {
			le.WithError(err).Warn("json unmarshal response failed")
			return nil
		}
		return resp
	} else if strings.HasPrefix(contentType, "text/plain") ||
		strings.HasPrefix(contentType, "text/xml") {
		return string(bs)
	} else {
		return "<unsupported capture content-type>"
	}
}

type responseWrapper struct {
	gin.ResponseWriter
	buffer *bytes.Buffer
}

func (rw *responseWrapper) Write(data []byte) (int, error) {
	// Capture the response body
	written, err := rw.ResponseWriter.Write(data)
	rw.buffer.Write(data)
	return written, err
}
