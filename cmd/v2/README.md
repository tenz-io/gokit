# cmd

基于 urfave/cli/v2 的 CLI 启动框架，与 app 模块提供相同的基础能力（配置加载、日志初始化、admin 端点、优雅退出），但以 CLI 命令和标志位的形式暴露，适合需要子命令、参数解析的可执行程序。

## 功能特性

- `Run` 一行启动一个 CLI App，自动挂载 `-h/--help`、`-v/--verbose` 等公共标志，支持自定义标志和子命令
- `WithYamlConfig` / `WithJsonConfig` 按命令行 `-c/--config` 指定的路径读取配置文件并反序列化到配置结构体，读取前自动应用 annotation 默认值、读取后自动做 annotation 校验
- `WithDotEnvConfig` 读取 `.env` 文件并注入进程环境变量
- `WithLogger` 按 `-v` `-l` `-s` 标志初始化 logger（日志级别、输出目录、是否打印到控制台、是否记录 traffic 日志）
- `WithAdminHTTPServer` 启动一个后台 HTTP 端口，挂载 `/ping`、`/metrics`、`/debug/pprof/*`，用于健康检查、Prometheus 采集与性能分析
- `WithUpdateConfigByEnv` / `UpdateConfig` 递归扫描配置结构体，把形如 `${ENV_NAME}` 的字符串字段替换成对应环境变量的值
- `GetConfig[Ptr]` 在命令处理函数中按类型安全地取回启动时加载的配置指针

## 快速开始

```go
import "github.com/tenz-io/gokit/cmd/v2"
```

```go
package main

import "github.com/tenz-io/gokit/cmd/v2"

type Config struct {
	Addr string `yaml:"addr" default:":8080"`
}

func main() {
	confPtr := &Config{}

	_ = cmd.Run(cmd.App{
		Name:    "my-cli",
		Usage:   "示例 CLI 服务",
		ConfPtr: confPtr,
		Inits: []cmd.InitFunc{
			cmd.WithYamlConfig(),
			cmd.WithLogger(true),
			cmd.WithAdminHTTPServer(),
		},
		Run: func(c *cmd.Context, confPtr any, errC chan<- error) {
			cfg := confPtr.(*Config)
			// 在这里启动业务服务，例如监听 cfg.Addr
			_ = cfg
			errC <- nil
		},
	}, nil)
}
```

在子命令的 Action 中取回配置：

```go
func action(c *cmd.Context) error {
	cfg, err := cmd.GetConfig[*Config](c)
	if err != nil {
		return err
	}
	_ = cfg
	return nil
}
```

## API 速查

| 符号 | 说明 |
|---|---|
| `App` | CLI 应用定义：名称、用法、配置指针、初始化函数列表、运行函数 |
| `Run(app App, extraFlags []Flag, extraCommands ...*Command) error` | 构建并运行 CLI App，执行 Inits、调用 Run，处理信号与退出 |
| `InitFunc` / `RunFunc` / `CleanFunc` | 初始化函数、运行函数、清理函数的类型定义 |
| `Command` / `Context` / `Flag` | `cli.Command` / `cli.Context` / `cli.Flag` 的类型别名 |
| `StringFlag` / `BoolFlag` / `IntFlag` | 常用标志类型别名，用于定义 `extraFlags` |
| `WithYamlConfig() InitFunc` | 读取 YAML 配置文件（`-c` 标志指定路径）并注入 confPtr |
| `WithJsonConfig() InitFunc` | 读取 JSON 配置文件并注入 confPtr |
| `WithDotEnvConfig() InitFunc` | 读取 `.env` 文件（`-e` 标志指定路径）并加载到环境变量 |
| `WithLogger(trafficEnabled bool) InitFunc` | 按标志初始化 logger，可选开启 traffic 日志 |
| `WithAdminHTTPServer() InitFunc` | 启动 admin HTTP 端口（`-a` 标志指定端口），挂载 ping/metrics/pprof |
| `WithUpdateConfigByEnv() InitFunc` | 用环境变量覆盖配置中 `${ENV_NAME}` 格式的字符串字段 |
| `ReadConfig(path string, confPtr any, unmarshalFn func([]byte, any) error) error` | 读取文件、应用默认值、反序列化、校验的完整流程 |
| `UpdateConfig(ptr any) error` | 递归替换结构体中 `${ENV_NAME}` 格式的字符串字段 |
| `GetConfig[Ptr any](c *Context) (Ptr, error)` | 从 App Metadata 中按类型取回配置指针 |
| `AddPingHandler(m *http.ServeMux)` | 注册 `/ping` 健康检查接口 |
| `AddPrometheusHandler(m *http.ServeMux)` | 注册 `/metrics` Prometheus 采集接口 |
| `AddProfilingHandler(m *http.ServeMux)` | 注册 `/debug/pprof/*` 性能分析接口 |
| `PingHandler(w http.ResponseWriter, r *http.Request)` | `/ping` 的 HandlerFunc 实现，返回主机名与运行时长 |
| `FlagNameConfig/FlagNameEnv/FlagNameLog/FlagNameVerbose/FlagNameConsole/FlagNameAdmin/FlagNameHelp` | 内置标志名常量 |

依赖 `github.com/tenz-io/gokit/annotation/v3`（配置默认值与校验）与 `github.com/tenz-io/gokit/logger/v2`（日志初始化）。

引入路径：`github.com/tenz-io/gokit/cmd/v2`
