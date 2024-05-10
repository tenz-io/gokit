package ginext

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

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

// Define a test struct that matches expected usage
type TestUpload struct {
	File     []byte `protobuf:"bytes,1,opt,name=file,proto3" json:"file,omitempty" form:"file"`
	Filename string `protobuf:"bytes,2,opt,name=filename,proto3" json:"filename,omitempty"`
}

func (t *TestUpload) GetFile() []byte {
	return t.File
}

func (t *TestUpload) GetFilename() string {
	return t.Filename
}

// TestShouldBindFile checks if the file binding works correctly
func TestShouldBindFile(t *testing.T) {
	// Set up a test file content
	fileContents := []byte("test file content")
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	part.Write(fileContents)
	writer.Close()

	// Create a test request and recorder
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// Create a new Gin context from the http request
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Create an instance of the struct that ShouldBindFile will populate
	var testUpload TestUpload

	// Call the function under test
	err = ShouldBindFile(c, &testUpload)
	assert.NoError(t, err, "ShouldBindFile should not return an error")

	// Check if the file was correctly bound
	assert.Equal(t, fileContents, testUpload.File, "The file contents should match the uploaded file")

	// Optionally, test other parts of the response or additional conditions
	assert.Equal(t, "test.txt", testUpload.Filename, "The filename should match the uploaded filename")
}

// Define a struct that matches your URI parameters
type TestUri struct {
	ID   string `uri:"id" binding:"required"`
	Name string `uri:"name" binding:"required"`
}

// TestShouldBindUri checks if URI binding works correctly
func TestShouldBindUri(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Setup route and params
	router := gin.Default()
	router.GET("/:id/:name", func(c *gin.Context) {
		var uri TestUri
		err := ShouldBindUri(c, &uri)
		// Here you can handle errors as per your application needs
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, uri)
	})

	// Create a request to pass to our handler.
	req, _ := http.NewRequest("GET", "/123/john", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check to ensure the response was what you expected
	assert.Equal(t, http.StatusOK, w.Code, "HTTP status should be 200")
	assert.Contains(t, w.Body.String(), "123", "Response should contain '123'")
	assert.Contains(t, w.Body.String(), "john", "Response should contain 'john'")
}

type TestRequest struct {
	// @inject_tag: form:"title"
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty" form:"title"`
	// @inject_tag: form:"page"
	Page int32 `protobuf:"varint,2,opt,name=page,proto3" json:"page,omitempty" form:"page"`
	// @inject_tag: form:"page_size" binding:"required"
	PageSize int32 `protobuf:"varint,3,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty" form:"page_size" binding:"required"`
	// @inject_tag: form:"author_id" uri:"author_id"
	AuthorId int32 `protobuf:"varint,4,opt,name=author_id,json=authorId,proto3" json:"author_id,omitempty" form:"author_id" uri:"author_id"`
}

func TestShouldBind_form(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Setup route and params
	router := gin.Default()
	router.POST("/v1/author/:author_id/articles", func(c *gin.Context) {
		var req TestRequest
		err := ShouldBind(c, &req)
		// Here you can handle errors as per your application needs
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	// Create a request to pass to our handler.
	body := bytes.NewBufferString("title=test&page=1&page_size=10&author_id=123")
	req, _ := http.NewRequest("POST", "/v1/author/123/articles", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Form = make(map[string][]string)
	req.Form.Add("title", "test")
	req.Form.Add("page", "1")
	req.Form.Add("page_size", "10")
	req.Form.Add("author_id", "123")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check to ensure the response was what you expected
	assert.Equal(t, http.StatusOK, w.Code, "HTTP status should be 200")
	t.Logf("response: %s", w.Body.String())

	assert.Contains(t, w.Body.String(), "test", "Response should contain 'test'")
	assert.Contains(t, w.Body.String(), "1", "Response should contain '1'")
	assert.Contains(t, w.Body.String(), "10", "Response should contain '10'")
	assert.Contains(t, w.Body.String(), "123", "Response should contain '123'")
}

func TestShouldBind_json(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Setup route and params
	router := gin.Default()
	router.POST("/v1/author/:author_id/articles", func(c *gin.Context) {
		var req TestRequest
		err := ShouldBind(c, &req)
		// Here you can handle errors as per your application needs
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	// Create a request to pass to our handler.
	body := bytes.NewBufferString(`{"title":"test","page":1,"page_size":10,"author_id":123}`)
	req, _ := http.NewRequest("POST", "/v1/author/123/articles", body)
	req.Header.Set("Content-Type", "application/json")
	req.Form = make(map[string][]string)
	req.Form.Add("title", "test")
	req.Form.Add("page", "1")
	req.Form.Add("page_size", "10")
	req.Form.Add("author_id", "123")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check to ensure the response was what you expected
	assert.Equal(t, http.StatusOK, w.Code, "HTTP status should be 200")
	t.Logf("response: %s", w.Body.String())

	assert.Contains(t, w.Body.String(), "test", "Response should contain 'test'")
	assert.Contains(t, w.Body.String(), "1", "Response should contain '1'")
	assert.Contains(t, w.Body.String(), "10", "Response should contain '10'")
	assert.Contains(t, w.Body.String(), "123", "Response should contain '123'")
}

type TestQuery struct {
	// @inject_tag: form:"title"
	Title string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty" form:"title"`
	// @inject_tag: form:"page"
	Page int32 `protobuf:"varint,2,opt,name=page,proto3" json:"page,omitempty" form:"page"`
	// @inject_tag: form:"page_size" binding:"required"
	PageSize int32 `protobuf:"varint,3,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty" form:"page_size" binding:"required"`
}

func TestShouldBindQuery(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Setup route and params
	router := gin.Default()
	router.GET("/v1/author/:author_id/articles", func(c *gin.Context) {
		var req TestQuery
		err := ShouldBindQuery(c, &req)
		// Here you can handle errors as per your application needs
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	// Create a request to pass to our handler.
	req, _ := http.NewRequest("GET", "/v1/author/123/articles?title=test&page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check to ensure the response was what you expected
	assert.Equal(t, http.StatusOK, w.Code, "HTTP status should be 200")
	t.Logf("response: %s", w.Body.String())

	assert.Contains(t, w.Body.String(), "test", "Response should contain 'test'")
	assert.Contains(t, w.Body.String(), "1", "Response should contain '1'")
	assert.Contains(t, w.Body.String(), "10", "Response should contain '10'")
}
