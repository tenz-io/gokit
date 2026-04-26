package logger

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestConfigure(t *testing.T) {
	Configure(Config{
		Level:   DebugLevel,
		Console: false,
	})
	defer Configure(defaultConfig) // restore

	// Should not panic
	Debug("hello")
	Info("world")
	Warn("warning")
	Error("error")
}

func TestEntry(t *testing.T) {
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

func TestEntry_With(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	child := e.With("user", "bob", "age", 42)
	child.Info("hello")

	if e == child {
		t.Error("With should return a new Entry")
	}
}

func TestEntry_WithError(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	child := e.WithError(errors.New("boom"))
	child.Warn("something failed")
}

func TestEntry_WithRequestID(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	child := e.WithRequestID("req-123")
	child.Info("processing")
}

func TestEntry_WithRequestID_Empty(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	child := e.WithRequestID("") // should return same entry
	if child.Enabled(DebugLevel) != e.Enabled(DebugLevel) {
		t.Error("empty request ID should not change entry")
	}
}

func TestEntry_StartTraffic_Disabled(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	rec := e.StartTraffic("test")
	if rec != nil {
		t.Error("traffic should be nil when not configured")
	}
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
	e := FromContext(nil)
	if e == nil {
		t.Fatal("FromContext(nil) should return global logger")
	}
}

func TestCopyToContext(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	ctx1 := WithLogger(context.Background(), e)
	ctx2 := CopyToContext(ctx1, context.Background())
	e2 := FromContext(ctx2)
	if e2 == nil {
		t.Fatal("CopyToContext failed")
	}
}

func TestSetLevel(t *testing.T) {
	old := GetLevel()
	SetLevel(ErrorLevel)
	if GetLevel() != ErrorLevel {
		t.Error("SetLevel should change level")
	}
	SetLevel(old)
}

func TestTraffic_End(t *testing.T) {
	e := NewEntry(Config{
		Level:   DebugLevel,
		Console: false,
		Traffic: true,
	})
	rec := e.StartTraffic("testCmd")
	if rec == nil {
		t.Skip("traffic requires filesystem")
	}
	time.Sleep(time.Millisecond)
	rec.End(map[string]any{"result": "ok"}, "200")

	rec2 := e.StartTraffic("failCmd")
	rec2.EndWithError(errors.New("timeout"))
}

func TestData(t *testing.T) {
	// Should not panic even when traffic is not configured
	Data(&Traffic{
		Typ:  TrafficTypRecv,
		Cmd:  "test",
		Code: "200",
		Msg:  "ok",
		Cost: time.Second,
	})
	Data(nil) // no panic
}

func TestPackageFunctions(t *testing.T) {
	Configure(Config{Level: DebugLevel, Console: false})
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
