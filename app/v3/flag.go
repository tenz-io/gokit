// Package app provides a small application-lifecycle framework: flag parsing,
// ordered initialization with LIFO cleanup, an optional admin HTTP server
// (pprof / metrics / ping), and graceful, signal-driven shutdown.
//
// v3 is a clean rewrite on top of logger/v3 and annotation/v3. It fixes the
// startup hazards of v2:
//   - Flags are immutable value specs; parsing snapshots results into a Flags
//     value instead of mutating caller-defined Flag structs (v2 aliased the
//     caller's fields via &f.Value).
//   - Parse returns an error rather than calling os.Exit, so bad args and
//     -h/--help are testable and never abort the process.
//   - Cleanup runs in LIFO order and is invoked even when a later init fails,
//     so already-acquired resources never leak.
//   - The admin server uses a dedicated *http.ServeMux (never the package-level
//     DefaultServeMux) and shuts down gracefully via Shutdown on exit.
//   - Run returns an exit code instead of calling os.Exit internally; the
//     caller decides whether to os.Exit.
package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

// Flag name constants used by the built-in flags and by the With* initializers.
const (
	FlagNameConfig         = "config"          // path to the config file
	FlagNamePort           = "port"            // service HTTP port
	FlagNameAdminPort      = "admin-port"      // admin HTTP port
	FlagNameLog            = "log"             // log output directory
	FlagNameLoggingFile    = "logging-file"    // enable file logging
	FlagNameLoggingConsole = "logging-console" // enable console logging
	FlagNameVerbose        = "verbose"         // verbose (debug-level) logging
)

// FlagKind identifies a flag's value type so Parse can route the parsed token
// to the right setter and Flags can return a typed value.
type FlagKind int

const (
	FlagKindString FlagKind = iota
	FlagKindInt
	FlagKindBool
	FlagKindDuration
)

// FlagSpec describes one command-line flag as an immutable value: the flag's
// name, kind, default, usage text, and an optional "env" hook for defaulting
// from an environment variable. Unlike v2's *Flag structs, a FlagSpec carries
// no mutable destination — parsing writes into a fresh Flags snapshot, so the
// same FlagSpec slice is safe to reuse across calls and concurrent parses.
type FlagSpec struct {
	Name    string
	Kind    FlagKind
	Default any
	Usage   string
	// Env, when non-empty, supplies the default from the named environment
	// variable when the flag is absent from argv. Precedence: argv > env > Default.
	Env string
}

// StringFlag, IntFlag, BoolFlag and DurationFlag are convenience constructors
// so callers build specs without spelling out FlagKind.
func StringFlag(name, def, usage string) FlagSpec {
	return FlagSpec{Name: name, Kind: FlagKindString, Default: def, Usage: usage}
}

func IntFlag(name string, def int, usage string) FlagSpec {
	return FlagSpec{Name: name, Kind: FlagKindInt, Default: def, Usage: usage}
}

func BoolFlag(name string, def bool, usage string) FlagSpec {
	return FlagSpec{Name: name, Kind: FlagKindBool, Default: def, Usage: usage}
}

func DurationFlag(name string, def time.Duration, usage string) FlagSpec {
	return FlagSpec{Name: name, Kind: FlagKindDuration, Default: def, Usage: usage}
}

// DefaultFlags is the built-in flag set registered by every App unless the
// caller supplies an overriding slice. The With* initializers read these names.
var DefaultFlags = []FlagSpec{
	StringFlag(FlagNameConfig, "config/app.yaml", "Conf file"),
	IntFlag(FlagNamePort, 8080, "HTTP port"),
	IntFlag(FlagNameAdminPort, 8085, "Admin HTTP port"),
	StringFlag(FlagNameLog, "./log", "Log output directory"),
	BoolFlag(FlagNameLoggingFile, true, "Log to file(true/false)"),
	BoolFlag(FlagNameLoggingConsole, false, "Log to console(true/false)"),
	BoolFlag(FlagNameVerbose, false, "Verbose mode(true/false)"),
}

// Flags is the immutable, parsed snapshot returned by ParseFlags. Each lookup
// returns the resolved value and a bool indicating the flag was registered;
// callers that ignore the bool get the zero value for an unknown name, which is
// the common "just read it" path. Flags is safe for concurrent reads.
type Flags struct {
	values map[string]flagValue
}

type flagValue struct {
	kind FlagKind
	str  string
	num  int64
	b    bool
	dur  time.Duration
}

// Lookup returns the raw flagValue and whether the flag was registered. Most
// callers use the typed String/Int/Bool/Duration accessors instead.
func (fs *Flags) Lookup(name string) (flagValue, bool) {
	v, ok := fs.values[name]
	return v, ok
}

// String returns the string value of name, or "" when the flag is absent.
func (fs *Flags) String(name string) string { v, _ := fs.values[name]; return v.str }

// Int returns the int value of name, or 0 when the flag is absent.
func (fs *Flags) Int(name string) int { v, _ := fs.values[name]; return int(v.num) }

// Bool returns the bool value of name, or false when the flag is absent.
func (fs *Flags) Bool(name string) bool { v, _ := fs.values[name]; return v.b }

// Duration returns the duration value of name, or zero when the flag is absent.
func (fs *Flags) Duration(name string) time.Duration {
	v, _ := fs.values[name]
	return v.dur
}

// IsSet reports whether name was a registered flag.
func (fs *Flags) IsSet(name string) bool { _, ok := fs.values[name]; return ok }

// Print writes the resolved flag values to w in a compact "name: value" block.
// Output goes to a caller-supplied io.Writer so tests can capture it; Run
// routes it through the logger.
func (fs *Flags) Print(w io.Writer) {
	fmt.Fprintln(w, "args: ==================")
	for _, f := range DefaultFlags {
		v, ok := fs.values[f.Name]
		if !ok {
			continue
		}
		fmt.Fprintf(w, "%s: %s\n", f.Name, v.display())
	}
	fmt.Fprintln(w, "==================")
}

func (v flagValue) display() string {
	switch v.kind {
	case FlagKindString:
		return v.str
	case FlagKindInt:
		return strconv.FormatInt(v.num, 10)
	case FlagKindBool:
		return strconv.FormatBool(v.b)
	case FlagKindDuration:
		return v.dur.String()
	}
	return v.str
}

// ParseFlags resolves specs against argv (defaulting to os.Args[1:]) and
// returns an immutable Flags snapshot. It never calls os.Exit: a parse error,
// an unknown flag, or -h/--help surfaces as an error (flag.ErrHelp for help) so
// the caller can print usage and decide an exit code.
//
// specs is combined with DefaultFlags when nil; otherwise the caller's specs
// extend DefaultFlags (caller specs take precedence on name collision).
func ParseFlags(name string, specs []FlagSpec, args []string) (*Flags, error) {
	if specs == nil {
		specs = DefaultFlags
	} else {
		// Reject caller-side duplicates before mergeSpecs folds overrides into a
		// single entry (which would hide the dup). A spec may collide with a
		// built-in default (override), but two caller specs with the same name
		// is an error.
		if dup := firstDuplicate(specs); dup != "" {
			return nil, fmt.Errorf("app: duplicate flag %q", dup)
		}
		specs = mergeSpecs(specs)
	}
	if args == nil {
		args = os.Args[1:]
	}

	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	// Silence the default error output; ParseFlags reports errors as values so
	// the caller controls formatting and the exit code.
	fs.SetOutput(io.Discard)

	holders := make(map[string]*flagValue, len(specs))
	for _, s := range specs {
		if s.Name == "" {
			return nil, fmt.Errorf("app: flag with empty name")
		}
		if _, dup := holders[s.Name]; dup {
			return nil, fmt.Errorf("app: duplicate flag %q", s.Name)
		}
		v := &flagValue{kind: s.Kind}
		applySpec(fs, s, v)
		holders[s.Name] = v
	}

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("app: parse flags: %w", err)
	}

	out := make(map[string]flagValue, len(holders))
	for n, h := range holders {
		out[n] = *h
	}
	return &Flags{values: out}, nil
}

// mergeSpecs returns DefaultFlags augmented with specs; a specs entry whose name
// already exists in DefaultFlags replaces it (caller override), otherwise it is
// appended. The returned slice is a fresh copy so the caller's slice is not
// mutated. Callers must have already de-duplicated their own specs.
func mergeSpecs(specs []FlagSpec) []FlagSpec {
	out := make([]FlagSpec, 0, len(DefaultFlags)+len(specs))
	out = append(out, DefaultFlags...)

	seen := make(map[string]int, len(out))
	for i, s := range out {
		seen[s.Name] = i
	}
	for _, s := range specs {
		if idx, ok := seen[s.Name]; ok {
			out[idx] = s
		} else {
			out = append(out, s)
			seen[s.Name] = len(out) - 1
		}
	}
	return out
}

// firstDuplicate returns the name of the first flag appearing twice in specs,
// or "" when all names are distinct. Used by ParseFlags to reject caller-side
// dups before mergeSpecs folds overrides.
func firstDuplicate(specs []FlagSpec) string {
	seen := make(map[string]struct{}, len(specs))
	for _, s := range specs {
		if s.Name == "" {
			continue
		}
		if _, ok := seen[s.Name]; ok {
			return s.Name
		}
		seen[s.Name] = struct{}{}
	}
	return ""
}

// applySpec registers one flag on fs, writing the resolved default (env > spec
// default) into v's fields and binding the flag's pointer to v so Parse fills
// in the argv value.
func applySpec(fs *flag.FlagSet, s FlagSpec, v *flagValue) {
	def := s.Default
	if s.Env != "" {
		if e, ok := os.LookupEnv(s.Env); ok {
			def = e
		}
	}
	switch s.Kind {
	case FlagKindString:
		str, _ := def.(string)
		v.str = str
		fs.Func(s.Name, s.Usage, func(val string) error {
			v.str = val
			return nil
		})
	case FlagKindInt:
		num := toInt(def)
		v.num = num
		fs.Func(s.Name, s.Usage, func(val string) error {
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			v.num = n
			return nil
		})
	case FlagKindBool:
		b := toBool(def)
		v.b = b
		fs.BoolFunc(s.Name, s.Usage, func(val string) error {
			b, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			v.b = b
			return nil
		})
	case FlagKindDuration:
		d := toDuration(def)
		v.dur = d
		fs.Func(s.Name, s.Usage, func(val string) error {
			d, err := time.ParseDuration(val)
			if err != nil {
				return err
			}
			v.dur = d
			return nil
		})
	}
}

func toInt(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int64:
		return x
	case string:
		if n, err := strconv.ParseInt(x, 10, 64); err == nil {
			return n
		}
	}
	return 0
}

func toBool(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		if b, err := strconv.ParseBool(x); err == nil {
			return b
		}
	}
	return false
}

func toDuration(v any) time.Duration {
	switch x := v.(type) {
	case time.Duration:
		return x
	case string:
		if d, err := time.ParseDuration(x); err == nil {
			return d
		}
	}
	return 0
}
