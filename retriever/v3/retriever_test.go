package retriever

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// errSentinel 用于测试的错误哨兵,以同一性比较,绝不内联 fmt.Errorf 重建。
var (
	errBoom      = errors.New("boom")
	errPerm      = errors.New("permanent")
	errTransient = errors.New("transient")
)

func TestNew_Defaults(t *testing.T) {
	r := New[int]()
	if r == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNew_GenericReturn(t *testing.T) {
	r := New[string]()
	// 无需任何类型断言即可拿到 string —— 这是 v3 相对 v2 的核心易用性改进。
	result, err := r.Do(context.Background(), func(ctx context.Context) (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("Do() err = %v", err)
	}
	if result != "ok" {
		t.Errorf("Do() = %q, want ok", result)
	}
}

func TestDo_Success(t *testing.T) {
	r := New[int](WithMaxAttempts(3))
	result, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		return 42, nil
	})
	if err != nil {
		t.Fatalf("Do() err = %v", err)
	}
	if result != 42 {
		t.Errorf("Do() = %v, want 42", result)
	}
}

func TestDo_Retry(t *testing.T) {
	attempts := 0
	r := New[string](WithMaxAttempts(3), WithBackoff(Constant(0)))
	result, err := r.Do(context.Background(), func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", fmt.Errorf("fail %d", attempts)
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("Do() err = %v", err)
	}
	if result != "ok" {
		t.Errorf("Do() = %q, want ok", result)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestDo_RetriesEveryErrorByDefault(t *testing.T) {
	// 默认行为:对任何错误都重试,无需 DoFunc 返回 retry 标志。
	attempts := 0
	r := New[int](WithMaxAttempts(3), WithBackoff(Constant(0)))
	_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		attempts++
		return 0, errBoom
	})
	if !errors.Is(err, ErrMaxAttempts) {
		t.Errorf("err = %v, want ErrMaxAttempts", err)
	}
	if !errors.Is(err, errBoom) {
		t.Errorf("err = %v, want wrapping errBoom", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestDo_NonRetryable(t *testing.T) {
	// NonRetryable 标记的错误立即返回,即便它本应可重试。
	attempts := 0
	r := New[int](WithMaxAttempts(5), WithBackoff(Constant(0)))
	_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		attempts++
		return 0, NonRetryable(errPerm)
	})
	if !errors.Is(err, errPerm) {
		t.Errorf("err = %v, want errPerm", err)
	}
	if !IsNonRetryable(err) {
		t.Errorf("IsNonRetryable(err) = false, want true")
	}
	if err.Error() != errPerm.Error() {
		t.Errorf("err.Error() = %q, want %q (string-transparent)", err.Error(), errPerm.Error())
	}
	if attempts != 1 {
		t.Errorf("non-retryable should not retry, got %d attempts", attempts)
	}
}

func TestNonRetryable_Nil(t *testing.T) {
	if got := NonRetryable(nil); got != nil {
		t.Errorf("NonRetryable(nil) = %v, want nil", got)
	}
}

func TestNonRetryable_TransparentString(t *testing.T) {
	wrapped := NonRetryable(errPerm)
	// 标记对 Error() 文本完全透明。
	if wrapped.Error() != errPerm.Error() {
		t.Errorf("Error() = %q, want %q", wrapped.Error(), errPerm.Error())
	}
	// errors.Is 穿透到内层 errPerm,也能判别 ErrNonRetryable。
	if !errors.Is(wrapped, errPerm) {
		t.Errorf("errors.Is(wrapped, errPerm) = false, want true")
	}
	if !errors.Is(wrapped, ErrNonRetryable) {
		t.Errorf("errors.Is(wrapped, ErrNonRetryable) = false, want true")
	}
}

func TestDo_RetryableClassifier(t *testing.T) {
	// 仅对 errTransient 重试;其余错误立即返回。
	transient, perm := 0, 0
	r := New[int](
		WithMaxAttempts(4),
		WithBackoff(Constant(0)),
		WithRetryable(func(err error) bool {
			return errors.Is(err, errTransient)
		}),
	)
	// transient 重试到第 3 次成功。
	transientAttempts := 0
	_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		transient++
		transientAttempts++
		if transientAttempts < 3 {
			return 0, errTransient
		}
		return 1, nil
	})
	if err != nil {
		t.Fatalf("Do() transient err = %v", err)
	}
	if transientAttempts != 3 {
		t.Errorf("transient attempts = %d, want 3", transientAttempts)
	}

	// non-transient 立即返回,不重试。
	permAttempts := 0
	_, err = r.Do(context.Background(), func(ctx context.Context) (int, error) {
		perm++
		permAttempts++
		return 0, errPerm
	})
	if !errors.Is(err, errPerm) {
		t.Errorf("err = %v, want errPerm", err)
	}
	if permAttempts != 1 {
		t.Errorf("perm attempts = %d, want 1", permAttempts)
	}
	_ = transient
	_ = perm
}

func TestDo_ClassifierIgnoredForNonRetryable(t *testing.T) {
	// NonRetryable 标记优先于分类器,即便分类器本会重试。
	attempts := 0
	r := New[int](
		WithMaxAttempts(5),
		WithBackoff(Constant(0)),
		WithRetryable(func(error) bool { return true }), // 本会重试一切
	)
	_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		attempts++
		return 0, NonRetryable(errBoom)
	})
	if !errors.Is(err, errBoom) {
		t.Errorf("err = %v, want errBoom", err)
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (NonRetryable wins over classifier)", attempts)
	}
}

func TestDo_MaxAttempts(t *testing.T) {
	attempts := 0
	r := New[int](WithMaxAttempts(3), WithBackoff(Constant(0)))
	_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		attempts++
		return 0, fmt.Errorf("always fail")
	})
	if !errors.Is(err, ErrMaxAttempts) {
		t.Errorf("err = %v, want ErrMaxAttempts", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestDo_DefaultMaxAttempts(t *testing.T) {
	// WithMaxAttempts 未调用时,默认 3。
	attempts := 0
	r := New[int](WithBackoff(Constant(0)))
	_, _ = r.Do(context.Background(), func(ctx context.Context) (int, error) {
		attempts++
		return 0, errBoom
	})
	if attempts != 3 {
		t.Errorf("attempts = %d, want default 3", attempts)
	}
}

func TestDo_NilFn(t *testing.T) {
	r := New[int]()
	_, err := r.Do(context.Background(), nil)
	if err == nil {
		t.Error("Do(nil fn) should return error")
	}
}

func TestDo_NilCtx(t *testing.T) {
	r := New[int]()
	_, err := r.Do(nil, func(ctx context.Context) (int, error) { return 1, nil })
	if err == nil {
		t.Error("Do(nil ctx) should return error")
	}
}

func TestDo_PreCancelledCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := New[int](WithMaxAttempts(3), WithBackoff(Constant(time.Second)))
	_, err := r.Do(ctx, func(ctx context.Context) (int, error) {
		return 0, errBoom
	})
	// ctx 在首次尝试前已取消 => 返回 ctx.Err(),不调用 fn。
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestDo_CancelDuringBackoff(t *testing.T) {
	// 退避期间取消 ctx => 立即返回 ctx.Err(),不等退避结束。
	ctx, cancel := context.WithCancel(context.Background())
	r := New[int](WithMaxAttempts(10), WithBackoff(Constant(time.Second)))

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	_, err := r.Do(ctx, func(ctx context.Context) (int, error) {
		return 0, errBoom
	})
	elapsed := time.Since(start)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
	// 应远小于一次 1s 退避。
	if elapsed > 500*time.Millisecond {
		t.Errorf("elapsed = %v, want < 500ms (interrupted during backoff)", elapsed)
	}
}

func TestDo_Timeout(t *testing.T) {
	// WithTimeout 为全部尝试设置统一截止时间。
	r := New[int](
		WithMaxAttempts(10),
		WithTimeout(50*time.Millisecond),
		WithBackoff(Constant(20*time.Millisecond)),
	)
	_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		return 0, errBoom
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
}

func TestDo_FnSeesTimeoutCtx(t *testing.T) {
	// WithTimeout 包裹的 ctx 应透传到 fn,使其能感知到截止时间。
	r := New[int](WithMaxAttempts(1), WithTimeout(10*time.Millisecond))
	var sawDeadline bool
	_, _ = r.Do(context.Background(), func(ctx context.Context) (int, error) {
		_, ok := ctx.Deadline()
		sawDeadline = ok
		return 0, nil
	})
	if !sawDeadline {
		t.Error("fn should see a deadline-bearing ctx when WithTimeout is set")
	}
}

func TestDo_ReturnsLastErrOnMaxAttempts(t *testing.T) {
	// ErrMaxAttempts 包裹的是最后一次失败的错误。
	r := New[int](WithMaxAttempts(2), WithBackoff(Constant(0)))
	_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		return 0, errBoom
	})
	if !errors.Is(err, errBoom) {
		t.Errorf("err = %v, want wrapping errBoom", err)
	}
}

func TestDo_ConcurrentSafe(t *testing.T) {
	// retrier 无共享可变状态,Do 应可被多 goroutine 并发调用。
	r := New[int](WithMaxAttempts(3), WithBackoff(Constant(0)))
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			v, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
				return 7, nil
			})
			if err != nil || v != 7 {
				t.Errorf("Do() = (%d, %v), want (7, nil)", v, err)
			}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

// --- 构造期 option 边界行为 ------------------------------------------------

func TestWithMaxAttempts_NonPositiveFallsBack(t *testing.T) {
	// n<=0 应回退到默认 3,而非生效为 0(那会使循环一次都不执行)。
	for _, n := range []int{0, -1, -100} {
		attempts := 0
		r := New[int](WithMaxAttempts(n), WithBackoff(Constant(0)))
		_, _ = r.Do(context.Background(), func(ctx context.Context) (int, error) {
			attempts++
			return 0, errBoom
		})
		if attempts != 3 {
			t.Errorf("WithMaxAttempts(%d): attempts = %d, want fallback 3", n, attempts)
		}
	}
}

func TestWithTimeout_NonPositiveMeansNone(t *testing.T) {
	// d<=0 表示不设全局截止时间:仅受 ctx 自身约束。d=0 不应让 ctx 立刻过期。
	for _, d := range []time.Duration{0, -time.Second} {
		var sawDeadline bool
		r := New[int](WithMaxAttempts(1), WithTimeout(d))
		_, _ = r.Do(context.Background(), func(ctx context.Context) (int, error) {
			_, ok := ctx.Deadline()
			sawDeadline = ok
			return 1, nil
		})
		if sawDeadline {
			t.Errorf("WithTimeout(%v): fn saw a deadline, want none", d)
		}
	}
}

func TestWithBackoff_NilIgnored(t *testing.T) {
	// nil backoff 应被忽略,保留默认的指数退避(非 nil)。
	r := New[int](WithBackoff(nil))
	// 间接验证:默认 backoff 非 nil,Do 能正常跑通一次成功路径。
	if _, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		return 1, nil
	}); err != nil {
		t.Fatalf("Do() with nil backoff override err = %v", err)
	}
}

func TestWithRetryable_NilClearsClassifier(t *testing.T) {
	// 先设一个分类器,再用 nil 覆盖:应退回"对所有错误重试"的默认行为。
	r := New[int](
		WithMaxAttempts(3),
		WithBackoff(Constant(0)),
		WithRetryable(func(error) bool { return false }), // 本应使一切不重试
		WithRetryable(nil),                               // 清除分类器
	)
	attempts := 0
	_, _ = r.Do(context.Background(), func(ctx context.Context) (int, error) {
		attempts++
		return 0, errBoom
	})
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3 (nil classifier => retry all)", attempts)
	}
}

func TestNew_AllOptionsCombined(t *testing.T) {
	// 四个 option 叠加:验证它们互不干扰、各自生效。
	var sawDeadline bool
	r := New[int](
		WithMaxAttempts(5),
		WithTimeout(10*time.Second),
		WithBackoff(Constant(0)),
		WithRetryable(func(err error) bool { return errors.Is(err, errTransient) }),
	)
	// transient 重试 2 次后成功(<=5 次),且 fn 能看到 deadline。
	n := 0
	v, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		n++
		if _, ok := ctx.Deadline(); ok {
			sawDeadline = true
		}
		if n < 3 {
			return 0, errTransient
		}
		return 99, nil
	})
	if err != nil || v != 99 {
		t.Fatalf("Do() = (%d, %v), want (99, nil)", v, err)
	}
	if n != 3 {
		t.Errorf("attempts = %d, want 3", n)
	}
	if !sawDeadline {
		t.Error("fn should see deadline from WithTimeout")
	}
}

// --- 错误模型边界 ----------------------------------------------------------

func TestNonRetryable_NestedWrapping(t *testing.T) {
	// NonRetryable 的标记即便被外层 fmt.Errorf 进一步包装,errors.Is 仍能穿透识别。
	inner := errors.New("inner boom")
	wrapped := fmt.Errorf("outer: %w", NonRetryable(inner))
	if !IsNonRetryable(wrapped) {
		t.Error("IsNonRetryable should see through fmt.Errorf wrapping")
	}
	if !errors.Is(wrapped, inner) {
		t.Error("errors.Is should still reach inner error")
	}
}

func TestNonRetryable_NonMarkedNotFlagged(t *testing.T) {
	// 普通错误不应被误判为 NonRetryable。
	if IsNonRetryable(errBoom) {
		t.Error("plain error should not be NonRetryable")
	}
	if IsNonRetryable(nil) {
		t.Error("nil should not be NonRetryable")
	}
}

func TestDo_MaxAttemptsErrorMessage(t *testing.T) {
	// ErrMaxAttempts 的文本应含尝试次数与内层错误,便于排查。
	r := New[int](WithMaxAttempts(3), WithBackoff(Constant(0)))
	_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		return 0, errBoom
	})
	msg := err.Error()
	if !strings.Contains(msg, "max attempts") {
		t.Errorf("err msg = %q, want substring 'max attempts'", msg)
	}
	if !strings.Contains(msg, "boom") {
		t.Errorf("err msg = %q, want substring 'boom'", msg)
	}
}

func TestDo_MaxAttemptsReturnsFnZeroValue(t *testing.T) {
	// 达到上限时返回 fn 最后一次返回的值;此处 fn 返回 nil,故 result 为 nil。
	r := New[[]int](WithMaxAttempts(2), WithBackoff(Constant(0)))
	result, _ := r.Do(context.Background(), func(ctx context.Context) ([]int, error) {
		return nil, errBoom
	})
	if result != nil {
		t.Errorf("result = %v, want nil (fn's last value)", result)
	}
}

func TestDo_ClassifierExhaustionReturnsMaxAttempts(t *testing.T) {
	// 分类器判定可重试的错误,耗尽尝试后应返回 ErrMaxAttempts(而非裸错误)。
	r := New[int](
		WithMaxAttempts(2),
		WithBackoff(Constant(0)),
		WithRetryable(func(error) bool { return true }),
	)
	_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		return 0, errTransient
	})
	if !errors.Is(err, ErrMaxAttempts) {
		t.Errorf("err = %v, want ErrMaxAttempts", err)
	}
	if !errors.Is(err, errTransient) {
		t.Errorf("err = %v, want wrapping errTransient", err)
	}
}

// --- 终止路径携带 fn 的最后一次 result(资源释放契约) -----------------------

func TestDo_NonRetryableCarriesResult(t *testing.T) {
	// NonRetryable 终止路径应原样带回 fn 返回的 result,而非零值。
	// 这是 README 演示的 return resp, NonRetryable(err) 能被调用方关闭 resp.Body 的前提。
	r := New[*resource](WithMaxAttempts(5), WithBackoff(Constant(0)))
	res, err := r.Do(context.Background(), func(ctx context.Context) (*resource, error) {
		return &resource{id: 7}, NonRetryable(errPerm)
	})
	if !errors.Is(err, errPerm) {
		t.Fatalf("err = %v, want errPerm", err)
	}
	if res == nil || res.id != 7 {
		t.Errorf("result = %v, want &{id:7} (fn's last value)", res)
	}
}

func TestDo_ClassifierRejectCarriesResult(t *testing.T) {
	// 分类器判定不可重试时,同样带回 fn 的 result。
	r := New[*resource](
		WithMaxAttempts(5),
		WithBackoff(Constant(0)),
		WithRetryable(func(error) bool { return false }), // 一律不重试
	)
	res, err := r.Do(context.Background(), func(ctx context.Context) (*resource, error) {
		return &resource{id: 9}, errBoom
	})
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want errBoom", err)
	}
	if res == nil || res.id != 9 {
		t.Errorf("result = %v, want &{id:9}", res)
	}
}

func TestDo_ExhaustionCarriesLastResult(t *testing.T) {
	// 耗尽尝试时,带回最后一次(最后一次失败)的 result。
	r := New[*resource](WithMaxAttempts(2), WithBackoff(Constant(0)))
	res, err := r.Do(context.Background(), func(ctx context.Context) (*resource, error) {
		return &resource{id: 42}, errBoom
	})
	if !errors.Is(err, ErrMaxAttempts) {
		t.Fatalf("err = %v, want ErrMaxAttempts", err)
	}
	if res == nil || res.id != 42 {
		t.Errorf("result = %v, want &{id:42} (last attempt's value)", res)
	}
}

// --- context 终止优先于尝试耗尽(分类与序号解耦) --------------------------

func TestDo_TimeoutClassificationConsistent(t *testing.T) {
	// 同一超时,不论 maxAttempts 是 1 还是 3,都应统一归为 ctx 终止
	// (DeadlineExceeded),且不匹配 ErrMaxAttempts —— 便于监控/分支一致。
	body := func(ctx context.Context) (int, error) {
		<-ctx.Done() // 阻塞到 ctx 超时
		return 0, ctx.Err()
	}
	for _, max := range []int{1, 3} {
		r := New[int](
			WithMaxAttempts(max),
			WithTimeout(20*time.Millisecond),
			WithBackoff(Constant(0)),
		)
		_, err := r.Do(context.Background(), body)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("max=%d: err = %v, want DeadlineExceeded", max, err)
		}
		if errors.Is(err, ErrMaxAttempts) {
			t.Errorf("max=%d: err = %v, must NOT match ErrMaxAttempts (ctx termination takes priority)", max, err)
		}
	}
}

func TestDo_CancelClassificationConsistent(t *testing.T) {
	// 同上,但用外部取消:不论 maxAttempts,统一归为 context.Canceled。
	for _, max := range []int{1, 3} {
		ctx, cancel := context.WithCancel(context.Background())
		r := New[int](WithMaxAttempts(max), WithBackoff(Constant(0)))
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()
		_, err := r.Do(ctx, func(ctx context.Context) (int, error) {
			<-ctx.Done()
			return 0, ctx.Err()
		})
		if !errors.Is(err, context.Canceled) {
			t.Errorf("max=%d: err = %v, want Canceled", max, err)
		}
		if errors.Is(err, ErrMaxAttempts) {
			t.Errorf("max=%d: err = %v, must NOT match ErrMaxAttempts", max, err)
		}
	}
}

func TestDo_TimeoutCarriesResult(t *testing.T) {
	// ctx 终止路径也应带回 fn 的最后一次 result。
	r := New[*resource](
		WithMaxAttempts(5),
		WithTimeout(15*time.Millisecond),
		WithBackoff(Constant(0)),
	)
	res, err := r.Do(context.Background(), func(ctx context.Context) (*resource, error) {
		<-ctx.Done()
		return &resource{id: 5}, ctx.Err()
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}
	if res == nil || res.id != 5 {
		t.Errorf("result = %v, want &{id:5} (carried on ctx termination)", res)
	}
}

func TestDo_NonRetryableBeatsCtxTermination(t *testing.T) {
	// NonRetryable 标记优先级仍高于 ctx 终止(标记在 ctx 检查之前判定)。
	r := New[int](WithMaxAttempts(3), WithBackoff(Constant(0)))
	_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		return 0, NonRetryable(errPerm)
	})
	if !errors.Is(err, errPerm) {
		t.Fatalf("err = %v, want errPerm (NonRetryable wins)", err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		t.Error("NonRetryable must not surface as DeadlineExceeded")
	}
}

// resource 是测试用的可识别资源,模拟 *http.Response 这类需调用方释放的对象。
type resource struct{ id int }

func TestDo_DefaultBackoffIsExponential(t *testing.T) {
	// 不显式 WithBackoff 时,默认应为 Exponential{Base:100ms, Factor:2}。
	// 通过测量首次重试前的等待时长来间接验证(应在 ~100ms 量级,远小于 200ms 的第二次)。
	r := New[int](WithMaxAttempts(3)) // 不覆盖 backoff
	start := time.Now()
	_, _ = r.Do(context.Background(), func(ctx context.Context) (int, error) {
		return 0, errBoom
	})
	// 3 次尝试 + 2 次退避:100ms(attempt0) + 200ms(attempt1) ≈ 300ms。
	// 容忍 ±30ms 抖动与调度开销,但应明显 > 仅一次 100ms 退避。
	elapsed := time.Since(start)
	if elapsed < 250*time.Millisecond {
		t.Errorf("elapsed = %v, want >= 250ms (default exponential 100+200ms)", elapsed)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("elapsed = %v, want <= 500ms (sanity upper bound)", elapsed)
	}
}

func TestDo_SucceedsOnNthAttemptBoundary(t *testing.T) {
	// 恰好在最后一次允许尝试时成功 —— 验证边界:不提前 break、不超 1 次。
	for _, max := range []int{1, 2, 5} {
		n := 0
		r := New[int](WithMaxAttempts(max), WithBackoff(Constant(0)))
		v, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
			n++
			if n < max {
				return 0, errBoom
			}
			return max, nil
		})
		if err != nil {
			t.Errorf("max=%d: Do() err = %v", max, err)
		}
		if v != max {
			t.Errorf("max=%d: result = %v, want %d", max, v, max)
		}
		if n != max {
			t.Errorf("max=%d: attempts = %d, want %d", max, n, max)
		}
	}
}

func TestDo_FailsExactlyAtMaxAttempts(t *testing.T) {
	// 恰好在第 max 次(最后一次)失败:应返回 ErrMaxAttempts,且只尝试 max 次。
	for _, max := range []int{1, 2, 5} {
		n := 0
		r := New[int](WithMaxAttempts(max), WithBackoff(Constant(0)))
		_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
			n++
			return 0, errBoom
		})
		if !errors.Is(err, ErrMaxAttempts) {
			t.Errorf("max=%d: err = %v, want ErrMaxAttempts", max, err)
		}
		if n != max {
			t.Errorf("max=%d: attempts = %d, want %d", max, n, max)
		}
	}
}

func TestDo_SuccessFirstCallZeroWait(t *testing.T) {
	// 首次即成功:不应触发任何退避等待(即便 backoff 很大)。
	r := New[int](WithMaxAttempts(3), WithBackoff(Constant(time.Second)))
	start := time.Now()
	_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		return 1, nil
	})
	if err != nil {
		t.Fatalf("Do() err = %v", err)
	}
	if d := time.Since(start); d > 100*time.Millisecond {
		t.Errorf("success path waited %v, want ~0 (no backoff on first-call success)", d)
	}
}

func TestDo_TimeoutFiresDuringAttempts(t *testing.T) {
	// WithTimeout 截止时间短于单次 fn 耗时:fn 自身应看到 ctx 超时。
	r := New[int](
		WithMaxAttempts(10),
		WithTimeout(20*time.Millisecond),
		WithBackoff(Constant(0)),
	)
	called := 0
	_, err := r.Do(context.Background(), func(ctx context.Context) (int, error) {
		called++
		<-ctx.Done() // 阻塞到 ctx 超时
		return 0, ctx.Err()
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want DeadlineExceeded", err)
	}
}

func TestDo_NonRetryableWrappedErrorString(t *testing.T) {
	// 业务错误带前缀信息时,NonRetryable 标记也不应改变其显示文本。
	bizErr := fmt.Errorf("service: %w", errPerm)
	marked := NonRetryable(bizErr)
	if marked.Error() != bizErr.Error() {
		t.Errorf("Error() = %q, want %q", marked.Error(), bizErr.Error())
	}
}

// --- 表驱动:Backoff 策略汇总(纯策略测试见 backoff_test.go) ----------------

// 注:Constant/Linear/Exponential/Jitter 的纯单元测试统一收敛在
// backoff_test.go;本文件仅保留与 [Do] 交互相关的集成式 backoff 测试。

// --- 集成:退避实际被应用到重试间隔 ----------------------------------------

func TestDo_BackoffActuallyApplied(t *testing.T) {
	// 验证 Do 确实用 backoff 的时长在两次尝试间等待,而非立即重试。
	r := New[int](WithMaxAttempts(3), WithBackoff(Constant(50*time.Millisecond)))
	var timestamps []time.Time
	_, _ = r.Do(context.Background(), func(ctx context.Context) (int, error) {
		timestamps = append(timestamps, time.Now())
		return 0, errBoom
	})
	if len(timestamps) != 3 {
		t.Fatalf("got %d timestamps, want 3", len(timestamps))
	}
	// 两次尝试间隔应 >= 50ms(退避生效),远小于则未等待。
	for i := 1; i < len(timestamps); i++ {
		gap := timestamps[i].Sub(timestamps[i-1])
		if gap < 45*time.Millisecond {
			t.Errorf("gap %d = %v, want >= ~45ms (backoff applied)", i, gap)
		}
	}
}

func TestDo_BackoffGrowsAcrossAttempts(t *testing.T) {
	// 线性退避:第 1->2 次间隔应 < 第 2->3 次间隔(Step 使其增长)。
	r := New[int](
		WithMaxAttempts(3),
		WithBackoff(Linear{Base: 20 * time.Millisecond, Step: 30 * time.Millisecond}),
	)
	var ts []time.Time
	_, _ = r.Do(context.Background(), func(ctx context.Context) (int, error) {
		ts = append(ts, time.Now())
		return 0, errBoom
	})
	gap1 := ts[1].Sub(ts[0]) // ~20ms
	gap2 := ts[2].Sub(ts[1]) // ~50ms
	if gap2 <= gap1 {
		t.Errorf("gap2 (%v) should exceed gap1 (%v) for growing backoff", gap2, gap1)
	}
}
