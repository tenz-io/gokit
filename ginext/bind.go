package ginext

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/annotation"
	function "github.com/tenz-io/gokit/functional"
)

// BindAndValidate binds the passed struct pointer using the specified binding engine.
// e.g: /path/:id -> struct { ID int64 `bind:"uri,name=id"` }
// e.g: /path?offset=1 -> struct { Offset int `bind:"query,name=offset" validate:"gt=0"` }
// e.g. body json -> struct { Name string `json:"name"` }
// e.g. body form -> struct { Name string `bind:"form,name=username"` }
// e.g. body multipart -> struct { File []byte `bind:"file,name=file"` }
func BindAndValidate(c *gin.Context, ptr any) (err error) {
	defer func() {
		if err != nil {
			err = warpError(c, err)
		}
	}()

	err = annotation.ParseDefault(ptr)
	if err != nil {
		return fmt.Errorf("error parsing default value: %s", err.Error())
	}

	if has, err := tryBindUri(c, ptr); has && err != nil {
		return annotation.NewValidationError("uri", err.Error())
	}

	if has, err := tryBindQuery(c, ptr); has && err != nil {
		return annotation.NewValidationError("query", err.Error())
	}

	if has, err := tryBindHeader(c, ptr); has && err != nil {
		return annotation.NewValidationError("header", err.Error())
	}
	if has, err := tryBindMultipart(c, ptr); has && err != nil {
		return annotation.NewValidationError("multipart", err.Error())
	}

	if has, err := tryBindForm(c, ptr); has && err != nil {
		return annotation.NewValidationError("form", err.Error())
	}

	if has, err := tryBindJSON(c, ptr); has && err != nil {
		return annotation.NewValidationError("json", err.Error())
	}

	err = annotation.ValidateStruct(ptr)
	if err != nil {
		return err
	}

	return nil
}

// Validate the struct pointer using the specified binding engine.
func Validate(c *gin.Context, ptr any) (err error) {
	err = annotation.ValidateStruct(ptr)
	if err != nil {
		return warpError(c, err)
	}
	return nil

}

// tryBindUri tries to bind a uri request to a struct
// e.g: /path/:id -> struct { ID int64 `bind:"uri,name=id"` }
func tryBindUri(c *gin.Context, ptr any) (isUri bool, err error) {
	requestFields := annotation.GetRequestFields(ptr)
	if len(requestFields.Values()) == 0 {
		return false, nil
	}

	uriFields := function.Filter(requestFields.Values(), func(field annotation.RequestField) bool {
		return field.IsUri
	})

	if len(uriFields) == 0 {
		return false, nil
	}

	for _, field := range uriFields {
		value := c.Param(field.TagName)
		if value == "" {
			continue
		}
		if err := field.SetString(value); field.IsRequired && err != nil {
			return true, err
		}

	}

	return true, nil
}

// tryBindQuery tries to bind a query request to a struct
// e.g: /path?offset=1 -> struct { Offset int `bind:"query:name=offset"` }
func tryBindQuery(c *gin.Context, ptr any) (isQuery bool, err error) {
	requestFields := annotation.GetRequestFields(ptr)
	if len(requestFields.Values()) == 0 {
		return false, nil
	}

	queryFields := function.Filter(requestFields.Values(), func(field annotation.RequestField) bool {
		return field.IsQuery
	})

	if len(queryFields) == 0 {
		return false, nil
	}

	for _, field := range queryFields {
		value := c.Query(field.TagName)
		if value == "" {
			continue
		}
		if err := field.SetString(value); field.IsRequired && err != nil {
			return true, err
		}
	}
	return true, nil
}

// tryBindHeader tries to bind a header request to a struct
// e.g: header: Authorization: Bearer token -> struct { Authorization string `bind:"header,name=Authorization"` }
func tryBindHeader(c *gin.Context, ptr any) (isHeader bool, err error) {
	requestFields := annotation.GetRequestFields(ptr)
	if len(requestFields.Values()) == 0 {
		return false, nil
	}

	headerFields := function.Filter(requestFields.Values(), func(field annotation.RequestField) bool {
		return field.IsHeader
	})

	if len(headerFields) == 0 {
		return false, nil
	}

	for _, field := range headerFields {
		// get header value
		value := c.GetHeader(field.TagName)
		if value == "" {
			continue
		}
		if err := field.SetString(value); field.IsRequired && err != nil {
			return true, err
		}
	}
	return true, nil
}

// tryBindForm tries to bind a form request to a struct
// content-type: application/x-www-form-urlencoded
func tryBindForm(c *gin.Context, ptr any) (isForm bool, err error) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		// is not form
		return false, nil
	}

	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		return true, annotation.NewValidationError(
			"method",
			fmt.Sprintf("invalid method %s for form request, should be POST or PUT", c.Request.Method),
		)
	}

	// read form fields
	requestFields := annotation.GetRequestFields(ptr)
	if len(requestFields.Values()) == 0 {
		return false, nil
	}
	formFields := function.Filter(requestFields.Values(), func(field annotation.RequestField) bool {
		return field.IsForm
	})
	if len(formFields) == 0 {
		return false, nil
	}

	for _, field := range formFields {
		// ignore because the field maybe is optional
		if err := readFormAndSetField(c, &field); field.IsRequired && err != nil {
			return true, err
		}
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

	// if method is GET, HEAD, DELETE, ignore json body
	if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodDelete {
		return false, nil
	}

	// read request body into byte slice
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return true, annotation.NewValidationError(
			"body",
			fmt.Sprintf("error reading request body: %s", err.Error()),
		)
	}

	if err = json.Unmarshal(body, ptr); err != nil {
		return true, annotation.NewValidationError(
			"json_format",
			fmt.Sprintf("error unmarshalling request body: %s", err.Error()),
		)
	}

	return true, nil
}

// TryBindMultipart tries to bind a multipart request to a struct,
// content-type: multipart/form-data
func tryBindMultipart(c *gin.Context, ptr any) (isMultipart bool, err error) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
		// is not multipart form
		return false, nil
	}

	// if method is not POST or PUT, return error
	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		return true, annotation.NewValidationError(
			"method",
			fmt.Sprintf("invalid method %s for multipart request, should be POST or PUT", c.Request.Method),
		)
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
		return true, annotation.NewProtoError("multipart", "no file field found in struct")
	}

	// Parse the multipart form
	if err = c.Request.ParseMultipartForm(10 << 20); err != nil {
		return true, annotation.NewValidationError(
			"multipart",
			fmt.Sprintf("error parsing multipart form: %s", err.Error()),
		)
	}

	// read files
	for _, field := range fileFields {
		if err = readFileAndSetField(c, &field); field.IsRequired && err != nil {
			return true, err
		}
	}

	// read form fields
	for _, field := range formFields {
		if err = readFormAndSetField(c, &field); field.IsRequired && err != nil {
			return true, err
		}
	}

	return true, nil
}

func readFileAndSetField(c *gin.Context, field *annotation.RequestField) error {
	if err := (*field).Validate(); err != nil {
		return err
	}

	// Get the file from the form data
	file, _, err := c.Request.FormFile(field.TagName)
	if err != nil {
		return annotation.NewValidationError(
			field.TagName,
			fmt.Sprintf("error getting file: %s, err: %s", field.TagName, err.Error()),
		)
	}
	defer func() {
		_ = file.Close()
	}()

	// Read the file into a byte slice
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return annotation.NewValidationError(
			field.TagName,
			fmt.Sprintf("error reading file: %s, err: %s", field.TagName, err.Error()),
		)
	}

	if len(fileBytes) == 0 {
		return nil
	}

	err = field.Set(fileBytes)
	if err != nil {
		return err
	}
	return nil
}

func readFormAndSetField(c *gin.Context, field *annotation.RequestField) error {
	if err := (*field).Validate(); err != nil {
		return err
	}

	// Get the form value from the form data
	value := c.Request.FormValue(field.TagName)
	if value == "" {
		return nil
	}

	if err := field.SetString(value); err != nil {
		return err

	}
	return nil
}
