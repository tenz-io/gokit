# gokit

Go 1.24+ 的 monorepo，提供应用启动引导、可观测性、通信、数据结构与并发韧性等工具模块，用于构建生产级 Go 服务。

## 模块概览

当前仓库共 13 个模块目录，全部为 `v3` 主版本。

### 基础设施

| 模块                       | 用途                                                                |
|----------------------------|---------------------------------------------------------------------|
| [functional](functional/)  | 泛型函数式编程工具：`Map`、`Filter`、`Reduce`、`GroupBy`、`TopK` 等  |
| [collection](collection/)  | 泛型数据结构：`Stack`、`Queue`、`PriorityQueue`、`Set`               |
| [tracer](tracer/)          | 基于 context 的请求 ID 传递及 debug/stress/shadow 标志位管理         |
| [annotation](annotation/)  | struct-tag 驱动：声明式默认值、可插拔校验器、缓存字段 plan             |

### 可观测性

| 模块                | 用途                                                                  |
|---------------------|-----------------------------------------------------------------------|
| [logger](logger/)   | 封装 `go.uber.org/zap` 的结构化日志，支持日志轮转、限流和输出裁剪      |
| [monitor](monitor/) | 面向 single-flight 模式的 Prometheus 指标：直方图、计数器、仪表盘、摘要 |

### 应用启动引导

| 模块        | 用途                                                                                                |
|-------------|-------------------------------------------------------------------------------------------------------|
| [app](app/) | 应用生命周期框架：配置加载（YAML/JSON/dotenv）、admin HTTP 端点（pprof、Prometheus、健康检查）、优雅退出 |

### 通信

| 模块                 | 用途                                                                              |
|----------------------|-------------------------------------------------------------------------------------|
| [ginext](ginext/)    | Gin 扩展：请求绑定 + 校验、JWT 鉴权、RPC 拦截器、结构化响应                          |
| [httpext](httpext/)  | HTTP 客户端，提供可组合的传输层拦截器链：请求头注入、流量采集、指标和慢请求日志      |

### 数据与持久化

| 模块                | 用途                                                                                    |
|---------------------|-------------------------------------------------------------------------------------------|
| [gormext](gormext/) | GORM 插件：通过 GORM 回调注入追踪、流量日志、Prometheus 指标、错误日志和慢查询检测         |
| [cache](cache/)     | 缓存抽象（`Manager` 接口），支持可插拔后端：内存 map、泛型 LRU（带 TTL）、Redis（含可观测性钩子） |

### 并发与韧性

| 模块                     | 用途                                                                                       |
|--------------------------|----------------------------------------------------------------------------------------------|
| [async](async/)         | 泛型并发任务执行器：`AllOf`（收集全部结果）、`AnyOf`（首个成功即返回）、panic 安全、errgroup 模式 |
| [retriever](retriever/) | 可配置的重试库，支持多种退避策略（指数 + 抖动、线性、无退避）、最大尝试次数和 context 超时       |

## 依赖关系

```
        annotation          functional   collection
              |                       (无内部依赖)
            app

        ginext ──┬── annotation
                  ├── logger
                  ├── monitor
                  └── tracer

  httpext / gormext / cache
        ├── logger
        ├── monitor
        └── tracer

  async / retriever / logger / monitor / tracer  (无内部依赖)
```

**内部依赖明细**（直接依赖，数据来自各模块源码内 `import` 语句的实际扫描）：

| 模块          | 内部依赖                                    |
|---------------|----------------------------------------------|
| `ginext`      | `annotation`、`logger`、`monitor`、`tracer` |
| `app`         | `annotation`、`logger`                        |
| `cache`       | `logger`、`monitor`、`tracer`                  |
| `gormext`     | `logger`、`monitor`、`tracer`                  |
| `httpext`     | `logger`、`monitor`、`tracer`                  |
| `annotation`  | _(无)_                                        |
| `functional`  | _(无)_                                        |
| `collection`  | _(无)_                                        |
| `tracer`      | _(无)_                                        |
| `logger`      | _(无)_                                        |
| `monitor`     | _(无)_                                        |
| `async`       | _(无)_                                        |
| `retriever`   | _(无)_                                        |

> 所有模块均为 `v3` 主版本。

## 分层架构

1. **基础设施层** — `functional`、`collection`、`tracer`、`annotation`，无任何内部依赖，供其他模块自由组合。
2. **可观测性层** — `logger`、`monitor`、`tracer` 构成可观测性三元组，被绝大多数中间层模块依赖，用于统一日志、指标和链路追踪。
3. **中间层** — `app`、`ginext`、`httpext`、`gormext`、`cache`，组合基础设施与可观测性能力，直接面向业务服务的启动、通信与数据访问场景。

`async`、`retriever` 作为通用并发/韧性工具，不依赖以上任何内部模块，可在各层中独立引用。

## 关键外部依赖

- `go.uber.org/zap` — 结构化日志（`logger`）
- `github.com/gin-gonic/gin` — HTTP 框架（`ginext`）
- `gorm.io/gorm` — ORM（`gormext`）
- `github.com/go-redis/redis/v8` — Redis 客户端（`cache`）
- `github.com/prometheus/client_golang` — 指标采集（`monitor`、`app`）

## 开发

### 环境要求

- `wire` — DI 代码生成。安装：https://github.com/google/wire
- `go-enum` — 枚举代码生成。安装：https://github.com/abice/go-enum
- `gci` — Go import 格式化。安装：https://github.com/daixiang0/gci
- `mockery` — Mock 代码生成。安装：https://github.com/vektra/mockery

### 工作区

本仓库使用 Go workspace（`go.work`）管理，包含 27 个工作区成员（13 个模块主目录及其 example 子模块）。所有子模块在工作区中统一声明，可同时进行开发；每个模块维护独立的 `go.mod`，独立进行版本管理。

### 发版

发版流程详见 [release.md](release.md)。所有模块统一为 `v3` 主版本轨道，由 `scripts/tag-all.sh`（一键批量打 tag + GitHub Release，全程经 `gh` API）和 `scripts/version-check.sh`（版本一致性检查）两个脚本管理。一键发布：`gh auth login` 后执行 `./scripts/tag-all.sh v3.0.1 --release`（或 `make release VERSION=v3.0.1`），发版前请先运行版本检查脚本确认无误。
