package annotation

import (
	"reflect"
	"sort"
)

type (
	Annotation  string
	BindingType string
)

const (
	// Bind annotation used for binding request.
	// contains the categories: source, name
	// binding source including:
	// - uri: uri parameters
	// - query: query parameters
	// - header: header parameters
	// - form: form parameters
	// - file: file parameters
	// e.g: uri -> struct { ID int `bind:"uri,name=id"` }
	// e.g: query -> struct { Title string `bind:"query,name=title"` }
	// e.g: header -> struct { Authorization string `bind:"header,name=Authorization"` }
	// e.g: form -> struct { Username string `bind:"form,name=username"` }
	// e.g: file -> struct { File []byte `bind:"file,name=file"` }
	Bind Annotation = "bind"
	// Default annotation used for setting the default value of a field.
	// e.g. struct { Title string `default:"default title"` }
	Default Annotation = "default"
	// Protobuf annotation used for marking a field as protobuf.
	// e.g. struct { Title string `protobuf:"bytes,1,opt,name=title,proto3" }
	Protobuf Annotation = "protobuf"
	// JSON annotation used for marking a field as json.
	// e.g. struct { Name string `json:"name,omitempty"` }
	JSON Annotation = "json"
	// YAML annotation used for marking a field as yaml.
	// e.g. struct { Name string `yaml:"name"` }
	YAML Annotation = "yaml"
	// Validate annotation used for marking a field as validate.
	// contains validation rules.
	// - required: field is required
	// - lt: less than (int, float) or each element is less than ([]int, []float)
	// - lte: less than or equal to (int, float) or each element is less than or equal to ([]int, []float)
	// - gt: greater than (int, float) or each element is greater than ([]int, []float)
	// - gte: greater than or equal to (int, float) or each element is greater than or equal to ([]int, []float)
	// - len: exact length (string, []any)
	// - min_len: minimum length (string, []any)
	// - max_len: maximum length (string, []any)
	// - non_blank: not blank (string) or each element is not blank ([]string)
	// - pattern: regular expression pattern (string)
	// e.g. struct { Age int `validate:"required,gt=0,lte=300"` }
	// e.g. struct { BoundingBox []int `validate:"len=4"` }
	// e.g. struct { Locations []string `validate:"required,min_len=1,max_len=4"` }
	// e.g. struct { Names []string `validate:"non_blank"` }
	// e.g. struct { Email string `validate:"required,pattern=^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$"` }
	Validate Annotation = "validate"
)

const (
	// URI binding type.
	URI BindingType = "uri"
	// Query binding type.
	Query BindingType = "query"
	// Header binding type.
	Header BindingType = "header"
	// Form binding type.
	Form BindingType = "form"
	// File binding type.
	File BindingType = "file"
	// None binding type.
	None BindingType = ""
)

// GetAnnotations returns the annotations of fields in a struct.
func GetAnnotations(field reflect.StructField) []Annotation {
	var (
		annotations = []Annotation{}
	)
	for _, tag := range []Annotation{
		Bind, Default, Protobuf, JSON, YAML, Validate,
	} {
		if _, ok := field.Tag.Lookup(string(tag)); ok {
			annotations = append(annotations, tag)
		}
	}

	sort.Slice(annotations, func(i, j int) bool {
		return annotations[i] < annotations[j]
	})

	return annotations
}
