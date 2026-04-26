# monitor

Prometheus 指标监控。histogram、counter、gauge、summary，context 传递。

```go
import "github.com/tenz-io/gokit/monitor/v2"
```

## 快速开始

```go
ctx = monitor.InitSingleFlight(ctx, "getUser")
rec := monitor.BeginRecord(ctx, "total")
defer rec.EndWithCode("200")
```
