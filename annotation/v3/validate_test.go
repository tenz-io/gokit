package annotation

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func reflectValue(ptr any) reflect.Value { return reflect.ValueOf(ptr).Elem() }

type inner struct {
	InnerField int `validate:"required,gt=0"`
}

type sampleConfig struct {
	Host string `default:"localhost" validate:"required"`
	Port int    `default:"8080" validate:"required,gt=0,lte=65535"`
	// optional pointer, NO default tag -> must stay nil after ApplyDefaults
	OptionalPtr *int
	// pointer WITH default -> populated
	WithDefaultPtr *string `default:"hi"`
	Nested         inner   `validate:"required"`
	NestedPtr      *inner  `validate:"required"`
}

func TestApplyDefaults_DoesNotAllocateUnrelatedPointers(t *testing.T) {
	cfg := &sampleConfig{}
	require.NoError(t, ApplyDefaults(cfg))

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	// Optional pointer untouched.
	assert.Nil(t, cfg.OptionalPtr)
	// Defaulted pointer populated.
	require.NotNil(t, cfg.WithDefaultPtr)
	assert.Equal(t, "hi", *cfg.WithDefaultPtr)
	// Nested value field: its inner default applies.
	// (inner has no default tags; nothing to set, but recursion must not error)
	assert.Equal(t, 0, cfg.Nested.InnerField)
	// Required nested pointer: ApplyDefaults must NOT instantiate it (no
	// default anywhere in inner). Validation will catch it.
	assert.Nil(t, cfg.NestedPtr)
}

func TestApplyDefaults_KeepsCallerValue(t *testing.T) {
	cfg := &sampleConfig{Host: "override.example", Port: 9000}
	require.NoError(t, ApplyDefaults(cfg))
	assert.Equal(t, "override.example", cfg.Host)
	assert.Equal(t, 9000, cfg.Port)
}

func TestValidate_CollectsAllErrors(t *testing.T) {
	cfg := &sampleConfig{Port: -1} // Host missing, Port out of range, NestedPtr nil, inner.InnerField 0
	err := Validate(cfg)
	require.Error(t, err)
	verrs, ok := AsErrors(err)
	require.True(t, ok)
	// Host required, Port gt+lte, Nested.InnerField required+gt, NestedPtr required.
	assert.Greater(t, len(verrs), 1, "should report multiple failures, got %v", verrs)

	fields := map[string]bool{}
	for _, e := range verrs {
		fields[e.Field] = true
	}
	assert.True(t, fields["Host"])
	assert.True(t, fields["Port"])
	assert.True(t, fields["NestedPtr"])
}

func TestValidate_PassesWhenValid(t *testing.T) {
	cfg := &sampleConfig{Nested: inner{InnerField: 5}, NestedPtr: &inner{InnerField: 5}}
	require.NoError(t, ApplyDefaults(cfg))
	assert.NoError(t, Validate(cfg))
}

type sliceRules struct {
	Tags    []string `validate:"min_len=1,dive:non_blank"`
	Counts  []int    `validate:"gt=0"`
	JustOne []int    `validate:"len=3"`
}

func TestValidate_SliceAndDive(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		s := &sliceRules{Tags: []string{"a", "b"}, Counts: []int{1, 2}, JustOne: []int{1, 2, 3}}
		assert.NoError(t, Validate(s))
	})
	t.Run("dive catches blank element with index", func(t *testing.T) {
		s := &sliceRules{Tags: []string{"a", "  "}, Counts: []int{1, 2}, JustOne: []int{1, 2, 3}}
		err := Validate(s)
		require.Error(t, err)
		verrs, _ := AsErrors(err)
		found := false
		for _, e := range verrs {
			if e.Rule == "dive" && contains(e.Msg, "element 1") {
				found = true
			}
		}
		assert.True(t, found, "dive should report element index, got %v", verrs)
	})
	t.Run("slice numeric compares each element", func(t *testing.T) {
		s := &sliceRules{Tags: []string{"a"}, Counts: []int{0, 2}, JustOne: []int{1, 2, 3}}
		err := Validate(s)
		require.Error(t, err)
		verrs, _ := AsErrors(err)
		assert.True(t, hasRule(verrs, "gt"))
	})
	t.Run("len mismatch", func(t *testing.T) {
		s := &sliceRules{Tags: []string{"a"}, Counts: []int{1}, JustOne: []int{1, 2}}
		err := Validate(s)
		require.Error(t, err)
		verrs, _ := AsErrors(err)
		assert.True(t, hasRule(verrs, "len"))
	})
}

type msgConfig struct {
	Age int `validate:"gte=0,lte=150,msg=Age must be 0..150"`
}

func TestValidate_CustomMessage(t *testing.T) {
	err := Validate(&msgConfig{Age: 200})
	require.Error(t, err)
	verrs, _ := AsErrors(err)
	var matched FieldError
	for _, e := range verrs {
		if e.Rule == "lte" {
			matched = e
		}
	}
	assert.Equal(t, "Age must be 0..150", matched.Msg)
}

type patternConfig struct {
	Email string `validate:"required,email"`
	Code  string `validate:"pattern=^[A-Z]{3}$"`
}

func TestValidate_Patterns(t *testing.T) {
	t.Run("email", func(t *testing.T) {
		assert.NoError(t, Validate(&patternConfig{Email: "a@b.com", Code: "ABC"}))
	})
	t.Run("bad email", func(t *testing.T) {
		err := Validate(&patternConfig{Email: "nope", Code: "ABC"})
		require.Error(t, err)
		verrs, _ := AsErrors(err)
		assert.True(t, hasRule(verrs, "email"))
	})
	t.Run("bad pattern", func(t *testing.T) {
		err := Validate(&patternConfig{Email: "a@b.com", Code: "ab"})
		require.Error(t, err)
		verrs, _ := AsErrors(err)
		assert.True(t, hasRule(verrs, "pattern"))
	})
}

func TestValidate_ConfigErrorSurfaced(t *testing.T) {
	// "gt=abc" is a bad parameter; plan build must fail rather than silently
	// treating it as 0.
	type bad struct {
		N int `validate:"gt=abc"`
	}
	err := Validate(&bad{})
	require.Error(t, err)
}

func TestValidate_UnknownRuleBecomesFailure(t *testing.T) {
	type weird struct {
		N int `validate:"frobnicate=1"`
	}
	err := Validate(&weird{N: 1})
	require.Error(t, err)
	verrs, _ := AsErrors(err)
	assert.True(t, hasRule(verrs, "frobnicate"))
}

// Custom registered rule.
func TestRegister_CustomRule(t *testing.T) {
	Register("even", func(_ string, _ reflect.StructField) (Rule, error) {
		return func(rv reflect.Value) (bool, string) {
			if rv.Kind() == reflect.Int && rv.Int()%2 != 0 {
				return false, "must be even"
			}
			return true, ""
		}, nil
	})
	type even struct {
		N int `validate:"even"`
	}
	assert.NoError(t, Validate(&even{N: 2}))
	err := Validate(&even{N: 3})
	require.Error(t, err)
	verrs, _ := AsErrors(err)
	assert.True(t, hasRule(verrs, "even"))
}

func TestSetString_ScalarsAndDuration(t *testing.T) {
	type s struct {
		I     int
		F     float64
		B     bool
		Str   string
		D     time.Duration
		P     *int
		Bytes []byte
	}
	obj := &s{}
	root := reflectValue(obj)
	p, err := PlanFor(obj)
	require.NoError(t, err)
	for _, f := range p.Fields() {
		fv := root.FieldByIndex(f.Index)
		switch f.Name {
		case "I":
			require.NoError(t, SetString(fv, "42"))
			assert.Equal(t, 42, obj.I)
		case "F":
			require.NoError(t, SetString(fv, "1.5"))
			assert.Equal(t, 1.5, obj.F)
		case "B":
			require.NoError(t, SetString(fv, "true"))
			assert.True(t, obj.B)
		case "Str":
			require.NoError(t, SetString(fv, "hello"))
			assert.Equal(t, "hello", obj.Str)
		case "D":
			require.NoError(t, SetString(fv, "5s"))
			assert.Equal(t, 5*time.Second, obj.D)
		case "P":
			require.NoError(t, SetString(fv, "7"))
			require.NotNil(t, obj.P)
			assert.Equal(t, 7, *obj.P)
		case "Bytes":
			require.NoError(t, SetString(fv, "raw"))
			assert.Equal(t, []byte("raw"), obj.Bytes)
		}
	}
}

type durationCfg struct {
	Timeout time.Duration `default:"5s" validate:"gt=0"`
}

func TestApplyDefaults_Duration(t *testing.T) {
	d := &durationCfg{}
	require.NoError(t, ApplyDefaults(d))
	assert.Equal(t, 5*time.Second, d.Timeout)
	assert.NoError(t, Validate(d))
}

// --- helpers ---

func hasRule(errs ValidationErrors, rule string) bool {
	for _, e := range errs {
		if e.Rule == rule {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// guard: errors.As on the concrete type works.
func TestAsErrors_NonValidationError(t *testing.T) {
	_, ok := AsErrors(errors.New("other"))
	assert.False(t, ok)
}
