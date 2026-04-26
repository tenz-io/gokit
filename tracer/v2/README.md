# tracer

基于 context 的请求 ID 传递及 debug/stress/shadow 标志位。

```go
import "github.com/tenz-io/gokit/tracer/v2"
```

## 快速开始

```go
ctx = tracer.WithRequestId(ctx, "req-123")
id := tracer.RequestIdFromCtx(ctx)
ctx = tracer.WithFlag(ctx, tracer.FlagDebug|tracer.FlagShadow)
```
