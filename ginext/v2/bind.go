package ginext

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/annotation/v3"
)

// BindAndValidate binds a request into ptr and validates it.
//
// Binding sources are declared per field with the `bind` tag, e.g.
//
//	GET /path/:id?offset=1
//	type Req struct {
//	    ID     int64  `bind:"uri,name=id"   validate:"required,gt=0"`
//	    Offset int    `bind:"query,name=offset" default:"0" validate:"gte=0"`
//	    Auth   string `bind:"header,name=Authorization"`
//	    Title  string `bind:"form,name=title" validate:"required,min_len=1"`
//	    File   []byte `bind:"file,name=file" validate:"required"`
//	}
//
// For JSON bodies (`Content-Type: application/json`) the body is unmarshalled
// into ptr directly; the `json` tag controls field names.
//
// Defaults (`default` tag) are applied first, then each source is read and the
// field set, then Validate runs and collects every failure.
func BindAndValidate(c *gin.Context, ptr any) (err error) {
	defer func() {
		if err != nil {
			err = warpError(c, err)
		}
	}()

	if err = annotation.ApplyDefaults(ptr); err != nil {
		return fmt.Errorf("parse default value: %w", err)
	}

	if has, e := tryBindURI(c, ptr); has && e != nil {
		return annotation.Err("uri", "", e.Error())
	}
	if has, e := tryBindQuery(c, ptr); has && e != nil {
		return annotation.Err("query", "", e.Error())
	}
	if has, e := tryBindHeader(c, ptr); has && e != nil {
		return annotation.Err("header", "", e.Error())
	}
	if has, e := tryBindMultipart(c, ptr); has && e != nil {
		return annotation.Err("multipart", "", e.Error())
	}
	if has, e := tryBindForm(c, ptr); has && e != nil {
		return annotation.Err("form", "", e.Error())
	}
	if has, e := tryBindJSON(c, ptr); has && e != nil {
		return e
	}

	if err = annotation.Validate(ptr); err != nil {
		return err
	}
	return nil
}

// Validate runs annotation validation on ptr.
func Validate(c *gin.Context, ptr any) (err error) {
	defer func() {
		if err != nil {
			err = warpError(c, err)
		}
	}()
	if err = annotation.Validate(ptr); err != nil {
		return err
	}
	return nil
}

// rootValue resolves the struct value behind ptr for field-by-index access.
func rootValue(ptr any) (reflect.Value, error) {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return reflect.Value{}, fmt.Errorf("expected a non-nil pointer to a struct")
	}
	return rv.Elem(), nil
}

// fieldByIndex returns the settable value of f on root.
func fieldByIndex(root reflect.Value, f *annotation.Field) reflect.Value {
	return root.FieldByIndex(f.Index)
}

// tryBindURI binds route parameters, e.g. /path/:id -> ID int64 `bind:"uri,name=id"`.
func tryBindURI(c *gin.Context, ptr any) (has bool, err error) {
	plan, e := annotation.PlanFor(ptr)
	if e != nil {
		return false, nil
	}
	uriFields := plan.FieldsBySource(annotation.BindURI)
	if len(uriFields) == 0 {
		return false, nil
	}
	root, e := rootValue(ptr)
	if e != nil {
		return true, e
	}
	for _, f := range uriFields {
		val := c.Param(f.BindName)
		if val == "" {
			if f.IsRequired() {
				return true, annotation.Err(f.BindName, "required", "is required")
			}
			continue
		}
		if e := annotation.SetString(fieldByIndex(root, f), val); e != nil {
			return true, annotation.Err(f.BindName, "", e.Error())
		}
	}
	return true, nil
}

// tryBindQuery binds query parameters, e.g. ?offset=1 -> Offset `bind:"query,name=offset"`.
func tryBindQuery(c *gin.Context, ptr any) (has bool, err error) {
	plan, e := annotation.PlanFor(ptr)
	if e != nil {
		return false, nil
	}
	qFields := plan.FieldsBySource(annotation.BindQuery)
	if len(qFields) == 0 {
		return false, nil
	}
	root, e := rootValue(ptr)
	if e != nil {
		return true, e
	}
	for _, f := range qFields {
		val := c.Query(f.BindName)
		if val == "" {
			if f.IsRequired() {
				return true, annotation.Err(f.BindName, "required", "is required")
			}
			continue
		}
		if e := annotation.SetString(fieldByIndex(root, f), val); e != nil {
			return true, annotation.Err(f.BindName, "", e.Error())
		}
	}
	return true, nil
}

// tryBindHeader binds headers, e.g. Authorization: Bearer ... -> Auth `bind:"header,name=Authorization"`.
func tryBindHeader(c *gin.Context, ptr any) (has bool, err error) {
	plan, e := annotation.PlanFor(ptr)
	if e != nil {
		return false, nil
	}
	hFields := plan.FieldsBySource(annotation.BindHeader)
	if len(hFields) == 0 {
		return false, nil
	}
	root, e := rootValue(ptr)
	if e != nil {
		return true, e
	}
	for _, f := range hFields {
		val := c.GetHeader(f.BindName)
		if val == "" {
			if f.IsRequired() {
				return true, annotation.Err(f.BindName, "required", "is required")
			}
			continue
		}
		if e := annotation.SetString(fieldByIndex(root, f), val); e != nil {
			return true, annotation.Err(f.BindName, "", e.Error())
		}
	}
	return true, nil
}

// tryBindForm binds application/x-www-form-urlencoded bodies (POST/PUT only).
func tryBindForm(c *gin.Context, ptr any) (has bool, err error) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		return false, nil
	}
	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		return true, annotation.Err("method", "",
			fmt.Sprintf("invalid method %s for form request, should be POST or PUT", c.Request.Method))
	}

	plan, e := annotation.PlanFor(ptr)
	if e != nil {
		return false, nil
	}
	formFields := plan.FieldsBySource(annotation.BindForm)
	if len(formFields) == 0 {
		return false, nil
	}
	root, e := rootValue(ptr)
	if e != nil {
		return true, e
	}
	for _, f := range formFields {
		if e := readFormAndSetField(c, root, f); e != nil && f.IsRequired() {
			return true, e
		}
	}
	return true, nil
}

// tryBindJSON unmarshals a JSON body into ptr (POST/PUT/PATCH).
func tryBindJSON(c *gin.Context, ptr any) (has bool, err error) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		return false, nil
	}
	if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodDelete {
		return false, nil
	}
	body, e := io.ReadAll(c.Request.Body)
	if e != nil {
		return true, annotation.Err("body", "", fmt.Sprintf("error reading request body: %s", e.Error()))
	}
	if e = json.Unmarshal(body, ptr); e != nil {
		return true, annotation.Err("json_format", "", fmt.Sprintf("error unmarshalling request body: %s", e.Error()))
	}
	return true, nil
}

// tryBindMultipart binds a multipart/form-data body (files + form fields).
func tryBindMultipart(c *gin.Context, ptr any) (has bool, err error) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
		return false, nil
	}
	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		return true, annotation.Err("method", "",
			fmt.Sprintf("invalid method %s for multipart request, should be POST or PUT", c.Request.Method))
	}

	plan, e := annotation.PlanFor(ptr)
	if e != nil {
		return false, nil
	}
	fileFields := plan.FieldsBySource(annotation.BindFile)
	formFields := plan.FieldsBySource(annotation.BindForm)
	if len(fileFields) == 0 {
		return true, annotation.Err("multipart", "", "no file field found in struct")
	}

	if e = c.Request.ParseMultipartForm(10 << 20); e != nil {
		return true, annotation.Err("multipart", "", fmt.Sprintf("error parsing multipart form: %s", e.Error()))
	}

	root, e := rootValue(ptr)
	if e != nil {
		return true, e
	}

	for _, f := range fileFields {
		if e = readFileAndSetField(c, root, f); e != nil && f.IsRequired() {
			return true, e
		}
	}
	for _, f := range formFields {
		if e = readFormAndSetField(c, root, f); e != nil && f.IsRequired() {
			return true, e
		}
	}
	return true, nil
}

func readFileAndSetField(c *gin.Context, root reflect.Value, f *annotation.Field) error {
	file, _, err := c.Request.FormFile(f.BindName)
	if err != nil {
		return annotation.Err(f.BindName, "", fmt.Sprintf("error getting file %s: %s", f.BindName, err.Error()))
	}
	defer func() { _ = file.Close() }()

	bs, err := io.ReadAll(file)
	if err != nil {
		return annotation.Err(f.BindName, "", fmt.Sprintf("error reading file %s: %s", f.BindName, err.Error()))
	}
	if len(bs) == 0 {
		return nil
	}
	if err := annotation.Set(fieldByIndex(root, f), bs); err != nil {
		return annotation.Err(f.BindName, "", err.Error())
	}
	return nil
}

func readFormAndSetField(c *gin.Context, root reflect.Value, f *annotation.Field) error {
	val := c.Request.FormValue(f.BindName)
	if val == "" {
		return nil
	}
	if err := annotation.SetString(fieldByIndex(root, f), val); err != nil {
		return annotation.Err(f.BindName, "", err.Error())
	}
	return nil
}
