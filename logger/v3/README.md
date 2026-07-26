# logger

基于 [go.uber.org/zap](https://github.com/uber-go/zap) 的结构化日志库。v3 是一次无兼容包袱的干净重写，提供分级日志、控制台/文件输出（按级别分流）、context 串联、流量日志与输出裁剪。

```go
import "github.com/tenz-io/gokit/logger/v3"
```

## 模块介绍

logger 解决四类问题：

- **分级输出**（`Debug/Info/Warn/Error`）：四级日志；普通方法是 print 风格，`*f` 是 printf 风格，`*w` 接收结构化 key-value 字段。
- **双通道输出**（`Console` + `FilePath`）：默认控制台输出（Info/Debug→stdout，Warn/Error→stderr）；开启文件后按 `debug.log`/`info.log`/`warn.log`/`error.log` 拆分写入，并基于 lumberjack 自动轮转压缩。
- **Context 串联**（`WithLogger`/`FromContext`/`CopyToContext`）：把携带字段的 `Entry` 挂到 `context.Context`，在调用链、goroutine 之间贯穿 `request_id` 等字段而不丢失。
- **流量日志**（`StartTraffic`/`TrafficRec`）：记录每次请求/响应的 `cmd`、耗时、状态码与响应体，写入独立的 `traffic.log`，与业务日志分离。

核心能力：

- 四级日志 + 运行时动态调级（`SetLevel`/`GetLevel` 真正生效，AtomicLevel 已接入 core）
- 控制台默认开启，文件按级别分流（`error.log` 不再混进 Info 行）
- 输出格式可切（`Encoding`：`console` 默认 / `json`）
- `With`/`WithField`/`WithFields`/`WithError`/`WithRequestID`/`WithTracing` 链式附加字段，返回新 `Entry` 不改原实例
- `OutputTrimmer` 对超长字符串、超深结构体/数组自动截断，防止大字段刷屏
- 独立流量日志，记录请求耗时、返回码、响应体
- context 跨调用链传递日志实例

## 快速开始

```go
package main

import (
	"context"
	"errors"
	"time"

	"github.com/tenz-io/gokit/logger/v3"
)

func main() {
	// 1. 配置全局日志：默认控制台输出，开启文件按级别分流
	logger.Configure(logger.Config{
		Level:    logger.InfoLevel,
		Console:  true,
		FilePath: "log",
		Traffic:  true,
	})

	// 2. 分级输出 + 结构化字段
	logger.Infow("server started", "port", 8080, "env", "production")
	logger.Warnw("low disk space", "disk", "/dev/sda1", "percent", 85)
	logger.Errorw("connection refused", "addr", "db.internal:5432")

	// 3. 链式附加字段，返回新 Entry，不改全局实例
	userLog := logger.With("user_id", "usr_123", "session", "abc")
	userLog.Info("user logged in")

	// 4. error 字段
	logger.WithError(errors.New("timeout")).Warn("retrying request")

	// 5. request_id 贯穿调用链
	reqLog := logger.WithRequestID("req_abc123")
	reqLog.Info("processing request")

	// 6. 流量日志：start/end 记录耗时与响应
	rec := logger.StartTraffic("getUser")
	if rec != nil {
		time.Sleep(10 * time.Millisecond)
		rec.End(map[string]any{"name": "bob"}, "200")
	}

	// 7. context 串联：把带字段的 Entry 挂到 ctx，下游取出
	ctx := logger.WithLogger(context.Background(), reqLog)
	logger.FromContext(ctx).Info("log from context")
}
```

## 能力清单

| 能力 | 含义 |
|------|------|
| 按级别分流到不同文件 | 开启 `FilePath` 后自动按 `debug.log`/`info.log`/`warn.log`/`error.log` 拆分，每个文件只写**自身级别及以上**的日志（`error.log` 只含 Error 行），便于按级别分类排查 |
| 日志文件自动轮转与压缩 | 基于 lumberjack，按 `MaxSize`/`MaxAge`/`MaxBackups` 限制单文件大小、保留天数和备份数，超出后自动压缩旧文件 |
| 运行时动态调级 | `SetLevel`/`GetLevel` 基于 `zap.AtomicLevel` 并已接入 core，可在服务运行中不重启地临时调低/调高全局日志级别 |
| 输出格式可切 | `Encoding`（`console` 默认 / `json`），console 对 grep 友好，json 适合日志聚合（Loki/ELK） |
| 链式附加结构化字段 | `With`/`WithField`/`WithFields`/`WithError`/`WithRequestID`/`WithTracing` 返回新 `Entry`，不改原实例，适合按请求/用户维度派生带上下文字段的日志器 |
| 输出裁剪防止大字段刷屏 | `TrimConfig`（`ArrLimit`/`StrLimit`/`DeepLimit`/`Ignores`）对超长字符串、超深结构体/map/slice 自动截断或跳过指定字段 |
| 独立的流量日志 | `Traffic` 配置开启后，`StartTraffic`+`TrafficRec.End`/`EndWithError` 把每次请求的耗时、返回码、响应体写入单独的 `traffic.log`，与业务日志分离 |
| Context 跨调用链传递日志实例 | `WithLogger`/`FromContext`/`CopyToContext` 把携带字段的 `Entry` 挂到 `context.Context`，在函数调用链、goroutine 之间传递而不丢失已附加字段 |
| 独立实例与全局实例并存 | `NewEntry`/`NewEntryWithOpts` 可创建不影响全局配置的独立日志实例，适合多租户或不同日志策略的子模块 |

## API 速查

| 符号 | 说明 |
|------|------|
| `Configure(cfg Config)` | 用 `Config` 初始化全局日志实例 |
| `ConfigureWithOpts(opts ...ConfigOption)` | 用函数式选项初始化全局日志实例 |
| `NewEntry(cfg Config) Entry` | 创建独立的日志实例，不影响全局 |
| `NewEntryWithOpts(opts ...ConfigOption) Entry` | 用函数式选项创建独立日志实例 |
| `Config` | 日志配置（级别/编码/控制台/文件路径/轮转/流量/裁剪等） |
| `TrimConfig` | 输出裁剪配置（数组长度/字符串长度/嵌套深度/忽略字段） |
| `Level` / `DebugLevel`/`InfoLevel`/`WarnLevel`/`ErrorLevel` | 日志级别类型及常量 |
| `Encoding` / `ConsoleEncoding`/`JSONEncoding` | 输出编码选择 |
| `Fields` | `map[string]any`，用于批量传入结构化字段 |
| `L() Entry` | 获取全局日志实例 |
| `SetLevel(lvl Level)` / `GetLevel() Level` | 运行时设置/读取日志级别（接入 core，真生效） |
| `Entry` | 日志接口：`Debug/Info/Warn/Error`（print 风格）、`*f`（printf 风格）、`*w`（结构化字段）+ `SetLevel`/`GetLevel` |
| `Debug/Info/Warn/Error`、`*f`、`*w` | 包级便捷函数，作用于全局日志实例；例如 `Infow("started", "port", 8080)` |
| `With(args ...any) Entry` | 追加一组 key-value 结构化字段，返回新 `Entry` |
| `WithField(k, v) Entry` / `WithFields(Fields) Entry` | 追加单个/多个结构化字段 |
| `WithError(err error) Entry` | 追加 `error` 字段 |
| `WithRequestID(id) Entry` / `WithTracing(id) Entry` | 追加请求 ID / 追踪字段（别名） |
| `Enabled(lvl Level) bool` | 判断当前级别是否会输出日志 |
| `StartTraffic(cmd string) *TrafficRec` | 开始一次流量记录（未启用流量日志时返回 nil，nil-safe） |
| `TrafficRec.WithTyp(t TrafficTyp) *TrafficRec` | 设置方向（recv/send），默认 recv |
| `TrafficRec.End(resp any, code string, fields ...any)` | 结束流量记录，写入耗时/返回码/响应体 |
| `TrafficRec.EndWithError(err error, fields ...any)` | 以错误结束流量记录 |
| `TrafficTyp` / `TrafficTypRecv`/`TrafficTypSend` | 流量方向常量 |
| `FromContext(ctx) Entry` | 从 context 取出挂载的日志实例，取不到则回退全局 |
| `WithLogger(ctx, e Entry) context.Context` | 把 `Entry` 挂载到 context |
| `CopyToContext(srcCtx, dstCtx) context.Context` | 把 srcCtx 上的日志实例复制到 dstCtx |
| `ConfigOption` | 函数式选项（`WithLevel`/`WithConsole`/`WithFilePath`/`WithEncoding`/`WithTraffic`/`WithTrimConfig` 等） |

引入路径：`import "github.com/tenz-io/gokit/logger/v3"`
