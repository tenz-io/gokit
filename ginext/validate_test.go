package ginext

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tenz-io/gokit/ginext/errcode"
)

func Test_warpError(t *testing.T) {
	type args struct {
		c   *gin.Context
		err error
	}
	tests := []struct {
		name   string
		args   args
		assert func(t *testing.T, err error)
	}{
		{
			name: "when error is validation error",
			args: args{
				c: &gin.Context{},
				err: &ValidateError{
					Key:     "test",
					Message: "oops",
				},
			},
			assert: func(t *testing.T, err error) {
				var codeErr *errcode.Error
				if !errors.As(err, &codeErr) {
					t.Errorf("error is not errcode.Error")
					return
				}

				if codeErr.Code != 400 {
					t.Errorf("error code is not 400")
					return
				}

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := warpError(tt.args.c, tt.args.err)
			t.Log(err)
			tt.assert(t, err)
		})
	}
}

type TestResponseFrame[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type TestRequest struct {
	// @inject_tag: form:"title"
	Title string `json:"title,omitempty" form:"title" binding:"required"`
	// @inject_tag: form:"page"
	Page int32 `json:"page,omitempty" form:"page"`
	// @inject_tag: form:"page_size"
	PageSize int32 `json:"page_size,omitempty" form:"page_size"`
	// @inject_tag: form:"author_id" uri:"author_id"
	AuthorId int32 `json:"author_id,omitempty" uri:"author_id"`
}

func TestShouldBind_form(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/v1/:author_id/articles", func(c *gin.Context) {
		var in TestRequest
		if err := ShouldBind(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		if err := ShouldBindUri(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, &in)
	})

	body := []byte(`title=test&page=1&page_size=10&foo=bar`)
	req, _ := http.NewRequest("POST", "/v1/123/articles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	respContent := w.Body.String()
	t.Logf("response content: %s", respContent)

	status := w.Code
	if status != http.StatusOK {
		t.Errorf("status code is not 200")
		return
	}

	out := &TestResponseFrame[*TestRequest]{}
	err := json.Unmarshal([]byte(respContent), out)
	if err != nil {
		t.Errorf("failed to unmarshal response")
		return
	}

	expectedResp := &TestResponseFrame[*TestRequest]{
		Code:    0,
		Message: "success",
		Data: &TestRequest{
			Title:    "test",
			Page:     1,
			PageSize: 10,
			AuthorId: 123,
		},
	}

	if !reflect.DeepEqual(out, expectedResp) {
		t.Errorf("response is not expected, got: %+v, expected: %+v", out.Data, expectedResp)
		return
	}
}

func TestShouldBind_json(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/v1/:author_id/articles", func(c *gin.Context) {
		var in TestRequest
		if err := ShouldBind(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		if err := ShouldBindUri(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, &in)
	})

	body := []byte(`{"title":"test","page":1,"page_size":10}`)
	req, _ := http.NewRequest("POST", "/v1/123/articles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	respContent := w.Body.String()
	t.Logf("response content: %s", respContent)

	status := w.Code
	if status != http.StatusOK {
		t.Errorf("status code is not 200")
		return
	}

	out := &TestResponseFrame[*TestRequest]{}
	err := json.Unmarshal([]byte(respContent), out)
	if err != nil {
		t.Errorf("failed to unmarshal response")
		return
	}

	expectedResp := &TestResponseFrame[*TestRequest]{
		Code:    0,
		Message: "success",
		Data: &TestRequest{
			Title:    "test",
			Page:     1,
			PageSize: 10,
			AuthorId: 123,
		},
	}

	if !reflect.DeepEqual(out, expectedResp) {
		t.Errorf("response is not expected, got: %+v, expected: %+v", out.Data, expectedResp)
		return
	}
}

type TestFileRequest struct {
	// @inject_tag: form:"userid"
	UserId   int64  `json:"userid,omitempty" uri:"userid"`
	Username string `json:"username,omitempty" form:"username" binding:"required"`
	// @inject_tag: form:"page"
	File     []byte `json:"file,omitempty" form:"file" binding:"required"`
	Filename string `json:"filename,omitempty" form:"filename" binding:"required"`
}

func (tf *TestFileRequest) GetFile() []byte {
	return tf.File
}

func (tf *TestFileRequest) GetFilename() string {
	return tf.Filename
}

func TestShouldBind_file(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/v1/:userid/upload", func(c *gin.Context) {
		var in TestFileRequest
		if err := ShouldBind(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		if err := ShouldBindUri(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, &map[string]any{
			"userid":    in.UserId,
			"username":  in.Username,
			"filename":  in.Filename,
			"file_size": len(in.File),
		})
	})

	fileContents := []byte("test file content")
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	part.Write(fileContents)
	writer.Close()
	req, _ := http.NewRequest("POST", "/v1/123/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	respContent := w.Body.String()
	t.Logf("response content: %s", respContent)

}

func Test_tryBindMultipart(t *testing.T) {
	type args struct {
		c   *gin.Context
		ptr any
	}
	tests := []struct {
		name          string
		args          args
		wantMultipart bool
		wantErr       assert.ErrorAssertionFunc
	}{
		{
			name: "when content type is not multipart",
			args: args{
				c: &gin.Context{
					Request: &http.Request{
						Header: http.Header{
							"Content-Type": []string{"application/json"},
						},
					},
				},
				ptr: &TestFileRequest{},
			},
			wantMultipart: false,
			wantErr:       assert.NoError,
		},
		{
			name: "when method is not POST or PUT",
			args: args{
				c: &gin.Context{
					Request: &http.Request{
						Header: http.Header{
							"Content-Type": []string{"multipart/form-data"},
						},
						Method: http.MethodGet,
					},
				},
			},
			wantMultipart: true,
			wantErr:       assert.Error,
		},
		{
			name: "when field file not found",
			args: args{
				c: &gin.Context{
					Request: &http.Request{
						Header: http.Header{
							"Content-Type": []string{"multipart/form-data"},
						},
						Method: http.MethodPost,
					},
				},
				ptr: &TestFileRequest{},
			},
			wantMultipart: true,
			wantErr:       assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMultipart, err := tryBindMultipart(tt.args.c, tt.args.ptr)
			t.Logf("gotMultipart: %v, err: %v", gotMultipart, err)
			if !tt.wantErr(t, err, fmt.Sprintf("tryBindMultipart(%v, %v)", tt.args.c, tt.args.ptr)) {
				return
			}
			assert.Equalf(t, tt.wantMultipart, gotMultipart, "tryBindMultipart(%v, %v)", tt.args.c, tt.args.ptr)
		})
	}
}
