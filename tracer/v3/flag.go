package tracer

// Flag is a bitmask of traffic-mode flags carried through a request's
// context.
//
// A Flag is cheap (uint8) and safe to share; the zero value FlagNone carries
// no mode. Build a combination with the bitwise OR operator:
//
//	f := tracer.FlagDebug | tracer.FlagShadow
//
// The known flags, their names and aliases live in flagTable — that single
// table is the source of truth for parsing ([ParseFlag]) and rendering
// ([Flag.String], [Flag.Names]) so adding a mode is a one-line change. uint8
// (unsigned) gives eight usable flag bits and avoids the sign-bit overflow
// trap that a signed int8 would hit at bit 7.
type Flag uint8

// Predefined traffic-mode flags.
const (
	// FlagNone is the zero Flag: no mode set.
	FlagNone Flag = 0
	// FlagDebug enables debug mode: more verbose logging / tracing.
	FlagDebug Flag = 1 << iota
	// FlagStress marks stress-test traffic: isolate metrics, skip side
	// effects, etc.
	FlagStress
	// FlagShadow marks shadow traffic (record/replay): the request should
	// not produce real side effects.
	FlagShadow
)

// flagDef describes one known flag for the registry table.
type flagDef struct {
	flag Flag
	// name is the canonical string form, used by ParseFlag and String.
	name string
	// aliases are accepted by ParseFlag in addition to name (case-folded).
	// May be nil.
	aliases []string
}

// flagTable is the single source of truth for known flags. Order matters:
// String renders bits in this order, and the highest set bit wins the "all
// known" fast path in String. Add a flag here and parsing/rendering pick it
// up automatically.
var flagTable = []flagDef{
	{flag: FlagDebug, name: "debug"},
	{flag: FlagStress, name: "stress"},
	{flag: FlagShadow, name: "shadow"},
}

// Is reports whether flag f has every bit of x set. FlagNone is not a real
// flag: Is(FlagNone) returns false (a flag set is never "none"). To test for
// the absence of all modes, compare f == FlagNone directly.
func (f Flag) Is(x Flag) bool {
	if x == FlagNone {
		return false
	}
	return f&x == x
}

// HasAll reports whether f contains every bit of flags. It is an alias for
// Is kept for readability at call sites that pass a combined mask.
// HasAll(FlagNone) returns false for the same reason Is does.
func (f Flag) HasAll(flags Flag) bool { return f.Is(flags) }

// IsDebug reports whether FlagDebug is set.
func (f Flag) IsDebug() bool { return f.Is(FlagDebug) }

// IsStress reports whether FlagStress is set.
func (f Flag) IsStress() bool { return f.Is(FlagStress) }

// IsShadow reports whether FlagShadow is set.
func (f Flag) IsShadow() bool { return f.Is(FlagShadow) }

// Names returns the canonical names of every set bit in f, in flagTable
// order. Returns nil for FlagNone. It is the structured counterpart to
// String (which joins the same names with "|").
func (f Flag) Names() []string {
	if f == FlagNone {
		return nil
	}
	var names []string
	for _, d := range flagTable {
		if f.Is(d.flag) {
			names = append(names, d.name)
		}
	}
	return names
}
