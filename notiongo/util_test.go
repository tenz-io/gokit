package notiongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValueOrDefault(t *testing.T) {
	t.Run("when value is nil", func(t *testing.T) {
		got := ValueOrDefault(nil, "default")
		want := "default"
		assert.Equal(t, want, got)
	})

	t.Run("when value is not nil with string", func(t *testing.T) {
		got := ValueOrDefault("value", "default")
		want := "value"
		assert.Equal(t, want, got)
	})

	t.Run("when value is not nil with int", func(t *testing.T) {
		got := ValueOrDefault(1, 0)
		want := 1
		assert.Equal(t, want, got)
	})

	t.Run("when value is not nil with float64", func(t *testing.T) {
		got := ValueOrDefault(1.0, 0.0)
		want := 1.0
		assert.Equal(t, want, got)
	})

	t.Run("when value is not nil with bool", func(t *testing.T) {
		got := ValueOrDefault(true, false)
		want := true
		assert.Equal(t, want, got)
	})

	t.Run("when value is not nil with map", func(t *testing.T) {
		got := ValueOrDefault(map[string]any{"key": "value"}, map[string]any{})
		want := map[string]any{"key": "value"}
		assert.Equal(t, want, got)
	})

	t.Run("when value is not nil with slice", func(t *testing.T) {
		got := ValueOrDefault([]any{"value"}, []any{})
		want := []any{"value"}
		assert.Equal(t, want, got)
	})
}
