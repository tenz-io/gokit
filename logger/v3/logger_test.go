package logger

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// quietConfig builds a config that writes to a tmp directory and stays quiet
// on the console so test output stays clean.
func quietConfig(t *testing.T) (Config, string) {
	t.Helper()
	dir := t.TempDir()
	return Config{
		Level:    DebugLevel,
		Console:  false,
		FilePath: dir,
	}, dir
}

func TestConfigure(t *testing.T) {
	cfg, _ := quietConfig(t)
	Configure(cfg)
	defer Configure(defaultConfig)

	Debug("hello")
	Info("world")
	Warn("warning")
	Error("error")
}

func TestEntry_AllLevels(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})

	e.Debug("debug")
	e.Debugf("debug %s", "formatted")
	e.Info("info")
	e.Infof("info %d", 42)
	e.Warn("warn")
	e.Warnf("warn %v", true)
	e.Error("error")
	e.Errorf("error %s", "details")

	if !e.Enabled(DebugLevel) {
		t.Error("should be enabled at DebugLevel")
	}
}

func TestEntry_Disabled(t *testing.T) {
	e := NewEntry(Config{Level: ErrorLevel, Console: false})
	if e.Enabled(DebugLevel) || e.Enabled(InfoLevel) || e.Enabled(WarnLevel) {
		t.Error("only ErrorLevel should be enabled")
	}
	if !e.Enabled(ErrorLevel) {
		t.Error("ErrorLevel should be enabled")
	}
}

func TestEntry_WithReturnsNewInstance(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	child := e.With("user", "bob", "age", 42)
	child.Info("hello")

	if e == child {
		t.Error("With should return a new Entry")
	}
}

func TestEntry_WithError(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	if child := e.WithError(nil); child != e {
		t.Error("WithError(nil) should return the same entry")
	}
	e.WithError(errors.New("boom")).Warn("something failed")
}

func TestEntry_WithRequestID_Empty(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	if child := e.WithRequestID(""); child != e {
		t.Error("WithRequestID('') should return the same entry")
	}
	e.WithRequestID("req-123").Info("processing")
}

func TestEntry_StartTraffic_Disabled(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	if rec := e.StartTraffic("test"); rec != nil {
		t.Error("traffic should be nil when not configured")
	}
}

func TestEntry_StartTraffic_NilSafe(t *testing.T) {
	var rec *TrafficRec                      // nil
	rec.End(map[string]any{"ok": true}, "0") // must not panic
	rec.EndWithError(errors.New("x"))        // must not panic
}

// TestSetLevel_TakesEffect is the regression test for the v2 bug where
// SetLevel was a no-op because the AtomicLevel was not wired into the core.
func TestSetLevel_TakesEffect(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	if got := e.GetLevel(); got != DebugLevel {
		t.Fatalf("initial level = %v, want DebugLevel", got)
	}

	e.SetLevel(ErrorLevel)
	if got := e.GetLevel(); got != ErrorLevel {
		t.Fatalf("after SetLevel = %v, want ErrorLevel", got)
	}
	if e.Enabled(DebugLevel) || e.Enabled(InfoLevel) || e.Enabled(WarnLevel) {
		t.Error("after SetLevel(ErrorLevel), lower levels should be disabled")
	}
	if !e.Enabled(ErrorLevel) {
		t.Error("ErrorLevel should still be enabled")
	}
}

// TestFileRouting_ByLevel verifies the v2 bug fix: each file only contains
// entries at or above its threshold, so error.log never holds Info lines.
func TestFileRouting_ByLevel(t *testing.T) {
	cfg, dir := quietConfig(t)
	e := NewEntry(cfg)

	e.Debug("dbg-msg")
	e.Info("info-msg")
	e.Warn("warn-msg")
	e.Error("err-msg")

	// Sync underlying writers before reading.
	if err := e.Logger().Sync(); err != nil && !strings.Contains(err.Error(), "inappropriate") {
		t.Logf("sync: %v", err)
	}

	cases := []struct {
		file string
		want []string // substrings that MUST appear
		skip []string // substrings that MUST NOT appear
	}{
		{"debug.log", []string{"dbg-msg", "info-msg", "warn-msg", "err-msg"}, nil},
		{"info.log", []string{"info-msg", "warn-msg", "err-msg"}, []string{"dbg-msg"}},
		{"warn.log", []string{"warn-msg", "err-msg"}, []string{"dbg-msg", "info-msg"}},
		{"error.log", []string{"err-msg"}, []string{"dbg-msg", "info-msg", "warn-msg"}},
	}
	for _, c := range cases {
		path := filepath.Join(dir, c.file)
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", c.file, err)
		}
		body := string(got)
		for _, s := range c.want {
			if !strings.Contains(body, s) {
				t.Errorf("%s: want substring %q", c.file, s)
			}
		}
		for _, s := range c.skip {
			if strings.Contains(body, s) {
				t.Errorf("%s: must not contain %q", c.file, s)
			}
		}
	}
}

func TestTraffic_End(t *testing.T) {
	dir := t.TempDir()
	e := NewEntry(Config{
		Level:       DebugLevel,
		Console:     false,
		Traffic:     true,
		TrafficPath: dir,
	})
	rec := e.StartTraffic("testCmd")
	if rec == nil {
		t.Skip("traffic requires filesystem")
	}
	time.Sleep(time.Millisecond)
	rec.End(map[string]any{"result": "ok"}, "200")

	rec2 := e.StartTraffic("failCmd").WithTyp(TrafficTypSend)
	rec2.EndWithError(errors.New("timeout"))
}

func TestContext(t *testing.T) {
	ctx := WithLogger(context.Background(), L())
	e := FromContext(ctx)
	if e == nil {
		t.Fatal("FromContext returned nil")
	}
	e.Info("from context")
}

func TestContext_Nil(t *testing.T) {
	if e := FromContext(nil); e == nil {
		t.Fatal("FromContext(nil) should return global logger")
	}
}

func TestCopyToContext(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	ctx1 := WithLogger(context.Background(), e)
	ctx2 := CopyToContext(ctx1, context.Background())
	if FromContext(ctx2) == nil {
		t.Fatal("CopyToContext failed")
	}
}

func TestPackageFunctions(t *testing.T) {
	cfg, _ := quietConfig(t)
	Configure(cfg)
	defer Configure(defaultConfig)

	Info("package info")
	Infof("package infof %d", 1)
	Debug("package debug")
	Warn("package warn")
	Error("package error")

	With("key", "val").Info("with fields")
	WithError(errors.New("err")).Warn("with error")
	WithRequestID("rid_1").Info("with request id")
}

func TestJSONEncoding(t *testing.T) {
	dir := t.TempDir()
	e := NewEntry(Config{
		Level:    DebugLevel,
		Encoding: JSONEncoding,
		Console:  false,
		FilePath: dir,
	})
	e.Infow("json-msg", "k", "v")

	if err := e.Logger().Sync(); err != nil && !strings.Contains(err.Error(), "inappropriate") {
		t.Logf("sync: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "info.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "{") {
		t.Error("JSON encoding should produce JSON object lines")
	}
}
