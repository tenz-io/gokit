# app

应用生命周期框架：配置加载、日志初始化、admin HTTP 端点（pprof/Prometheus/健康检查）、优雅退出。

## 功能特性

- `app.Run` 统一编排：解析命令行参数 -> 依次执行 `Inits` 初始化函数 -> 启动 `Run` 主逻辑 -> 阻塞等待退出信号并执行清理
- `WithYamlConfig` / `WithJsonConfig` 读取配置文件并反序列化到业务配置结构体，内置默认值填充与校验（依赖 `annotation`）
- `WithDotEnvConfig` 加载 `.env` 文件到环境变量
- `WithLogger` 一行初始化全局日志（依赖 `logger`），支持文件/控制台输出、verbose 级别、流量日志开关
- `WithAdminHTTPServer` 启动独立的 admin HTTP 端口，自动挂载 `/debug/pprof`、`/metrics`、`/ping`
- 内置命令行 flag：`config`、`port`、`admin-port`、`log`、`logging-file`、`logging-console`、`verbose`
- `WaitSignal` 监听系统信号 / context 取消 / 运行错误，统一触发清理钩子并退出进程

## 快速开始

```go
import "github.com/tenz-io/gokit/app/v2"
```

```go
package main

import (
	"github.com/tenz-io/gokit/app/v2"
	"github.com/tenz-io/gokit/logger/v2"
)

type AppConfig struct {
	Name string `yaml:"name"`
}

func main() {
	app.Run(app.Config{
		Name: "my-service",
		Conf: &AppConfig{},
		Inits: []app.InitFunc{
			app.WithYamlConfig(),
			app.WithLogger(true),
			app.WithAdminHTTPServer(),
		},
		Run: func(c *app.Context, confPtr any, errC chan<- error) {
			cfg := confPtr.(*AppConfig)
			logger.Infof("service %s started", cfg.Name)
			<-c.Done()
		},
	}, nil)
}
```

## API 速查

| 名称 | 说明 |
| --- | --- |
| `Run(cfg Config, flags []Flag)` | 应用入口：解析 flag、执行初始化、运行主逻辑、等待退出 |
| `Config` | 应用配置：`Name`/`Usage`/`Conf`/`Inits`/`Run` |
| `InitFunc` | 初始化函数签名 `func(c *Context, confPtr any) (CleanFunc, error)` |
| `RunFunc` | 主运行函数签名 `func(c *Context, confPtr any, errC chan<- error)` |
| `CleanFunc` | 清理函数签名 `func()` |
| `Context` | 内嵌 `context.Context`，附带解析后的命令行 `Flags` |
| `NewContext(ctx, flags) *Context` | 构造 `Context` |
| `(*Context).GetFlags() *Flags` | 获取命令行 flag 集合 |
| `WithYamlConfig() InitFunc` | 读取 YAML 配置文件并反序列化到 `Conf` |
| `WithJsonConfig() InitFunc` | 读取 JSON 配置文件并反序列化到 `Conf` |
| `WithDotEnvConfig(filenames ...string) InitFunc` | 加载 `.env` 文件到环境变量 |
| `WithLogger(trafficEnabled bool) InitFunc` | 按 flag 配置初始化全局日志 |
| `WithAdminHTTPServer() InitFunc` | 启动 admin HTTP 服务（pprof/metrics/ping） |
| `ReadConfig(confPath string, confPtr any, unmarshalFn func([]byte, any) error) error` | 通用配置文件读取+默认值+校验 |
| `AddPrometheusHandler(m *http.ServeMux)` | 挂载 `/metrics` 处理器 |
| `AddPingHandler(m *http.ServeMux)` | 挂载 `/ping` 健康检查处理器 |
| `AddProfilingHandler(m *http.ServeMux)` | 挂载 `/debug/pprof/*` 系列处理器 |
| `PingHandler(w http.ResponseWriter, r *http.Request)` | `/ping` 的具体处理逻辑，返回主机名和运行时长 |
| `Flag` / `StringFlag` / `IntFlag` / `BoolFlag` | 命令行 flag 定义接口及实现 |
| `NewFlags(flags []Flag) (*Flags, error)` | 构造 flag 集合（去重校验） |
| `Parse(name string, flags *Flags) error` | 解析命令行参数到 flag 集合 |
| `(*Flags).String/Int/Bool(name) (T, error)` | 按名称、类型读取 flag 值 |
| `(*Flags).IsSet(name) bool` | 判断 flag 是否存在 |
| `(*Flags).Print()` | 打印当前所有 flag 值 |
| `FlagNameConfig` / `FlagNamePort` / `FlagNameAdminPort` / `FlagNameLog` / `FlagNameLoggingFile` / `FlagNameLoggingConsole` / `FlagNameVerbose` | 内置 flag 名称常量 |
| `WaitSignal(ctx, errC, hook func())` | 等待信号/取消/错误并触发清理后退出进程 |
| `If[T any](cond bool, a, b T) T` | 三元表达式风格的条件选择 |
| `PrettyString(v any) string` | 将任意值格式化为 JSON 字符串用于打印 |

引入路径：`github.com/tenz-io/gokit/app/v2`
