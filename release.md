# 发布流程

本项目含 20 个 Go 子模块，通过 `go.work` 管理。每个子模块独立版本号，发布时需为每个子模块打 Git tag。

## 一键批量发布

### 本地执行

```bash
# 1. 确认所有测试通过
make test

# 2. 预览即将创建的 tag
./scripts/tag-all.sh v2.0.1 --dry-run

# 3. 打 tag
./scripts/tag-all.sh v2.0.1

# 4. 推送所有 tag 到远端
./scripts/tag-all.sh v2.0.1 --push

# 也可以只给部分模块打 tag
./scripts/tag-all.sh v2.0.1 tracer,logger,async --push
```

### GitHub Actions

1. 打开 [Actions → Tag Release](https://github.com/tenz-io/gokit/actions/workflows/tag-release.yml)
2. 点击 **Run workflow**
3. 填写参数：
   - `version`：版本号，如 `v2.0.1`（必填）
   - `modules`：限定模块（逗号分隔），留空 = 全部 20 个非 example 模块
   - `dry_run`：`true` = 仅预览不执行
4. 点击 **Run workflow**

## Tag 命名规则

| go.work 路径 | Module 路径 | Tag 示例 |
|---|---|---|
| `./annotation` | `github.com/tenz-io/gokit/annotation` | `annotation/v2.0.1` |
| `./logger/v2` | `github.com/tenz-io/gokit/logger/v2` | `logger/v2.0.1` |
| `./async` | `github.com/tenz-io/gokit/async` | `async/v2.0.1` |

规则：`/vN` 后缀从目录名中移除后作为 tag 前缀。即 `logger/v2` 打 `logger/v2.0.1`，不是 `logger/v2/v2.0.1`。

## 子模块清单

以下为 `go.work` 中所有非 example 子模块（共 20 个），打 tag 时会全部覆盖：

| # | 子模块 | 功能 |
|---|--------|------|
| 1 | `annotation` | Struct tag 解析、默认值、校验 |
| 2 | `app` | 应用启动引导 |
| 3 | `async` | 并发任务执行 |
| 4 | `bloom` | 布隆过滤器 |
| 5 | `cache` | 缓存抽象（内存/LRU/Redis） |
| 6 | `cmd` | CLI 启动框架 |
| 7 | `collection` | 数据结构（Stack/Queue/Set） |
| 8 | `functional` | 函数式编程工具 |
| 9 | `genproto` | 共享 protobuf 定义 |
| 10 | `ginext` | Gin 框架扩展 |
| 11 | `gormext` | GORM 插件 |
| 12 | `grpcetcd` | etcd 服务注册发现 |
| 13 | `grpcext` | gRPC 拦截器 |
| 14 | `httpext` | HTTP 客户端扩展 |
| 15 | `logger/v2` | 结构化日志 |
| 16 | `monitor` | Prometheus 指标 |
| 17 | `notionx` | Markdown → Notion 转换 |
| 18 | `protoc-gen-go-gin` | Proto → Gin 代码生成 |
| 19 | `retriever` | 重试/退避库 |
| 20 | `tracer` | 请求追踪 & 标志位 |

## 发布后

Go 模块消费者可通过 `go get` 拉取新版本：

```bash
go get github.com/tenz-io/gokit/logger/v2@v2.0.1
go get github.com/tenz-io/gokit/tracer@v2.0.1
```

消费者 `go.mod` 中版本引用示例：

```
require (
    github.com/tenz-io/gokit/logger/v2 v2.0.1
    github.com/tenz-io/gokit/tracer v2.0.1
    github.com/tenz-io/gokit/annotation v2.0.1
)
```

## 注意事项

- **tag 不可变**：Git tag 一旦推送不可修改（如需修正，需打新版本号）
- **Go 模块版本规则**：`v2+` 版本要求 module path 以 `/v2` 结尾（如 `logger/v2`），否则只能用 `v0` 或 `v1`
- **go.work 不参与发布**：`go.work` 仅用于本地开发，外部消费者通过 tag + `go get` 拉取
- **example 模块不打 tag**：`go.work` 中的 `*/example` 目录仅用于本地示例，不对外发布
