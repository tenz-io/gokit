package tracer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlag_Basic(t *testing.T) {
	cases := map[string]Flag{
		"":                    FlagNone,
		"   ":                 FlagNone,
		"normal":              FlagNone,
		"debug":               FlagDebug,
		"DEBUG":               FlagDebug, // case-folded
		" stress ":            FlagStress,
		"shadow":              FlagShadow,
		"debug|shadow":        FlagDebug | FlagShadow,
		"debug|stress|shadow": FlagDebug | FlagStress | FlagShadow,
		"|debug|":             FlagDebug, // empty tokens ignored
		"debug|bogus":         FlagDebug, // unknown tokens ignored, not zeroed
		"bogus|alsobogus":     FlagNone,
	}
	for in, want := range cases {
		assert.Equal(t, want, ParseFlag(in), "ParseFlag(%q)", in)
	}
}

func TestParseFlag_RoundTripsString(t *testing.T) {
	// For every set of known bits, ParseFlag(f.String()) == f.
	flags := []Flag{
		FlagNone,
		FlagDebug,
		FlagStress,
		FlagShadow,
		FlagDebug | FlagStress,
		FlagDebug | FlagShadow,
		FlagStress | FlagShadow,
		FlagDebug | FlagStress | FlagShadow,
	}
	for _, f := range flags {
		assert.Equal(t, f, ParseFlag(f.String()), "ParseFlag(%q)", f.String())
	}
}

func TestFlagString(t *testing.T) {
	assert.Equal(t, "none", FlagNone.String())
	assert.Equal(t, "debug", FlagDebug.String())
	assert.Equal(t, "shadow", FlagShadow.String())
	assert.Equal(t, "debug|shadow", (FlagDebug | FlagShadow).String())
	assert.Equal(t, "debug|stress|shadow", (FlagDebug | FlagStress | FlagShadow).String())
}

func TestFlagString_StableForUnknownBits(t *testing.T) {
	// A reserved/unknown bit is dropped from the rendered name, not panicked.
	// 1<<6 sits outside allFlagBits (which only covers bits 0..2) yet stays
	// inside int8's range.
	unknown := Flag(1 << 6)
	assert.Equal(t, "none", unknown.String())
	assert.Equal(t, "debug", (FlagDebug | unknown).String())
}

func TestFlagGoString(t *testing.T) {
	assert.Equal(t, "Flag(debug|shadow)", (FlagDebug | FlagShadow).GoString())
	assert.Equal(t, "Flag(none)", FlagNone.GoString())
}

func TestFlagNames(t *testing.T) {
	assert.Nil(t, FlagNone.Names())
	assert.Equal(t, []string{"debug"}, FlagDebug.Names())
	assert.Equal(t, []string{"debug", "shadow"}, (FlagDebug | FlagShadow).Names())
	assert.Equal(t,
		[]string{"debug", "stress", "shadow"},
		(FlagDebug | FlagStress | FlagShadow).Names(),
	)
}

func TestFlagNames_IndependentOfBitOrder(t *testing.T) {
	// Order of OR-ing must not change Names (flagTable order governs).
	assert.Equal(t,
		(FlagShadow | FlagDebug).Names(),
		(FlagDebug | FlagShadow).Names(),
	)
}
