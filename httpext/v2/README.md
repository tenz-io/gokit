# httpext

HTTP 客户端扩展。可组合传输层拦截器链：header 注入、流量、指标、慢请求日志。

```go
import "github.com/tenz-io/gokit/httpext/v2"
```

## 快速开始

```go
interceptor := httpext.NewInterceptorWithOpts(
    httpext.WithEnableTraffic(true),
    httpext.WithEnableMetrics(true),
)
client := &http.Client{}
interceptor.Apply(client)
```
