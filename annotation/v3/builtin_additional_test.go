package annotation

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func compileTestRule(t *testing.T, validator Validator, param string, value any) Rule {
	t.Helper()
	rule, err := validator(param, reflect.StructField{Type: reflect.TypeOf(value)})
	require.NoError(t, err)
	return rule
}

func runTestRule(t *testing.T, validator Validator, param string, value any) (bool, string) {
	t.Helper()
	rule := compileTestRule(t, validator, param, value)
	return rule(reflect.ValueOf(value))
}

func TestBuiltins_RequiredAndLengthRules(t *testing.T) {
	emptyValues := []any{"", []int(nil), map[string]int{}, [0]int{}, false, (*int)(nil), time.Time{}}
	for _, value := range emptyValues {
		ok, msg := runTestRule(t, requiredValidator, "", value)
		assert.False(t, ok, "%T", value)
		assert.NotEmpty(t, msg)
	}
	presentValues := []any{"x", []int{1}, map[string]int{"x": 1}, true, 0, uint(0), 0.0, time.Now()}
	for _, value := range presentValues {
		ok, _ := runTestRule(t, requiredValidator, "", value)
		assert.True(t, ok, "%T", value)
	}

	lengthCases := []struct {
		validator Validator
		param     string
		value     any
		want      bool
	}{
		{lenValidator, "3", "abc", true},
		{lenValidator, "2", "abc", false},
		{minLenValidator, "2", []int{1, 2}, true},
		{minLenValidator, "3", []int{1, 2}, false},
		{maxLenValidator, "2", map[string]int{"a": 1}, true},
		{maxLenValidator, "0", map[string]int{"a": 1}, false},
		{lenValidator, "3", 99, true},
	}
	for _, tc := range lengthCases {
		ok, _ := runTestRule(t, tc.validator, tc.param, tc.value)
		assert.Equal(t, tc.want, ok)
	}
	for _, validator := range []Validator{lenValidator, minLenValidator, maxLenValidator} {
		_, err := validator("bad", reflect.StructField{})
		assert.Error(t, err)
	}
}

func TestBuiltins_NumericComparisons(t *testing.T) {
	validators := []struct {
		name string
		fn   Validator
		pass float64
		fail float64
	}{
		{"min", minValidator, 2, 1},
		{"max", maxValidator, 2, 3},
		{"gt", gtValidator, 3, 2},
		{"lt", ltValidator, 1, 2},
		{"gte", gteValidator, 2, 1},
		{"lte", lteValidator, 2, 3},
	}
	for _, tc := range validators {
		t.Run(tc.name, func(t *testing.T) {
			ok, _ := runTestRule(t, tc.fn, "2", tc.pass)
			assert.True(t, ok)
			ok, msg := runTestRule(t, tc.fn, "2", tc.fail)
			assert.False(t, ok)
			assert.NotEmpty(t, msg)
			_, err := tc.fn("bad", reflect.StructField{})
			assert.Error(t, err)
		})
	}

	ok, msg := runTestRule(t, gtValidator, "0", []int{1, 0, 2})
	assert.False(t, ok)
	assert.Contains(t, msg, "element 1")
	ok, _ = runTestRule(t, gtValidator, "0", []string{"not numeric"})
	assert.True(t, ok)
	ok, _ = runTestRule(t, gtValidator, "0", "not numeric")
	assert.True(t, ok)
	ok, _ = runTestRule(t, gtValidator, "1", "2s")
	assert.True(t, ok)

	// Integer comparisons must not lose precision through float64 conversion.
	const beyondFloatPrecision = uint64(9007199254740993)
	ok, _ = runTestRule(t, gtValidator, "9007199254740992", beyondFloatPrecision)
	assert.True(t, ok)
	ok, _ = runTestRule(t, lteValidator, "9007199254740992", beyondFloatPrecision)
	assert.False(t, ok)

}

func TestBuiltins_EqualityAndOneOf(t *testing.T) {
	ok, _ := runTestRule(t, eqValidator, "2", 2)
	assert.True(t, ok)
	ok, _ = runTestRule(t, eqValidator, "2", 3)
	assert.False(t, ok)
	ok, _ = runTestRule(t, neValidator, "2", 3)
	assert.True(t, ok)
	ok, _ = runTestRule(t, neValidator, "2", 2)
	assert.False(t, ok)
	ok, _ = runTestRule(t, eqValidator, "9007199254740993", uint64(9007199254740993))
	assert.True(t, ok)

	// A duration-looking string is still a string for eq/ne.
	ok, _ = runTestRule(t, eqValidator, "1s", "1s")
	assert.True(t, ok)
	ok, _ = runTestRule(t, neValidator, "1s", "1s")
	assert.False(t, ok)

	_, err := eqValidator("not-number", reflect.StructField{Type: reflect.TypeOf(int(0))})
	assert.Error(t, err)
	_, err = neValidator("not-number", reflect.StructField{Type: reflect.TypeOf((*int)(nil))})
	assert.Error(t, err)

	ok, _ = runTestRule(t, oneofValidator, "red green blue", "green")
	assert.True(t, ok)
	ok, msg := runTestRule(t, oneofValidator, "red green blue", "black")
	assert.False(t, ok)
	assert.Contains(t, msg, "red green blue")
}

func TestBuiltins_StringAndPatternRules(t *testing.T) {
	ok, _ := runTestRule(t, nonBlankValidator, "", " text ")
	assert.True(t, ok)
	ok, _ = runTestRule(t, nonBlankValidator, "", "   ")
	assert.False(t, ok)
	ok, msg := runTestRule(t, nonBlankValidator, "", []string{"ok", " "})
	assert.False(t, ok)
	assert.Contains(t, msg, "element 1")
	ok, _ = runTestRule(t, nonBlankValidator, "", 1)
	assert.True(t, ok)

	stringRules := []struct {
		fn    Validator
		param string
		pass  string
		fail  string
	}{
		{containsValidator, "mid", "a-middle", "outside"},
		{prefixValidator, "pre", "prefix", "xpre"},
		{suffixValidator, "end", "the-end", "ending"},
	}
	for _, tc := range stringRules {
		ok, _ := runTestRule(t, tc.fn, tc.param, tc.pass)
		assert.True(t, ok)
		ok, msg := runTestRule(t, tc.fn, tc.param, tc.fail)
		assert.False(t, ok)
		assert.NotEmpty(t, msg)
		ok, _ = runTestRule(t, tc.fn, tc.param, 123)
		assert.True(t, ok)
	}

	namedCases := []struct {
		name, pass, fail string
	}{
		{"email", "a@b.com", "a @b.com"},
		{"url", "https://example.com/a", "ftp://example.com"},
		{"uuid", "123e4567-e89b-12d3-a456-426614174000", "no"},
		{"ipv4", "127.0.0.1", "127.0.0"},
		{"ipv6", "2001:db8::1", "xyz"},
		{"alpha", "Abc", "A1"},
		{"alphanum", "A1", "A-1"},
		{"numeric", "123", "12a"},
		{"hex", "aB09", "xz"},
		{"date", "2026-07-26", "26-07-26"},
		{"base64", "YWJjZA==", "bad!"},
	}

	semanticFailures := map[string]string{
		"url":    "https://",
		"ipv4":   "999.999.999.999",
		"ipv6":   ":",
		"date":   "2026-99-99",
		"base64": "a",
	}
	for name, value := range semanticFailures {
		ok, _ := runTestRule(t, namedPatternValidator(name), "", value)
		assert.False(t, ok, name)
	}
	for _, tc := range namedCases {
		validator := namedPatternValidator(tc.name)
		ok, _ := runTestRule(t, validator, "", tc.pass)
		assert.True(t, ok, tc.name)
		ok, msg := runTestRule(t, validator, "", tc.fail)
		assert.False(t, ok, tc.name)
		assert.Contains(t, msg, tc.name)
		ok, _ = runTestRule(t, validator, "", 123)
		assert.True(t, ok)
	}

	ok, _ = runTestRule(t, patternValidator, "^[A-Z]+$", "ABC")
	assert.True(t, ok)
	ok, _ = runTestRule(t, patternValidator, "^[A-Z]+$", "abc")
	assert.False(t, ok)
	ok, _ = runTestRule(t, patternValidator, "^[A-Z]+$", 1)
	assert.True(t, ok)
	_, err := patternValidator("[", reflect.StructField{})
	assert.Error(t, err)
	_, err = compilePattern("#missing")
	assert.Error(t, err)
	re1, err := compilePattern("^[a-z]+$")
	require.NoError(t, err)
	re2, err := compilePattern("^[a-z]+$")
	require.NoError(t, err)
	assert.Same(t, re1, re2)
}

func TestBuiltins_DiveParameterizedAndMessageIsolation(t *testing.T) {
	type parameterized struct {
		Values []int `validate:"dive:gt=0"`
	}
	assert.NoError(t, Validate(&parameterized{Values: []int{1, 2}}))
	err := Validate(&parameterized{Values: []int{1, 0}})
	require.Error(t, err)
	errs, _ := AsErrors(err)
	require.Len(t, errs, 1)
	assert.Equal(t, "dive", errs[0].Rule)
	assert.Contains(t, errs[0].Msg, "element 1")

	rule := compileTestRule(t, diveValidator, "non_blank", map[string]string{})
	ok, msg := rule(reflect.ValueOf(map[string]string{"bad": " "}))
	assert.False(t, ok)
	assert.Contains(t, msg, "key bad")
	ok, _ = rule(reflect.ValueOf(42))
	assert.True(t, ok)
	_, err = diveValidator("missing", reflect.StructField{})
	assert.Error(t, err)
	_, err = diveValidator("msg", reflect.StructField{})
	assert.Error(t, err)
	_, err = diveValidator("gt=bad", reflect.StructField{Type: reflect.TypeOf([]int{})})
	assert.Error(t, err)

	type messages struct {
		Value int `validate:"gt=0,msg=positive,lt=10"`
	}
	err = Validate(&messages{Value: 11})
	require.Error(t, err)
	errs, _ = AsErrors(err)
	require.Len(t, errs, 1)
	assert.Equal(t, "lt", errs[0].Rule)
	assert.NotEqual(t, "positive", errs[0].Msg, "msg applies only to the preceding rule")
}

func TestBuiltins_ValidatePointedValues(t *testing.T) {
	type pointerValues struct {
		Count *int    `validate:"required,gt=0"`
		Email *string `validate:"required,email"`
		Code  *string `validate:"eq=ok"`
	}
	count, email, code := -1, "not-an-email", "bad"
	err := Validate(&pointerValues{Count: &count, Email: &email, Code: &code})
	require.Error(t, err)
	errs, _ := AsErrors(err)
	assert.True(t, hasRule(errs, "gt"))
	assert.True(t, hasRule(errs, "email"))
	assert.True(t, hasRule(errs, "eq"))

	count, email, code = 1, "a@b.com", "ok"
	assert.NoError(t, Validate(&pointerValues{Count: &count, Email: &email, Code: &code}))

	// Optional nil pointers skip value rules; required remains responsible for
	// presence checks.
	type optional struct {
		Count *int `validate:"gt=0"`
	}
	assert.NoError(t, Validate(&optional{}))
}

func TestBuiltinHelpers(t *testing.T) {
	name, param, ok := splitRule("dive:gt=0")
	assert.True(t, ok)
	assert.Equal(t, "dive", name)
	assert.Equal(t, "gt=0", param)
	name, param, ok = splitRule("gt=1")
	assert.True(t, ok)
	assert.Equal(t, "gt", name)
	assert.Equal(t, "1", param)
	name, param, ok = splitRule("required")
	assert.False(t, ok)
	assert.Equal(t, "required", name)
	assert.Empty(t, param)

	assert.True(t, isSliceOrArray(reflect.Slice))
	assert.True(t, isSliceOrArray(reflect.Array))
	assert.False(t, isSliceOrArray(reflect.Map))
	_, ok = lengthOf(reflect.ValueOf(1))
	assert.False(t, ok)
}
