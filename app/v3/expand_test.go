package app

import (
	"strings"
	"testing"
)

// lookup builds a func(string)(string,bool) from a map for tests.
func lookup(env map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := env[k]
		return v, ok
	}
}

func TestExpand(t *testing.T) {
	tests := []struct {
		name string
		in   string
		env  map[string]string
		want string
		err  string // non-empty substring the error must contain, "" = no error
	}{
		{name: "no dollar", in: "plain config", want: "plain config"},
		{name: "no placeholder dollar", in: "price is $5", want: "price is $5"},
		{name: "bare var no braces left as-is", in: "use $HOME here", want: "use $HOME here"},
		{name: "simple set", in: "${VAR}", env: map[string]string{"VAR": "v"}, want: "v"},
		{name: "simple unset errors", in: "${MISSING}", env: nil, err: "unset or empty"},
		{name: "empty value errors", in: "${VAR}", env: map[string]string{"VAR": ""}, err: "unset or empty"},
		{name: "default when unset", in: "${VAR:-fallback}", env: nil, want: "fallback"},
		{name: "default when empty", in: "${VAR:-fallback}", env: map[string]string{"VAR": ""}, want: "fallback"},
		{name: "default not used when set", in: "${VAR:-fallback}", env: map[string]string{"VAR": "real"}, want: "real"},
		{name: "default with colons (url port)", in: "${URL:-http://host:8080}", env: nil, want: "http://host:8080"},
		{name: "required unset errors with msg", in: "${VAR:?must be set}", env: nil, err: "must be set"},
		{name: "required empty errors", in: "${VAR:?required}", env: map[string]string{"VAR": ""}, err: "required"},
		{name: "required set ok", in: "${VAR:?required}", env: map[string]string{"VAR": "ok"}, want: "ok"},
		{name: "required no msg", in: "${VAR:?}", env: nil, err: "required but unset"},
		{name: "multiple placeholders", in: "a=${A} b=${B:-b}", env: map[string]string{"A": "1"}, want: "a=1 b=b"},
		{name: "nested default", in: "${A:-${B}}", env: map[string]string{"B": "nested"}, want: "nested"},
		{name: "nested default both unset errors", in: "${A:-${B}}", env: nil, err: "unset or empty"},
		{name: "placeholder mid string", in: "host=redis:6379 pass=${PASS}", env: map[string]string{"PASS": "s3cret"}, want: "host=redis:6379 pass=s3cret"},
		{name: "invalid name errors", in: "${1A}", env: nil, err: "invalid variable name"},
		{name: "unterminated errors", in: "${VAR", env: nil, err: "unterminated"},
		{name: "dollar at end stays", in: "trailing$", env: nil, want: "trailing$"},
		{name: "literal brace no dollar", in: "{not a placeholder}", env: nil, want: "{not a placeholder}"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Expand([]byte(tc.in), lookup(tc.env))
			if tc.err != "" {
				if err == nil {
					t.Fatalf("Expand(%q) want error containing %q, got nil", tc.in, tc.err)
				}
				if !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("Expand(%q) error = %v, want substring %q", tc.in, err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Expand(%q) unexpected error: %v", tc.in, err)
			}
			if string(got) != tc.want {
				t.Fatalf("Expand(%q) = %q, want %q", tc.in, string(got), tc.want)
			}
		})
	}
}

func TestExpand_FastPathReturnsSameSlice(t *testing.T) {
	in := []byte("no placeholders here")
	out, err := Expand(in, lookup(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Fast path returns the input slice header unchanged (no alloc, no copy).
	if &out[0] != &in[0] {
		t.Errorf("fast path should return the same backing array")
	}
}

func TestExpand_CircularDefaultErrors(t *testing.T) {
	// ${A:-${A}} — A defaults to itself; must error, not recurse forever.
	_, err := Expand([]byte("${A:-${A}}"), lookup(nil))
	if err == nil || !strings.Contains(err.Error(), "circular") {
		t.Fatalf("expected circular error, got %v", err)
	}
}

func TestExpand_DeepNestingErrors(t *testing.T) {
	// 40 levels of ${A:-${A:-...${A:-x}...}} exceeds the 32-deep cap.
	var b strings.Builder
	for i := 0; i < 40; i++ {
		b.WriteString("${A:-")
	}
	b.WriteString("x")
	for i := 0; i < 40; i++ {
		b.WriteString("}")
	}
	_, err := Expand([]byte(b.String()), lookup(nil))
	if err == nil {
		t.Fatal("expected depth error, got nil")
	}
}

func TestExpand_NilLookupUsesOSEnv(t *testing.T) {
	t.Setenv("APP_V3_TEST_VAR", "from-os-env")
	out, err := Expand([]byte("v=${APP_V3_TEST_VAR}"), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "v=from-os-env" {
		t.Errorf("got %q, want v=from-os-env", string(out))
	}
}
