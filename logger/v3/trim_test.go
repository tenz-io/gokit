package logger

import (
	"errors"
	"reflect"
	"testing"
	"time"
	"unicode/utf8"
)

type testStringer string

func (s testStringer) String() string { return string(s) }

func TestTrimFieldsIgnoresAndPreservesMalformedInput(t *testing.T) {
	ot := newTrimmer(&TrimConfig{StrLimit: 3, Ignores: []string{"secret"}})
	got := ot.TrimFields([]any{
		"name", "abcdef",
		"secret", "hidden",
		42, "unmodified",
		"dangling",
	})
	want := []any{"name", "abc...", 42, "unmodified", "dangling"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("TrimFields() = %#v, want %#v", got, want)
	}

	one := []any{"only"}
	if got := ot.TrimFields(one); !reflect.DeepEqual(got, one) {
		t.Errorf("single argument = %#v, want unchanged", got)
	}
}

func TestTrimAnyScalarAndSpecialTypes(t *testing.T) {
	ot := newTrimmer(&TrimConfig{StrLimit: 5})
	now := time.Date(2026, 7, 26, 10, 11, 12, 345000000, time.UTC)
	cases := []struct {
		name string
		in   any
		want any
	}{
		{"nil", nil, nil},
		{"bool", true, true},
		{"int", int32(-2), int64(-2)},
		{"uint", uint16(3), uint64(3)},
		{"float", float32(1.5), float64(1.5)},
		{"string", "abcdef", "abcde..."},
		{"duration", 1500 * time.Millisecond, "1.5s"},
		{"time", now, "2026-07-26T10:11:12.345"},
		{"error", errors.New("failure"), "failu..."},
		{"stringer", testStringer("stringer"), "strin..."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ot.trimAny(tc.in, ot.deepLimit); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("trimAny(%#v) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}

func TestTrimStringKeepsValidUTF8(t *testing.T) {
	ot := newTrimmer(&TrimConfig{StrLimit: 4})
	got := ot.trimString("你好")
	if got != "你..." {
		t.Fatalf("trimString = %q, want %q", got, "你...")
	}
	if !utf8.ValidString(got) {
		t.Fatalf("trimString returned invalid UTF-8: %q", got)
	}
}

func TestTrimStructHonorsTagsIgnoresAndDepth(t *testing.T) {
	type nested struct {
		Value string `json:"value"`
	}
	type payload struct {
		Name       string `json:"name,omitempty"`
		Fallback   string `json:",omitempty"`
		Hidden     string `json:"-"`
		Ignored    string `json:"ignored"`
		Nested     nested `json:"nested"`
		unexported string
	}
	ot := newTrimmer(&TrimConfig{DeepLimit: 1, Ignores: []string{"ignored"}})
	got, ok := ot.trimAny(payload{
		Name: "alice", Fallback: "fallback", Hidden: "hidden", Ignored: "ignored",
		Nested: nested{Value: "deep"}, unexported: "private",
	}, ot.deepLimit).(map[string]any)
	if !ok {
		t.Fatalf("trimmed payload type = %T", got)
	}
	if got["name"] != "alice" || got["Fallback"] != "fallback" {
		t.Errorf("trimmed payload = %#v", got)
	}
	for _, key := range []string{"", "Hidden", "ignored", "unexported"} {
		if _, exists := got[key]; exists {
			t.Errorf("unexpected struct field %q in %#v", key, got)
		}
	}
	if got["nested"] != nil {
		t.Errorf("nested value at depth limit = %#v, want nil", got["nested"])
	}
}

func TestTrimMapSupportsNonStringKeys(t *testing.T) {
	ot := newTrimmer(nil)
	got := ot.trimAny(map[bool]string{true: "yes", false: "no"}, ot.deepLimit).(map[string]any)
	if got["true"] != "yes" || got["false"] != "no" {
		t.Fatalf("trimmed map = %#v", got)
	}
}

func TestTrimSlicesBytesAndCycles(t *testing.T) {
	ot := newTrimmer(&TrimConfig{ArrLimit: 2, StrLimit: 32, DeepLimit: 2})
	if got := ot.trimAny([]int{1, 2, 3}, 2); !reflect.DeepEqual(got, []any{int64(1), int64(2)}) {
		t.Errorf("trimmed slice = %#v", got)
	}
	if got := ot.trimAny([]byte{1, 2, 3}, 2); got != "AQI=..." {
		t.Errorf("trimmed []byte = %#v, want AQI=...", got)
	}
	if got := ot.trimAny([3]byte{1, 2, 3}, 2); got != "AQI=..." {
		t.Errorf("trimmed [3]byte = %#v, want AQI=...", got)
	}
	if got := ot.trimAny([]byte{}, 2); got != "[]" {
		t.Errorf("trimmed empty bytes = %#v, want []", got)
	}

	cyclic := make([]any, 1)
	cyclic[0] = cyclic
	got := ot.trimAny(cyclic, 2).([]any)
	nested, ok := got[0].([]any)
	if !ok || len(nested) != 1 || nested[0] != nil {
		t.Fatalf("trimmed cyclic slice = %#v", got)
	}
}

func TestTrimmerDefaults(t *testing.T) {
	ot := newTrimmer(&TrimConfig{})
	if ot.arrLimit != defaultArrLimit || ot.strLimit != defaultStrLimit || ot.deepLimit != defaultDeepLimit {
		t.Fatalf("defaults = arr:%d str:%d depth:%d", ot.arrLimit, ot.strLimit, ot.deepLimit)
	}
	if got := keySet([]string{"a", "b"}); !got["a"] || !got["b"] {
		t.Fatalf("keySet = %#v", got)
	}
}
