package async

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestGroup_EmptyWait(t *testing.T) {
	g := New[int](context.Background())
	if err := g.Wait(); err != nil {
		t.Errorf("Wait() empty = %v, want nil", err)
	}
	if got := g.Results(); got != nil {
		t.Errorf("Results() = %v, want nil", got)
	}
}

func TestGroup_NilTasksSkipped(t *testing.T) {
	g := New[int](context.Background())
	g.Go(nil)
	g.Go(nil)
	if err := g.Wait(); err != nil {
		t.Errorf("Wait() nil-only = %v, want nil", err)
	}
}

func TestGroup_SuccessResults(t *testing.T) {
	g := New[int](context.Background())
	g.Go(intTask(1, nil))
	g.Go(intTask(2, nil))
	g.Go(intTask(3, nil))
	if err := g.Wait(); err != nil {
		t.Fatalf("Wait() = %v, want nil", err)
	}
	got := g.Results()
	if len(got) != 3 {
		t.Fatalf("Results() len = %d, want 3", len(got))
	}
	seen := map[int]bool{}
	for _, r := range got {
		if r.Err != nil {
			t.Errorf("Results() err = %v, want nil", r.Err)
		}
		seen[r.Value] = true
	}
	for _, want := range []int{1, 2, 3} {
		if !seen[want] {
			t.Errorf("Results() missing %d", want)
		}
	}
}

func TestGroup_JoinAllErrors(t *testing.T) {
	g := New[int](context.Background())
	g.Go(intTask(0, errOne))
	g.Go(intTask(0, errTwo))
	err := g.Wait()
	if err == nil {
		t.Fatal("Wait() should join errors")
	}
	if !errors.Is(err, errOne) || !errors.Is(err, errTwo) {
		t.Errorf("Wait() should join both errors: %v", err)
	}
}

func TestGroup_PanicIsRecovered(t *testing.T) {
	g := New[int](context.Background())
	g.Go(func(context.Context) (int, error) { panic("kaboom") })
	g.Go(intTask(1, nil))
	err := g.Wait()
	if err == nil {
		t.Fatal("Wait() should surface panic error")
	}
	var pe *PanicError
	if !errors.As(err, &pe) {
		t.Errorf("Wait() err = %T, want *PanicError", err)
	}
}

func TestGroup_WithLimit_BoundsConcurrency(t *testing.T) {
	const limit = 2
	var inFlight, peak int32
	g := New[int](context.Background(), WithLimit[int](limit))
	for i := 0; i < 20; i++ {
		g.Go(func(ctx context.Context) (int, error) {
			cur := atomic.AddInt32(&inFlight, 1)
			for {
				p := atomic.LoadInt32(&peak)
				if cur <= p || atomic.CompareAndSwapInt32(&peak, p, cur) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt32(&inFlight, -1)
			return 0, nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatalf("Wait() = %v", err)
	}
	if got := atomic.LoadInt32(&peak); got > int32(limit) {
		t.Errorf("peak concurrency = %d, want <= %d", got, limit)
	}
}

func TestGroup_WithLimit_DropsNonPositive(t *testing.T) {
	for _, n := range []int{0, -1, -100} {
		g := New[int](context.Background(), WithLimit[int](n))
		if g.sem != nil {
			t.Errorf("WithLimit(%d) should be a no-op, got sem=%v", n, g.sem)
		}
		_ = g.Wait()
	}
}

func TestGroup_WithCancelOnError_ReturnsFirstAndStops(t *testing.T) {
	canceled := make(chan struct{})
	g := New[int](context.Background(), WithCancelOnError[int]())
	g.Go(intTask(0, errFirst))
	g.Go(func(ctx context.Context) (int, error) {
		<-ctx.Done()
		close(canceled)
		return 0, ctx.Err()
	})
	err := g.Wait()
	if err == nil {
		t.Fatal("Wait() should error")
	}
	if !errors.Is(err, errFirst) {
		t.Errorf("Wait() = %v, want first error", err)
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Error("cancel-on-error did not cancel the pending task")
	}
}

func TestGroup_CancelOnError_StillRunsPreDispatched(t *testing.T) {
	// Without limit, both tasks are dispatched immediately and may both fail
	// with their own (non-cancellation) errors before the group cancels.
	// Under cancel-on-error, Wait must report exactly one real failure —
	// whichever arrived first — and must NOT leak a context.Canceled from
	// any task that observed the cancellation.
	g := New[int](context.Background(), WithCancelOnError[int]())
	g.Go(intTask(0, errFirst))
	g.Go(intTask(0, errSecond))
	err := g.Wait()
	if err == nil {
		t.Fatal("Wait() should error")
	}
	if !errors.Is(err, errFirst) && !errors.Is(err, errSecond) {
		t.Errorf("Wait() = %v, want one of the real failures", err)
	}
	if errors.Is(err, context.Canceled) {
		t.Errorf("Wait() leaked a context.Canceled from the cancelled task: %v", err)
	}
	var pe *PanicError
	if errors.As(err, &pe) {
		t.Fatalf("Wait() should not contain a panic: %v", err)
	}
}

func TestGroup_ResultsAliasSafe(t *testing.T) {
	g := New[int](context.Background())
	g.Go(intTask(5, nil))
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
	r := g.Results()
	r[0] = Result[int]{Value: 999} // mutate the copy, not internal
	again := g.Results()
	if again[0].Value != 5 {
		t.Errorf("internal results mutated: got %d, want 5", again[0].Value)
	}
}
