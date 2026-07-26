# logger

基于 go.uber.org/zap 的结构化日志库，提供分级日志、文件轮转、输出裁剪、context 传递和请求/响应流量日志。

## 功能特性

- 四级日志（Debug/Info/Warn/Error），支持 `Configure`/`ConfigureWithOpts` 全局配置和 `NewEntry`/`NewEntryWithOpts` 创建独立实例
- 基于 lumberjack 的文件轮转（按大小/天数/备份数限制，自动压缩旧日志）
- `With`/`WithField`/`WithFields`/`WithError`/`WithRequestID` 链式附加结构化字段，返回新的 `Entry` 而不影响原实例
- 输出裁剪（`TrimConfig`）：对超长字符串、超深结构体/数组自动截断，避免大字段日志膨胀
- `StartTraffic`/`TrafficRec.End`/`EndWithError` 记录请求耗时、返回码和响应体，写入独立的流量日志文件
- `WithLogger`/`FromContext`/`CopyToContext` 支持把 `Entry` 挂载到 `context.Context` 并跨调用链传递
- `SetLevel`/`GetLevel` 支持运行时动态调整全局日志级别

## 能力清单

| 能力 | 含义 |
|------|------|
| 分级输出到不同文件 | `FilePath` 配置后自动按 `info.log`/`error.log`/`debug.log` 拆分写入，同时可选 `Console` 输出到 stdout，便于线上按级别分类排查 |
| 日志文件自动轮转与压缩 | 基于 lumberjack，按 `MaxSize`/`MaxAge`/`MaxBackups` 限制单文件大小、保留天数和备份数，超出后自动压缩旧文件，避免日志盘被写满 |
| 运行时动态调级 | `SetLevel`/`GetLevel` 基于 `zap.AtomicLevel`，可在服务运行中不重启地临时调低/调高全局日志级别，方便线上排障 |
| 链式附加结构化字段 | `With`/`WithField`/`WithFields`/`WithError`/`WithRequestID`/`WithTracing` 返回新的 `Entry`，不修改原实例，适合按请求/用户维度派生带上下文字段的日志器 |
| 输出裁剪防止大字段刷屏 | `TrimConfig`（`ArrLimit`/`StrLimit`/`DeepLimit`/`Ignores`）对超长字符串、超深结构体/map/slice 自动截断或跳过指定字段，避免大请求体、长数组把日志刷爆 |
| 独立的流量日志 | `Traffic`/`TrafficPath` 等配置开启后，`StartTraffic`+`TrafficRec.End`/`EndWithError` 会把每次请求的耗时、返回码、响应体写入单独的 `traffic.log`，与业务日志分离，方便统计接口耗时和成功率 |
| Context 跨调用链传递日志实例 | `WithLogger`/`FromContext`/`CopyToContext` 把携带了字段的 `Entry` 挂载到 `context.Context`，在函数调用链、goroutine 之间传递而不丢失已附加的字段 |
| 独立实例与全局实例并存 | `NewEntry`/`NewEntryWithOpts` 可创建不影响全局配置的独立日志实例，适合多租户或需要不同日志策略的子模块 |
| 兼容旧版 API | 保留 `WithLoggerLevel`/`WithTrafficEnabled`/`ConfigureTrafficWithOpts`/`StartTrafficRec` 等旧函数别名，便于从历史版本平滑升级而不用大改调用代码 |

## 快速开始

```go
import (
	"context"

	"github.com/tenz-io/gokit/logger/v2"
)

func main() {
	logger.Configure(logger.Config{
		Level:    logger.InfoLevel,
		Console:  true,
		FilePath: "log",
		Traffic:  true,
	})

	logger.Info("server started", "port", 8080)

	log := logger.With("user_id", "u123")
	log.Info("login")

	rec := logger.StartTraffic("GetUser")
	rec.End(map[string]any{"id": "u123"}, "0")

	ctx := logger.WithLogger(context.Background(), log)
	logger.FromContext(ctx).Warn("low balance")
}
```

## API 速查

| 符号 | 说明 |
|------|------|
| `Configure(cfg Config)` | 用 `Config` 初始化全局日志实例 |
| `ConfigureWithOpts(opts ...ConfigOption)` | 用函数式选项初始化全局日志实例 |
| `NewEntry(cfg Config) Entry` | 创建独立的日志实例，不影响全局 |
| `NewEntryWithOpts(opts ...ConfigOption) Entry` | 用函数式选项创建独立日志实例 |
| `Config` | 日志配置（级别/控制台/文件路径/轮转/流量/裁剪等） |
| `TrimConfig` | 输出裁剪配置（数组长度/字符串长度/嵌套深度/忽略字段） |
| `Level` / `DebugLevel`/`InfoLevel`/`WarnLevel`/`ErrorLevel` | 日志级别类型及常量 |
| `Fields` | `map[string]any`，用于批量传入结构化字段 |
| `L() Entry` | 获取全局日志实例 |
| `SetLevel(lvl Level)` / `GetLevel() Level` | 运行时设置/读取全局日志级别 |
| `Entry` | 日志接口：`Debug/Info/Warn/Error`（及 `xxxf` 格式化版本） |
| `Debug/Debugf/Info/Infof/Warn/Warnf/Error/Errorf` | 包级便捷函数，作用于全局日志实例 |
| `With(args ...any) Entry` | 追加一组 key-value 结构化字段，返回新 `Entry` |
| `WithField(k string, v any) Entry` / `WithFields(Fields) Entry` | 追加单个/多个结构化字段 |
| `WithError(err error) Entry` | 追加 `error` 字段 |
| `WithRequestID(id string) Entry` / `WithTracing(id string) Entry` | 追加请求 ID/追踪字段 |
| `Enabled(lvl Level) bool` | 判断当前级别是否会输出日志 |
| `StartTraffic(cmd string) *TrafficRec` | 开始一次流量记录（未启用流量日志时返回 nil） |
| `TrafficRec.End(resp any, code string, fields ...any)` | 结束流量记录，写入耗时/返回码/响应体 |
| `TrafficRec.EndWithError(err error, fields ...any)` | 以错误结束流量记录 |
| `Data(t *Traffic)` | 直接写入一条预构建的流量日志 |
| `Traffic` / `ReqEntity` / `RespEntity` / `TrafficTyp` | 流量日志相关数据结构（`TrafficTypRecv`/`TrafficTypSend`） |
| `FromContext(ctx context.Context) Entry` | 从 context 取出挂载的日志实例，取不到则回退全局 |
| `WithLogger(ctx, e Entry) context.Context` | 把 `Entry` 挂载到 context |
| `CopyToContext(srcCtx, dstCtx context.Context) context.Context` | 把 srcCtx 上的日志实例复制到 dstCtx |

引入路径：`import "github.com/tenz-io/gokit/logger/v2"`
