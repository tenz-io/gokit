package tracer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFlagTable_Invariants locks down the claim that flagTable is the single
// source of truth: each entry must be a single bit, distinct from every other
// entry's bit, and its name and aliases must be globally unique (so two
// entries can't both claim "debug"). If someone adds a flag, these checks
// catch a misregistration at test time rather than as a silent parse bug.
func TestFlagTable_Invariants(t *testing.T) {
	seenBits := make(map[Flag]string, len(flagTable))
	seenNames := make(map[string]Flag, len(flagTable)*2)

	for _, d := range flagTable {
		// Each flag must be a single bit (power of two) and non-zero.
		assert.NotZero(t, d.flag, "flagTable entry %+v has zero flag", d)
		assert.Zero(t, d.flag&(d.flag-1), "flagTable entry %+v is not a single bit", d)

		// The bit must not be claimed by another entry.
		if prev, dup := seenBits[d.flag]; dup {
			t.Errorf("flag bit %d claimed by both %q and %q", d.flag, prev, d.name)
		}
		seenBits[d.flag] = d.name

		// Name and every alias must be globally unique (case-folded).
		for _, tok := range append([]string{d.name}, d.aliases...) {
			key := lowerASCII(tok)
			if _, dup := seenNames[key]; dup {
				t.Errorf("token %q (flag %q) is registered more than once", tok, d.name)
			}
			seenNames[key] = d.flag
		}
	}
}

// TestFlagTable_CoversDefinedConstants ensures every iota-defined flag
// constant appears in flagTable, so ParseFlag/String actually know about it.
// Adding a const without a table entry would otherwise silently make that
// flag un-parseable.
func TestFlagTable_CoversDefinedConstants(t *testing.T) {
	defined := []Flag{FlagDebug, FlagStress, FlagShadow}
	known := allFlagBits
	for _, f := range defined {
		assert.True(t, f&known == f, "flag %d is defined but missing from flagTable", f)
	}
}

// lowerASCII is a tiny case-fold for ASCII tokens used here (the parse path
// uses strings.ToLower; we avoid importing strings just to keep this test
// dependency-free and explicit).
func lowerASCII(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}
