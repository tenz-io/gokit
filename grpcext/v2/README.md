# grpcext

gRPC 拦截器。请求追踪、流量日志、Prometheus 指标。

```go
import "github.com/tenz-io/gokit/grpcext/v2"
```

## 快速开始

```go
interceptor := grpcext.NewInterceptorWithOpts(
    grpcext.WithServerTraffic(true),
    grpcext.WithServerMetrics(true),
)
```
