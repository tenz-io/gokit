package logger

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestValidateConfigDefaultsAndOverrides(t *testing.T) {
	cfg := Config{MaxSize: -1, MaxAge: 0, MaxBackups: -2, CallerSkip: -1}
	validateConfig(&cfg)
	if cfg.MaxSize != 100 || cfg.MaxAge != 7 || cfg.MaxBackups != 10 {
		t.Fatalf("main defaults = size:%d age:%d backups:%d", cfg.MaxSize, cfg.MaxAge, cfg.MaxBackups)
	}
	if cfg.TrafficMaxSize != 100 || cfg.TrafficMaxAge != 7 || cfg.TrafficMaxBackups != 10 {
		t.Fatalf("traffic defaults = size:%d age:%d backups:%d", cfg.TrafficMaxSize, cfg.TrafficMaxAge, cfg.TrafficMaxBackups)
	}
	if cfg.Encoding != ConsoleEncoding {
		t.Fatalf("encoding = %q, want console", cfg.Encoding)
	}
	if cfg.CallerSkip != 0 {
		t.Fatalf("caller skip = %d, want 0", cfg.CallerSkip)
	}

	cfg = Config{
		MaxSize: 1, MaxAge: 2, MaxBackups: 3,
		TrafficMaxSize: 4, TrafficMaxAge: 5, TrafficMaxBackups: 6,
		Encoding: JSONEncoding,
	}
	validateConfig(&cfg)
	if cfg.MaxSize != 1 || cfg.MaxAge != 2 || cfg.MaxBackups != 3 ||
		cfg.TrafficMaxSize != 4 || cfg.TrafficMaxAge != 5 || cfg.TrafficMaxBackups != 6 ||
		cfg.Encoding != JSONEncoding {
		t.Fatalf("validateConfig overwrote explicit values: %#v", cfg)
	}
}

func TestFilePathFailureUsesNoOpCore(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	e := NewEntry(Config{
		Level:    DebugLevel,
		Console:  false,
		FilePath: filepath.Join(file, "child"),
	}).(*logEntry)
	if e.base.Desugar().Core().Enabled(zapcore.Level(DebugLevel)) {
		t.Fatal("failed file sink should not fall back to console")
	}
}

func TestAllConfigOptions(t *testing.T) {
	tc := &TrimConfig{ArrLimit: 1}
	cfg := defaultConfig
	opts := []ConfigOption{
		WithLevel(ErrorLevel), WithEncoding(JSONEncoding), WithConsole(false),
		WithFilePath("logs"), WithMaxSize(1), WithMaxAge(2), WithMaxBackups(3),
		WithCaller(true), WithCallerSkip(4), WithTraffic(true), WithTrafficPath("traffic"),
		WithTrafficMaxSize(5), WithTrafficMaxAge(6), WithTrafficMaxBackups(7),
		WithTrimConfig(tc),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.Level != ErrorLevel || cfg.Encoding != JSONEncoding || cfg.Console || cfg.FilePath != "logs" ||
		cfg.MaxSize != 1 || cfg.MaxAge != 2 || cfg.MaxBackups != 3 || !cfg.Caller || cfg.CallerSkip != 4 ||
		!cfg.Traffic || cfg.TrafficPath != "traffic" || cfg.TrafficMaxSize != 5 ||
		cfg.TrafficMaxAge != 6 || cfg.TrafficMaxBackups != 7 || cfg.Trimmer != tc {
		t.Fatalf("options produced unexpected config: %#v", cfg)
	}
}
