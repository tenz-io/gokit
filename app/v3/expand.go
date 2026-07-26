package app

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// Expand replaces ${...} placeholders in bs using lookup to resolve variable
// values. It is the environment-interpolation step ReadConfig runs on the raw
// config bytes before unmarshalling, so sensitive values (passwords, DSNs,
// API keys) stay out of the config file and come from the process environment
// (loaded by WithDotEnvConfig from .env, or set directly).
//
// Supported syntax (shell / docker-compose compatible):
//
//	${VAR}            the value of VAR
//	${VAR:-default}   default when VAR is unset OR empty, else VAR's value
//	${VAR:?msg}       error when VAR is unset OR empty (msg optional)
//
// VAR must match [A-Za-z_][A-Za-z0-9_]*. A bare $ not followed by '{' is left
// untouched (so "$5" or "$HOME" in prose is not mangled); only the braced form
// is interpolated. default may itself contain ${...} and is expanded
// recursively against the same lookup.
//
// Expansion is strict by default: a ${VAR} whose VAR is unset and has no
// :-default / :?error clause is an error, so a missing sensitive value fails
// startup rather than leaking the literal "${VAR}" into the decoded config.
//
// bs with no '$' is returned as-is (fast path, zero allocation).
func Expand(bs []byte, lookup func(string) (string, bool)) ([]byte, error) {
	if !bytes.ContainsRune(bs, '$') {
		return bs, nil // fast path: nothing to expand
	}
	if lookup == nil {
		lookup = os.LookupEnv
	}
	ex := expander{lookup: lookup, depth: 0, maxDepth: 32}
	out, err := ex.expand(string(bs), nil)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// expander carries the recursion guard for default-value expansion. The
// `seen` slice tracks placeholder variable names currently being expanded so
// "${A:-${A}}" or two mutually-referencing defaults can't loop forever.
type expander struct {
	lookup   func(string) (string, bool)
	seen     []string
	depth    int
	maxDepth int
}

// expand resolves placeholders in s. seen accumulates variable names entered
// on the current expansion path for cycle detection.
func (ex *expander) expand(s string, seen []string) (string, error) {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		// Find the next '$'. Everything up to it is literal.
		dollar := strings.IndexByte(s[i:], '$')
		if dollar < 0 {
			b.WriteString(s[i:])
			break
		}
		// Copy the literal run before '$'.
		b.WriteString(s[i : i+dollar])
		j := i + dollar // index of '$'
		// Require '{' immediately after '$'; otherwise leave '$' literal.
		if j+1 >= len(s) || s[j+1] != '{' {
			b.WriteByte(s[j])
			i = j + 1
			continue
		}
		// Find the matching '}'. Nesting matters: a default value may itself
		// contain ${...} (e.g. ${A:-${B}}), so we scan for the brace that
		// balances the opening ${ , counting nested ${ opens.
		close, ok := matchClose(s, j) // j = index of '$'
		if !ok {
			return "", fmt.Errorf("app: unterminated placeholder in config: %q", s[j:])
		}
		inner := s[j+2 : j+2+close] // text between ${ and }
		i = j + 2 + close + 1       // resume after '}'

		val, err := ex.resolve(inner, seen)
		if err != nil {
			return "", err
		}
		b.WriteString(val)
	}
	return b.String(), nil
}

// resolve parses one placeholder's inner text (between ${ and }) into its
// name and optional operator (:- or :?), then produces the final value.
func (ex *expander) resolve(inner string, seen []string) (string, error) {
	// Split on the first ':' if present, but only when followed by '-' or '?' —
	// otherwise ':' is part of a value (e.g. a URL with a port). Find :- or :?.
	name, op, arg := "", byte(0), ""
	if idx := indexOperator(inner); idx >= 0 {
		name = inner[:idx]
		op = inner[idx+1] // '-' or '?'
		arg = inner[idx+2:]
	} else {
		name = inner
	}

	if !validName(name) {
		return "", fmt.Errorf("app: invalid variable name %q in placeholder", name)
	}

	// Cycle guard: if this name is already being expanded on the current
	// path (we are inside its own default expression), it's a circular
	// reference like ${A:-${A}} — report rather than recurse.
	for _, n := range seen {
		if n == name {
			return "", fmt.Errorf("app: circular config placeholder reference to %q", name)
		}
	}

	val, set := ex.lookup(name)

	switch op {
	case 0: // plain ${VAR}
		if !set || val == "" {
			return "", fmt.Errorf("app: config placeholder ${%s} is unset or empty", name)
		}
		return val, nil
	case '-': // ${VAR:-default}
		if set && val != "" {
			return val, nil
		}
		return ex.expandDefault(arg, name, seen)
	case '?': // ${VAR:?msg}
		if set && val != "" {
			return val, nil
		}
		if strings.TrimSpace(arg) != "" {
			return "", fmt.Errorf("app: config variable %s: %s", name, arg)
		}
		return "", fmt.Errorf("app: config variable %s is required but unset or empty", name)
	}
	return val, nil
}

// expandDefault recursively expands the default expression, guarding against
// cycles through the variable being defaulted.
func (ex *expander) expandDefault(arg, name string, seen []string) (string, error) {
	if ex.depth >= ex.maxDepth {
		return "", fmt.Errorf("app: config placeholder nesting too deep (max %d)", ex.maxDepth)
	}
	for _, n := range seen {
		if n == name {
			return "", fmt.Errorf("app: circular config placeholder reference to %q", name)
		}
	}
	// Only recurse when the default itself contains a placeholder; otherwise
	// return the literal (common case: ${DB_PASS:-fallback}).
	if !strings.Contains(arg, "${") {
		return arg, nil
	}
	ex.depth++
	defer func() { ex.depth-- }()
	return ex.expand(arg, append(seen, name))
}

// indexOperator returns the index of ':' in inner when it is followed by '-'
// or '?', i.e. the ':-' or ':?' separator. Returns -1 when inner has no such
// operator. The separator is the first occurrence so a default value may
// contain colons (e.g. ${URL:-http://host:8080}).
func indexOperator(inner string) int {
	for i := 0; i < len(inner)-1; i++ {
		if inner[i] == ':' && (inner[i+1] == '-' || inner[i+1] == '?') {
			return i
		}
	}
	return -1
}

// matchClose scans s starting just after the '${' at index dollarPos and
// returns the byte offset (relative to s[dollarPos+2]) of the '}' that closes
// this placeholder. It tracks nested '${' opens so a default value that
// contains its own placeholder (e.g. ${A:-${B}}) is matched against the
// correct closing brace, not the inner one.
func matchClose(s string, dollarPos int) (int, bool) {
	depth := 1 // we are inside one open ${ already
	for k := dollarPos + 2; k < len(s); k++ {
		switch {
		case s[k] == '$' && k+1 < len(s) && s[k+1] == '{':
			depth++
			k++ // consume the '{'
		case s[k] == '}':
			depth--
			if depth == 0 {
				return k - (dollarPos + 2), true
			}
		}
	}
	return 0, false
}

// validName reports whether name is a shell-style identifier: first char
// letter or underscore, rest letters/digits/underscore. Empty is invalid.
func validName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			continue
		}
		if i > 0 && r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}
