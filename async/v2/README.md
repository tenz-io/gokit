# async

泛型并发任务执行器。Run、Wait、AllOf、AnyOf 四种模式，内置 panic 恢复。

```go
import "github.com/tenz-io/gokit/async/v2"
```

## 快速开始

```go
async.Run(ctx, fetchUser, fetchOrder)
result, err := async.AnyOf(ctx, fromCache, fromDB)
results := async.AllOf(ctx, []async.Fn[int]{taskA, taskB})
```
