// Package retriever 提供泛型、可配置的重试执行器,内置可组合的退避策略。
//
// 本包是对 retriever/v2 的全新重写,目标是更简洁、更易用:
//   - [Retriever] 是泛型接口,其 [Retriever.Do] 返回类型化的 (T, error),
//     不再需要任何类型断言;
//   - 重试模型单一化:[Do] 默认对任何错误都重试,直到成功或达到上限;
//     若某个错误不可重试(如参数错误),用 [NonRetryable] 包装后 [Do] 会
//     立即返回它;更精细的"哪些错误值得重试"判断可用 [WithRetryable];
//   - 退避策略 [Backoff] 统一以 [time.Duration] 表达,均为可比较字段、
//     可直接用复合字面量构造的值类型,并可通过 [Jitter] 装饰器叠加抖动;
//   - 仅通过 [New] 构造,采用函数式 [Option],无导出的 Config 结构体。
//
// 重试期间,所有等待都会监听 ctx 的取消/超时,从而及时返回,不会多等一次
// 退避时长。可选的 [WithTimeout] 为全部尝试设置统一截止时间。
//
// 返回值约定:[Do] 的所有终止路径都携带 fn 最后一次返回的 result(连同
// error),便于调用方释放 fn 已创建但未消费的资源(如 *http.Response.Body)。
// fn 自行负责在中间失败、即将重试时丢弃本应释放的资源 —— [Do] 不会在重试
// 之间代为回收。ctx 终止(超时/取消)被判定后优先于尝试耗尽返回,使同一超时
// 不论发生在哪次尝试都统一归为 ctx 错误而非 [ErrMaxAttempts]。
package retriever

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// 错误哨兵。它们仅在 [Do] 的返回路径上产生,业务代码可用 errors.Is 判别。
var (
	// ErrNonRetryable 标记一个被 [NonRetryable] 包装的错误。[Do] 遇到它时
	// 会立即返回该错误,不再重试。该标记对错误字符串是透明的:被包装错误
	// 的 Error() 文本保持不变,业务原因仍是内层 error。
	ErrNonRetryable = errors.New("retriever: non-retryable error")
	// ErrMaxAttempts 在耗尽全部尝试次数仍失败时,作为包裹错误返回,
	// 内层为最后一次失败的 error。
	ErrMaxAttempts = errors.New("retriever: max attempts reached")
)

// nonRetryableError 是 [NonRetryable] 的内部实现。它的 Error() 直接委托给
// 内层 error,因此被标记错误的字符串文本与原错误完全一致 —— 不引入任何
// "retriever: ..." 噪声;同时通过 [errors.Is] 穿透支持 [ErrNonRetryable] 判别。
type nonRetryableError struct{ err error }

func (e *nonRetryableError) Error() string { return e.err.Error() }
func (e *nonRetryableError) Unwrap() error { return e.err }
func (e *nonRetryableError) Is(target error) bool {
	return target == ErrNonRetryable
}

// NonRetryable 包装 err,标记它对 [Do] 不可重试。nil 输入返回 nil。
// 标记是字符串透明的:Error() 文本与 err 一致,不改变其显示。
// [Do] 遇到此错误会立即原样返回它,不再重试。
func NonRetryable(err error) error {
	if err == nil {
		return nil
	}
	return &nonRetryableError{err: err}
}

// IsNonRetryable 报告 err 是否被 [NonRetryable] 标记(经由 [errors.Is] 穿透,
// 即便 err 又被进一步包装也能识别)。
func IsNonRetryable(err error) bool {
	return errors.Is(err, ErrNonRetryable)
}

// DoFunc 是一个可重试操作。返回非空 error 时,[Retriever.Do] 会按配置
// 重试,直到成功、达到上限、ctx 取消,或错误被 [NonRetryable] 标记。
//
// fn 若在返回 error 的同时也返回了非零值 result(例如已打开的
// *http.Response),该 result 会被 Do 原样带回(连同 error),便于调用方
// 释放其资源。fn 自行负责在中间失败、即将重试时释放本应丢弃的资源 ——
// Do 只在最终返回路径上携带 result,不会在重试之间代为回收。
type DoFunc[T any] func(ctx context.Context) (T, error)

// Retriever 用退避策略执行 [DoFunc]。泛型参数 T 即 Do 的返回值类型。
type Retriever[T any] interface {
	// Do 执行 fn 并按需重试。终止路径及返回值约定:
	//
	//   - fn 返回 nil error:立即返回 (result, nil)。
	//   - fn 返回 [NonRetryable] 标记的错误:立即返回 (result, err)。
	//   - 分类器 ([WithRetryable]) 判定不可重试:立即返回 (result, err)。
	//   - ctx 在 fn 调用后已终止(超时/取消):返回 (result, ctx.Err()),
	//     优先于 [ErrMaxAttempts],使同一超时不论发生在哪次尝试都统一归类。
	//   - 退避等待期间 ctx 取消:返回 (result, ctx.Err())。
	//   - 耗尽全部尝试仍失败:返回 (result, 用 [ErrMaxAttempts] 包裹的 lastErr)。
	//
	// 各终止路径都携带 fn 最后一次返回的 result(可能为零值),见 [DoFunc]
	// 关于资源释放的约定。
	Do(ctx context.Context, fn DoFunc[T]) (T, error)
}

// Option 在构造期配置一个 [Retriever]。
type Option func(*options)

type options struct {
	maxAttempts int
	timeout     time.Duration
	backoff     Backoff
	retryable   func(error) bool
}

func defaultOptions() options {
	return options{
		maxAttempts: 3,
		timeout:     0,
		backoff: Exponential{
			Base:   100 * time.Millisecond,
			Factor: 2,
		},
		retryable: nil,
	}
}

// WithMaxAttempts 设置最大尝试次数(含首次调用)。n<=0 视为无效,保留当前值
// (默认为 3):即便在本 option 之后再传其它有效值,本调用也不会覆盖 ——
// 即无效值是"不修改"而非"重置为默认"。需恢复默认请显式传 3。
func WithMaxAttempts(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.maxAttempts = n
		}
	}
}

// WithTimeout 为全部尝试设置统一截止时间。d<=0 表示不设全局截止时间,
// 仅受 ctx 自身的取消/超时约束。
func WithTimeout(d time.Duration) Option {
	return func(o *options) { o.timeout = d }
}

// WithBackoff 设置退避策略。nil 输入被忽略,保留默认的指数退避。
func WithBackoff(b Backoff) Option {
	return func(o *options) {
		if b != nil {
			o.backoff = b
		}
	}
}

// WithRetryable 设置错误分类器:返回 true 表示该错误值得重试,
// false 表示应立即返回。[NonRetryable] 标记的错误总是立即返回,
// 优先于本分类器。fn 为 nil 时,[Do] 默认对所有非 NonRetryable 错误重试。
func WithRetryable(fn func(error) bool) Option {
	return func(o *options) { o.retryable = fn }
}

// New 返回一个泛型 [Retriever]。零个 option 即可用,默认:最多 3 次尝试、
// 100ms 起步的指数退避、无全局截止时间、对所有错误重试。
func New[T any](opts ...Option) Retriever[T] {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return &retrier[T]{
		maxAttempts: o.maxAttempts,
		timeout:     o.timeout,
		backoff:     o.backoff,
		retryable:   o.retryable,
	}
}

// retrier 是 [Retriever] 的唯一实现。其自身字段在构造后不可变,故 Do 可被
// 多个 goroutine 并发调用。但 [Backoff] 实现与 [WithRetryable] 闭包可能
// 含共享可变状态:内置的 [Constant]/[Linear]/[Exponential]/[Jitter] 均
// 并发安全,自定义 Backoff 或分类器须由调用方自行保证并发安全。
type retrier[T any] struct {
	maxAttempts int
	timeout     time.Duration
	backoff     Backoff
	retryable   func(error) bool
}

// Do 实现 [Retriever]。
func (r *retrier[T]) Do(ctx context.Context, fn DoFunc[T]) (T, error) {
	var zero T
	if fn == nil {
		return zero, errors.New("retriever: fn is nil")
	}
	if ctx == nil {
		return zero, errors.New("retriever: ctx is nil")
	}

	// 可选的全局截止时间:它包裹所有尝试与退避等待。
	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}

	var (
		lastErr    error
		lastResult T
	)
	for attempt := 0; attempt < r.maxAttempts; attempt++ {
		// 在每次尝试前检查 ctx,使外部取消/超时能在退避等待期间被及时感知。
		if err := ctx.Err(); err != nil {
			return zero, err
		}

		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}
		// 记录本次失败:即便要终止,也把 fn 返回的 result 带回去 —— 调用方可能
		// 需要它(如已打开但需关闭的 *http.Response)。中间成功但被分类器/
		// NonRetryable 终止的失败,其资源仍由 fn 自行释放(见包文档约定)。
		lastResult, lastErr = result, err

		// fn 调用后若 ctx 已终止(超时/取消),优先按 ctx 终止归类,而非尝试耗尽:
		// 这样同一超时不论发生在哪次尝试,都统一返回 ctx.Err(),分类与序号解耦。
		if cerr := ctx.Err(); cerr != nil {
			return lastResult, cerr
		}

		// NonRetryable 标记优先:立即返回该错误(标记本身字符串透明)。
		if IsNonRetryable(err) {
			return lastResult, err
		}

		// 若提供了分类器,且它判定当前错误不可重试,则立即返回。
		if r.retryable != nil && !r.retryable(err) {
			return lastResult, lastErr
		}

		// 最后一次尝试不再等待退避。
		if attempt == r.maxAttempts-1 {
			break
		}

		// 退避等待,期间监听 ctx 取消。
		timer := time.NewTimer(r.backoff.Next(attempt))
		select {
		case <-ctx.Done():
			timer.Stop()
			return lastResult, ctx.Err()
		case <-timer.C:
		}
		timer.Stop()
	}

	return lastResult, fmt.Errorf("%w: %w", ErrMaxAttempts, lastErr)
}
