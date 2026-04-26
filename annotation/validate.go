package annotation

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var _ = []error{ProtoError{}, (*ValidationError)(nil), (ValidationErrors)(nil)}

type predefinedPatternName = string

const (
	Email  predefinedPatternName = "#email"
	URL    predefinedPatternName = "#url"
	Abc    predefinedPatternName = "#abc"
	Digits predefinedPatternName = "#123"
	Abc123 predefinedPatternName = "#abc123"
	Hex    predefinedPatternName = "#hex"
	Base64 predefinedPatternName = "#base64"
	Date   predefinedPatternName = "#date"
)

const maxMatchString = 256

// precompiled regex patterns for predefined pattern names.
var precompiledPatterns = map[predefinedPatternName]*regexp.Regexp{
	Email:  regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`),
	URL:    regexp.MustCompile(`^(http|https)://[a-zA-Z0-9-.]+.[a-zA-Z]{2,3}(/\S*)?$`),
	Abc:    regexp.MustCompile(`^[a-zA-Z]+$`),
	Digits: regexp.MustCompile(`^\d+$`),
	Abc123: regexp.MustCompile(`^[a-zA-Z0-9]+$`),
	Hex:    regexp.MustCompile(`^[0-9a-fA-F]+$`),
	Base64: regexp.MustCompile(`^[a-zA-Z0-9+/]*={0,2}$`),
	Date:   regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
}

// ProtoError represents a protocol-level error on a specific field.
type ProtoError struct {
	Field   string
	Message string
}

func NewProtoError(field, message string) ProtoError {
	return ProtoError{Field: field, Message: message}
}

func (p ProtoError) Error() string { return fmt.Sprintf("%s: %s", p.Field, p.Message) }

// ValidationError represents a single validation failure on a field.
type ValidationError struct {
	Key     string
	Message string
}

func NewValidationError(key, message string) *ValidationError {
	return &ValidationError{Key: key, Message: message}
}

func (v *ValidationError) Error() string { return fmt.Sprintf("%s: %s", v.Key, v.Message) }

// ValidationErrors is a collection of validation errors.
type ValidationErrors []*ValidationError

// HasErrors reports whether there are any errors accumulated.
func (v ValidationErrors) HasErrors() bool { return len(v) > 0 }

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for i, err := range v {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// ValidateStruct validates struct fields against their validate tag rules.
func ValidateStruct(structPtr any) error {
	v := reflect.ValueOf(structPtr)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("ValidateStruct: expected a pointer to a struct")
	}
	return validateStructValue(v.Elem())
}

func validateStructValue(v reflect.Value) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		if fieldVal.Kind() == reflect.Ptr && !fieldVal.IsNil() {
			fieldVal = fieldVal.Elem()
		}
		if fieldVal.Kind() == reflect.Struct {
			if err := validateStructValue(fieldVal); err != nil {
				return err
			}
		}
		if err := validateField(field, fieldVal); err != nil {
			return err
		}
	}
	return nil
}

func validateField(field reflect.StructField, fieldVal reflect.Value) error {
	tag := field.Tag.Get(string(Validate))
	if tag == "" {
		return nil
	}

	var errs ValidationErrors
	for _, rule := range strings.Split(tag, ",") {
		if err := applyRule(rule, field, fieldVal); err != nil {
			var vErr *ValidationError
			if errors.As(err, &vErr) {
				errs = append(errs, vErr)
				continue
			}
			return err
		}
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

func isRequired(tag reflect.StructTag) bool {
	for _, rule := range strings.Split(tag.Get(string(Validate)), ",") {
		if rule == "required" {
			return true
		}
	}
	return false
}

func applyRule(rule string, field reflect.StructField, fieldVal reflect.Value) error {
	switch {
	case rule == "required":
		if isEmptyValue(fieldVal) {
			return NewValidationError(field.Name, "is required")
		}
	case strings.HasPrefix(rule, "lt="):
		val, _ := strconv.ParseFloat(strings.TrimPrefix(rule, "lt="), 64)
		if !isLessThan(fieldVal, val) {
			return NewValidationError(field.Name, fmt.Sprintf("must be < %v", val))
		}
	case strings.HasPrefix(rule, "lte="):
		val, _ := strconv.ParseFloat(strings.TrimPrefix(rule, "lte="), 64)
		if !isLessThanOrEqual(fieldVal, val) {
			return NewValidationError(field.Name, fmt.Sprintf("must be <= %v", val))
		}
	case strings.HasPrefix(rule, "gt="):
		val, _ := strconv.ParseFloat(strings.TrimPrefix(rule, "gt="), 64)
		if !isGreaterThan(fieldVal, val) {
			return NewValidationError(field.Name, fmt.Sprintf("must be > %v", val))
		}
	case strings.HasPrefix(rule, "gte="):
		val, _ := strconv.ParseFloat(strings.TrimPrefix(rule, "gte="), 64)
		if !isGreaterThanOrEqual(fieldVal, val) {
			return NewValidationError(field.Name, fmt.Sprintf("must be >= %v", val))
		}
	case strings.HasPrefix(rule, "len="):
		n, _ := strconv.Atoi(strings.TrimPrefix(rule, "len="))
		if !hasLength(fieldVal, n) {
			return NewValidationError(field.Name, fmt.Sprintf("must have length %d", n))
		}
	case strings.HasPrefix(rule, "min_len="):
		n, _ := strconv.Atoi(strings.TrimPrefix(rule, "min_len="))
		if !hasMinLength(fieldVal, n) {
			return NewValidationError(field.Name, fmt.Sprintf("min length is %d", n))
		}
	case strings.HasPrefix(rule, "max_len="):
		n, _ := strconv.Atoi(strings.TrimPrefix(rule, "max_len="))
		if !hasMaxLength(fieldVal, n) {
			return NewValidationError(field.Name, fmt.Sprintf("max length is %d", n))
		}
	case rule == "non_blank":
		if isBlank(fieldVal) {
			return NewValidationError(field.Name, "must not be blank")
		}
	case strings.HasPrefix(rule, "pattern="):
		pattern := strings.TrimPrefix(rule, "pattern=")
		if matched, msg := matchesPattern(fieldVal, pattern); !matched {
			return NewValidationError(field.Name, msg)
		}
	default:
		return fmt.Errorf("unknown validation rule: %s", rule)
	}
	return nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

func isLessThan(v reflect.Value, limit float64) bool {
	return cmpNumeric(v, limit, func(a, b float64) bool { return a < b })
}
func isLessThanOrEqual(v reflect.Value, limit float64) bool {
	return cmpNumeric(v, limit, func(a, b float64) bool { return a <= b })
}
func isGreaterThan(v reflect.Value, limit float64) bool {
	return cmpNumeric(v, limit, func(a, b float64) bool { return a > b })
}
func isGreaterThanOrEqual(v reflect.Value, limit float64) bool {
	return cmpNumeric(v, limit, func(a, b float64) bool { return a >= b })
}

func cmpNumeric(v reflect.Value, limit float64, op func(float64, float64) bool) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return op(float64(v.Int()), limit)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return op(float64(v.Uint()), limit)
	case reflect.Float32, reflect.Float64:
		return op(v.Float(), limit)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !cmpNumeric(v.Index(i), limit, op) {
				return false
			}
		}
		return true
	}
	return false
}

func hasLength(v reflect.Value, n int) bool {
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		return v.Len() == n
	}
	return false
}

func hasMinLength(v reflect.Value, n int) bool {
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		return v.Len() >= n
	}
	return false
}

func hasMaxLength(v reflect.Value, n int) bool {
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		return v.Len() <= n
	}
	return false
}

func isBlank(v reflect.Value) bool {
	switch {
	case v.Kind() == reflect.String:
		return strings.TrimSpace(v.String()) == ""
	case v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.String:
		for i := 0; i < v.Len(); i++ {
			if isBlank(v.Index(i)) {
				return true
			}
		}
	}
	return false
}

func matchesPattern(v reflect.Value, pattern string) (bool, string) {
	if v.Kind() != reflect.String {
		return false, "not a string"
	}
	return matchPattern(pattern, v.String())
}

func matchPattern(pattern, s string) (bool, string) {
	if strings.HasPrefix(pattern, "#") {
		if re, ok := precompiledPatterns[pattern]; ok {
			if len(s) > maxMatchString {
				return false, "string too long"
			}
			if re.MatchString(s) {
				return true, ""
			}
			return false, fmt.Sprintf("does not match %s", pattern)
		}
		return false, fmt.Sprintf("unknown predefined pattern: %s", pattern)
	}

	if !strings.HasPrefix(pattern, "^") || !strings.HasSuffix(pattern, "$") {
		return false, "pattern must start with ^ and end with $"
	}

	if len(s) > maxMatchString {
		return false, "string too long"
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Sprintf("invalid pattern: %v", err)
	}
	if re.MatchString(s) {
		return true, ""
	}
	return false, fmt.Sprintf("does not match %s", pattern)
}
