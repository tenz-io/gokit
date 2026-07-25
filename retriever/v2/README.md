# retriever

可配置重试库，支持指数退避（含抖动）、线性退避、无退避三种策略，并可设置最大尝试次数和跨全部尝试的 context 超时。

## 功能特性

- `Do` 支持按错误动态决定是否重试（业务函数自行返回 `retry bool`），可精确区分"可重试错误"与"直接失败"
- `DoAlwaysRetry` 简化场景：只要返回 err 就一直重试，无需自己判断 retry 标志
- `ExponentialBackoff` 指数退避，支持基础等待时间、增长因子和随机抖动，避免重试风暴
- `LinearBackoff` 固定间隔退避，适合对延迟敏感、退避策略要简单可预测的场景
- `NoBackoff` 立即重试，无等待
- `WithMaxAttempt` 控制最大尝试次数（含首次调用），`WithMaxTotalAttemptTime` 控制所有尝试合计的超时时间
- 通过 `New`（函数式选项）或 `NewRetriever`（显式 `Config`）两种方式创建 `Retriever`，均带有合理默认值

## 快速开始

```go
import "github.com/tenz-io/gokit/retriever/v2"

r := retriever.New(
    retriever.WithMaxAttempt(3),
    retriever.WithMaxTotalAttemptTime(5*time.Second),
    retriever.WithBackoff(retriever.NewExponentialBackoff(100, 2.0, 0.3)),
)

// 只要返回 err 就重试，直到成功或达到上限
result, err := r.DoAlwaysRetry(context.Background(), func(ctx context.Context) (any, error) {
    resp, err := http.Get("https://api.example.com")
    return resp, err
})

// 自行决定某个错误是否值得重试
result, err = r.Do(context.Background(), func(ctx context.Context) (any, bool, error) {
    resp, err := http.Get("https://api.example.com")
    if err != nil {
        return nil, true, err // 网络错误，重试
    }
    if resp.StatusCode == http.StatusBadRequest {
        return resp, false, fmt.Errorf("bad request") // 客户端错误，不重试
    }
    return resp, false, nil
})
```

## API 速查

| 符号 | 说明 |
|---|---|
| `Retriever` | 重试执行器接口，提供 `Do` 和 `DoAlwaysRetry` 两个方法 |
| `New(opts ...ConfigOpt) Retriever` | 通过函数式选项创建 `Retriever`，基于默认配置叠加 |
| `NewRetriever(config Config) Retriever` | 通过显式 `Config` 创建 `Retriever` |
| `Retriever.Do(ctx, fn DoFunc) (any, error)` | 执行函数，由 `fn` 返回值决定是否重试 |
| `Retriever.DoAlwaysRetry(ctx, fn DoFuncAlwaysRetry) (any, error)` | 执行函数，只要出错就一直重试直到成功或达到上限 |
| `DoFunc` | 可重试函数类型：`func(ctx) (resp any, retry bool, err error)` |
| `DoFuncAlwaysRetry` | 可重试函数类型：`func(ctx) (resp any, err error)` |
| `Config` | 重试参数：`MaxAttempt`、`MaxTotalAttemptTime`、`Backoff` |
| `ConfigOpt` | 配置 `Config` 的函数式选项类型 |
| `WithMaxAttempt(n int) ConfigOpt` | 设置最大尝试次数（含首次） |
| `WithMaxTotalAttemptTime(d time.Duration) ConfigOpt` | 设置所有尝试合计的超时时间 |
| `WithBackoff(b Backoff) ConfigOpt` | 设置退避策略 |
| `Backoff` | 退避策略接口：`Next(failCount int) time.Duration` |
| `NoBackoff` | 无退避，立即重试 |
| `LinearBackoff` | 固定间隔退避 |
| `NewLinearBackoff(ms int64) Backoff` | 创建 `LinearBackoff`，参数为毫秒 |
| `ExponentialBackoff` | 指数退避（支持抖动） |
| `NewExponentialBackoff(base, factor, jitter float64) Backoff` | 创建 `ExponentialBackoff`，base 为毫秒，jitter 建议在 `[0, 1.0)` |

引入路径：`import "github.com/tenz-io/gokit/retriever/v2"`
