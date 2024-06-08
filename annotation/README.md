## bind

`bind` annotation used for binding request.
contains the categories: source, name
binding source including:

- uri: uri parameters
- query: query parameters
- header: header parameters
- form: form parameters
- file: file parameters

examples:

- uri -> struct field
  /get_article/:id -> struct { ID int `bind:"uri,name=id"` }
- query -> struct field
  /get_articles?page=1 -> struct { Page int32 `bind:"query,name=page"` }
- header -> struct field
  Authorization: Bearer token -> struct { Authorization string `bind:"header,name=Authorization
- form -> struct field
  form-data: username=abc -> struct { Username string `bind:"form,name=username"` }
  also support
  json: {"username": "abc"} -> struct { Username string `bind:"form,name=username"` }
  json can have nested struct
- file -> struct field
  form-data: file=abc.txt -> struct { File []byte `bind:"file,name=file"` }

## validate

`validate` annotation used for marking a field as validate.
contains validation rules.

- required: field is required
- lt: less than (int, float) or each element is less than ([]int, []float)
- lte: less than or equal to (int, float) or each element is less than or equal to ([]int, []float)
- gt: greater than (int, float) or each element is greater than ([]int, []float)
- gte: greater than or equal to (int, float) or each element is greater than or equal to ([]int, []float)
- len: exact length (string, []any)
- min_len: minimum length (string, []any)
- max_len: maximum length (string, []any)
- non_blank: not blank (string) or each element is not blank ([]string)
- pattern: regular expression pattern (string)

examples:

- struct { Age int `validate:"required,gt=0,lte=300"` }
- struct { BoundingBox []int `validate:"len=4"` }
- struct { Locations []string `validate:"required,min_len=1,max_len=4"` }
- struct { Names []string `validate:"non_blank"` }
- struct { Email string `validate:"required,pattern=^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$"` }

## default

`default` annotation used for setting the default value of a field.
examples:

- struct { Title string `default:"default title"` }

```go
package test

type TestRequest struct {
	// @inject_tag: bind:"uri,name=author_id" validate:"required,gt=0"
	AuthorId int32 `json:"author_id,omitempty" bind:"uri,name=author_id" validate:"required,gt=0"`

	// @inject_tag: bind:"query,name=page" default:"1" validate:"required,gt=0"
	Page int32 `json:"page,omitempty" bind:"query,name=page" default:"1" validate:"required,gt=0"`

	// @inject_tag: bind:"query,name=page_size" default:"10" validate:"required,gt=0,lte=100"
	PageSize int32 `json:"page_size,omitempty" bind:"query,name=page_size" default:"10" validate:"required,gt=0,lte=100"`

	// @inject_tag: bind:"header,name=X-Request-ID"
	RequestID string `json:"request_id,omitempty" bind:"header,name=request_id"`

	// @inject_tag: bind:"file,name=image" validate:"required"
	Image []byte `json:"image,omitempty" bind:"file,name=image" validate:"required,min_len=0,max_len=102400"`

	// @inject_tag: bind:"form,name=title" validate:"required,min_len=1,max_len=100"
	Title string `json:"title,omitempty" bind:"form,name=title" validate:"required,min_len=1,max_len=100"`
}

```