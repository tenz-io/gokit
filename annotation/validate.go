package annotation

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	_ error = (*ProtoError)(nil)
	_ error = (*ValidationError)(nil)
	_ error = (*ValidationErrors)(nil)
)

type ProtoError struct {
	Field   string
	Message string
}

func NewProtoError(field, message string) ProtoError {
	return ProtoError{
		Field:   field,
		Message: message,
	}
}

func (p ProtoError) Error() string {
	return fmt.Sprintf("%s: %s", p.Field, p.Message)
}

type ValidationError struct {
	Key     string
	Message string
}

func NewValidationError(key, message string) *ValidationError {
	return &ValidationError{
		Key:     key,
		Message: message,
	}
}

type ValidationErrors []*ValidationError

func NewValidationErrors(errs ...*ValidationError) ValidationErrors {
	return errs
}

func (v ValidationErrors) Append(err *ValidationError) ValidationErrors {
	return append(v, err)
}

// HasErrors checks if there are any validation errors.
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for i, err := range v {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Key, v.Message)
}

// ValidateStruct validates the struct fields based on their validate tag values.
func ValidateStruct(structPtr any) error {
	v := reflect.ValueOf(structPtr)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct")
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

// validateField applies the validation rules to the given field.
func validateField(field reflect.StructField, fieldVal reflect.Value) error {
	tag := field.Tag.Get(string(Validate))
	if tag == "" {
		return nil
	}

	rules := strings.Split(tag, ",")
	invalidErrors := NewValidationErrors()
	for _, rule := range rules {
		if err := applyRule(rule, field, fieldVal); err != nil {
			if vErr := new(ValidationError); errors.As(err, &vErr) {
				invalidErrors = invalidErrors.Append(vErr)
				continue
			}
			return err
		}
	}
	if invalidErrors.HasErrors() {
		return invalidErrors
	}
	return nil
}

// applyRule applies a single validation rule to a field.
func applyRule(rule string, field reflect.StructField, fieldVal reflect.Value) error {
	switch {
	case rule == "required":
		if isEmptyValue(fieldVal) {
			return NewValidationError(field.Name, "is required")
		}
	case strings.HasPrefix(rule, "lt="):
		val, err := strconv.ParseFloat(strings.TrimPrefix(rule, "lt="), 64)
		if err != nil {
			return err
		}
		if !isLessThan(fieldVal, val) {
			return NewValidationError(field.Name, fmt.Sprintf("must be less than %f", val))
		}
	case strings.HasPrefix(rule, "lte="):
		val, err := strconv.ParseFloat(strings.TrimPrefix(rule, "lte="), 64)
		if err != nil {
			return err
		}
		if !isLessThanOrEqual(fieldVal, val) {
			return NewValidationError(field.Name, fmt.Sprintf("must be less than or equal to %f", val))
		}
	case strings.HasPrefix(rule, "gt="):
		val, err := strconv.ParseFloat(strings.TrimPrefix(rule, "gt="), 64)
		if err != nil {
			return err
		}
		if !isGreaterThan(fieldVal, val) {
			return NewValidationError(field.Name, fmt.Sprintf("must be greater than %f", val))
		}
	case strings.HasPrefix(rule, "gte="):
		val, err := strconv.ParseFloat(strings.TrimPrefix(rule, "gte="), 64)
		if err != nil {
			return err
		}
		if !isGreaterThanOrEqual(fieldVal, val) {
			return NewValidationError(field.Name, fmt.Sprintf("must be greater than or equal to %f", val))
		}
	case strings.HasPrefix(rule, "len="):
		length, err := strconv.Atoi(strings.TrimPrefix(rule, "len="))
		if err != nil {
			return err
		}
		if !hasLength(fieldVal, length) {
			return NewValidationError(field.Name, fmt.Sprintf("must have length %d", length))
		}
	case strings.HasPrefix(rule, "min_len="):
		minLength, err := strconv.Atoi(strings.TrimPrefix(rule, "min_len="))
		if err != nil {
			return err
		}
		if !hasMinLength(fieldVal, minLength) {
			return NewValidationError(field.Name, fmt.Sprintf("must have minimum length %d", minLength))
		}
	case strings.HasPrefix(rule, "max_len="):
		maxLength, err := strconv.Atoi(strings.TrimPrefix(rule, "max_len="))
		if err != nil {
			return err
		}
		if !hasMaxLength(fieldVal, maxLength) {
			return NewValidationError(field.Name, fmt.Sprintf("must have maximum length %d", maxLength))
		}
	case rule == "non_blank":
		if isBlank(fieldVal) {
			return NewValidationError(field.Name, "must not be blank")
		}
	case strings.HasPrefix(rule, "pattern="):
		pattern := strings.TrimPrefix(rule, "pattern=")
		matched, msg := matchesPattern(fieldVal, pattern)
		if !matched {
			return NewValidationError(field.Name, "not match: "+msg)
		}
	default:
		return fmt.Errorf("unknown validation rule: %s", rule)
	}
	return nil
}

// isEmptyValue checks if a value is considered empty.
// only checks for the following types: string, array, slice, map, ptr, interface
func isEmptyValue(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return val.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return false
	default:
		// including struct, complex, chan, func, unsafe.Pointer
		return false
	}
}

func isLessThan(v reflect.Value, limit float64) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()) < limit
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()) < limit
	case reflect.Float32, reflect.Float64:
		return v.Float() < limit
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isLessThan(v.Index(i), limit) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func isLessThanOrEqual(v reflect.Value, limit float64) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()) <= limit
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()) <= limit
	case reflect.Float32, reflect.Float64:
		return v.Float() <= limit
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isLessThanOrEqual(v.Index(i), limit) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func isGreaterThan(v reflect.Value, limit float64) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()) > limit
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()) > limit
	case reflect.Float32, reflect.Float64:
		return v.Float() > limit
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isGreaterThan(v.Index(i), limit) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func isGreaterThanOrEqual(v reflect.Value, limit float64) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()) >= limit
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()) >= limit
	case reflect.Float32, reflect.Float64:
		return v.Float() >= limit
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isGreaterThanOrEqual(v.Index(i), limit) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func hasLength(v reflect.Value, length int) bool {
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		return v.Len() == length
	default:
		return false
	}
}

func hasMinLength(v reflect.Value, minLength int) bool {
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		return v.Len() >= minLength
	default:
		return false
	}
}

func hasMaxLength(v reflect.Value, maxLength int) bool {
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		return v.Len() <= maxLength
	default:
		return false
	}
}

func isBlank(v reflect.Value) bool {
	kind := v.Kind()
	switch {
	case kind == reflect.String:
		return strings.TrimSpace(v.String()) == ""
	case (kind == reflect.Slice) &&
		v.Type().Elem().Kind() == reflect.String:
		for i := 0; i < v.Len(); i++ {
			if isBlank(v.Index(i)) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func matchesPattern(v reflect.Value, pattern string) (bool, string) {
	if v.Kind() != reflect.String {
		return false, "is not a string"
	}

	return matchString(pattern, v.String())
}

type (
	predefinedPatternName = string
	predefinedPattern     = string
)

const (
	Email  predefinedPatternName = "#email"
	URL    predefinedPatternName = "#url"
	Abc    predefinedPatternName = "#abc"
	Digits predefinedPatternName = "#123"
	Abc123 predefinedPatternName = "#abc123"
	Hex    predefinedPatternName = "#hex"
	Base64 predefinedPatternName = "#base64"
)

const (
	emailPattern  predefinedPattern = "^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$"
	urlPattern    predefinedPattern = `^(http|https)://[a-zA-Z0-9-.]+.[a-zA-Z]{2,3}(/S*)?$`
	abcPattern    predefinedPattern = `^[a-zA-Z]+$`
	digitsPattern predefinedPattern = `^\d+$`
	abc123Pattern predefinedPattern = `^[a-zA-Z0-9]+$`
	hexPattern    predefinedPattern = `^[0-9a-fA-F]+$`
	base64Pattern predefinedPattern = `^[a-zA-Z0-9+/]*={0,2}$`
)

const (
	maxMatchString = 256
)

// getPredefinedPattern returns the predefined pattern based on the pattern name.
func getPredefinedPattern(name predefinedPatternName) (pattern predefinedPattern, existing bool) {
	if !strings.HasPrefix(name, "#") {
		return "", false
	}
	switch name {
	case Email:
		return emailPattern, true
	case URL:
		return urlPattern, true
	case Digits:
		return digitsPattern, true
	case Hex:
		return hexPattern, true
	case Base64:
		return base64Pattern, true
	case Abc:
		return abcPattern, true
	case Abc123:
		return abc123Pattern, true
	default:
		return "", false
	}
}

// matchString checks if a string matches a pattern.
// if s head with #, it will use predefined pattern.
// otherwise, it will use the pattern as a regular expression, which always starts with ^ and ends with $.
func matchString(pattern, s string) (matched bool, msg string) {
	predefined, existing := getPredefinedPattern(pattern)
	if existing {
		return matchRegexp(predefined, s)
	}

	return matchRegexp(pattern, s)
}

// matchRegexp checks if a string matches a regular expression pattern.
func matchRegexp(pattern string, s string) (matched bool, errMsg string) {
	if len(s) > maxMatchString {
		return false, "is too long"
	}

	// not start with ^ and end with $, just skip as not match
	if !strings.HasPrefix(pattern, "^") || !strings.HasSuffix(pattern, "$") {
		return false, fmt.Sprintf("invalid pattern: %s", pattern)
	}

	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false, fmt.Sprintf("pattern is not valid: %v", err)
	}

	if !matched {
		return false, fmt.Sprintf("not match pattern: %s", pattern)
	}

	return true, ""

}
