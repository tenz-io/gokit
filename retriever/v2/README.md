# retriever

可配置重试库。支持指数退避（含抖动）、线性、无退避三种策略。

```go
import "github.com/tenz-io/gokit/retriever/v2"
```

## 快速开始

```go
r := retriever.New(
    retriever.WithMaxAttempt(3),
    retriever.WithBackoff(retriever.NewExponentialBackoff(100, 2.0, 0.3)),
)
result, err := r.DoAlwaysRetry(ctx, func(ctx context.Context) (any, error) {
    return http.Get("https://api.example.com")
})
```
