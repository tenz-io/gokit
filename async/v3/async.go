// Package async 提供带 panic 恢复的通用并发任务执行能力。
//
// 本包提供两种互补的使用风格：
//
//   - 无状态辅助函数 [Run]、[RunAll]、[Wait]、[AllOf]、[AnyOf],适用于任务集
//     已知的一次性 fan-out。
//   - [Group] 构建器(见 [New]),适用于任务被增量添加的开放式场景,可选并发
//     上限与"首个出错即取消"。
//
// 每个任务都在 panic 安全的包装器中执行,panic 会被转换为有类型的
// [*PanicError],因此 panic 的任务永远不会让进程崩溃。
package async

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
)

// Task 是共享的异步任务签名。所有辅助函数与 [Group] 都基于它操作。
type Task[T any] func(context.Context) (T, error)

// Result 保存单个任务的结果。切片位置即对应输入顺序,因此没有单独的 index 字段。
type Result[T any] struct {
	Value T
	Err   error
}

// PanicError 在任务 panic 时返回。它携带被 recover 的值以及捕获的调用栈,
// 调用方可通过 [errors.As] 进行检视或记录日志。
type PanicError struct {
	value any
	stack []byte
}

// Value 返回传给 panic 的值。
func (p *PanicError) Value() any { return p.value }

// Stack 返回在 panic 发生处捕获的 goroutine 调用栈。
func (p *PanicError) Stack() []byte { return p.stack }

// Error 实现 error 接口。
func (p *PanicError) Error() string {
	switch v := p.value.(type) {
	case error:
		return fmt.Sprintf("async: panic: %v", v)
	case string:
		return fmt.Sprintf("async: panic: %s", v)
	default:
		return fmt.Sprintf("async: panic: %v", v)
	}
}

// Unwrap 使 errors.Is/As 能够穿透到包装了 error 的 panic 内部。
func (p *PanicError) Unwrap() error {
	if err, ok := p.value.(error); ok {
		return err
	}
	return nil
}

// Run 并发执行任务,并返回遇到的第一个错误(已过滤 nil)。出错时不会取消
// 兄弟任务(不会派生 context)。panic 会以 [*PanicError] 形式上报。任务集为
// 空时返回 nil。
func Run[T any](ctx context.Context, tasks ...Task[T]) error {
	tasks = filterNil(tasks)
	if len(tasks) == 0 {
		return nil
	}

	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		first error
	)

	for _, task := range tasks {
		wg.Add(1)
		go func(f Task[T]) {
			defer wg.Done()
			if _, err := recoverTask(f)(ctx); err != nil {
				mu.Lock()
				if first == nil {
					first = err
				}
				mu.Unlock()
			}
		}(task)
	}
	wg.Wait()
	return first
}

// RunAll 并发执行任务,并用 [errors.Join] 合并每个非 nil 的错误。panic 会以
// [*PanicError] 形式上报。当所有任务都成功,或任务集为空时返回 nil。
func RunAll[T any](ctx context.Context, tasks ...Task[T]) error {
	tasks = filterNil(tasks)
	if len(tasks) == 0 {
		return nil
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	for _, task := range tasks {
		wg.Add(1)
		go func(f Task[T]) {
			defer wg.Done()
			if _, err := recoverTask(f)(ctx); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(task)
	}
	wg.Wait()
	return errors.Join(errs...)
}

// Wait 并发执行任务并阻塞至全部完成,同时丢弃返回值与错误。它是
// fire-and-forget 变体。panic 会被吞掉(recover 但不返回);若必须观测到
// panic,请改用 [Run] 或 [RunAll]。
func Wait[T any](ctx context.Context, tasks ...Task[T]) {
	tasks = filterNil(tasks)
	if len(tasks) == 0 {
		return
	}
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go func(f Task[T]) {
			defer wg.Done()
			_, _ = recoverTask(f)(ctx)
		}(task)
	}
	wg.Wait()
}

// AllOf 并发执行任务,并按输入顺序为每个任务返回一个 [Result]。某个任务
// 失败或 panic 不影响其他任务;其结果会承载该错误(panic → [*PanicError])。
// 任务集为空时返回 nil。
func AllOf[T any](ctx context.Context, tasks ...Task[T]) []Result[T] {
	tasks = filterNil(tasks)
	if len(tasks) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	results := make([]Result[T], len(tasks))
	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, f Task[T]) {
			defer wg.Done()
			v, err := recoverTask(f)(ctx)
			results[idx] = Result[T]{Value: v, Err: err}
		}(i, task)
	}
	wg.Wait()
	return results
}

// AnyOf 并发执行任务,并返回第一个成功的结果,在首次成功时通过派生的
// context 取消其余任务。若所有任务都失败,这些失败会用 [errors.Join] 合并。
// panic 视作失败([*PanicError])。任务集为空则视为错误。
func AnyOf[T any](ctx context.Context, tasks ...Task[T]) (T, error) {
	var zero T
	if len(tasks) == 0 {
		return zero, errors.New("async.AnyOf: empty task set")
	}
	for _, t := range tasks {
		if t == nil {
			return zero, errors.New("async.AnyOf: nil task")
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		errs   []error
		won    bool
		winner T
	)

	for _, task := range tasks {
		wg.Add(1)
		go func(f Task[T]) {
			defer wg.Done()
			v, err := recoverTask(f)(ctx)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			if !won {
				won = true
				winner = v
				cancel()
			}
			mu.Unlock()
		}(task)
	}

	wg.Wait()
	if won {
		return winner, nil
	}
	return zero, fmt.Errorf("async.AnyOf: all %d task(s) failed: %w", len(tasks), errors.Join(errs...))
}

// recoverTask 包装一个任务,使 panic 被 recover 并以 [*PanicError] 形式上抛,
// 而非让进程崩溃。调用栈在 panic 发生处捕获一次。
func recoverTask[T any](task Task[T]) Task[T] {
	return func(ctx context.Context) (result T, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				err = &PanicError{value: rec, stack: stack()}
			}
		}()
		return task(ctx)
	}
}

// stack 返回当前 goroutine 的调用栈,供 [PanicError] 复用。
func stack() []byte {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return buf[:n]
}

// filterNil 返回剔除 nil 任务后的切片,保持原顺序。
func filterNil[T any](tasks []Task[T]) []Task[T] {
	out := tasks[:0:0]
	for _, t := range tasks {
		if t != nil {
			out = append(out, t)
		}
	}
	return out
}
