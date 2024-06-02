package ginext

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/tenz-io/gokit/annotation"
	function "github.com/tenz-io/gokit/functional"
	"github.com/tenz-io/gokit/ginext/errcode"
)

var _ error = (*ValidateError)(nil)

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

// ShouldBindUri binds the passed struct pointer using the specified binding engine.
// e.g: /path/:id -> struct { ID string `uri:"id"` }
func ShouldBindUri(c *gin.Context, v any) error {
	if err := c.ShouldBindUri(v); err != nil {
		return warpError(c, err)
	}
	return nil
}

// ShouldBind binds the passed struct pointer using the specified binding engine.
// e.g. body json -> struct { Name string `json:"name"` }
// e.g. body form -> struct { Name string `form:"name"` }
// e.g. body multipart -> struct { File []byte `form:"file"` Filename string `form:"filename"` }
func ShouldBind(c *gin.Context, v any) error {
	isJson, err := tryBindJSON(c, v)
	if isJson {
		if err != nil {
			return warpError(c, err)
		}
		return nil
	}

	isForm, err := tryBindForm(c, v)
	if isForm {
		if err != nil {
			return warpError(c, err)
		}
		return nil
	}

	isMultipart, err := tryBindMultipart(c, v)
	if isMultipart {
		if err != nil {
			return warpError(c, err)
		}
		return nil
	}

	return c.ShouldBind(v)
}

// tryBindForm tries to bind a form request to a struct
// content-type: application/x-www-form-urlencoded
func tryBindForm(c *gin.Context, ptr any) (isForm bool, err error) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		// is not form
		return false, nil
	}

	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		return true, &ValidateError{
			Key:     "method",
			Message: fmt.Sprintf("invalid method %s for form request, should be POST or PUT", c.Request.Method),
		}
	}

	_, err = getPtrElem(ptr)
	if err != nil {
		return true, err
	}

	if err = c.ShouldBind(ptr); err != nil {
		return true, err
	}

	return true, nil

}

// tryBindJSON tries to bind a json request to a struct
// content-type: application/json
func tryBindJSON(c *gin.Context, ptr any) (isJson bool, err error) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		// is not json
		return false, nil
	}

	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		return true, &ValidateError{
			Key:     "method",
			Message: fmt.Sprintf("invalid method %s for json request, should be POST or PUT", c.Request.Method),
		}
	}

	_, err = getPtrElem(ptr)
	if err != nil {
		return true, err
	}

	if err = c.ShouldBind(ptr); err != nil {
		return true, err
	}

	return true, nil
}

// TryBindMultipart tries to bind a multipart request to a struct,
// content-type: multipart/form-data
//
// The struct should have two fields:
// 1) has a `File []byte` field;
// 2) and a `Filename string` field;
func tryBindMultipart(c *gin.Context, ptr any) (isMultipart bool, err error) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
		// is not multipart form
		return false, nil
	}

	// if method is not POST or PUT, return error
	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		return true, &ValidateError{
			Key:     "method",
			Message: fmt.Sprintf("invalid method %s for file upload, should be POST or PUT", c.Request.Method),
		}
	}

	requestFields := annotation.GetRequestFields(ptr)
	var (
		fileFields []annotation.RequestField
		formFields []annotation.RequestField
	)
	fileFields = function.Filter(requestFields.Values(), func(field annotation.RequestField) bool {
		return field.IsFile
	})
	formFields = function.Filter(requestFields.Values(), func(field annotation.RequestField) bool {
		return field.IsForm
	})

	if len(fileFields) == 0 {
		return true, fmt.Errorf("no file field found in struct")
	}

	// Parse the multipart form
	if err = c.Request.ParseMultipartForm(10 << 20); err != nil {
		return true, &ValidateError{
			Key:     "multipart",
			Message: fmt.Sprintf("error parsing multipart form: %s", err.Error()),
		}
	}

	// read files
	for _, field := range fileFields {
		if err := readAndSetFile(c, &field); err != nil {
			return true, err
		}
	}

	// read form fields
	for _, field := range formFields {
		if err := readAndSetForm(c, &field); err != nil {
			return true, err
		}
	}

	return true, nil
}

func readAndSetFile(c *gin.Context, field *annotation.RequestField) error {
	if err := (*field).Validate(); err != nil {
		return &ValidateError{
			Key:     field.TagName,
			Message: fmt.Sprintf("error validating file field: %s, err: %s", field.TagName, err.Error()),
		}
	}

	// Get the file from the form data
	file, _, err := c.Request.FormFile(field.TagName)
	if err != nil {
		return &ValidateError{
			Key:     field.TagName,
			Message: fmt.Sprintf("error getting file: %s, err: %s", field.TagName, err.Error()),
		}
	}
	defer file.Close()

	// Read the file into a byte slice
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return &ValidateError{
			Key:     field.TagName,
			Message: fmt.Sprintf("error reading file: %s, err: %s", field.TagName, err.Error()),
		}
	}

	err = field.Set(fileBytes)
	if err != nil {
		return &ValidateError{
			Key:     field.TagName,
			Message: fmt.Sprintf("error setting file: %s, err: %s", field.TagName, err.Error()),
		}
	}
	return nil
}

func readAndSetForm(c *gin.Context, field *annotation.RequestField) error {
	if err := (*field).Validate(); err != nil {
		return &ValidateError{
			Key:     field.TagName,
			Message: fmt.Sprintf("error validating file field: %s, err: %s", field.TagName, err.Error()),
		}
	}

	// Get the form value from the form data
	value := c.Request.FormValue(field.TagName)
	err := field.Set(value)
	if err != nil {
		return &ValidateError{
			Key:     field.TagName,
			Message: fmt.Sprintf("error setting form field: %s, err: %s", field.TagName, err.Error()),
		}

	}
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

	e := validator.ValidationErrors{}
	if errors.As(err, &e) {
		return errcode.New(http.StatusBadRequest, e.Error())
	}

	return errcode.New(http.StatusBadRequest, fmt.Sprintf("invalid request: %s", err.Error()))
}
