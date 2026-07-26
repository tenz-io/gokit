package tracer

import (
	"fmt"
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

// allFlagBits is the OR of every known flag bit. String uses it to separate
// the named bits of a Flag from any reserved/unknown bits (bits set on a Flag
// that no flagTable entry claims — e.g. a bit a newer version introduced that
// this build does not know about). Unknown bits are rendered as a trailing
// 0xNN suffix rather than silently dropped, so a Flag carrying them is never
// mistaken for FlagNone.
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
// This is the lenient parser, kept for compatibility with the per-transport
// switch logic callers (e.g. ginext) previously hand-rolled. For inbound
// trust boundaries where a typo must not silently downgrade shadow/stress
// traffic to real traffic, use [ParseFlagStrict].
//
// This is the inverse of [Flag.String]:
//
//	tracer.ParseFlag(tracer.FlagDebug.String()) == tracer.FlagDebug
func ParseFlag(s string) Flag {
	f, _ := parseFlag(s, false)
	return f
}

// ParseFlagStrict parses a "|" separated flag string like [ParseFlag] but
// returns an error naming the first unknown token instead of silently
// dropping it. Use this at HTTP/gRPC inbound boundaries where a misspelled
// mode ("shdow") must surface rather than degrade shadow traffic to real
// traffic. The no-op tokens ("normal", "none") and empty tokens are accepted,
// as is the empty string (yields FlagNone, nil error). "none" is the string
// form of FlagNone (see [Flag.String]).
func ParseFlagStrict(s string) (Flag, error) {
	return parseFlag(s, true)
}

// parseFlag is the shared engine for ParseFlag (lenient) and ParseFlagStrict.
// When strict is true, an unknown token produces a non-nil error and parsing
// stops at the first such token (its already-OR'd bits are still returned).
func parseFlag(s string, strict bool) (Flag, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return FlagNone, nil
	}
	var f Flag
	for _, tok := range strings.Split(s, "|") {
		key := strings.ToLower(strings.TrimSpace(tok))
		if key == "" || key == "normal" || key == "none" {
			continue // "normal" / "none" / empty are explicitly the no-op mode
		}
		bit, ok := flagByName[key]
		if !ok {
			if strict {
				return f, fmt.Errorf("tracer: unknown flag %q", tok)
			}
			continue // lenient: ignore unknown tokens
		}
		f |= bit
	}
	return f, nil
}

// String returns the canonical "|" joined form of f's set bits, in flagTable
// order, e.g. "debug|shadow". FlagNone renders as "none".
//
// Bits set on f that no flagTable entry claims (reserved/unknown bits, e.g.
// a flag a newer version introduced that this build does not know about) are
// NOT dropped: they are appended as a trailing 0xNN suffix, so a Flag that
// carries unknown bits is never mistaken for FlagNone when versions disagree.
func (f Flag) String() string {
	known := f & allFlagBits
	unknown := f &^ allFlagBits

	names := known.Names()
	switch {
	case len(names) == 0 && unknown == 0:
		return "none"
	case unknown == 0:
		// only known bits — join names
		return joinNames(names)
	default:
		// some unknown bits remain; render names (if any) then 0xNN.
		// Convert to uint8 BEFORE handing to fmt: Flag implements Stringer,
		// so fmt.Sprintf("%x", unknown) would re-enter Flag.String and
		// recurse forever. The raw uint8 prints as hex with no dispatch.
		prefix := ""
		if len(names) > 0 {
			prefix = joinNames(names) + "|"
		}
		return prefix + fmt.Sprintf("0x%02x", uint8(unknown))
	}
}

// joinNames joins flag names with "|" in flagTable order. Names is assumed
// non-empty and ordered.
func joinNames(names []string) string {
	if len(names) == 1 {
		return names[0]
	}
	var b strings.Builder
	b.Grow(len(names) * 8)
	b.WriteString(names[0])
	for _, n := range names[1:] {
		b.WriteByte('|')
		b.WriteString(n)
	}
	return b.String()
}

// GoString returns the debug-printf form, e.g. "Flag(debug|shadow)" or
// "Flag(none)", so Flag prints readably inside fmt.Printf("%#v", ...).
func (f Flag) GoString() string { return "Flag(" + f.String() + ")" }
