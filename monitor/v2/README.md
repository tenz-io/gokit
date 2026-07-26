# monitor

面向 single-flight 模式的 Prometheus 指标：直方图、计数器、仪表盘、摘要，支持 context 传递。

## 功能特性

- `SingleFlight` 接口统一暴露 Set/Incr/Decr（仪表盘）、Count/CountDelta（计数器）、Observe（直方图，用于延迟）、Sample（摘要，用于数据量）等指标操作
- `NewSingleFlight` 创建以 `cmd` 命名空间隔离的指标导出器，内部指标全部预先注册到 Prometheus 默认 Registry
- `BeginRecord`/`Recorder` 提供一次调用的计时器，`End`/`EndWithCode`/`EndWithOpt`/`EndWithError`/`EndWithErrorOpt`/`EndWithCodeOpt` 覆盖成功、错误、自定义 code/opt 等多种结束方式，自动记录耗时直方图与计数器，并维护活跃请求数（gauge）
- `InitSingleFlight` 在 context 中初始化（若已存在则复用）single flight 监控器，`WithMonitor`/`FromContext`/`HasSingleFlight` 用于显式注入与读取
- `CopyToContext` 在跨 goroutine/跨请求场景下，把已存在的监控器从源 context 拷贝到目标 context，避免丢失指标上下文
- 未初始化时 `FromContext` 返回空实现（no-op），调用方无需做 nil 判断即可安全埋点

## 能力清单

| 能力 | 含义 |
| --- | --- |
| 仪表盘打点（Set/Incr/Decr） | 记录瞬时值或当前活跃数，如并发请求数、队列长度等实时状态 |
| 计数器打点（Count/CountDelta） | 累计次数或按增量累加，如请求总数、成功/失败次数统计 |
| 延迟直方图观测（Observe） | 以毫秒为单位记录耗时分布，用于统计接口/下游调用的延迟分布与分桶（内置从 0.1ms 到 10s 的分桶区间） |
| 数据量摘要采样（Sample） | 以分位数（p50/p90/p95/p99）记录数据大小等数值分布，用于观察长尾情况 |
| 一次调用计时（BeginRecord/Recorder） | 开始一次调用记录起始时间，通过 End 系列方法结束时自动异步写入耗时直方图、计数器，并维护活跃请求数 gauge |
| 按结果结束记录（EndWithError/EndWithCode/EndWithOpt 等） | 根据 error 是否为 nil、自定义 code、opt 标签等多种方式结束记录，无需手动区分成功/失败分支 |
| 指标导出器命名隔离（NewSingleFlight） | 以 cmd 作为标签维度创建独立的指标导出器，区分不同业务/服务的同类指标 |
| Context 传递与复用（InitSingleFlight/WithMonitor/FromContext） | 将监控器绑定到 context 中随调用链传递，未初始化时自动降级为空实现，避免埋点代码判空 |
| 跨 Context 拷贝（CopyToContext） | 在跨 goroutine 或跨请求场景下把已有监控器从源 context 复制到目标 context，避免指标上下文丢失 |

## 快速开始

```go
import "github.com/tenz-io/gokit/monitor/v2"
```

```go
package main

import (
	"context"

	"github.com/tenz-io/gokit/monitor/v2"
)

func handleGetUser(ctx context.Context) error {
	// 在入口处初始化 single flight 监控器（已存在则直接复用）
	ctx = monitor.InitSingleFlight(ctx, "userService")

	// 开始一次记录，dsCmd 用于标识具体下游/操作名
	rec := monitor.BeginRecord(ctx, "getUser")

	err := doGetUser(ctx)

	// 根据 error 自动映射 code：nil -> "0"，非 nil -> "1"
	rec.EndWithError(err)
	return err
}

func doGetUser(ctx context.Context) error {
	// 手动打点：仪表盘 +1
	monitor.FromContext(ctx).Incr(ctx, "getUser", "0", "cache")
	return nil
}
```

## API 速查

| 符号 | 说明 |
| --- | --- |
| `SingleFlight` | 指标操作接口：Set/Incr/Decr/Count/CountDelta/Observe/Sample/BeginRecord |
| `NewSingleFlight(cmd string) SingleFlight` | 创建以 cmd 为标签的指标导出器实现 |
| `Recorder` | 一次调用的计时记录器，内部记录起始时间 |
| `(*Recorder) End()` | 以默认成功 code（"0"）结束记录 |
| `(*Recorder) EndWithCode(code string)` | 以指定 code 结束记录 |
| `(*Recorder) EndWithOpt(opt string)` | 以默认成功 code + 指定 opt 结束记录 |
| `(*Recorder) EndWithError(err error)` | 按 err 是否为 nil 自动映射 code 结束记录 |
| `(*Recorder) EndWithErrorOpt(err error, opt string)` | 同上，附带 opt 标签 |
| `(*Recorder) EndWithCodeOpt(code, opt string)` | 以指定 code 和 opt 结束记录，异步写入耗时/计数/活跃数 |
| `BeginRecord(ctx, dsCmd string) *Recorder` | 从 ctx 中取出监控器并开始一次记录 |
| `InitSingleFlight(ctx, cmd string) context.Context` | 在 ctx 中初始化（或复用已有）single flight 监控器 |
| `WithMonitor(ctx, sf SingleFlight) context.Context` | 显式将监控器注入 ctx |
| `FromContext(ctx) SingleFlight` | 从 ctx 取出监控器，不存在时返回空实现（不为 nil） |
| `HasSingleFlight(ctx) bool` | 判断 ctx 中是否已注入监控器 |
| `CopyToContext(srcCtx, dstCtx) context.Context` | 将监控器从源 ctx 拷贝到目标 ctx |

引入路径：`github.com/tenz-io/gokit/monitor/v2`
