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
	// A reserved/unknown bit (outside allFlagBits, which covers bits 0..2)
	// is preserved as a trailing 0xNN suffix rather than silently dropped,
	// so a Flag carrying bits this build doesn't know about is never mistaken
	// for FlagNone. uint8 makes 1<<7 a safe unknown-bit value.
	unknown := Flag(1 << 7) // 0x80
	assert.Equal(t, "0x80", unknown.String())
	assert.Equal(t, "debug|0x80", (FlagDebug | unknown).String())

	// A combination of several unknown bits shows the combined mask.
	multi := Flag(1<<7 | 1<<6) // 0xc0
	assert.Equal(t, "0xc0", multi.String())
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

func TestParseFlagStrict_AcceptsKnown(t *testing.T) {
	cases := map[string]Flag{
		"":                    FlagNone,
		"normal":              FlagNone,
		"debug":               FlagDebug,
		"DEBUG":               FlagDebug,
		"debug|shadow":        FlagDebug | FlagShadow,
		"debug|stress|shadow": FlagDebug | FlagStress | FlagShadow,
		"|debug|":             FlagDebug,
	}
	for in, want := range cases {
		got, err := ParseFlagStrict(in)
		assert.NoError(t, err, "ParseFlagStrict(%q)", in)
		assert.Equal(t, want, got, "ParseFlagStrict(%q)", in)
	}
}

func TestParseFlagStrict_RejectsUnknown(t *testing.T) {
	// A typo like "shdow" must surface, not silently degrade to FlagNone
	// (which would run shadow traffic as real traffic).
	f, err := ParseFlagStrict("shdow")
	assert.Error(t, err)
	assert.Equal(t, FlagNone, f) // nothing parsed before the error

	// Known-then-unknown: the bits parsed so far are returned alongside the
	// error so the caller still sees partial state if it wants.
	f, err = ParseFlagStrict("debug|shdow")
	assert.Error(t, err)
	assert.Equal(t, FlagDebug, f)

	// normal|debug is NOT a conflict at the parse layer: normal is the
	// no-op token, debug is set.
	f, err = ParseFlagStrict("normal|debug")
	assert.NoError(t, err)
	assert.Equal(t, FlagDebug, f)
}

func TestParseFlagStrict_RoundTripsString(t *testing.T) {
	flags := []Flag{
		FlagNone, FlagDebug, FlagStress, FlagShadow,
		FlagDebug | FlagShadow, FlagDebug | FlagStress | FlagShadow,
	}
	for _, f := range flags {
		got, err := ParseFlagStrict(f.String())
		assert.NoError(t, err, "ParseFlagStrict(%q)", f.String())
		assert.Equal(t, f, got)
	}
}
