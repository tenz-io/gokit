package tracer

import (
	"strings"
)

// flagByName maps every accepted string token (canonical names plus
// aliases, lowercased) to its flag. Built once at init from flagTable so the
// parse hot path is a single map lookup per token, not a linear scan.
var flagByName = func() map[string]Flag {
	m := make(map[string]Flag, len(flagTable)*2)
	for _, d := range flagTable {
		m[strings.ToLower(d.name)] = d.flag
		for _, a := range d.aliases {
			m[strings.ToLower(a)] = d.flag
		}
	}
	return m
}()

// allFlagBits is the OR of every known flag bit. Used to mask off unknown
// bits when rendering, so a Flag carrying a reserved-but-unnamed bit still
// produces a stable String (the unknown bit is dropped from the name list
// rather than panicking).
var allFlagBits = func() Flag {
	var all Flag
	for _, d := range flagTable {
		all |= d.flag
	}
	return all
}()

// ParseFlag parses a "|" separated flag string such as "debug|shadow" into a
// Flag. Tokens are case-folded and trimmed; unknown tokens are ignored (so
// "debug|bogus" yields FlagDebug, not FlagNone). An empty string yields
// FlagNone.
//
// This is the inverse of [Flag.String]:
//
//	tracer.ParseFlag(tracer.FlagDebug.String()) == tracer.FlagDebug
//
// ParseFlag absorbs the per-transport switch logic that callers (e.g. ginext)
// previously hand-rolled, so a flag string off the wire becomes one call.
func ParseFlag(s string) Flag {
	s = strings.TrimSpace(s)
	if s == "" {
		return FlagNone
	}
	var f Flag
	for _, tok := range strings.Split(s, "|") {
		key := strings.ToLower(strings.TrimSpace(tok))
		if key == "" || key == "normal" {
			continue // "normal" / empty is explicitly the no-op mode
		}
		if bit, ok := flagByName[key]; ok {
			f |= bit
		}
		// Unknown tokens are ignored, matching v2 ginext getTracerFlag semantics.
	}
	return f
}

// String returns the canonical "|" joined form of f's set bits, in flagTable
// order, e.g. "debug|shadow". FlagNone renders as "none". Unknown bits
// (outside allFlagBits) are dropped from the output, so a Flag built with
// reserved bits still renders stably.
func (f Flag) String() string {
	names := f.Names()
	switch len(names) {
	case 0:
		return "none"
	case 1:
		return names[0]
	default:
		var b strings.Builder
		// Each name is short and there are few of them; pre-size generously.
		b.Grow(len(names) * 8)
		b.WriteString(names[0])
		for _, n := range names[1:] {
			b.WriteByte('|')
			b.WriteString(n)
		}
		return b.String()
	}
}

// GoString returns the debug-printf form ("%!x(debug|shadow)") so Flag
// prints readably inside fmt.Printf("%#v", ...). The leading 0x-style hex
// of the bit value makes round-tripping unambiguous when debugging.
func (f Flag) GoString() string { return "Flag(" + f.String() + ")" }
