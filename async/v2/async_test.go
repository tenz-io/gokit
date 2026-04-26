package async

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestRun_Empty(t *testing.T) {
	if err := Run[int](context.Background()); err != nil {
		t.Errorf("Run() on empty = %v, want nil", err)
	}
}

func TestRun_Success(t *testing.T) {
	err := Run(context.Background(),
		func(ctx context.Context) (int, error) { return 1, nil },
		func(ctx context.Context) (int, error) { return 2, nil },
	)
	if err != nil {
		t.Errorf("Run() = %v", err)
	}
}

func TestRun_FirstError(t *testing.T) {
	err := Run(context.Background(),
		func(ctx context.Context) (int, error) { return 0, fmt.Errorf("boom") },
		func(ctx context.Context) (int, error) { time.Sleep(100 * time.Millisecond); return 1, nil },
	)
	if err == nil {
		t.Error("Run() should return error")
	}
}

func TestWait(t *testing.T) {
	var counter int32
	Wait(context.Background(),
		func(ctx context.Context) (int, error) { atomic.AddInt32(&counter, 1); return 1, nil },
		func(ctx context.Context) (int, error) { atomic.AddInt32(&counter, 1); return 2, nil },
	)
	if atomic.LoadInt32(&counter) != 2 {
		t.Errorf("Wait() counter = %v, want 2", counter)
	}
}

func TestWait_Empty(t *testing.T) {
	Wait[int](context.Background()) // should not panic
}

func TestAllOf_Empty(t *testing.T) {
	results := AllOf[int](context.Background(), nil)
	if len(results) != 0 {
		t.Errorf("AllOf() on nil = %v", results)
	}
}

func TestAllOf(t *testing.T) {
	fns := []Fn[int]{
		func(ctx context.Context) (int, error) { return 10, nil },
		func(ctx context.Context) (int, error) { return 20, nil },
		func(ctx context.Context) (int, error) { return 30, nil },
	}
	results := AllOf(context.Background(), fns)
	if len(results) != 3 {
		t.Fatalf("AllOf() len = %v, want 3", len(results))
	}
	if results[0].Val != 10 || results[0].Err != nil {
		t.Errorf("AllOf()[0] = {%v, %v}", results[0].Val, results[0].Err)
	}
	if results[1].Val != 20 || results[2].Val != 30 {
		t.Error("AllOf() order mismatch")
	}
}

func TestAllOf_WithErrors(t *testing.T) {
	fns := []Fn[int]{
		func(ctx context.Context) (int, error) { return 1, nil },
		func(ctx context.Context) (int, error) { return 0, fmt.Errorf("failed") },
	}
	results := AllOf(context.Background(), fns)
	if results[0].Err != nil || results[1].Err == nil {
		t.Error("AllOf() errors incorrect")
	}
}

func TestAnyOf_Empty(t *testing.T) {
	_, err := AnyOf[int](context.Background())
	if err == nil {
		t.Error("AnyOf() on empty should error")
	}
}

func TestAnyOf_FirstSuccess(t *testing.T) {
	result, err := AnyOf(context.Background(),
		func(ctx context.Context) (int, error) {
			time.Sleep(50 * time.Millisecond)
			return 1, nil
		},
		func(ctx context.Context) (int, error) {
			return 2, nil
		},
	)
	if err != nil {
		t.Errorf("AnyOf() err = %v", err)
	}
	if result != 2 {
		t.Errorf("AnyOf() = %v, want 2 (fastest)", result)
	}
}

func TestAnyOf_AllFail(t *testing.T) {
	_, err := AnyOf(context.Background(),
		func(ctx context.Context) (int, error) { return 0, fmt.Errorf("e1") },
		func(ctx context.Context) (int, error) { return 0, fmt.Errorf("e2") },
	)
	if err == nil {
		t.Error("AnyOf() should error when all fail")
	}
}

func TestPanicProof(t *testing.T) {
	var counter int32
	fn := panicProof(func(ctx context.Context) (int, error) {
		atomic.AddInt32(&counter, 1)
		panic("oh no")
	})
	_, err := fn(context.Background())
	if err == nil {
		t.Error("panicProof() should return error on panic")
	}
	if atomic.LoadInt32(&counter) != 1 {
		t.Error("panicProof() should have called the function")
	}
}
