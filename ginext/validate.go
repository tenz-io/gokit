package ginext

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext/errcode"
)

var _ error = (*ValidateError)(nil)

type FileRequest interface {
	GetFile() []byte
	GetFilename() string
}

type uploadInput struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type ValidateError struct {
	Key     string
	Message string
}

type ValidateErrors []*ValidateError

func (v *ValidateError) Error() string {
	return v.Message
}

func (v ValidateErrors) Error() string {
	return strings.Join(v.Errors(), ",")
}

func (v ValidateErrors) Errors() []string {
	errs := make([]string, len(v))
	for _, err := range v {
		errs = append(errs, err.Error())
	}
	return errs
}

// ShouldBind binds the passed struct pointer using the specified binding engine.
func ShouldBind(c *gin.Context, v any) error {
	if err := c.ShouldBind(v); err != nil {
		return warpError(c, err)
	}
	return nil
}

// ShouldBindUri binds the passed struct pointer using the specified binding engine.
func ShouldBindUri(c *gin.Context, v any) error {
	if err := c.ShouldBindUri(v); err != nil {
		return warpError(c, err)
	}
	return nil
}

// ShouldBindFile binds the passed struct pointer using the specified binding engine.
func ShouldBindFile(c *gin.Context, v any) error {
	// if content type is not multipart/form-data, return nil
	if !strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
		return nil
	}

	// if method is not POST or PUT, return error
	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		return warpError(c, fmt.Errorf("invalid method %s for file upload", c.Request.Method))
	}

	if v == nil {
		return errors.New("nil struct pointer passed to ShouldBindFile")
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return errors.New("must pass a pointer to a struct to ShouldBindFile")
	}

	structVal := val.Elem()

	// Bind the regular fields first if needed
	var upload uploadInput
	if err := c.ShouldBind(&upload); err != nil {
		return err
	}

	// Check for a File field and set it
	fileField := structVal.FieldByName("File")
	filenameField := structVal.FieldByName("Filename")
	if !fileField.IsValid() || fileField.Kind() != reflect.Slice || fileField.Type().Elem().Kind() != reflect.Uint8 {
		return errors.New("struct does not have a 'File []byte' field")
	}

	if !filenameField.IsValid() || filenameField.Kind() != reflect.String {
		return errors.New("struct does not have a 'Filename string' field")
	}

	// Open the file
	file, err := upload.File.Open()
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	// Read the file into a byte slice
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	fileField.SetBytes(fileBytes)
	filenameField.SetString(upload.File.Filename)

	return nil
}

func warpError(c *gin.Context, err error) error {
	if e := new(ValidateErrors); errors.As(err, &e) {
		return errcode.New(http.StatusBadRequest, e.Error())
	}

	if e := new(ValidateError); errors.As(err, &e) {
		return errcode.New(http.StatusBadRequest, e.Error())
	}

	if e := new(json.UnmarshalTypeError); errors.As(err, &e) {
		return errcode.New(http.StatusBadRequest, e.Error())
	}

	return errcode.New(http.StatusBadRequest, "invalid request")
}
