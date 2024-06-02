package annotation

import (
	"reflect"
	"sort"
)

type Annotation string

const (
	// Query annotation used for binding query parameters.
	// e.g. /users?name=foo -> struct { Name string `query:"name"` }
	Query Annotation = "query"
	// URI annotation used for binding uri parameters.
	// e.g. /users/{id} -> struct { ID int `uri:"id"` }
	URI Annotation = "uri"
	// Header annotation used for binding header parameters.
	// e.g. Authorization -> struct { Authorization string `header:"Authorization"` }
	Header Annotation = "header"
	// Form annotation used for binding form parameters.
	// e.g. name=foo -> struct { Name string `form:"name"` }
	Form Annotation = "form"
	// File annotation used for binding file parameters.
	// e.g. file -> struct { File []byte `file:"file"` }
	File Annotation = "file"
	// Default annotation used for marking a field as default.
	// e.g. struct { Limit int `default:"10"` }
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

// GetAnnotations returns the annotations of fields in a struct.
func GetAnnotations(field reflect.StructField) []Annotation {
	var (
		annotations = []Annotation{}
	)
	for _, tag := range []Annotation{
		Query, URI, Header, Form, File, Default, Protobuf, JSON, YAML, Validate,
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
