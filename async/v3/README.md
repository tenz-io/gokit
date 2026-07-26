# async

泛型并发任务执行器：一次性 fan-out 的 `Run`/`RunAll`/`Wait`/`AllOf`/`AnyOf`，加上可限流、可首错取消的 `Group` 构建器。所有任务经 panic 安全包装，panic 会被转成类型化 `*PanicError`，绝不拖垮进程。

```go
import "github.com/tenz-io/gokit/async/v3"
```

## 模块介绍

async 解决并发任务编排的几类典型场景：

- **一次性 fan-out**：任务集在调用前就已知，挑一个语义最贴的助手即可。
- **开放式构建**：任务按需加入、需要并发上限或首错即停时，用 `Group`。
- **panic 安全**：任意任务 panic 都被 recover 成 `*PanicError`（带原始值与堆栈），不写死日志、不依赖 `log`；调用方用 `errors.As` 决定怎么记。

V3 相对 V2 的核心变化：

- **统一可变参数**：`AllOf` 从 `[]Fn[T]` 改成 `...Task[T]`，与 `Run`/`Wait`/`AnyOf` 一致；手上有切片就 `tasks...` 展开。
- **结果模型精简**：`Holder[T]` → `Result[T]{Value, Err}`，去掉无用的未导出下标字段（位置即下标）。
- **panic 不再强写日志**：库不应绑定 `log`；panic 转成 `*PanicError`（实现 `Unwrap`），需要时 `errors.As` 自取堆栈。
- **新增 `RunAll`**：用 `errors.Join` 汇总**全部**错误（现代 Go idiom），区别于 `Run`（仅首个）。
- **新增 `Group`**：errgroup 风格的泛型构建器，`WithLimit` 限流、`WithCancelOnError` 首错取消，弥补 v2 无并发上限的痛点。

## 能力清单

| 能力 | 含义 |
| --- | --- |
| 并发取首个错误 | `Run` 并发跑全部任务，返回**首个**错误，不取消兄弟任务（errgroup 的非取消变体） |
| 并发汇总全部错误 | `RunAll` 并发跑全部任务，用 `errors.Join` 合并**所有**错误 |
| 并发等待、忽略错误 | `Wait` 并发跑全部任务并阻塞到完成，丢弃值与错误（fire-and-forget） |
| 按序收集全部结果 | `AllOf` 并发跑全部任务，按**输入顺序**返回每个任务的 `Result`，单任务失败不影响其他任务 |
| 抢首个成功结果 | `AnyOf` 并发跑全部任务，谁先成功返回谁，并通过取消派生 context 让其余任务尽快退出；全失败返回 joined 错误 |
| panic 转类型化错误 | 所有任务经 `recoverTask` 包装，panic 变成 `*PanicError`（带 `Value()` 与 `Stack()`），可 `errors.As` 提取、`errors.Is` 穿透被包裹的 error |
| 可限流构建器 | `Group` + `WithLimit(n)` 用信号量限定并发数，超限任务阻塞等位，避免一次开满 goroutine |
| 首错取消构建器 | `Group` + `WithCancelOnError()` 首次失败即取消派生 context，在跑任务可短路；`Wait` 返回该首错 |
| 空任务安全跳过 | `nil` 任务被静默过滤（`AnyOf` 例外，明确报错），不会触发 panic |

## 快速开始

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/tenz-io/gokit/async/v3"
)

func fetchUser(context.Context) (string, error) { return "user-1", nil }
func fetchOrder(context.Context) (string, error) { return "order-1", nil }

func main() {
	ctx := context.Background()

	// 1) 并发执行，只关心是否出错（首个错误）
	if err := async.Run(ctx, fetchUser, fetchOrder); err != nil {
		log.Fatal(err)
	}

	// 2) 按顺序收集每个任务的结果（值 + 错误互不影响）
	for i, r := range async.AllOf(ctx, fetchUser, fetchOrder) {
		fmt.Printf("[%d] val=%q err=%v\n", i, r.Value, r.Err)
	}

	// 3) 取最快成功的一个，其余被取消
	val, err := async.AnyOf(ctx, fetchUser, fetchOrder)
	fmt.Println("any:", val, err)

	// 4) 开放式构建：限流 + 首错取消
	g := async.New[int](ctx,
		async.WithLimit[int](4),       // 最多 4 个并发
		async.WithCancelOnError[int](), // 首次失败即取消其余
	)
	for i := 0; i < 100; i++ {
		i := i
		g.Go(func(ctx context.Context) (int, error) {
			if i < 0 { // 演示首错即停
				return 0, errors.New("bad")
			}
			return i, nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Println("group failed:", err)
	}
	_ = g.Results() // 成功结果（完成序，不含失败项）
}
```

### panic 处理

任务 panic 不会崩溃进程，而是以 `*PanicError` 形式出现在 `Result.Err` 或返回的 error 里：

```go
err := async.Run(ctx, func(context.Context) (int, error) { panic("boom") })
var pe *async.PanicError
if errors.As(err, &pe) {
	fmt.Println(pe.Value())      // 原始 panic 值
	_ = pe.Stack()               // 捕获的 goroutine 堆栈
}
// 若 panic 的是 error，errors.Is/As 可穿透到内层：
_ = errors.Is(err, context.Canceled)
```

## API 速查

| 符号 | 说明 |
| --- | --- |
| `type Task[T any] func(context.Context) (T, error)` | 统一任务签名 |
| `type Result[T any] struct{ Value T; Err error }` | 单任务结果，位置即下标 |
| `type PanicError struct{...}` | panic 转成的类型化错误，`Value()`/`Stack()`/`Unwrap()` |
| `func Run(ctx, tasks ...Task[T]) error` | 并发执行，返回首个错误（不取消兄弟） |
| `func RunAll(ctx, tasks ...Task[T]) error` | 并发执行，`errors.Join` 汇总全部错误 |
| `func Wait(ctx, tasks ...Task[T])` | 并发执行并等待完成，忽略值与错误 |
| `func AllOf(ctx, tasks ...Task[T]) []Result[T]` | 并发执行，按输入顺序返回全部结果 |
| `func AnyOf(ctx, tasks ...Task[T]) (T, error)` | 并发执行，返回首个成功结果，取消其余；全失败返回 joined 错误 |
| `type Group[T any]` | 泛型、panic 安全的 errgroup 风格构建器 |
| `func New[T](ctx, opts ...Option[T]) *Group[T]` | 创建 Group，派生可取消 context |
| `func (g) Go(Task[T])` | 提交任务；`WithLimit` 下阻塞等位；nil 被跳过 |
| `func (g) Wait() error` | 阻塞至全部完成；`WithCancelOnError` 返回首错，否则 `errors.Join` 全部 |
| `func (g) Results() []Result[T]` | 成功结果（完成序，拷贝返回，可安全改写） |
| `func WithLimit[T](n int) Option[T]` | 并发上限，非正数忽略（默认无限） |
| `func WithCancelOnError[T]() Option[T]` | 首次失败即取消派生 context、其余可短路 |

引入路径：`github.com/tenz-io/gokit/async/v3`
