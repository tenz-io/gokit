package ginext

import (
	"bytes"
	"errors"
	"mime/multipart"
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
