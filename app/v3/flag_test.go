package app

import (
	"errors"
	"flag"
	"strings"
	"testing"
	"time"
)

func TestParseFlags_Defaults(t *testing.T) {
	fs, err := ParseFlags("test", nil, []string{})
	if err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	if got := fs.String(FlagNameConfig); got != "config/app.yaml" {
		t.Errorf("config default = %q, want config/app.yaml", got)
	}
	if got := fs.Int(FlagNameAdminPort); got != 8085 {
		t.Errorf("admin port default = %d, want 8085", got)
	}
	if got := fs.Bool(FlagNameLoggingFile); got != true {
		t.Errorf("logging-file default = %v, want true", got)
	}
}

func TestParseFlags_ArgvOverridesDefault(t *testing.T) {
	args := []string{"-config", "prod.yaml", "-admin-port", "9090", "-verbose", "-logging-file=false"}
	fs, err := ParseFlags("test", nil, args)
	if err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	if got := fs.String(FlagNameConfig); got != "prod.yaml" {
		t.Errorf("config = %q, want prod.yaml", got)
	}
	if got := fs.Int(FlagNameAdminPort); got != 9090 {
		t.Errorf("admin port = %d, want 9090", got)
	}
	if got := fs.Bool(FlagNameVerbose); got != true {
		t.Errorf("verbose = %v, want true", got)
	}
	if got := fs.Bool(FlagNameLoggingFile); got != false {
		t.Errorf("logging-file = %v, want false", got)
	}
}

func TestParseFlags_CustomSpecsExtendDefaults(t *testing.T) {
	specs := []FlagSpec{StringFlag("env", "test", "Environment")}
	fs, err := ParseFlags("test", specs, []string{"-env", "prod"})
	if err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	if got := fs.String("env"); got != "prod" {
		t.Errorf("env = %q, want prod", got)
	}
	// built-in defaults still present
	if got := fs.Int(FlagNamePort); got != 8080 {
		t.Errorf("port = %d, want 8080 (default should remain)", got)
	}
}

func TestParseFlags_CustomSpecOverridesDefault(t *testing.T) {
	// A caller spec whose name collides with a built-in flag overrides it.
	specs := []FlagSpec{IntFlag(FlagNamePort, 3000, "override port")}
	fs, err := ParseFlags("test", specs, []string{})
	if err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	if got := fs.Int(FlagNamePort); got != 3000 {
		t.Errorf("port = %d, want overridden 3000", got)
	}
}

func TestParseFlags_BadValueIsErrorNotExit(t *testing.T) {
	// A non-int admin-port must surface as an error, not call os.Exit.
	_, err := ParseFlags("test", nil, []string{"-admin-port", "notanint"})
	if err == nil {
		t.Fatal("expected error for bad int, got nil")
	}
	if !strings.Contains(err.Error(), "parse flags") {
		t.Errorf("error should mention parse flags, got: %v", err)
	}
}

func TestParseFlags_HelpIsErrorNotExit(t *testing.T) {
	// -h/--help should come back as flag.ErrHelp wrapped, never os.Exit.
	_, err := ParseFlags("test", nil, []string{"-h"})
	if err == nil {
		t.Fatal("expected ErrHelp for -h, got nil")
	}
	if !errors.Is(err, flag.ErrHelp) {
		t.Errorf("error should wrap flag.ErrHelp, got: %v", err)
	}
}

func TestParseFlags_DuplicateSpecName(t *testing.T) {
	specs := []FlagSpec{
		StringFlag("env", "a", "x"),
		StringFlag("env", "b", "x"),
	}
	if _, err := ParseFlags("test", specs, []string{}); err == nil {
		t.Fatal("expected error for duplicate flag name, got nil")
	}
}

func TestParseFlags_EnvDefault(t *testing.T) {
	t.Setenv("MY_APP_ENV", "staging")
	specs := []FlagSpec{{Name: "env", Kind: FlagKindString, Default: "test", Env: "MY_APP_ENV", Usage: "env"}}
	fs, err := ParseFlags("test", specs, []string{})
	if err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	if got := fs.String("env"); got != "staging" {
		t.Errorf("env = %q, want staging from env", got)
	}
	// argv still wins over env.
	fs2, _ := ParseFlags("test", specs, []string{"-env", "prod"})
	if got := fs2.String("env"); got != "prod" {
		t.Errorf("env = %q, want prod (argv wins)", got)
	}
}

func TestParseFlags_DurationFlag(t *testing.T) {
	specs := []FlagSpec{DurationFlag("timeout", 5*time.Second, "timeout")}
	fs, err := ParseFlags("test", specs, []string{"-timeout", "30s"})
	if err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	if got := fs.Duration("timeout"); got != 30*time.Second {
		t.Errorf("timeout = %v, want 30s", got)
	}
}

func TestFlags_UnknownReturnsZero(t *testing.T) {
	fs, _ := ParseFlags("test", nil, []string{})
	if got := fs.String("nope"); got != "" {
		t.Errorf("unknown string = %q, want empty", got)
	}
	if got := fs.Int("nope"); got != 0 {
		t.Errorf("unknown int = %d, want 0", got)
	}
	if fs.IsSet("nope") {
		t.Errorf("IsSet(unknown) = true, want false")
	}
	if fs.IsSet(FlagNameConfig) == false {
		t.Errorf("IsSet(config) = false, want true")
	}
}

func TestFlags_Print(t *testing.T) {
	fs, _ := ParseFlags("test", nil, []string{"-config", "x.yaml"})
	var b strings.Builder
	fs.Print(&b)
	out := b.String()
	if !strings.Contains(out, "config: x.yaml") {
		t.Errorf("Print output missing config value:\n%s", out)
	}
}
