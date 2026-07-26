package monitor

import "testing"

func TestNormalizeOpt(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"":        valNA,
		"hit":     "hit",
		"miss":    "miss",
		"actives": "actives",
	}
	for in, want := range cases {
		if got := normalizeOpt(in); got != want {
			t.Errorf("normalizeOpt(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeCode(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"":    codeOK, // empty collapses to ok
		"0":   codeOK,
		"1":   codeErr,
		"2":   codeErr, // any non-zero maps to err
		"500": codeErr,
		"ok":  codeErr, // not "0" → err
	}
	for in, want := range cases {
		if got := normalizeCode(in); got != want {
			t.Errorf("normalizeCode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLabelsOf(t *testing.T) {
	t.Parallel()
	got := labelsOf("svc", "getUser", "0", "hit")
	want := map[string]string{
		labelCmd: "svc", labelDsCmd: "getUser", labelCode: "0", labelOpt: "hit",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("labelsOf[%q] = %q, want %q", k, got[k], v)
		}
	}
}
