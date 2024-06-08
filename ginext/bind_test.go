package ginext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type TestResponseFrame[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type TestRequest struct {
	state int

	// @inject_tag: bind:"uri,name=userid" validate:"required,gt=0"
	UserID int64 `json:"userid,omitempty" bind:"uri,name=userid" validate:"required,gt=0"`

	// @inject_tag: bind:"query,name=page" default:"1" validate:"required,gt=0"
	Page int32 `json:"page,omitempty" bind:"query,name=page" default:"1" validate:"required,gt=0"`

	// @inject_tag: bind:"query,name=page_size" default:"10" validate:"required,gt=0,lte=100"
	PageSize int32 `json:"page_size,omitempty" bind:"query,name=page_size" default:"10" validate:"required,gt=0,lte=100"`

	// @inject_tag: bind:"header,name=X-Request-ID"
	RequestID string `json:"request_id" bind:"header,name=X-Request-ID"`

	// @inject_tag: bind:"form,name=title" validate:"required,min_len=1,max_len=100"
	Title string `json:"title,omitempty" bind:"form,name=title" validate:"required,min_len=1,max_len=100"`
}

type TestFileRequest struct {
	state int

	// @inject_tag: bind:"uri,name=userid" validate:"required,gt=0"
	UserID int64 `json:"userid,omitempty" bind:"uri,name=userid" validate:"required,gt=0"`

	// @inject_tag: bind:"query,name=region" default:"sg" validate:"required,len=2,pattern=#digits"
	Region string `json:"region,omitempty" bind:"query,name=region" default:"sg" validate:"required,len=2,pattern=#abc"`

	// @inject_tag: bind:"header,name=X-Request-ID"
	RequestID string `json:"request_id" bind:"header,name=X-Request-ID"`

	// @inject_tag: bind:"form,name=title" validate:"required,min_len=1,max_len=100"
	Title string `json:"title,omitempty" bind:"form,name=title" validate:"required,min_len=1,max_len=100"`

	// @inject_tag: bind:"file,name=txt_file" validate:"required,min_len=1,max_len=102400"
	TextFile []byte `json:"txt_file,omitempty" bind:"file,name=txt_file" validate:"required,min_len=1,max_len=102400"`
}

type TestFileResponseBody struct {
	UserID    int64  `json:"userid"`
	Region    string `json:"region"`
	RequestID string `json:"requestid"`
	Title     string `json:"title"`
	FileSize  int    `json:"txt_file__size"`
}

func Test_BindAndValidate_file(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/v1/:userid/upload", func(c *gin.Context) {
		var in TestFileRequest
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}

		Response(c, &TestFileResponseBody{
			UserID:    in.UserID,
			Region:    in.Region,
			RequestID: in.RequestID,
			Title:     in.Title,
			FileSize:  len(in.TextFile),
		})
	})

	fileContents := []byte("test file content")
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("txt_file", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	part.Write(fileContents)
	part, err = writer.CreateFormField("title")
	if err != nil {
		t.Fatal(err)
		return
	}
	part.Write([]byte("test title"))

	writer.Close()
	req, _ := http.NewRequest("POST", "/v1/123/upload?region=us", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Request-ID", "123456")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	respContent := w.Body.String()
	t.Logf("response content: %s", respContent)

	// assert response
	status := w.Code
	if status != http.StatusOK {
		t.Errorf("status code is not 200")
		return
	}

	out := &TestResponseFrame[*TestFileResponseBody]{}
	err = json.Unmarshal([]byte(respContent), out)
	if err != nil {
		t.Errorf("failed to unmarshal response")
		return
	}

	// deep equal
	expectedResp := &TestResponseFrame[*TestFileResponseBody]{
		Code:    0,
		Message: "success",
		Data: &TestFileResponseBody{
			UserID:    123,
			Region:    "us",
			RequestID: "123456",
			Title:     "test title",
			FileSize:  len(fileContents),
		},
	}

	// assert equal ignore order
	assert.Equalf(t, out, expectedResp, "response data is not expected")

}

func Test_BindAndValidate_form(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/v1/:userid/articles", func(c *gin.Context) {
		var in TestRequest
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}

		Response(c, &in)
	})

	body := []byte(`title=test`)
	req, _ := http.NewRequest("POST", "/v1/123/articles?page=5&page_size=20", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Request-ID", "123456")
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
			UserID:    123,
			Page:      5,
			PageSize:  20,
			RequestID: "123456",
			Title:     "test",
		},
	}

	assert.Equalf(t, out, expectedResp, "response data is not expected")
}

func Test_BindAndValidate_json(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/v1/:userid/articles", func(c *gin.Context) {
		var in TestRequest
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}

		Response(c, &in)
	})

	body := []byte(`{"title":"test"}`)
	req, _ := http.NewRequest("POST", "/v1/123/articles?page=2&page_size=30", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "123456")
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
			UserID:    123,
			RequestID: "123456",
			Page:      2,
			PageSize:  30,
			Title:     "test",
		},
	}

	assert.Equalf(t, out, expectedResp, "response data is not expected")
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
