// Package annotation is a struct-tag-driven toolkit for Go structs:
// declarative defaults, a pluggable validator, and a cached field plan that
// HTTP layers (or any transport) can reuse for binding.
//
// One walk of a struct produces a cached Plan; Validate, ApplyDefaults and
// external binders all consume that same plan instead of re-reflecting.
package annotation

import (
	"fmt"
	"strings"
)

// FieldError is a single validation failure on a struct field.
type FieldError struct {
	// Field is the dotted path to the field, e.g. "Config.Addr.Street".
	Field string
	// Rule is the name of the rule that failed, e.g. "required", "gt".
	// Empty for ad-hoc errors such as bind failures.
	Rule string
	// Param is the raw parameter of the rule, e.g. "0", "^a-z+$".
	Param string
	// Msg is a human-readable explanation.
	Msg string
}

// Error implements error.
func (e FieldError) Error() string {
	switch {
	case e.Msg == "":
		return e.Field + ": invalid"
	case e.Rule == "":
		return e.Field + ": " + e.Msg
	default:
		return e.Field + ": " + e.Msg
	}
}

// ValidationErrors is the ordered collection of every failure found while
// validating a struct. Unlike returning on the first error, it accumulates
// all problems so callers can report them at once.
type ValidationErrors []FieldError

// Has reports whether any errors were collected.
func (v ValidationErrors) Has() bool { return len(v) > 0 }

// Error implements error.
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, e := range v {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(e.Error())
	}
	return sb.String()
}

// NewFieldError builds a FieldError.
func NewFieldError(field, rule, param, msg string) FieldError {
	return FieldError{Field: field, Rule: rule, Param: param, Msg: msg}
}

// Err is a convenience constructor returning a single-error ValidationErrors,
// useful for ad-hoc failures (e.g. a malformed request body) outside the rule
// engine. The rule may be empty.
func Err(field, rule, msg string) ValidationErrors {
	return ValidationErrors{{Field: field, Rule: rule, Msg: msg}}
}

// Errf is like Err with a formatted message.
func Errf(field, rule, format string, args ...any) ValidationErrors {
	return ValidationErrors{{Field: field, Rule: rule, Msg: fmt.Sprintf(format, args...)}}
}
