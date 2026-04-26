# gokit

Go 通用工具包 — 一个 Go 1.21+ 的 monorepo，提供应用启动引导、可观测性、通信、数据结构和代码生成等模块，用于构建生产级服务。

## 模块概览

### 基础设施

| 模块                                | 用途                                                                                  |
|-------------------------------------|---------------------------------------------------------------------------------------|
| [functional](functional/)           | 泛型函数式编程工具：`Map`、`Filter`、`Reduce`、`GroupBy`、`TopK` 等                    |
| [collection](collection/)           | 泛型数据结构：`Stack`、`Queue`、`PriorityQueue`、`Set`                                 |
| [tracer](tracer/)                   | 基于 context 的请求 ID 传递及 debug/stress/shadow 标志位管理                           |
| [bloom](bloom/)                     | 概率型布隆过滤器，基于 Murmur3 哈希，可调节预期元素数量和误判率                          |
| [annotation](annotation/)           | 基于 struct tag 的请求字段绑定（`bind`/`json`/`protobuf`）、默认值注入和校验规则          |

### 可观测性

| 模块                  | 用途                                                                                |
|-----------------------|-------------------------------------------------------------------------------------|
| [logger](logger/)     | 封装 `go.uber.org/zap` 的结构化日志，支持日志轮转、限流和输出裁剪                       |
| [monitor](monitor/)   | 面向 single-flight 模式的 Prometheus 指标：直方图、计数器、仪表盘、摘要                |

### 应用启动引导

| 模块            | 用途                                                                                                                       |
|-----------------|----------------------------------------------------------------------------------------------------------------------------|
| [app](app/)     | 应用生命周期框架：配置加载（YAML/JSON/dotenv）、admin HTTP 端点（pprof、Prometheus、健康检查）、优雅退出                        |
| [cmd](cmd/)     | 基于 `urfave/cli/v2` 的 CLI 启动框架 — 与 `app` 相同的基础能力，以 CLI 命令和标志位形式暴露                                   |

### 通信

| 模块                                | 用途                                                                                     |
|-------------------------------------|------------------------------------------------------------------------------------------|
| [ginext](ginext/)                   | Gin 扩展：请求绑定 + 校验、JWT 鉴权、RPC 拦截器、结构化响应                                 |
| [grpcext](grpcext/)                 | gRPC 一元/流式拦截器，提供服务端和客户端的请求追踪、流量日志和指标采集                         |
| [grpcetcd](grpcetcd/)               | 基于 etcd 的服务注册（租约 + 心跳保活）和客户端服务发现解析器（轮询负载均衡）                 |
| [httpext](httpext/)                 | HTTP 客户端，提供可组合的传输层拦截器链：请求头注入、流量采集、指标和慢请求日志               |
| [genproto](genproto/)               | 共享 protobuf 类型定义：`Auth`、`RequestHeader`、`ResponseHeader`                          |

### 数据与持久化

| 模块                  | 用途                                                                                              |
|-----------------------|---------------------------------------------------------------------------------------------------|
| [gormext](gormext/)   | GORM 插件：通过 GORM 回调注入追踪、流量日志、Prometheus 指标、错误日志和慢查询检测                     |
| [cache](cache/)       | 缓存抽象（`Manager` 接口），支持可插拔后端：内存 map、泛型 LRU（带 TTL）、Redis（含可观测性钩子）      |

### 并发与韧性

| 模块                        | 用途                                                                                |
|-----------------------------|-------------------------------------------------------------------------------------|
| [async](async/)             | 泛型并发任务执行器：`AllOf`（收集全部结果）、`AnyOf`（首个成功即返回）、panic 安全、errgroup 模式 |
| [retriever](retriever/)     | 可配置的重试库，支持多种退避策略（指数 + 抖动、线性、无退避）、最大尝试次数和 context 超时  |

### 代码生成

| 模块                                      | 用途                                                                                         |
|-------------------------------------------|----------------------------------------------------------------------------------------------|
| [protoc-gen-go-gin](protoc-gen-go-gin/)   | `protoc` 插件，根据带 `google.api.http` 注解的 protobuf 服务定义自动生成 Gin HTTP handler 代码   |

### 第三方集成

| 模块                    | 用途                                                              |
|-------------------------|-------------------------------------------------------------------|
| [notionx](notionx/)     | 将 Markdown 转换为 Notion API block 结构，用于程序化创建 Notion 页面 |

## 依赖关系图

```
                        protoc-gen-go-gin
                           /         \
                       ginext       genproto
                      /  |   \
        annotation  functional  logger -- monitor -- tracer
           |                        |
          app        ┌─────────────┼─────────────┐
          cmd   grpcetcd         grpcext       cache
         (CLI)  (etcd gRPC)   (gRPC 指标)    (Redis/LRU)
                                 httpext
                              (HTTP 客户端)
                                 gormext
                               (GORM 插件)
```

**内部依赖明细**（直接依赖）：

| 模块                 | 内部依赖                                          |
|----------------------|--------------------------------------------------|
| `app`                | `annotation`、`logger`                            |
| `cmd`                | `annotation`、`logger`                            |
| `ginext`             | `annotation`、`functional`、`logger`、`monitor`、`tracer` |
| `grpcext`            | `logger`、`monitor`、`tracer`                      |
| `grpcetcd`           | `logger`                                           |
| `httpext`            | `logger`、`monitor`、`tracer`                      |
| `gormext`            | `logger`、`monitor`、`tracer`                      |
| `cache`              | `logger`、`monitor`、`tracer`                      |
| `protoc-gen-go-gin`  | `ginext`、`genproto`                               |
| `functional`         | _(无)_                                            |
| `collection`         | _(无)_                                            |
| `tracer`             | _(无)_                                            |
| `bloom`              | _(无)_                                            |
| `annotation`         | _(无)_                                            |
| `async`              | _(无)_                                            |
| `retriever`          | _(无)_                                            |
| `notionx`            | _(无)_                                            |

**分层架构：**

1. **基础设施层** — `functional`、`collection`、`tracer`、`bloom`、`annotation`，无任何内部依赖。
2. **可观测性层** — `logger`、`monitor`、`tracer` 构成可观测性三元组，被大多数中层模块依赖。
3. **中间层** — `app`、`cmd`、`ginext`、`grpcext`、`grpcetcd`、`httpext`、`gormext`、`cache`，组合基础设施与可观测性能力。
4. **顶层** — `protoc-gen-go-gin` 位于最高层，依赖 `ginext` 和 `genproto`。

**关键外部依赖：**

- `go.uber.org/zap` — 结构化日志（`logger`）
- `github.com/gin-gonic/gin` — HTTP 框架（`ginext`）
- `google.golang.org/grpc` — gRPC（`grpcext`、`grpcetcd`）
- `go.etcd.io/etcd/client/v3` — 服务发现（`grpcetcd`）
- `gorm.io/gorm` — ORM（`gormext`）
- `github.com/go-redis/redis/v8` — Redis 客户端（`cache`）
- `github.com/prometheus/client_golang` — 指标采集（`monitor`、`app`、`cmd`）
- `github.com/urfave/cli/v2` — CLI 框架（`cmd`）
- `google.golang.org/protobuf` — protobuf 运行时（`genproto`、`grpcext`、`protoc-gen-go-gin`）

## 开发

### 环境要求

- `wire` — DI 代码生成。安装：https://github.com/google/wire
- `go-enum` — 枚举代码生成。安装：https://github.com/abice/go-enum
- `gci` — Go import 格式化。安装：https://github.com/daixiang0/gci
- `mockery` — Mock 代码生成。安装：https://github.com/vektra/mockery

### 工作区

本仓库使用 Go workspace（`go.work`）管理。所有子模块均在工作区中声明，可同时进行开发。每个模块维护独立的 `go.mod`，独立进行版本管理。
