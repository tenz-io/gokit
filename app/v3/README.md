# app

应用生命周期框架：命令行 flag 解析、有序初始化（带 LIFO 清理）、admin HTTP 端点（pprof / metrics / ping）与信号驱动的优雅退出。v3 基于 `logger/v3` 与 `annotation/v3` 的一次无兼容包袱的干净重写，重点修复了 v2 的启动缺陷。

```go
import "github.com/tenz-io/gokit/app/v3"
```

## 模块介绍

app 解决四类问题：

- **统一编排**（`Run`）：解析 flag → 依次执行 `Inits` → 启动 `Run` 主逻辑 → 阻塞等待退出信号并执行清理。`Run` 返回退出码而不是内部 `os.Exit`，调用方决定是否退出进程。
- **有序初始化与 LIFO 清理**：`Inits` 顺序执行，每个 `InitFunc` 可返回一个 `CleanFunc`；退出或中途失败时按**逆序**调用已收集的清理函数，不会因后续初始化失败而泄漏已获取的资源。
- **配置加载**（`WithYamlConfig` / `WithJsonConfig` / `WithDotEnvConfig`）：读取文件、应用 `annotation` 默认值、**展开环境变量占位符**、反序列化、校验，一步完成。
- **环境变量占位符**（`${VAR}`）：配置文件里写 `${VAR}` / `${VAR:-default}` / `${VAR:?msg}`，实际值来自 `.env` 注入的环境变量——敏感配置（密码、DSN、API key）不落盘。
- **Admin HTTP 端点**（`WithAdminHTTPServer`）：独立 `*http.ServeMux`（不再污染全局 `DefaultServeMux`），挂载 `/debug/pprof`、`/metrics`、`/ping`，退出时通过 `Shutdown` 优雅关闭。

核心能力：

- flag 是**不可变值规格**（`FlagSpec`），解析结果快照进 `Flags`，不再可变地写回调用方定义的 flag 结构体
- 解析错误 / `-h` / 非法值**以 error 返回**，绝不 `os.Exit`，可测试
- 清理按 LIFO 调用；初始化失败也触发已收集的清理
- admin server 用独立 mux + `Server.Shutdown`，带超时优雅关闭
- `Run` 返回退出码（`ExitOK` / `ExitSetup` / `ExitRunError` / `ExitSignal`），不调用 `os.Exit`
- `errC` 缓冲，避免 v2 无缓冲 channel 在 Run 提前返回/panic 时挂起 `WaitSignal`

## 快速开始

```go
package main

import (
	"os"

	"github.com/tenz-io/gokit/app/v3"
	"github.com/tenz-io/gokit/logger/v3"
	"github.com/tenz-io/gokit/tracer/v3"
)

type AppConfig struct {
	Name     string `yaml:"name" default:"sample"`
	PageSize int    `yaml:"page_size" default:"10" validate:"required,gt=0,lte=100"`
}

func main() {
	cfg := app.Config{
		Name: "my-service",
		Conf: &AppConfig{},
		Inits: []app.InitFunc{
			app.WithYamlConfig(),
			app.WithLogger(true), // 开启流量日志
			app.WithAdminHTTPServer(),
		},
		Run: func(c *app.Context, conf any, errC chan<- error) {
			ctx, reqID := tracer.EnsureRequestID(c.Context)
			log := logger.WithRequestID(reqID)
			logger.FromContext(logger.WithLogger(ctx, log)).Infow("service started")
			<-c.Done() // 阻塞直到收到信号
		},
	}
	os.Exit(int(app.Run(cfg, nil)))
}
```

## 环境变量占位符

配置文件里写 `${VAR}` 占位符，实际值来自进程环境变量（`.env` 由 `WithDotEnvConfig` 注入，或真实环境提供）。这样敏感配置（密码、DSN、API key）不落盘到配置文件。

**语法**（shell / docker-compose 兼容）：

| 写法 | 含义 |
| --- | --- |
| `${VAR}` | 替换为环境变量 VAR 的值 |
| `${VAR:-default}` | VAR 未设**或为空**时用 `default`，否则用 VAR 的值 |
| `${VAR:?msg}` | VAR 未设**或为空**时报错（错误信息为 `msg`，可省略） |

**语义**：默认**严格**——配置里写了 `${VAR}` 而 VAR 未设且无 `:-` 兜底 / `:?` 报错时，`ReadConfig` 返回 error，启动失败。这样"漏配敏感配置"会在启动期就被发现，而不是把字面量 `${VAR}` 写进数据库连接串。不含 `${}` 的配置完全不受影响（快速路径，零开销、零行为变化）。

```yaml
# config/app.yaml
db_password: ${DB_PASSWORD:-dev-secret}   # 本地回退 dev-secret，生产由 .env 注入
api_key: ${API_KEY}                       # 必填，未设则启动失败
```

```go
// main.go：WithDotEnvConfig 必须排在 WithYamlConfig 之前，这样 .env 注入的环境
// 变量在展开占位符时已可见。
Inits: []app.InitFunc{
    app.WithDotEnvConfig(),   // 1. 注入 .env
    app.WithYamlConfig(),     // 2. 读配置 + 展开 ${...} + 反序列化 + 校验
    app.WithLogger(true),
    app.WithAdminHTTPServer(),
},
```

说明：
- 只支持 `${}` 花括号形式；裸 `$VAR` 会被原样保留（避免歧义）。
- `default` 内部若含 `${...}` 会递归展开（同一环境变量集），带环检测与深度上限，防止 `${A:-${A}}` 无限递归。
- 替换在**原始字节层、反序列化之前**进行，所以 `${PORT}` 填进 `int` 字段时由现有 unmarshal 负责类型转换，无需特例。

## 能力清单

| 能力 | 含义 |
| --- | --- |
| 统一应用编排 | `app.Run` 一次完成 flag 解析、`Inits` 顺序执行、`Run` 启动、等待退出与 LIFO 清理，返回退出码 |
| 有序初始化与 LIFO 清理 | `Inits` 顺序执行，每个可返回 `CleanFunc`；退出或中途失败时**逆序**调用已收集的清理，不泄漏资源 |
| 配置加载 | `WithYamlConfig`/`WithJsonConfig` 读文件 + 应用默认值 + **展开占位符** + 反序列化 + 校验；`WithDotEnvConfig` 注入环境变量 |
| 环境变量占位符 | 配置文件写 `${VAR}`/`${VAR:-x}`/`${VAR:?m}`，实际值来自 `.env`/环境；默认严格，未设兜底则启动失败 |
| 全局日志初始化 | `Run` 默认按 flag 初始化全局日志；`WithLogger(true)` 进一步开启流量日志 |
| Admin HTTP 端点 | `WithAdminHTTPServer` 用独立 mux 挂载 `/debug/pprof`、`/metrics`、`/ping`，退出时优雅 `Shutdown` |
| 健康检查探活 | `PingHandler`/`AddPingHandler` 返回主机名、启动时间与运行时长 |
| 可测试的 flag | `FlagSpec` 不可变值、`ParseFlags` 返回 error 不 `os.Exit`，支持 env 兜底默认值 |

## API 速查

| 符号 | 说明 |
| --- | --- |
| `Run(cfg Config, flags []FlagSpec, argv ...[]string) ExitCode` | 应用入口：解析 flag、执行初始化、运行主逻辑、等待退出，返回退出码（不 `os.Exit`） |
| `Config` | 应用配置：`Name`/`Usage`/`Conf`/`Inits`/`Run` |
| `InitFunc` | 初始化函数 `func(c *Context, conf any) (CleanFunc, error)` |
| `RunFunc` | 主运行函数 `func(c *Context, conf any, errC chan<- error)` |
| `CleanFunc` | 清理函数 `func(context.Context) error`，按 LIFO 调用 |
| `ExitCode` / `ExitOK`/`ExitSetup`/`ExitRunError`/`ExitSignal` | 退出码类型及常量 |
| `Context` | 内嵌 `context.Context`（应用上下文，退出时取消）+ 解析后的 `Flags` |
| `NewContext(ctx, flags) *Context` | 构造 `Context`（主要用于测试） |
| `(*Context).Flags() *Flags` | 获取命令行 flag 集合 |
| `WithYamlConfig() InitFunc` | 读取 YAML 配置文件并反序列化到 `Conf` |
| `WithJsonConfig() InitFunc` | 读取 JSON 配置文件并反序列化到 `Conf` |
| `WithDotEnvConfig(filenames ...string) InitFunc` | 加载 `.env` 文件到环境变量 |
| `WithLogger(trafficEnabled bool) InitFunc` | 按 flag 配置全局日志，可开启流量日志 |
| `WithAdminHTTPServer() InitFunc` | 启动 admin HTTP 服务（pprof/metrics/ping）并返回优雅关闭清理 |
| `ReadConfig(path string, conf any, unmarshal UnmarshalFunc) error` | 通用配置文件读取+默认值+**占位符展开**+校验 |
| `Expand(bs []byte, lookup func(string)(string,bool)) ([]byte, error)` | 在字节层展开 `${VAR}`/`${VAR:-x}`/`${VAR:?m}` 占位符；`ReadConfig` 内部已调用，亦可单独使用 |
| `AddPrometheusHandler(mux *http.ServeMux)` | 挂载 `/metrics` 处理器 |
| `AddPingHandler(mux *http.ServeMux)` | 挂载 `/ping` 健康检查处理器 |
| `AddProfilingHandler(mux *http.ServeMux)` | 挂载 `/debug/pprof/*` 系列处理器 |
| `PingHandler(w http.ResponseWriter, r *http.Request)` | `/ping` 处理逻辑，返回主机名和运行时长 |
| `FlagSpec` | 不可变的命令行 flag 规格（`Name`/`Kind`/`Default`/`Usage`/`Env`） |
| `FlagKind` / `FlagKindString`/`FlagKindInt`/`FlagKindBool`/`FlagKindDuration` | flag 值类型 |
| `StringFlag`/`IntFlag`/`BoolFlag`/`DurationFlag` | flag 规格便捷构造器 |
| `DefaultFlags` | 内置 flag 集（`config`/`port`/`admin-port`/`log`/`logging-file`/`logging-console`/`verbose`） |
| `ParseFlags(name string, specs []FlagSpec, args []string) (*Flags, error)` | 解析 flag 成不可变快照，返回 error 不 `os.Exit` |
| `(*Flags).String/Int/Bool/Duration(name)` | 按名称、类型读取 flag 值（未注册返回零值） |
| `(*Flags).IsSet(name) bool` | 判断 flag 是否注册 |
| `(*Flags).Print(w io.Writer)` | 打印当前 flag 值 |
| `FlagNameConfig` / `FlagNamePort` / `FlagNameAdminPort` / `FlagNameLog` / `FlagNameLoggingFile` / `FlagNameLoggingConsole` / `FlagNameVerbose` | 内置 flag 名称常量 |
| `PrettyString(v any) string` | 将任意值格式化为 JSON 字符串 |

引入路径：`import "github.com/tenz-io/gokit/app/v3"`

## 与 v2 的行为差异

v3 不保证与 v2 兼容，以下是显式的行为差异：

| 差异点 | v2 | v3 |
| --- | --- | --- |
| 退出方式 | `Run` 内部 `log.Fatalf`/`os.Exit`，无法测试、跳过清理 | `Run` 返回 `ExitCode`，不 `os.Exit`；调用方决定 |
| flag 定义 | `*StringFlag` 等指针结构，解析时 `&f.Value` 直接写回调用方字段，存在别名与重复运行问题 | `FlagSpec` 不可变值规格；解析结果快照进 `Flags`，不写回调用方 |
| 解析失败 | `flag.ExitOnError` → `os.Exit` | `flag.ContinueOnError` → 返回 error（`-h` 返回 `flag.ErrHelp` 包装） |
| 清理顺序 | `cleanFns` 顺序追加、顺序调用；初始化失败直接 `return` 不跑已收集清理 | LIFO 调用；初始化失败也逆序跑已收集的清理 |
| admin server | 全局 `http.DefaultServeMux` + `http.ListenAndServe`，无优雅关闭 | 独立 `*http.ServeMux` + `Server.Shutdown` 带超时 |
| 信号处理 | `WaitSignal` 监听后 `os.Exit`，路径间逻辑重复 | `wait` 返回退出码，信号统一处理 |
| `errC` | 无缓冲，Run 提前返回/panic 会让 `WaitSignal` 永久阻塞 | 缓冲大小 1，避免挂起 |
| 全局可变状态 | 包级 `start = time.Now()`、`http.DefaultServeMux` | 包级 `startTime` 只读；不用 `DefaultServeMux` |
| 配置占位符 | 无（敏感配置需手写代码从 env 读回填） | `${VAR}`/`${VAR:-x}`/`${VAR:?m}` 在反序列化前自动展开，默认严格 |
| 依赖 | `logger/v2` | `logger/v3`（context 串联、流量日志、运行时调级） |
