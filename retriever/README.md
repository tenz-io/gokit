# retriever

可配置的重试库。支持自定义退避策略、最大尝试次数、总超时时间。零依赖。

## 快速开始

```go
import "github.com/tenz-io/gokit/retriever"

// 使用默认配置（3 次尝试，指数退避 100ms 起）
r := retriever.New()

result, err := r.Do(ctx, func(ctx context.Context) (any, bool, error) {
    resp, err := http.Get("https://api.example.com")
    if err != nil {
        return nil, true, err // retryable
    }
    return resp, false, nil
})

// 自定义配置
r = retriever.New(
    retriever.WithMaxAttempt(5),
    retriever.WithMaxTotalAttemptTime(10 * time.Second),
    retriever.WithBackoff(retriever.NewExponentialBackoff(200, 2.0, 0.3)),
)

// 简单的"一直重试直到成功"模式
result, err = r.DoAlwaysRetry(ctx, func(ctx context.Context) (any, error) {
    return connect("tcp://backend:5432")
})
```

## API 参考

### 构造

| 函数 | 签名 | 说明 |
|------|------|------|
| `New` | `func New(opts ...ConfigOpt) Retriever` | 函数式选项构造 |
| `NewRetriever` | `func NewRetriever(config Config) Retriever` | 结构体配置构造 |

### Retriever 接口

| 方法 | 签名 | 说明 |
|------|------|------|
| `Do` | `func Do(ctx context.Context, fn DoFunc) (any, error)` | 执行 fn，根据 `(retry, err)` 决定是否重试 |
| `DoAlwaysRetry` | `func DoAlwaysRetry(ctx context.Context, fn DoFuncAlwaysRetry) (any, error)` | 执行 fn，任何错误都会重试 |

### 配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `MaxAttempt` | int | 3 | 最大尝试次数（含首次） |
| `MaxTotalAttemptTime` | time.Duration | 0（无限制） | 所有尝试的总时间上限 |
| `Backoff` | Backoff | `ExponentialBackoff(100, 2.0, 0.3)` | 退避策略 |

### 退避策略

| 策略 | 构造函数 | 行为 |
|------|----------|------|
| `NoBackoff` | `&NoBackoff{}` | 不等待，立即重试 |
| `LinearBackoff` | `NewLinearBackoff(ms)` | 每次等待固定时长 |
| `ExponentialBackoff` | `NewExponentialBackoff(base, factor, jitter)` | `base * (factor^failCount + random(0, jitter))` ms |

## 参数配置默认值

| 参数 | 默认值 | 含义 |
|------|--------|------|
| `MaxAttempt` | 3 | 最多尝试 3 次 |
| `MaxTotalAttemptTime` | 0 | 不限总时间 |
| `Backoff.Base` | 100ms | 首次退避 100ms |
| `Backoff.Factor` | 2.0 | 每次翻倍 |
| `Backoff.Jitter` | 0.3 | 最多 30% 随机抖动 |

实际退避序列（默认配置，无 jitter 情况下）：100ms → 200ms → 400ms。

## 最佳实践

### 选择合适的退避策略

```go
// 内部 API 调用，快速重试
retriever.New(
    retriever.WithMaxAttempt(3),
    retriever.WithBackoff(retriever.NewLinearBackoff(50)),
)

// 外部 API，逐步增大间隔
retriever.New(
    retriever.WithMaxAttempt(5),
    retriever.WithBackoff(retriever.NewExponentialBackoff(1000, 2.0, 0.5)),
)

// 带总超时的关键路径
retriever.New(
    retriever.WithMaxAttempt(10),
    retriever.WithMaxTotalAttemptTime(5 * time.Second),
    retriever.WithBackoff(retriever.NewExponentialBackoff(50, 1.5, 0.2)),
)
```

### Do vs DoAlwaysRetry

- `Do`：函数返回 `(result, false, err)` 表示不可重试的错误（如业务逻辑错误、数据验证失败），立即返回。
- `DoAlwaysRetry`：所有错误都是可重试的（如网络超时、服务不可用）。

```go
// 区分可重试与不可重试
r.Do(ctx, func(ctx context.Context) (any, bool, error) {
    if errors.Is(err, &ValidationError{}) {
        return nil, false, err // 不可重试
    }
    return nil, true, err // 可重试
})
```

### 退避抖动

抖动（jitter）是所有客户端分散重试时间的关键。生产环境中始终设置 >0 的 jitter 值，避免惊群效应（thundering herd）：

```go
// 推荐 jitter 范围：0.1 ~ 0.5
retriever.WithBackoff(retriever.NewExponentialBackoff(100, 2.0, 0.3))
```

## 与 v1 的区别

| 变更项 | v1 | v2 |
|--------|----|----|
| 构造函数 | `NewRetrieverWithConfig` | `New(...ConfigOpt)` (更简洁) |
| `UseGoroutine` | 存在 | 移除（语义混乱，少见场景） |
| `pow` 手写幂函数 | 存在 | 使用 `math.Pow` |
| 错误信息 | 无包名前缀 | `retriever:` 前缀便于定位 |
