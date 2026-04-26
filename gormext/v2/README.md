# gormext

GORM 插件。追踪、流量日志、Prometheus 指标、慢查询检测。

```go
import "github.com/tenz-io/gokit/gormext/v2"
```

## 快速开始

```go
tracker := gormext.NewTrackerWithOpts(
    gormext.WithTraffic(true),
    gormext.WithMetrics(true),
)
tracker.Apply(db)
```
