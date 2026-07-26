# monitor

面向 **single-flight 注入式**的 Prometheus 指标采集库：计数器、直方图、仪表盘、摘要，支持 context 传递。

V3 相对 V2 的核心变化：

- **single-flight 注入式**：请求入口 `Init(ctx, cmd)` 创建/复用一个 `cmd` 作用域的导出器并注入 context，链路下游的 `Begin` 复用同一份导出器，请求结束时聚批写一次。
- **Registry 可注入**：`Configure(opts...)` 在启动期注入自定义 Registry / 命名空间；未调用则兜底 `prometheus.DefaultRegisterer`；测试可用独立 Registry 隔离。彻底消除 v2 全局 `init()` 重复注册 panic。晚于首次使用或重复调用时 `Configure` 返回错误而非静默改写 Registry。
- **End 同步写**：去掉 v2 每点一个 goroutine 的异步写法；`Begin`/`End` 同步维护活跃数 gauge 的 `Inc`/`Dec` 配对，保证有序、不丢点。

```go
import "github.com/tenz-io/gokit/monitor/v3"
```

## 模块介绍

monitor 解决两类问题：

- **指标采集**：以 `cmd`（入口命令）+ `dsCmd`（下游/操作）+ `code`（结果码）+ `opt`（可选子维度）四元标签，覆盖仪表盘瞬时值、计数器累计、延迟直方图、数据量摘要四类指标。
- **single-flight 注入**：一次请求只创建一次 `Exporter`，随 context 在调用链传递，避免下游每个埋点重复 `Init`；`FromContext` 未初始化时返回 no-op 实现，调用方无需判空。

## 快速开始

```go
package main

import (
	"context"

	"github.com/tenz-io/gokit/monitor/v3"
)

func main() {
	// 启动期配置一次（未调用则兜底默认 Registry）
	monitor.Configure()

	ctx := context.Background()
	// 单飞：创建 userService 作用域的导出器并注入 ctx（已存在则复用）
	ctx = monitor.Init(ctx, "userService")

	// 主记录器
	rec := monitor.Begin(ctx, "total")
	err := handleGetUser(ctx)
	rec.EndWithError(err)
}

func handleGetUser(ctx context.Context) error {
	// 子记录器：复用上游注入的单飞 Exporter
	rec := monitor.Begin(ctx, "getUser")
	defer rec.EndWithError(nil)
	return nil
}
```

## 标签模型

所有指标共享统一标签集，基数策略一致：

| 标签 | 含义 | 归一化 |
| --- | --- | --- |
| `cmd` | 入口命令（Exporter 作用域） | 空 → `NA` |
| `dsCmd` | 下游/操作名 | 调用方传入，如 `getUser`、`db_query` |
| `code` | 结果码 | 空 → `0`(ok)；非 `0` → `1`(err) |
| `opt` | 可选子维度（如 `hit`/`miss`/`error`） | 空 → `NA` |

`code` 的归一化保证 `Observe`/`Sample` 的基数受控：非零一律映射为 `1`，避免业务自定义 code 爆基数。

## 能力清单

| 能力 | 含义 |
| --- | --- |
| 仪表盘打点（`Set`/`Incr`/`Decr`） | 记录瞬时值或当前活跃数，如并发请求数、队列长度 |
| 计数器打点（`Count`/`CountDelta`） | 累计次数或按增量累加，如请求总数、成功/失败次数 |
| 延迟直方图观测（`Observe`） | 以毫秒为单位记录耗时分布，内置从 0.1ms 到 10s 的分桶 |
| 数据量摘要采样（`Sample`） | 以 p50/p90/p95/p99 记录数据大小等数值分布 |
| 一次调用计时（`Begin`/`Recorder`） | `Begin` 起计时 + active +1，`End` 同步写耗时直方图、计数器、active -1 |
| 按结果结束（`EndWithError`/`EndWithCode`/`EndWithOpt` 等） | 按 err 是否 nil、自定义 code、opt 多种方式结束 |
| 单飞注入与复用（`Init`/`WithExporter`/`FromContext`） | 请求入口注入 Exporter，下游随 ctx 复用；未初始化降级为 no-op |
| 跨 Context 拷贝（`CopyToContext`） | 跨 goroutine/跨请求场景把已有 Exporter 从源 ctx 复制到目标 ctx |

## API 速查

| 符号 | 说明 |
| --- | --- |
| `Configure(opts ...Option) error` | 启动期配置（Registry/命名空间），须在首次 Exporter 构造前调用一次；重复调用或晚于首次使用返回 `ErrAlreadyConfigured`/`ErrAlreadyInUse`，不会静默改写 Registry |
| `WithNamespace(string)` `WithSubsystem(string)` `WithRegistry(prometheus.Registerer)` | 配置项 |
| `Exporter`（接口） | cmd 作用域指标导出器；`FromContext` 未注入时返回 no-op 实现，不为 nil |
| `NewExporter(cmd string) Exporter` | 创建以 cmd 为标签的导出器，cmd 为空 → `NA` |
| `Exporter` 的 `Set/Incr/Decr/Count/CountDelta/Observe/Sample` | 同步打点；`CountDelta` 入参 `uint64`（Counter 只增不减，负值会在 Prometheus 内 panic，故类型层杜绝） |
| `Recorder` | 单次调用计时器 |
| `Begin(ctx, dsCmd string) *Recorder` | 从 ctx 取 Exporter 并开始记录（active +1），并把 ctx 的值（trace/tenant）取消关联地保存到 Recorder 供 End 使用 |
| `(*Recorder).End()` `EndWithCode(code)` `EndWithOpt(opt)` `EndWithError(err)` `EndWithErrorOpt(err, opt)` `EndWithCodeOpt(code, opt)` | 结束记录（同步写，幂等） |
| `Init(ctx, cmd string) context.Context` | 单飞：在 ctx 中创建/复用 Exporter |
| `WithExporter(ctx, Exporter) context.Context` | 显式注入 |
| `FromContext(ctx) Exporter` | 取出，不存在返回 no-op（不为 nil） |
| `HasExporter(ctx) bool` | 判断是否已注入 |
| `CopyToContext(srcCtx, dstCtx) context.Context` | 跨 ctx 拷贝 |

引入路径：`github.com/tenz-io/gokit/monitor/v3`
