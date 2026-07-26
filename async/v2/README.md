# async

泛型并发任务执行器：Run、Wait、AllOf、AnyOf，内置 panic 恢复和 errgroup 模式。

## 功能特性

- `Run`：并发执行多个任务，采用 errgroup 语义，返回首个出现的错误，但不会因某个任务失败而取消其余任务的 context
- `Wait`：并发执行多个任务并等待全部完成，忽略所有错误，适合“fire and forget”场景
- `AllOf`：并发执行多个任务，按输入顺序收集全部结果（含每个任务的返回值和错误），一个任务出错不影响其他任务继续执行
- `AnyOf`：并发执行多个任务，只要有一个成功即返回该结果，并通过取消共享 context 让其余任务尽快退出；全部失败时返回汇总错误
- 内置 `panic` 恢复机制：任意任务内部发生 panic 都会被捕获并转换为 `error`，同时打印堆栈日志，不会导致整个进程崩溃
- 基于泛型的统一任务签名 `Fn[T]`，可直接复用同一批任务函数分别传给 `Run`、`Wait`、`AllOf`、`AnyOf`
- `nil` 任务会被安全跳过（`Run`/`Wait`/`AllOf`）或返回明确错误（`AnyOf`），避免因误传空函数导致 panic

## 能力清单

| 能力 | 含义 |
|---|---|
| 并发执行并取首个错误 | `Run` 并发跑多个任务，errgroup 语义，返回第一个出现的错误，但不取消其他任务的 context，适合“都要跑完，但只关心是否有失败”的场景 |
| 并发执行并忽略错误 | `Wait` 并发跑多个任务并等待全部结束，不关心结果和错误，适合日志上报、缓存预热等“fire and forget”场景 |
| 按序收集全部结果 | `AllOf` 并发跑多个任务，用 `Holder[T]` 按输入顺序返回每个任务的值和错误，一个任务失败不影响其他任务继续跑完，适合批量拉取多个数据源后统一处理 |
| 抢首个成功结果 | `AnyOf` 并发跑多个任务，谁先成功就返回谁的结果，并通过取消共享 context 让其余任务尽快退出，全部失败时返回汇总错误，适合多副本/多数据源的“谁快用谁” |
| panic 自动恢复 | 所有任务统一经过 `panicProof` 包装，任务内部 panic 会被捕获并转换成 `error`（同时打印堆栈日志），避免一个任务的 panic 拖垂整个进程 |
| 空任务安全跳过或报错 | `Run`/`Wait`/`AllOf` 会安全跳过 `nil` 任务（`AllOf` 记录为对应位置的错误），`AnyOf` 遇到 `nil` 任务直接返回明确错误，避免误传空函数引发 panic |
| 统一泛型任务签名复用 | 所有函数共用 `Fn[T] func(context.Context) (T, error)` 签名，同一批任务函数可以直接分别传给 `Run`、`Wait`、`AllOf`、`AnyOf` 而无需改造 |
| 支持 context 取消传播 | 各函数均接收外部 `ctx`，任务函数内部可感知外部取消/超时；`AnyOf` 额外派生内部 context 用于提前终止未完成任务 |

## 快速开始

```go
import "github.com/tenz-io/gokit/async/v2"

func fetchUser(ctx context.Context) (string, error) {
	return "user-1", nil
}

func fetchOrder(ctx context.Context) (string, error) {
	return "order-1", nil
}

func main() {
	ctx := context.Background()

	// 并发执行，只关心是否出错
	if err := async.Run(ctx, fetchUser, fetchOrder); err != nil {
		log.Fatal(err)
	}

	// 并发执行，等待全部完成，忽略错误
	async.Wait(ctx, fetchUser, fetchOrder)

	// 并发执行，按顺序收集全部结果
	results := async.AllOf(ctx, []async.Fn[string]{fetchUser, fetchOrder})
	for _, r := range results {
		fmt.Println(r.Val, r.Err)
	}

	// 并发执行，取最快成功的一个
	val, err := async.AnyOf(ctx, fetchUser, fetchOrder)
	fmt.Println(val, err)
}
```

## API 速查

| 符号 | 说明 |
|---|---|
| `type Fn[T any] func(context.Context) (T, error)` | 统一的泛型异步任务签名 |
| `type Holder[T any]` | `AllOf` 的单个结果容器，包含 `Val` 和 `Err`，按输入顺序保存 |
| `func Run[T any](ctx, fns ...Fn[T]) error` | 并发执行所有任务，返回首个错误（errgroup 语义，不取消 context） |
| `func Wait[T any](ctx, fns ...Fn[T])` | 并发执行所有任务并等待完成，忽略错误 |
| `func AllOf[T any](ctx, fns []Fn[T]) []Holder[T]` | 并发执行所有任务，按顺序返回全部结果 |
| `func AnyOf[T any](ctx, fns ...Fn[T]) (T, error)` | 并发执行所有任务，返回首个成功结果，全部失败则返回汇总错误 |

使用前先 `import "github.com/tenz-io/gokit/async/v2"`。
