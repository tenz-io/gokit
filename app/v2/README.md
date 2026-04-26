# app

应用生命周期框架。配置加载、日志初始化、admin HTTP 端点、优雅退出。

```go
import "github.com/tenz-io/gokit/app/v2"
```

## 快速开始

```go
app.Run(app.Config{
    Name:  "my-service",
    Inits: []app.InitFunc{
        app.WithYamlConfig(),
        app.WithLogger(true),
        app.WithAdminHTTPServer(),
    },
    RunFn: func(ctx *app.Context, cfg any) error {
        app.FromContext(ctx).Info("service started")
        <-ctx.Done()
        return nil
    },
})
```
