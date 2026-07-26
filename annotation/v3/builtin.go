package annotation

import (
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"net"
	"net/url"
	"reflect"
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

func numericValidator(name, param, sym string) (Rule, error) {
	limit, err := parseNumericLimit(param)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid number %q", name, param)
	}
	return func(rv reflect.Value) (bool, string) {
		rv, ok := indirectValue(rv)
		if !ok {
			return true, ""
		}
		if isSliceOrArray(rv.Kind()) {
			// Compare every element.
			for i := 0; i < rv.Len(); i++ {
				if cmp, ok := compareNumeric(rv.Index(i), limit); ok && !orderedComparison(name, cmp) {
					return false, fmt.Sprintf("element %d must be %s %s", i, sym, param)
				}
			}
			return true, ""
		}
		cmp, ok := compareNumeric(rv, limit)
		if !ok {
			return true, "" // not a numeric field; skip
		}
		if !orderedComparison(name, cmp) {
			return false, fmt.Sprintf("must be %s %s", sym, param)
		}
		return true, ""
	}, nil
}

func minValidator(param string, ft reflect.StructField) (Rule, error) {
	return numericValidator("min", param, ">=")
}
func maxValidator(param string, ft reflect.StructField) (Rule, error) {
	return numericValidator("max", param, "<=")
}
func gtValidator(param string, ft reflect.StructField) (Rule, error) {
	return numericValidator("gt", param, ">")
}
func ltValidator(param string, ft reflect.StructField) (Rule, error) {
	return numericValidator("lt", param, "<")
}
func gteValidator(param string, ft reflect.StructField) (Rule, error) {
	return numericValidator("gte", param, ">=")
}
func lteValidator(param string, ft reflect.StructField) (Rule, error) {
	return numericValidator("lte", param, "<=")
}

func eqValidator(param string, ft reflect.StructField) (Rule, error) {
	limit, isNumber, err := equalityLimit(param, ft)
	if err != nil {
		return nil, err
	}
	return func(rv reflect.Value) (bool, string) {
		if isNumber {
			cmp, ok := compareNumeric(rv, limit)
			if ok && cmp != 0 {
				return false, fmt.Sprintf("must equal %v", param)
			}
			return true, ""
		}
		rv, ok := indirectValue(rv)
		if !ok {
			return true, ""
		}
		if rv.Kind() == reflect.String && rv.String() != param {
			return false, fmt.Sprintf("must equal %v", param)
		}
		return true, ""
	}, nil
}

func neValidator(param string, ft reflect.StructField) (Rule, error) {
	limit, isNumber, err := equalityLimit(param, ft)
	if err != nil {
		return nil, err
	}
	return func(rv reflect.Value) (bool, string) {
		if isNumber {
			cmp, ok := compareNumeric(rv, limit)
			if ok && cmp == 0 {
				return false, fmt.Sprintf("must not equal %v", param)
			}
			return true, ""
		}
		rv, ok := indirectValue(rv)
		if !ok {
			return true, ""
		}
		if rv.Kind() == reflect.String && rv.String() == param {
			return false, fmt.Sprintf("must not equal %q", param)
		}
		return true, ""
	}, nil
}

func equalityLimit(param string, ft reflect.StructField) (numericLimit, bool, error) {
	t := ft.Type
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Float32, reflect.Float64:
		limit, err := parseNumericLimit(param)
		if err != nil {
			return numericLimit{}, true, fmt.Errorf("invalid number %q", param)
		}
		return limit, true, nil
	default:
		return numericLimit{}, false, nil
	}
}

type numericLimit struct {
	float float64
	exact *big.Float
}

func parseNumericLimit(param string) (numericLimit, error) {
	f, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return numericLimit{}, err
	}
	exact, _, err := big.ParseFloat(param, 10, 256, big.ToNearestEven)
	if err != nil {
		return numericLimit{}, err
	}
	return numericLimit{float: f, exact: exact}, nil
}

func compareNumeric(rv reflect.Value, limit numericLimit) (int, bool) {
	rv, ok := indirectValue(rv)
	if !ok {
		return 0, false
	}
	var value *big.Float
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = new(big.Float).SetPrec(256).SetInt64(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		value = new(big.Float).SetPrec(256).SetUint64(rv.Uint())
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		if math.IsNaN(f) {
			return 2, true
		}
		return compareFloat(f, limit.float), true
	case reflect.String:
		d, err := time.ParseDuration(rv.String())
		if err != nil {
			return 0, false
		}
		return compareFloat(d.Seconds(), limit.float), true
	default:
		return 0, false
	}
	return value.Cmp(limit.exact), true
}

func compareFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func orderedComparison(name string, cmp int) bool {
	switch name {
	case "min", "gte":
		return cmp >= 0 && cmp != 2
	case "max", "lte":
		return cmp <= 0
	case "gt":
		return cmp > 0 && cmp != 2
	case "lt":
		return cmp < 0
	default:
		return false
	}
}

// --- oneof ---

func oneofValidator(param string, _ reflect.StructField) (Rule, error) {
	options := strings.Split(param, " ")
	return func(rv reflect.Value) (bool, string) {
		rv, ok := indirectValue(rv)
		if !ok {
			return true, ""
		}
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
		rv, ok := indirectValue(rv)
		if !ok {
			return true, ""
		}
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
		rv, ok := indirectValue(rv)
		if !ok {
			return true, ""
		}
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
			rv, ok := indirectValue(rv)
			if !ok {
				return true, ""
			}
			if rv.Kind() != reflect.String {
				return true, ""
			}
			value := rv.String()
			if !re.MatchString(value) || !validNamedValue(name, value) {
				return false, fmt.Sprintf("must be a valid %s", name)
			}
			return true, ""
		}, nil
	}
}

func validNamedValue(name, value string) bool {
	switch name {
	case "url":
		u, err := url.ParseRequestURI(value)
		return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
	case "ipv4":
		ip := net.ParseIP(value)
		return ip != nil && ip.To4() != nil
	case "ipv6":
		ip := net.ParseIP(value)
		return ip != nil && ip.To4() == nil
	case "date":
		_, err := time.Parse("2006-01-02", value)
		return err == nil
	case "base64":
		if _, err := base64.StdEncoding.DecodeString(value); err == nil {
			return true
		}
		_, err := base64.RawStdEncoding.DecodeString(value)
		return err == nil
	default:
		return true
	}
}

// --- contains / prefix / suffix ---

func containsValidator(param string, _ reflect.StructField) (Rule, error) {
	return func(rv reflect.Value) (bool, string) {
		rv, ok := indirectValue(rv)
		if !ok {
			return true, ""
		}
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
		rv, ok := indirectValue(rv)
		if !ok {
			return true, ""
		}
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
		rv, ok := indirectValue(rv)
		if !ok {
			return true, ""
		}
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
	name, nestedParam, _ := splitRule(param)
	v, ok := lookupValidator(name)
	if !ok || v == nil {
		return nil, fmt.Errorf("dive: unknown rule %q", name)
	}
	elemRule, err := v(nestedParam, ft)
	if err != nil {
		return nil, fmt.Errorf("dive: %v", err)
	}
	return func(rv reflect.Value) (bool, string) {
		rv, ok := indirectValue(rv)
		if !ok {
			return true, ""
		}
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
	eq, colon := strings.Index(raw, "="), strings.Index(raw, ":")
	switch {
	case eq >= 0 && (colon < 0 || eq < colon):
		return raw[:eq], raw[eq+1:], true
	case colon >= 0:
		return raw[:colon], raw[colon+1:], true
	}
	return raw, "", false
}

func compileRules(f *Field) error {
	tag := f.StructField.Tag.Get(tagValidate)
	if tag == "" {
		return nil
	}
	for _, raw := range strings.Split(tag, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		name, param, _ := splitRule(raw)

		if name == "msg" {
			// Applies to the previously compiled rule.
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
			continue
		}
		check, err := v(param, f.StructField)
		if err != nil {
			return fmt.Errorf("field %s: rule %s: %w", f.Path, name, err)
		}
		r := namedRule{name: name, param: param, run: check}
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
	rv, ok := indirectValue(rv)
	if !ok {
		return 0, false
	}
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return rv.Len(), true
	}
	return 0, false
}

func indirectValue(rv reflect.Value) (reflect.Value, bool) {
	for rv.IsValid() && (rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface) {
		if rv.IsNil() {
			return reflect.Value{}, false
		}
		rv = rv.Elem()
	}
	return rv, rv.IsValid()
}

func isSliceOrArray(k reflect.Kind) bool {
	return k == reflect.Slice || k == reflect.Array
}
