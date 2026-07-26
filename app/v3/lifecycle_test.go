package app

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tenz-io/gokit/logger/v3"
)

// quietLogger swaps the global logger for a silent one for the duration of a
// test so app.Run's startup/shutdown logs don't clutter test output. It also
// redirects flag printing to io.Discard.
func quietLogger(t *testing.T) {
	t.Helper()
	prev := logger.L()
	logger.ConfigureWithOpts(logger.WithLevel(logger.ErrorLevel), logger.WithConsole(false))
	prevOut := flagOutput
	flagOutput = io.Discard
	t.Cleanup(func() {
		logger.ConfigureWithOpts(logger.WithLevel(logger.InfoLevel), logger.WithConsole(false))
		flagOutput = prevOut
		_ = prev
	})
}

func TestRun_InitOrderAndLIFOCleanup(t *testing.T) {
	quietLogger(t)
	var mu atomic.Pointer[[]string]
	empty := []string{}
	mu.Store(&empty)

	rec := func(s string) { cur := append(*mu.Load(), s); mu.Store(&cur) }

	inits := []InitFunc{
		func(_ *Context, _ any) (CleanFunc, error) {
			rec("init1")
			return func(context.Context) error { rec("clean1"); return nil }, nil
		},
		func(_ *Context, _ any) (CleanFunc, error) {
			rec("init2")
			return func(context.Context) error { rec("clean2"); return nil }, nil
		},
		func(_ *Context, _ any) (CleanFunc, error) {
			rec("init3")
			return func(context.Context) error { rec("clean3"); return nil }, nil
		},
	}

	cfg := Config{
		Name:  "lifo",
		Conf:  &struct{}{},
		Inits: inits,
		Run: func(_ *Context, _ any, errC chan<- error) {
			rec("run-start")
			rec("run-done")
			errC <- nil
		},
	}

	// Send a context-done by having Run complete: we signal via errC-nil.
	// To exercise the signal path we instead cancel through Run completing.
	// Run here completes on its own (errC<-nil), which should trigger cleanup.
	code := Run(cfg, nil, []string{})

	if code != ExitOK {
		t.Fatalf("Run exit = %d, want ExitOK", code)
	}

	got := *mu.Load()
	// want: inits in order, run starts, run done (errC<-nil path), then LIFO cleanup.
	want := []string{"init1", "init2", "init3", "run-start", "run-done", "clean3", "clean2", "clean1"}
	if !equalStringSlices(got, want) {
		t.Errorf("lifecycle order = %v\nwant %v", got, want)
	}
}

func TestRun_InitFailureRunsCollectedCleanups(t *testing.T) {
	quietLogger(t)
	var mu atomic.Pointer[[]string]
	empty := []string{}
	mu.Store(&empty)
	rec := func(s string) { cur := append(*mu.Load(), s); mu.Store(&cur) }

	boom := errors.New("boom")
	inits := []InitFunc{
		func(_ *Context, _ any) (CleanFunc, error) {
			rec("init1")
			return func(context.Context) error { rec("clean1"); return nil }, nil
		},
		func(_ *Context, _ any) (CleanFunc, error) {
			rec("init2")
			return func(context.Context) error { rec("clean2"); return nil }, nil
		},
		func(_ *Context, _ any) (CleanFunc, error) { rec("init3-fail"); return nil, boom },
	}

	cfg := Config{
		Name:  "fail",
		Conf:  &struct{}{},
		Inits: inits,
		Run:   func(*Context, any, chan<- error) {}, // never reached
	}

	code := Run(cfg, nil, []string{})
	if code != ExitSetup {
		t.Fatalf("Run exit = %d, want ExitSetup", code)
	}

	got := *mu.Load()
	want := []string{"init1", "init2", "init3-fail", "clean2", "clean1"}
	if !equalStringSlices(got, want) {
		t.Errorf("cleanup after init failure = %v\nwant %v", got, want)
	}
}

func TestRun_RunErrorReturnsExitRunError(t *testing.T) {
	quietLogger(t)
	cfg := Config{
		Name: "runerr",
		Conf: &struct{}{},
		Run: func(_ *Context, _ any, errC chan<- error) {
			errC <- errors.New("fatal")
		},
	}
	code := Run(cfg, nil, []string{})
	if code != ExitRunError {
		t.Fatalf("Run exit = %d, want ExitRunError", code)
	}
}

func TestRun_SignalReturnsExitSignal(t *testing.T) {
	quietLogger(t)
	started := make(chan struct{})
	cfg := Config{
		Name: "sig",
		Conf: &struct{}{},
		Run: func(c *Context, _ any, errC chan<- error) {
			close(started)
			<-c.Done()
		},
	}
	go func() { <-started; time.Sleep(20 * time.Millisecond); sendInterrupt() }()
	code := Run(cfg, nil, []string{})
	if code != ExitSignal {
		t.Fatalf("Run exit = %d, want ExitSignal", code)
	}
}

func TestWithAdminHTTPServer_GracefulShutdown(t *testing.T) {
	quietLogger(t)
	port := freePort(t)
	specs := []FlagSpec{
		StringFlag(FlagNameConfig, "", ""),
		IntFlag(FlagNamePort, 0, ""),
		IntFlag(FlagNameAdminPort, port, ""),
	}
	flags, err := ParseFlags("admin", specs, []string{})
	if err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := NewContext(ctx, flags)

	clean, err := WithAdminHTTPServer()(c, nil)
	if err != nil {
		t.Fatalf("WithAdminHTTPServer init: %v", err)
	}
	if clean == nil {
		t.Fatal("expected cleanup func, got nil")
	}

	// Sanity-check the resolved port flag before exercising shutdown.
	if got := c.Flags().Int(FlagNameAdminPort); got != port {
		t.Fatalf("admin port = %d, want %d", got, port)
	}

	// Give the listener a tick to bind.
	time.Sleep(50 * time.Millisecond)

	shutdownCtx, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	if err := clean(shutdownCtx); err != nil {
		t.Errorf("admin cleanup: %v", err)
	}
	cancel()
}

func TestWithAdminHTTPServer_InvalidPort(t *testing.T) {
	quietLogger(t)
	specs := []FlagSpec{IntFlag(FlagNameAdminPort, 99999, "")}
	flags, err := ParseFlags("admin", specs, []string{})
	if err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	c := NewContext(context.Background(), flags)
	if _, err := WithAdminHTTPServer()(c, nil); err == nil {
		t.Fatal("expected error for out-of-range admin port, got nil")
	}
}

func TestWithYamlConfig_DecodesAndDefaults(t *testing.T) {
	quietLogger(t)
	dir := t.TempDir()
	path := dir + "/app.yaml"
	if err := writeFile(path, "page_size: 50\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	flags, _ := ParseFlags("yaml", []FlagSpec{
		StringFlag(FlagNameConfig, path, ""),
	}, []string{})
	c := NewContext(context.Background(), flags)

	var conf struct {
		Name     string `yaml:"name" default:"default-name"`
		PageSize int    `yaml:"page_size" default:"10" validate:"required,gt=0,lte=100"`
	}
	if _, err := WithYamlConfig()(c, &conf); err != nil {
		t.Fatalf("WithYamlConfig: %v", err)
	}
	if conf.Name != "default-name" {
		t.Errorf("Name default not applied: %q", conf.Name)
	}
	if conf.PageSize != 50 {
		t.Errorf("PageSize = %d, want 50 from file", conf.PageSize)
	}
}

func TestReadConfig_ValidateFailure(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/bad.yaml"
	if err := writeFile(path, "page_size: 500\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	var conf struct {
		PageSize int `yaml:"page_size" default:"10" validate:"required,gt=0,lte=100"`
	}
	err := ReadConfig(path, &conf, yamlUnmarshal)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestWithYamlConfig_ExpandsEnvPlaceholders(t *testing.T) {
	quietLogger(t)
	dir := t.TempDir()
	path := dir + "/app.yaml"
	// ${SECRET} from env; ${PORT:-8080} uses default; ${REQUIRED:?msg} set.
	content := "name: ${NAME}\nport: ${PORT:-8080}\nrequired: ${REQUIRED:?must set}\n"
	if err := writeFile(path, content); err != nil {
		t.Fatalf("write: %v", err)
	}

	t.Setenv("NAME", "env-name")
	t.Setenv("REQUIRED", "yes")
	// PORT deliberately unset -> exercises :-default.

	flags, _ := ParseFlags("yaml", []FlagSpec{
		StringFlag(FlagNameConfig, path, ""),
	}, []string{})
	c := NewContext(context.Background(), flags)

	var conf struct {
		Name     string `yaml:"name"`
		Port     string `yaml:"port"`
		Required string `yaml:"required"`
	}
	if _, err := WithYamlConfig()(c, &conf); err != nil {
		t.Fatalf("WithYamlConfig: %v", err)
	}
	if conf.Name != "env-name" {
		t.Errorf("Name = %q, want env-name (from env)", conf.Name)
	}
	if conf.Port != "8080" {
		t.Errorf("Port = %q, want 8080 (default)", conf.Port)
	}
	if conf.Required != "yes" {
		t.Errorf("Required = %q, want yes (from env)", conf.Required)
	}
}

func TestReadConfig_UnsetPlaceholderErrors(t *testing.T) {
	quietLogger(t)
	dir := t.TempDir()
	path := dir + "/app.yaml"
	if err := writeFile(path, "db_password: ${DB_PASSWORD}\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	// DB_PASSWORD deliberately unset — must error, not leak the literal.
	osUnsetenv("DB_PASSWORD")
	var conf struct {
		DBPassword string `yaml:"db_password"`
	}
	err := ReadConfig(path, &conf, yamlUnmarshal)
	if err == nil || !strings.Contains(err.Error(), "expand") {
		t.Fatalf("expected expand error, got: %v", err)
	}
	if conf.DBPassword == "${DB_PASSWORD}" {
		t.Error("literal placeholder leaked into decoded config")
	}
}

func TestReadConfig_PlaceholderValueConvertedToType(t *testing.T) {
	quietLogger(t)
	// A numeric field fed by an env var through ${...}: the replaced bytes are
	// consumed by the existing unmarshal, so type conversion needs no special
	// handling in Expand.
	dir := t.TempDir()
	path := dir + "/app.yaml"
	if err := writeFile(path, "page_size: ${PAGESIZE}\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Setenv("PAGESIZE", "42")

	var conf struct {
		PageSize int `yaml:"page_size" default:"10" validate:"required,gt=0,lte=100"`
	}
	if err := ReadConfig(path, &conf, yamlUnmarshal); err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}
	if conf.PageSize != 42 {
		t.Errorf("PageSize = %d, want 42", conf.PageSize)
	}
}

// helpers

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// writeFile is a tiny helper so tests don't import os/io/ioutil directly.
func writeFile(path, content string) error {
	return osWriteFile(path, []byte(content))
}
