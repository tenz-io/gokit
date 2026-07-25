# monitor

面向 single-flight 模式的 Prometheus 指标：直方图、计数器、仪表盘、摘要，支持 context 传递。

## 功能特性

- `SingleFlight` 接口统一暴露 Set/Incr/Decr（仪表盘）、Count/CountDelta（计数器）、Observe（直方图，用于延迟）、Sample（摘要，用于数据量）等指标操作
- `NewSingleFlight` 创建以 `cmd` 命名空间隔离的指标导出器，内部指标全部预先注册到 Prometheus 默认 Registry
- `BeginRecord`/`Recorder` 提供一次调用的计时器，`End`/`EndWithCode`/`EndWithOpt`/`EndWithError`/`EndWithErrorOpt`/`EndWithCodeOpt` 覆盖成功、错误、自定义 code/opt 等多种结束方式，自动记录耗时直方图与计数器，并维护活跃请求数（gauge）
- `InitSingleFlight` 在 context 中初始化（若已存在则复用）single flight 监控器，`WithMonitor`/`FromContext`/`HasSingleFlight` 用于显式注入与读取
- `CopyToContext` 在跨 goroutine/跨请求场景下，把已存在的监控器从源 context 拷贝到目标 context，避免丢失指标上下文
- 未初始化时 `FromContext` 返回空实现（no-op），调用方无需做 nil 判断即可安全埋点

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
