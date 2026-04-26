package retriever

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewRetriever_Defaults(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDo_Success(t *testing.T) {
	r := New(WithMaxAttempt(3))
	result, err := r.Do(context.Background(), func(ctx context.Context) (any, bool, error) {
		return "ok", false, nil
	})
	if err != nil {
		t.Errorf("Do() err = %v", err)
	}
	if result != "ok" {
		t.Errorf("Do() = %v, want ok", result)
	}
}

func TestDo_Retry(t *testing.T) {
	attempts := 0
	r := New(WithMaxAttempt(3), WithBackoff(&NoBackoff{}))
	result, err := r.Do(context.Background(), func(ctx context.Context) (any, bool, error) {
		attempts++
		if attempts < 3 {
			return nil, true, fmt.Errorf("fail %d", attempts)
		}
		return "ok", false, nil
	})
	if err != nil {
		t.Errorf("Do() err = %v", err)
	}
	if result != "ok" {
		t.Errorf("Do() = %v", result)
	}
	if attempts != 3 {
		t.Errorf("attempts = %v, want 3", attempts)
	}
}

func TestDo_NonRetryable(t *testing.T) {
	attempts := 0
	r := New(WithMaxAttempt(5), WithBackoff(&NoBackoff{}))
	_, err := r.Do(context.Background(), func(ctx context.Context) (any, bool, error) {
		attempts++
		return nil, false, fmt.Errorf("permanent")
	})
	if err == nil {
		t.Error("Do() should return error")
	}
	if attempts != 1 {
		t.Errorf("non-retryable should not retry, got %d attempts", attempts)
	}
}

func TestDo_MaxAttempts(t *testing.T) {
	attempts := 0
	r := New(WithMaxAttempt(3), WithBackoff(&NoBackoff{}))
	_, err := r.Do(context.Background(), func(ctx context.Context) (any, bool, error) {
		attempts++
		return nil, true, fmt.Errorf("always fail")
	})
	if err == nil {
		t.Error("Do() should return error after max attempts")
	}
	if attempts != 3 {
		t.Errorf("attempts = %v, want 3", attempts)
	}
}

func TestDoAlwaysRetry(t *testing.T) {
	attempts := 0
	r := New(WithMaxAttempt(4), WithBackoff(&NoBackoff{}))
	result, err := r.DoAlwaysRetry(context.Background(), func(ctx context.Context) (any, error) {
		attempts++
		if attempts < 4 {
			return nil, fmt.Errorf("fail")
		}
		return "done", nil
	})
	if err != nil {
		t.Errorf("DoAlwaysRetry() err = %v", err)
	}
	if result != "done" {
		t.Errorf("DoAlwaysRetry() = %v", result)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := New(WithMaxAttempt(3), WithBackoff(NewLinearBackoff(1000)))
	_, err := r.Do(ctx, func(ctx context.Context) (any, bool, error) {
		return nil, true, fmt.Errorf("fail")
	})
	if err == nil {
		t.Error("Do() should return context error")
	}
}

func TestDo_TotalTimeout(t *testing.T) {
	r := New(
		WithMaxAttempt(10),
		WithMaxTotalAttemptTime(50*time.Millisecond),
		WithBackoff(NewLinearBackoff(20)),
	)
	_, err := r.Do(context.Background(), func(ctx context.Context) (any, bool, error) {
		return nil, true, fmt.Errorf("fail")
	})
	if err == nil {
		t.Error("Do() should return timeout error")
	}
}

func TestBackoff_NoBackoff(t *testing.T) {
	b := &NoBackoff{}
	if b.Next(5) != 0 {
		t.Error("NoBackoff should return 0")
	}
}

func TestBackoff_Linear(t *testing.T) {
	b := NewLinearBackoff(100)
	if b.Next(0) != 100*time.Millisecond {
		t.Error("LinearBackoff.Next(0) should be 100ms")
	}
	if b.Next(5) != 100*time.Millisecond {
		t.Error("LinearBackoff.Next(5) should be 100ms")
	}
}

func TestBackoff_Exponential(t *testing.T) {
	b := NewExponentialBackoff(100, 2.0, 0)
	d0 := b.Next(0)  // base * (2^0 + 0) = 100ms
	d1 := b.Next(1)  // base * (2^1 + 0) = 200ms
	d2 := b.Next(2)  // base * (2^2 + 0) = 400ms

	if d0 != 100*time.Millisecond {
		t.Errorf("Next(0) = %v, want 100ms", d0)
	}
	if d1 != 200*time.Millisecond {
		t.Errorf("Next(1) = %v, want 200ms", d1)
	}
	if d2 != 400*time.Millisecond {
		t.Errorf("Next(2) = %v, want 400ms", d2)
	}
}

func TestBackoff_ExponentialWithJitter(t *testing.T) {
	// With jitter the value should be >= base and < base*(factor^count + jitter)
	b := NewExponentialBackoff(100, 2.0, 0.5)
	for i := 0; i < 20; i++ {
		d := b.Next(0)
		if d < 100*time.Millisecond {
			t.Errorf("Next(0) = %v, should be >= 100ms", d)
		}
		if d > 150*time.Millisecond { // 100*(1+0.5) = 150
			t.Errorf("Next(0) = %v, should be <= 150ms", d)
		}
	}
}
