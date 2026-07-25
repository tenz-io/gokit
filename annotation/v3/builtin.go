package annotation

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})

// Register the built-in rules. Order does not matter; each is keyed by name.
func init() {
	Register("required", requiredValidator)
	Register("len", lenValidator)
	Register("min_len", minLenValidator)
	Register("max_len", maxLenValidator)
	Register("min", minValidator)
	Register("max", maxValidator)
	Register("gt", gtValidator)
	Register("lt", ltValidator)
	Register("gte", gteValidator)
	Register("lte", lteValidator)
	Register("eq", eqValidator)
	Register("ne", neValidator)
	Register("oneof", oneofValidator)
	Register("in", oneofValidator) // alias
	Register("non_blank", nonBlankValidator)
	Register("pattern", patternValidator)
	Register("email", namedPatternValidator("email"))
	Register("url", namedPatternValidator("url"))
	Register("uuid", namedPatternValidator("uuid"))
	Register("ipv4", namedPatternValidator("ipv4"))
	Register("ipv6", namedPatternValidator("ipv6"))
	Register("alpha", namedPatternValidator("alpha"))
	Register("alphanum", namedPatternValidator("alphanum"))
	Register("numeric", namedPatternValidator("numeric"))
	Register("hex", namedPatternValidator("hex"))
	Register("date", namedPatternValidator("date"))
	Register("base64", namedPatternValidator("base64"))
	Register("contains", containsValidator)
	Register("prefix", prefixValidator)
	Register("suffix", suffixValidator)
	Register("dive", diveValidator)
	Register("msg", nil) // modifier, not a rule; handled in compileRules
}

// --- required ---

func requiredValidator(_ string, _ reflect.StructField) (Rule, error) {
	return func(rv reflect.Value) (bool, string) {
		if isEmptyValue(rv) {
			return false, "is required"
		}
		return true, ""
	}, nil
}

// --- lengths ---

func lenValidator(param string, _ reflect.StructField) (Rule, error) {
	n, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf("len: invalid integer %q", param)
	}
	return func(rv reflect.Value) (bool, string) {
		if l, ok := lengthOf(rv); ok && l != n {
			return false, fmt.Sprintf("must have length %d", n)
		}
		return true, ""
	}, nil
}

func minLenValidator(param string, _ reflect.StructField) (Rule, error) {
	n, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf("min_len: invalid integer %q", param)
	}
	return func(rv reflect.Value) (bool, string) {
		if l, ok := lengthOf(rv); ok && l < n {
			return false, fmt.Sprintf("must have length >= %d", n)
		}
		return true, ""
	}, nil
}

func maxLenValidator(param string, _ reflect.StructField) (Rule, error) {
	n, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf("max_len: invalid integer %q", param)
	}
	return func(rv reflect.Value) (bool, string) {
		if l, ok := lengthOf(rv); ok && l > n {
			return false, fmt.Sprintf("must have length <= %d", n)
		}
		return true, ""
	}, nil
}

// --- numeric comparisons ---

func numericValidator(name, param string, op func(a, b float64) bool, sym string) (Rule, error) {
	limit, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid number %q", name, param)
	}
	return func(rv reflect.Value) (bool, string) {
		if isSliceOrArray(rv.Kind()) {
			// Compare every element.
			for i := 0; i < rv.Len(); i++ {
				if a, ok := numeric(rv.Index(i)); ok && !op(a, limit) {
					return false, fmt.Sprintf("element %d must be %s %v", i, sym, limit)
				}
			}
			return true, ""
		}
		a, ok := numeric(rv)
		if !ok {
			return true, "" // not a numeric field; skip
		}
		if !op(a, limit) {
			return false, fmt.Sprintf("must be %s %v", sym, limit)
		}
		return true, ""
	}, nil
}

func minValidator(param string, ft reflect.StructField) (Rule, error) {
	return numericValidator("min", param, func(a, b float64) bool { return a >= b }, ">=")
}
func maxValidator(param string, ft reflect.StructField) (Rule, error) {
	return numericValidator("max", param, func(a, b float64) bool { return a <= b }, "<=")
}
func gtValidator(param string, ft reflect.StructField) (Rule, error) {
	return numericValidator("gt", param, func(a, b float64) bool { return a > b }, ">")
}
func ltValidator(param string, ft reflect.StructField) (Rule, error) {
	return numericValidator("lt", param, func(a, b float64) bool { return a < b }, "<")
}
func gteValidator(param string, ft reflect.StructField) (Rule, error) {
	return numericValidator("gte", param, func(a, b float64) bool { return a >= b }, ">=")
}
func lteValidator(param string, ft reflect.StructField) (Rule, error) {
	return numericValidator("lte", param, func(a, b float64) bool { return a <= b }, "<=")
}

func eqValidator(param string, _ reflect.StructField) (Rule, error) {
	return func(rv reflect.Value) (bool, string) {
		if a, ok := numeric(rv); ok && a != mustFloat(param) {
			return false, fmt.Sprintf("must equal %v", param)
		}
		// For strings, compare directly.
		if rv.Kind() == reflect.String && rv.String() != param {
			return false, fmt.Sprintf("must equal %q", param)
		}
		return true, ""
	}, nil
}

func neValidator(param string, _ reflect.StructField) (Rule, error) {
	return func(rv reflect.Value) (bool, string) {
		if a, ok := numeric(rv); ok && a == mustFloat(param) {
			return false, fmt.Sprintf("must not equal %v", param)
		}
		if rv.Kind() == reflect.String && rv.String() == param {
			return false, fmt.Sprintf("must not equal %q", param)
		}
		return true, ""
	}, nil
}

// --- oneof ---

func oneofValidator(param string, _ reflect.StructField) (Rule, error) {
	options := strings.Split(param, " ")
	return func(rv reflect.Value) (bool, string) {
		cur := fmt.Sprintf("%v", rv.Interface())
		for _, o := range options {
			if o == cur {
				return true, ""
			}
		}
		return false, fmt.Sprintf("must be one of %s", strings.Join(options, " "))
	}, nil
}

// --- non_blank ---

func nonBlankValidator(_ string, _ reflect.StructField) (Rule, error) {
	return func(rv reflect.Value) (bool, string) {
		switch {
		case rv.Kind() == reflect.String:
			if strings.TrimSpace(rv.String()) == "" {
				return false, "must not be blank"
			}
		case isSliceOrArray(rv.Kind()) && rv.Type().Elem().Kind() == reflect.String:
			for i := 0; i < rv.Len(); i++ {
				if strings.TrimSpace(rv.Index(i).String()) == "" {
					return false, fmt.Sprintf("element %d must not be blank", i)
				}
			}
		}
		return true, ""
	}, nil
}

// --- patterns ---

func patternValidator(param string, _ reflect.StructField) (Rule, error) {
	re, err := compilePattern(param)
	if err != nil {
		return nil, fmt.Errorf("pattern: %v", err)
	}
	return func(rv reflect.Value) (bool, string) {
		if rv.Kind() != reflect.String {
			return true, "" // skip non-strings
		}
		if !re.MatchString(rv.String()) {
			return false, fmt.Sprintf("does not match %s", param)
		}
		return true, ""
	}, nil
}

func namedPatternValidator(name string) Validator {
	return func(_ string, _ reflect.StructField) (Rule, error) {
		re, err := compilePattern("#" + name)
		if err != nil {
			return nil, err
		}
		return func(rv reflect.Value) (bool, string) {
			if rv.Kind() != reflect.String {
				return true, ""
			}
			if !re.MatchString(rv.String()) {
				return false, fmt.Sprintf("must be a valid %s", name)
			}
			return true, ""
		}, nil
	}
}

// --- contains / prefix / suffix ---

func containsValidator(param string, _ reflect.StructField) (Rule, error) {
	return func(rv reflect.Value) (bool, string) {
		if rv.Kind() != reflect.String {
			return true, ""
		}
		if !strings.Contains(rv.String(), param) {
			return false, fmt.Sprintf("must contain %q", param)
		}
		return true, ""
	}, nil
}

func prefixValidator(param string, _ reflect.StructField) (Rule, error) {
	return func(rv reflect.Value) (bool, string) {
		if rv.Kind() != reflect.String {
			return true, ""
		}
		if !strings.HasPrefix(rv.String(), param) {
			return false, fmt.Sprintf("must start with %q", param)
		}
		return true, ""
	}, nil
}

func suffixValidator(param string, _ reflect.StructField) (Rule, error) {
	return func(rv reflect.Value) (bool, string) {
		if rv.Kind() != reflect.String {
			return true, ""
		}
		if !strings.HasSuffix(rv.String(), param) {
			return false, fmt.Sprintf("must end with %q", param)
		}
		return true, ""
	}, nil
}

// --- dive ---

// dive applies an element-level rule to every element of a slice/array/map.
// e.g. validate:"min_len=1,dive:non_blank".
func diveValidator(param string, ft reflect.StructField) (Rule, error) {
	v, ok := lookupValidator(param)
	if !ok {
		return nil, fmt.Errorf("dive: unknown rule %q", param)
	}
	elemRule, err := v("", ft)
	if err != nil {
		return nil, fmt.Errorf("dive: %v", err)
	}
	return func(rv reflect.Value) (bool, string) {
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < rv.Len(); i++ {
				if ok, msg := elemRule(rv.Index(i)); !ok {
					return false, fmt.Sprintf("element %d: %s", i, msg)
				}
			}
		case reflect.Map:
			for _, k := range rv.MapKeys() {
				if ok, msg := elemRule(rv.MapIndex(k)); !ok {
					return false, fmt.Sprintf("key %v: %s", k.Interface(), msg)
				}
			}
		}
		return true, ""
	}, nil
}

// --- helpers ---

// splitRule separates a rule token into its name and parameter. It accepts
// '=' as the canonical separator and ':' as a convenience (so "dive:non_blank"
// and "dive=non_blank" both work). A token with no separator is a bare rule.
func splitRule(raw string) (name, param string, ok bool) {
	if i := strings.Index(raw, "="); i >= 0 {
		return raw[:i], raw[i+1:], true
	}
	if i := strings.Index(raw, ":"); i >= 0 {
		return raw[:i], raw[i+1:], true
	}
	return raw, "", false
}

func compileRules(f *Field) error {
	tag := f.StructField.Tag.Get(tagValidate)
	if tag == "" {
		return nil
	}
	var pendingMsg string
	for _, raw := range strings.Split(tag, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		name, param, _ := splitRule(raw)

		if name == "msg" {
			// Applies to the previously compiled rule.
			pendingMsg = param
			if len(f.rules) > 0 {
				f.rules[len(f.rules)-1].msg = param
			}
			continue
		}

		v, ok := lookupValidator(name)
		if !ok {
			// Unknown rule becomes a permanent failure so the misconfiguration
			// is visible instead of silently passing.
			f.rules = append(f.rules, namedRule{
				name:  name,
				param: param,
				run:   func(reflect.Value) (bool, string) { return false, "unknown rule" },
			})
			pendingMsg = ""
			continue
		}
		check, err := v(param, f.StructField)
		if err != nil {
			return fmt.Errorf("field %s: rule %s: %w", f.Path, name, err)
		}
		r := namedRule{name: name, param: param, run: check, msg: pendingMsg}
		pendingMsg = ""
		f.rules = append(f.rules, r)
	}
	return nil
}

func isEmptyValue(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array, reflect.String, reflect.Chan:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr:
		return false // presence, not zero
	case reflect.Float32, reflect.Float64:
		return false
	case reflect.Struct:
		return rv.Type() == timeType && rv.Interface().(time.Time).IsZero()
	default:
		return !rv.IsValid()
	}
}

func lengthOf(rv reflect.Value) (int, bool) {
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return rv.Len(), true
	}
	return 0, false
}

func numeric(rv reflect.Value) (float64, bool) {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	}
	// Allow string-encoded durations to be compared as seconds.
	if rv.Kind() == reflect.String {
		if d, err := time.ParseDuration(rv.String()); err == nil {
			return d.Seconds(), true
		}
	}
	return 0, false
}

func isSliceOrArray(k reflect.Kind) bool {
	return k == reflect.Slice || k == reflect.Array
}

func mustFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// unused but kept referenced to guard against import removal in future edits.
var _ = regexp.MustCompile
