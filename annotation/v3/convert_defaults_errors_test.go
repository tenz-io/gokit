package annotation

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyDefaults_AllSupportedKindsAndNestedPointers(t *testing.T) {
	type nested struct {
		Name string `default:"nested"`
	}
	type defaults struct {
		String string        `default:"text"`
		Int8   int8          `default:"12"`
		Uint16 uint16        `default:"34"`
		Float  float32       `default:"1.25"`
		Bool   bool          `default:"true"`
		Bytes  []byte        `default:"bytes"`
		Delay  time.Duration `default:"250ms"`
		Ptr    **int         `default:"7"`
		Nested *nested
		Empty  *struct{ Value string }
		Keep   string `default:"replacement"`
	}

	v := &defaults{Keep: "caller"}
	require.NoError(t, ApplyDefaults(v))
	assert.Equal(t, "text", v.String)
	assert.Equal(t, int8(12), v.Int8)
	assert.Equal(t, uint16(34), v.Uint16)
	assert.InDelta(t, 1.25, v.Float, 0.001)
	assert.True(t, v.Bool)
	assert.Equal(t, []byte("bytes"), v.Bytes)
	assert.Equal(t, 250*time.Millisecond, v.Delay)
	require.NotNil(t, v.Ptr)
	require.NotNil(t, *v.Ptr)
	assert.Equal(t, 7, **v.Ptr)
	require.NotNil(t, v.Nested)
	assert.Equal(t, "nested", v.Nested.Name)
	assert.Nil(t, v.Empty)
	assert.Equal(t, "caller", v.Keep)
}

func TestApplyDefaults_ReportsConversionAndArgumentErrors(t *testing.T) {
	type badInt struct {
		Value int `default:"not-int"`
	}
	err := ApplyDefaults(&badInt{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "field Value")

	type badSlice struct {
		Value []int `default:"1"`
	}
	err = ApplyDefaults(&badSlice{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported slice")

	for _, input := range []any{nil, struct{}{}, new(int)} {
		require.Error(t, ApplyDefaults(input))
	}
}

func TestSetString_SuccessErrorsOverflowAndPointers(t *testing.T) {
	type target struct {
		I8    int8
		U8    uint8
		F32   float32
		Flag  bool
		Delay time.Duration
		Text  string
		Bytes []byte
		Ints  []int
		Obj   struct{}
		Ptr   **int
	}
	v := &target{}
	root := reflect.ValueOf(v).Elem()

	success := map[string]string{
		"I8": "127", "U8": "255", "F32": "3.5", "Flag": "true",
		"Delay": "2s", "Text": "hello", "Bytes": "raw", "Ptr": "9",
	}
	for field, input := range success {
		require.NoError(t, SetString(root.FieldByName(field), input), field)
	}
	assert.Equal(t, int8(127), v.I8)
	assert.Equal(t, uint8(255), v.U8)
	assert.Equal(t, float32(3.5), v.F32)
	assert.True(t, v.Flag)
	assert.Equal(t, 2*time.Second, v.Delay)
	assert.Equal(t, "hello", v.Text)
	assert.Equal(t, []byte("raw"), v.Bytes)
	require.NotNil(t, v.Ptr)
	require.NotNil(t, *v.Ptr)
	assert.Equal(t, 9, **v.Ptr)

	failures := []struct {
		field string
		input string
	}{
		{"I8", "128"},
		{"I8", "x"},
		{"U8", "256"},
		{"U8", "-1"},
		{"F32", "not-float"},
		{"Flag", "not-bool"},
		{"Delay", "not-duration"},
		{"Ints", "1"},
		{"Obj", "x"},
	}
	for _, tc := range failures {
		assert.Error(t, SetString(root.FieldByName(tc.field), tc.input), tc.field)
	}
	assert.Error(t, SetString(reflect.Value{}, "x"))
	assert.Error(t, SetString(reflect.ValueOf(1), "2"))
	var nilInt *int
	assert.Error(t, SetString(reflect.ValueOf(nilInt), "2"))
}

func TestSet_TypedAssignments(t *testing.T) {
	type namedInt int
	type target struct {
		Number int64
		Text   string
		Ptr    *int
	}
	v := &target{}
	root := reflect.ValueOf(v).Elem()

	require.NoError(t, Set(root.FieldByName("Text"), "value"))
	require.NoError(t, Set(root.FieldByName("Number"), namedInt(12)))
	require.NoError(t, Set(root.FieldByName("Ptr"), 8))
	require.NotNil(t, v.Ptr)
	assert.Equal(t, 8, *v.Ptr)
	assert.Equal(t, int64(12), v.Number)
	assert.Equal(t, "value", v.Text)

	assert.NoError(t, Set(root.FieldByName("Text"), nil))
	var typedNil *string
	assert.NoError(t, Set(root.FieldByName("Text"), typedNil))
	assert.Error(t, Set(root.FieldByName("Number"), "no"))
	assert.Error(t, Set(reflect.Value{}, 1))
	assert.Error(t, Set(reflect.ValueOf(1), 2))
}

func TestValidationErrorModel(t *testing.T) {
	plain := NewFieldError("Name", "required", "", "")
	assert.Equal(t, "Name: invalid", plain.Error())
	withMessage := NewFieldError("Age", "gt", "0", "must be positive")
	assert.Equal(t, "Age: must be positive", withMessage.Error())
	adHoc := NewFieldError("Body", "", "", "malformed")
	assert.Equal(t, "Body: malformed", adHoc.Error())

	var empty ValidationErrors
	assert.False(t, empty.Has())
	assert.Empty(t, empty.Error())

	combined := ValidationErrors{plain, withMessage}
	assert.True(t, combined.Has())
	assert.Equal(t, "Name: invalid; Age: must be positive", combined.Error())
	assert.Equal(t, "Field: bad", Err("Field", "bind", "bad").Error())
	assert.Equal(t, "Field: bad 2", Errf("Field", "bind", "bad %d", 2).Error())

	got, ok := AsErrors(combined)
	require.True(t, ok)
	assert.Equal(t, combined, got)
	_, ok = AsErrors(errors.New("plain"))
	assert.False(t, ok)
	_, ok = AsErrors(nil)
	assert.False(t, ok)
}
