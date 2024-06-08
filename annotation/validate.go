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
	_ error = (*ValidationError)(nil)
	_ error = (*ValidationErrors)(nil)
)

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
		if !matchesPattern(fieldVal, pattern) {
			return NewValidationError(field.Name, fmt.Sprintf("must match pattern %s", pattern))
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

func matchesPattern(v reflect.Value, pattern string) bool {
	if v.Kind() != reflect.String {
		return false
	}

	return matchString(pattern, v.String())
}

type (
	predefinedPatternName = string
	predefinedPattern     = string
)

const (
	Email        predefinedPatternName = "#email"
	URL          predefinedPatternName = "#url"
	Digits       predefinedPatternName = "#digits"
	Hex          predefinedPatternName = "#hex"
	Base64       predefinedPatternName = "#base64"
	Alphabets    predefinedPatternName = "#alphabets"
	Alphanumeric predefinedPatternName = "#alphanumeric"
)

const (
	emailPattern        predefinedPattern = "^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$"
	urlPattern          predefinedPattern = `^(http|https)://[a-zA-Z0-9-.]+.[a-zA-Z]{2,3}(/S*)?$`
	alphabetsPattern    predefinedPattern = `^[a-zA-Z]+$`
	digitsPattern       predefinedPattern = `^\d+$`
	alphanumericPattern predefinedPattern = `^[a-zA-Z0-9]+$`
	hexPattern          predefinedPattern = `^[0-9a-fA-F]+$`
	base64Pattern       predefinedPattern = `^[a-zA-Z0-9+/]*={0,2}$`
)

// matchString checks if a string matches a pattern.
// if s head with #, it will use predefined pattern.
// otherwise, it will use the pattern as a regular expression, which always starts with ^ and ends with $.
func matchString(pattern, s string) bool {
	if strings.HasPrefix(pattern, "#") {
		return matchPredefinedPattern(pattern, s)
	}
	return matchRegexp(pattern, s)
}

// matchRegexp checks if a string matches a regular expression pattern.
func matchPredefinedPattern(pattern string, s string) bool {
	switch pattern {
	case Email:
		return matchRegexp(emailPattern, s)
	case URL:
		return matchRegexp(urlPattern, s)
	case Digits:
		return matchRegexp(digitsPattern, s)
	case Hex:
		return matchRegexp(hexPattern, s)
	case Base64:
		return matchRegexp(base64Pattern, s)
	case Alphabets:
		return matchRegexp(alphabetsPattern, s)
	case Alphanumeric:
		return matchRegexp(alphanumericPattern, s)
	default:
		return false
	}
}

// matchRegexp checks if a string matches a regular expression pattern.
func matchRegexp(pattern predefinedPattern, s string) bool {
	// not start with ^ and end with $, just skip as not match
	if !strings.HasPrefix(pattern, "^") || !strings.HasSuffix(pattern, "$") {
		return false
	}

	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}
	return matched

}
