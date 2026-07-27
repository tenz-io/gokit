package async

import (
	"context"
	"errors"
	"sync"
)

// Option 用于配置 [Group]。
type Option[T any] func(*Group[T])

// WithLimit 将并发任务数上限设为 n。超出上限的任务会阻塞,直到有空位释放。
// 非正数将被忽略(默认无上限)。本选项不影响 cancel-on-error 行为。
func WithLimit[T any](n int) Option[T] {
	return func(g *Group[T]) {
		if n > 0 {
			g.limit = n
		}
	}
}

// WithCancelOnError 使 [Group.Wait] 在首次失败时取消其余任务(panic 也算
// 失败)并返回该错误。未设置本选项时,group 会把每个任务都运行到完成,并
// 用 [errors.Join] 合并所有错误。传给任务的派生 context 会在首次出错时被
// 取消,以便在途任务可以短路退出。
func WithCancelOnError[T any]() Option[T] {
	return func(g *Group[T]) {
		g.cancelOnError = true
	}
}

// Group 是一个通用、panic 安全的 errgroup 风格构建器,服务于共享同一类型
// 参数 T 的任务。用 [Group.Go] 添加任务;收集合并后的错误([Group.Wait])
// 或有序结果([Group.Results])。零值 Group 不可用——务必通过 [New] 获取。
type Group[T any] struct {
	ctx           context.Context
	derived       context.Context
	cancel        context.CancelFunc
	limit         int
	cancelOnError bool

	sem      chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex
	errs     []error
	firstErr error // 仅在设置了 cancelOnError 时填充

	results []Result[T]
	started bool
}

// New 返回一个绑定到 ctx 的 [Group]。通过 [Group.Go] 添加的任务会收到一个
// 派生的 context:仅当设置了 [WithCancelOnError] 时,它才会在首次出错时被
// 取消;否则它仅跟踪父 context 的取消。
func New[T any](ctx context.Context, opts ...Option[T]) *Group[T] {
	derived, cancel := context.WithCancel(ctx)
	g := &Group[T]{
		ctx:     ctx,
		derived: derived,
		cancel:  cancel,
		results: nil, // 按需增长
	}
	for _, opt := range opts {
		opt(g)
	}
	if g.limit > 0 {
		g.sem = make(chan struct{}, g.limit)
	}
	return g
}

// Go 提交一个任务以并发执行。当设置了 [WithLimit] 时,它会阻塞直到获得
// 并发空位。不支持 Go 与 Wait 并发调用:先添加完所有任务,再调用 Wait。
//
// nil 任务为 no-op。结果按完成顺序收集;仅当未使用
// [WithCancelOnError] 时,才应在 [Group.Wait] 之后调用 [Group.Results]
// (被取消的任务不会产生结果)。
func (g *Group[T]) Go(task Task[T]) {
	if task == nil {
		return
	}
	g.started = true
	g.wg.Add(1)
	if g.sem != nil {
		g.sem <- struct{}{}
	}
	go func() {
		defer g.wg.Done()
		if g.sem != nil {
			defer func() { <-g.sem }()
		}
		v, err := recoverTask(task)(g.derived)
		if err != nil {
			g.mu.Lock()
			// 在 cancel-on-error 模式下,因 group 已被取消(ctx.Err())而失败
			// 的任务属于下游症状,而非独立失败:丢弃它,使 Wait 只上报首个错误。
			if g.cancelOnError && g.firstErr != nil && errors.Is(err, context.Canceled) {
				g.mu.Unlock()
				return
			}
			g.errs = append(g.errs, err)
			if g.cancelOnError && g.firstErr == nil {
				g.firstErr = err
				g.cancel()
			}
			g.mu.Unlock()
			return
		}
		g.mu.Lock()
		g.results = append(g.results, Result[T]{Value: v})
		g.mu.Unlock()
	}()
}

// Wait 阻塞至所有已提交任务完成,并返回合并后的错误。设置
// [WithCancelOnError] 时返回首个失败;否则合并每个错误。没有任务(或仅有
// nil 任务)的 group 返回 nil。
func (g *Group[T]) Wait() error {
	g.wg.Wait()
	g.cancel()
	if g.cancelOnError && g.firstErr != nil {
		return g.firstErr
	}
	return errors.Join(g.errs...)
}

// Results 返回在 [Group.Wait] 期间收集的成功结果,按完成顺序排列。失败的
// 任务不贡献任何结果。在 Wait 完成之前、或没有任务成功时返回 nil。返回
// 的切片为副本,可随意修改而不影响 group。
func (g *Group[T]) Results() []Result[T] {
	if len(g.results) == 0 {
		return nil
	}
	out := make([]Result[T], len(g.results))
	copy(out, g.results)
	return out
}
