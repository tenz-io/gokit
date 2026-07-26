package async

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func intTask(v int, err error) Task[int] {
	return func(context.Context) (int, error) { return v, err }
}

// errSentinel values compared by identity in tests — never reconstruct them
// inline with fmt.Errorf, or errors.Is on a join will compare different pointers.
var (
	errOne    = errors.New("e1")
	errTwo    = errors.New("e2")
	errFirst  = errors.New("first")
	errSecond = errors.New("second")
)

func TestRun_Empty(t *testing.T) {
	if err := Run[int](context.Background()); err != nil {
		t.Errorf("Run() empty = %v, want nil", err)
	}
	if err := Run[int](context.Background(), nil, nil); err != nil {
		t.Errorf("Run() all-nil = %v, want nil", err)
	}
}

func TestRun_Success(t *testing.T) {
	if err := Run(context.Background(), intTask(1, nil), intTask(2, nil)); err != nil {
		t.Errorf("Run() success = %v, want nil", err)
	}
}

func TestRun_FirstError(t *testing.T) {
	err := Run(context.Background(),
		intTask(0, errors.New("boom")),
		intTask(1, nil),
	)
	if err == nil || err.Error() != "boom" {
		t.Errorf("Run() = %v, want boom", err)
	}
}

func TestRun_JoinsNoMoreThanFirst(t *testing.T) {
	// Run returns the first error that arrives; under concurrency that may be
	// either sentinel. The contract is: exactly one error (no join), and it
	// must be one of the real failures.
	err := Run(context.Background(),
		intTask(0, errOne),
		intTask(0, errTwo),
	)
	if err == nil {
		t.Fatal("Run() should return an error")
	}
	if !errors.Is(err, errOne) && !errors.Is(err, errTwo) {
		t.Errorf("Run() = %v, want one of the sentinels", err)
	}
}

func TestRunAll_JoinAllErrors(t *testing.T) {
	err := RunAll(context.Background(),
		intTask(0, errOne),
		intTask(0, errTwo),
	)
	if err == nil {
		t.Fatal("RunAll() should return an error")
	}
	if !errors.Is(err, errOne) {
		t.Errorf("RunAll() should join e1: %v", err)
	}
	if !errors.Is(err, errTwo) {
		t.Errorf("RunAll() should join e2: %v", err)
	}
}

func TestWait(t *testing.T) {
	var counter int32
	Wait(context.Background(),
		func(context.Context) (int, error) { atomic.AddInt32(&counter, 1); return 1, nil },
		func(context.Context) (int, error) { atomic.AddInt32(&counter, 1); return 2, nil },
	)
	if got := atomic.LoadInt32(&counter); got != 2 {
		t.Errorf("Wait() counter = %d, want 2", got)
	}
}

func TestWait_Empty(t *testing.T) {
	Wait[int](context.Background()) // must not panic
	Wait[int](context.Background(), nil, nil)
}

func TestAllOf_Empty(t *testing.T) {
	if r := AllOf[int](context.Background()); r != nil {
		t.Errorf("AllOf() empty = %v, want nil", r)
	}
}

func TestAllOf_OrderedResults(t *testing.T) {
	results := AllOf(context.Background(),
		intTask(10, nil),
		intTask(20, nil),
		intTask(30, nil),
	)
	if len(results) != 3 {
		t.Fatalf("AllOf() len = %d, want 3", len(results))
	}
	want := []int{10, 20, 30}
	for i, w := range want {
		if results[i].Value != w || results[i].Err != nil {
			t.Errorf("AllOf()[%d] = {Val:%d, Err:%v}, want {%d, nil}", i, results[i].Value, results[i].Err, w)
		}
	}
}

func TestAllOf_PreservesErrorsPerSlot(t *testing.T) {
	results := AllOf(context.Background(),
		intTask(1, nil),
		intTask(0, errors.New("failed")),
	)
	if results[0].Err != nil || results[0].Value != 1 {
		t.Errorf("AllOf()[0] = %v, want {1, nil}", results[0])
	}
	if results[1].Err == nil {
		t.Error("AllOf()[1].Err should be set")
	}
}

func TestAllOf_SkipsNilButKeepsLengthDropped(t *testing.T) {
	// nil tasks are dropped from the slice (filtered), so a single nil yields nil.
	results := AllOf[int](context.Background(), nil)
	if results != nil {
		t.Errorf("AllOf(nil-only) = %v, want nil", results)
	}
}

func TestAnyOf_Empty(t *testing.T) {
	if _, err := AnyOf[int](context.Background()); err == nil {
		t.Error("AnyOf() empty should error")
	}
}

func TestAnyOf_NilTask(t *testing.T) {
	if _, err := AnyOf[int](context.Background(), nil); err == nil {
		t.Error("AnyOf() nil task should error")
	}
}

func TestAnyOf_FirstSuccess(t *testing.T) {
	got, err := AnyOf(context.Background(),
		func(context.Context) (int, error) { time.Sleep(50 * time.Millisecond); return 1, nil },
		intTask(2, nil),
	)
	if err != nil {
		t.Fatalf("AnyOf() err = %v", err)
	}
	if got != 2 {
		t.Errorf("AnyOf() = %d, want 2 (fastest success)", got)
	}
}

func TestAnyOf_AllFail(t *testing.T) {
	_, err := AnyOf(context.Background(),
		intTask(0, errOne),
		intTask(0, errTwo),
	)
	if err == nil {
		t.Fatal("AnyOf() should error when all fail")
	}
	if !errors.Is(err, errOne) || !errors.Is(err, errTwo) {
		t.Errorf("AnyOf() joined error should contain both: %v", err)
	}
}

func TestAnyOf_CancelsRemainingOnFirstSuccess(t *testing.T) {
	canceled := make(chan struct{})
	got, err := AnyOf(context.Background(),
		intTask(42, nil),
		func(ctx context.Context) (int, error) {
			<-ctx.Done()
			close(canceled)
			return 0, ctx.Err()
		},
	)
	if err != nil || got != 42 {
		t.Fatalf("AnyOf() = (%d, %v), want (42, nil)", got, err)
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Error("AnyOf() did not cancel the losing task")
	}
}

func TestPanic_BecomesTypedError(t *testing.T) {
	err := Run(context.Background(), func(context.Context) (int, error) {
		panic("boom")
	})
	if err == nil {
		t.Fatal("Run() should surface a panic as an error")
	}
	var pe *PanicError
	if !errors.As(err, &pe) {
		t.Fatalf("Run() panic err = %T, want *PanicError", err)
	}
	if pe.Value() != "boom" {
		t.Errorf("PanicError.Value() = %v, want boom", pe.Value())
	}
	if len(pe.Stack()) == 0 {
		t.Error("PanicError.Stack() should be non-empty")
	}
	if got := pe.Error(); got == "" {
		t.Error("PanicError.Error() should be non-empty")
	}
}

func TestPanic_UnwrapsWrappedError(t *testing.T) {
	wrapped := fmt.Errorf("inner")
	err := Run(context.Background(), func(context.Context) (int, error) {
		panic(wrapped)
	})
	if !errors.Is(err, wrapped) {
		t.Errorf("PanicError should unwrap to inner error: %v", err)
	}
}

func TestPanic_AllOfIsolatesPanic(t *testing.T) {
	results := AllOf(context.Background(),
		intTask(1, nil),
		func(context.Context) (int, error) { panic(123) },
	)
	if results[0].Value != 1 || results[0].Err != nil {
		t.Errorf("AllOf()[0] = %v, want {1, nil}", results[0])
	}
	var pe *PanicError
	if !errors.As(results[1].Err, &pe) || pe.Value() != 123 {
		t.Errorf("AllOf()[1].Err = %v, want *PanicError{123}", results[1].Err)
	}
}

func TestAnyOf_PanicCountsAsFailure(t *testing.T) {
	got, err := AnyOf(context.Background(),
		func(context.Context) (int, error) { panic("nope") },
		intTask(7, nil),
	)
	if err != nil || got != 7 {
		t.Fatalf("AnyOf() = (%d, %v), want (7, nil)", got, err)
	}
}
