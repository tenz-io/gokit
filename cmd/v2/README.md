# cmd

CLI 启动框架，基于 urfave/cli。配置加载、日志初始化、admin 端点、优雅退出。

```go
import "github.com/tenz-io/gokit/cmd/v2"
```

## 快速开始

```go
cmd.Run(cmd.App{
    Name: "my-cli",
    Inits: []cmd.InitFunc{
        cmd.WithYamlConfig(),
        cmd.WithLogger(true),
    },
    Run: func(c *cmd.Context) error { return nil },
})
```
