package app

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

type testConf struct {
	Name string
}

func TestAdaptInit_ConfTypeMismatchIsError(t *testing.T) {
	called := false
	init := AdaptInit[testConf](func(_ *Context, _ *testConf) (CleanFunc, error) {
		called = true
		return nil, nil
	})
	// Pass a *struct{} — not *testConf.
	_, err := init(NewContext(context.Background(), &Flags{}), &struct{}{})
	if err == nil {
		t.Fatal("expected type-mismatch error, got nil")
	}
	if called {
		t.Fatal("typed init body must not run on mismatch")
	}
}

func TestAdaptInit_DelegatesAndReturnsCleanup(t *testing.T) {
	var ran atomic.Bool
	init := AdaptInit[testConf](func(_ *Context, _ *testConf) (CleanFunc, error) {
		ran.Store(true)
		return func(context.Context) error { return nil }, nil
	})
	clean, err := init(NewContext(context.Background(), &Flags{}), &testConf{Name: "x"})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if !ran.Load() {
		t.Fatal("typed init body did not run")
	}
	if clean == nil {
		t.Fatal("expected cleanup func, got nil")
	}
}

func TestAdaptRun_ConfTypeMismatchSignalsError(t *testing.T) {
	run := AdaptRun[testConf](func(_ *Context, _ *testConf) error {
		t.Fatal("typed run body must not run on mismatch")
		return nil
	})
	errC := make(chan error, 1)
	run(NewContext(context.Background(), &Flags{}), &struct{}{}, errC)
	if err := <-errC; err == nil {
		t.Fatal("expected error on errC, got nil")
	}
}

func TestAdaptRun_NilErrIsCleanCompletion(t *testing.T) {
	var ran atomic.Bool
	run := AdaptRun[testConf](func(_ *Context, _ *testConf) error {
		ran.Store(true)
		return nil // clean completion
	})
	errC := make(chan error, 1)
	run(NewContext(context.Background(), &Flags{}), &testConf{Name: "x"}, errC)
	if err := <-errC; err != nil {
		t.Fatalf("expected nil on errC, got %v", err)
	}
	if !ran.Load() {
		t.Fatal("typed run body did not run")
	}
}

func TestAdaptRun_PropagatesFatalError(t *testing.T) {
	boom := errors.New("boom")
	run := AdaptRun[testConf](func(_ *Context, _ *testConf) error { return boom })
	errC := make(chan error, 1)
	run(NewContext(context.Background(), &Flags{}), &testConf{}, errC)
	if err := <-errC; !errors.Is(err, boom) {
		t.Fatalf("errC = %v, want boom", err)
	}
}
