package annotation

import (
	"reflect"
	"sort"
)

type Annotation string

const (
	// Query annotation used for binding query parameters.
	Query Annotation = "query"
	// URI annotation used for binding uri parameters.
	URI Annotation = "uri"
	// Header annotation used for binding header parameters.
	Header Annotation = "header"
	// Form annotation used for binding form parameters.
	Form Annotation = "form"
	// File annotation used for binding file parameters.
	File Annotation = "file"
	// Required annotation used for marking a field as required.
	Required Annotation = "required"
	// Default annotation used for marking a field as default.
	Default Annotation = "default"
	// Protobuf annotation used for marking a field as protobuf.
	Protobuf Annotation = "protobuf"
	// JSON annotation used for marking a field as json.
	JSON Annotation = "json"
)

// GetAnnotations returns the annotations of fields in a struct.
func GetAnnotations(field reflect.StructField) []Annotation {
	var (
		annotations = []Annotation{}
	)
	for _, tag := range []Annotation{Query, URI, Header, Form, File, Required, Default, Protobuf, JSON} {
		if _, ok := field.Tag.Lookup(string(tag)); ok {
			annotations = append(annotations, tag)
		}
	}

	sort.Slice(annotations, func(i, j int) bool {
		return annotations[i] < annotations[j]
	})

	return annotations
}
