# async

泛型并发任务执行器。提供 `Run`、`Wait`、`AllOf`、`AnyOf` 四种模式，内置 panic 恢复。依赖 `golang.org/x/sync`（errgroup 模式已内联为轻量 channel）。

## 快速开始

```go
import "github.com/tenz-io/gokit/async"

// 并发执行，遇到第一个错误立即返回
err := async.Run(ctx,
    fetchUser,
    fetchOrder,
    fetchInventory,
)

// 并发执行，等待全部完成（忽略错误）
async.Wait(ctx, task1, task2, task3)

// 收集所有结果
fns := []async.Fn[int]{taskA, taskB, taskC}
results := async.AllOf(ctx, fns)
for _, r := range results {
    if r.Err != nil { /* handle */ }
    fmt.Println(r.Val)
}

// 取第一个成功的结果
user, err := async.AnyOf(ctx, fetchFromCache, fetchFromDb, fetchFromRemote)
```

## API 参考

### 类型

| 类型 | 定义 | 说明 |
|------|------|------|
| `Fn[T]` | `func(context.Context) (T, error)` | 泛型异步函数签名 |
| `Holder[T]` | `struct{ Val T; Err error }` | 携带结果或错误 |

### 执行函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `Run` | `func Run[T any](ctx context.Context, fns ...Fn[T]) error` | 并发执行，返回第一个错误。空列表返回 nil |
| `Wait` | `func Wait[T any](ctx context.Context, fns ...Fn[T])` | 并发执行，等待所有完成，忽略错误 |
| `AllOf` | `func AllOf[T any](ctx context.Context, fns []Fn[T]) []Holder[T]` | 并发执行，收集所有结果（保持输入顺序） |
| `AnyOf` | `func AnyOf[T any](ctx context.Context, fns ...Fn[T]) (T, error)` | 并发执行，返回第一个成功结果。所有失败则返回组合错误 |

## 执行模式对比

| 模式 | 错误处理 | 返回值 | 提前退出 | 适用场景 |
|------|----------|--------|----------|----------|
| `Run` | 收第一个错误 | 无 | 否 | 并行任务，一个失败就够 |
| `Wait` | 全部忽略 | 无 | 否 | 只关心副作用，不在乎成败 |
| `AllOf` | 每项独立 | `[]Holder[T]` | 否 | 需要所有任务的各自结果 |
| `AnyOf` | 聚合所有失败 | 首个成功值 | 是（成功时取消其余） | 竞速模式 |

## 最佳实践

### Panic 安全

所有函数自动捕获 goroutine 内的 panic，转换为 error 返回。日志会包含完整堆栈。

```go
// 即使 fn 内部 panic，Run 也不会崩溃
async.Run(ctx, func(ctx context.Context) (int, error) {
    panic("unexpected")
})
// 返回 error: "panic: unexpected"
```

### AnyOf 用于降级

利用 `AnyOf` 的竞速机制实现优雅降级：

```go
result, err := async.AnyOf(ctx,
    func(ctx context.Context) (*Data, error) {
        return cache.Get(ctx, key)   // 最快
    },
    func(ctx context.Context) (*Data, error) {
        return db.Query(ctx, key)    // 中等
    },
    func(ctx context.Context) (*Data, error) {
        return remoteAPI.Fetch(ctx, key) // 最慢
    },
)
```

### 避免泄露 goroutine

`AnyOf` 成功后会通过 context 取消通知其他 goroutine。确保被调用的函数响应 context 取消：

```go
func fetchFromDb(ctx context.Context) (*Data, error) {
    // 检测到 ctx 取消时尽快返回
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case result := <-dbChan:
        return result, nil
    }
}
```

## 与 v1 的区别

| 变更项 | v1 | v2 |
|--------|----|----|
| `Builder` 类型 | 存在 | 移除（冗余包装） |
| `Job[T]` 类型 | 存在 | 移除（直接使用 `Fn[T]`） |
| `NewJob` | 存在 | 移除 |
| `Run` 错误聚合 | 空任务列表返回 error | 空任务列表返回 nil |
| `AnyOf` 错误聚合 | 返回单个错误 | 聚合所有错误信息 |
| `golang.org/x/sync` | 依赖 | 移除（内联实现） |
