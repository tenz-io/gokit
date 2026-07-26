# tracer

基于 context 的请求 ID 传递及 debug/stress/shadow 标志位管理。v3 是一次无兼容包袱的干净重写：标志位由一张数据驱动的注册表统一管理，字符串解析/字符串化/遍历都走表，调用方不再需要手写 `switch`。

```go
import "github.com/tenz-io/gokit/tracer/v3"
```

## 模块介绍

tracer 解决两类问题：

- **请求 ID 透传**（`WithRequestID` / `RequestIDFromCtx`）：把请求 ID 写入 `context.Context`，在调用链/跨函数之间贯穿，便于日志关联和链路追踪；读取时若未设置会自动生成一个去掉短横线的 UUID 作为兜底，避免下游拿到空 ID。
- **流量标志位管理**（`WithFlag` / `WithFlags` / `FromContext`）：把 debug/stress/shadow 等流量标志位写入 context，支持位掩码组合；同时提供 `ParseFlag`/`Flag.String()` 把"字符串 ↔ 标志位"的转换收口到本模块，替代调用方重复的 `switch` 逻辑。

核心能力：

- 请求 ID：显式写入、读取兜底生成（UUID 去短横线）、不兜底的"是否已存在"判断
- 标志位：位掩码组合、单一/组合判断、debug/stress/shadow 便捷方法
- 字符串互转：`ParseFlag("debug|shadow")` / `(Flag).String()`，大小写不敏感、未知项忽略
- context 透传：标志位与请求 ID 用类型化空结构体作 key，零碰撞、无字符串比较

## 快速开始

```go
package main

import (
	"context"
	"fmt"

	"github.com/tenz-io/gokit/tracer/v3"
)

func main() {
	ctx := context.Background()

	// 1. 入口中间件：保证 context 一定带请求 ID（没有就生成并写回）
	//    这样日志、响应头、下游调用在整条链路上拿到的是同一个 ID
	ctx, id := tracer.EnsureRequestID(ctx)
	fmt.Println(id)

	// 显式写入请求 ID，供跨调用链透传
	ctx = tracer.WithRequestID(ctx, "req-123")
	fmt.Println(tracer.RequestIDFromCtx(ctx)) // req-123

	// 2. 标记本次请求为 debug + shadow 模式
	ctx = tracer.WithFlag(ctx, tracer.FlagDebug|tracer.FlagShadow)

	flag := tracer.FromContext(ctx)
	if flag.IsDebug() {
		// 输出更详细的调试信息
	}
	if flag.IsShadow() {
		// 走影子流量逻辑（录制/回放）
	}

	// 3. 把请求头里的字符串解析成 Flag（替代手写 switch）
	//    入站边界用 ParseFlagStrict：拼写错误(shdow)会报错，避免影子流量静默降级成真实流量
	f, err := tracer.ParseFlagStrict("debug|stress|shadow")
	if err != nil {
		// 拒绝该请求 / 记录告警
	}
	ctx = tracer.WithFlag(ctx, f)
	fmt.Println(tracer.FromContext(ctx)) // debug|stress|shadow

	// 兼容旧调用的宽松版：未知 token 静默忽略
	_ = tracer.ParseFlag("debug|bogus") // FlagDebug

	// 也可以用多个独立标志组合
	ctx = tracer.WithFlags(ctx, tracer.FlagStress, tracer.FlagShadow)
}
```

## 能力清单

| 能力 | 含义 |
| --- | --- |
| 入口保证请求 ID 唯一 | `EnsureRequestID` 在 context 未带 ID 时生成并写回，返回新 context 与 ID；已有则原样返回。入口中间件调用一次，整条链路（日志/响应头/下游）拿到同一个 ID |
| 生成并透传请求 ID | `WithRequestID` 将请求 ID 写入 context，供跨函数/跨调用链传递，便于日志关联和链路追踪 |
| 缺省自动兜底生成请求 ID | `RequestIDFromCtx` 在 context 中未设置请求 ID 时，自动生成一个去掉短横线的 UUID 作为兜底值，避免下游拿到空 ID（注意：生成值不写回 context，重复读取会得到不同 ID，需要稳定请用 `EnsureRequestID`） |
| 判断请求 ID 是否已存在 | `RequestIDFromCtxOr` 未设置时返回空字符串而不做兜底生成，适合"仅在已有 ID 时才透传，否则自行处理"的场景 |
| 设置单个流量标志位 | `WithFlag` 将 debug/stress/shadow 中的某一个（或组合）标志写入 context，用于标记当前请求的特殊处理模式 |
| 组合设置多个流量标志位 | `WithFlags` 一次传入多个标志并按位或组合后写入 context，适合同时开启例如 debug+shadow 的场景 |
| 读取当前流量标志位 | `FromContext` 从 context 中取出当前生效的标志位组合，未设置时返回 `FlagNone`，供后续分支判断使用 |
| 判断是否为调试模式 | `Flag.IsDebug` 判断是否命中 `FlagDebug`，用于决定是否输出更详细的调试信息 |
| 判断是否为压测流量 | `Flag.IsStress` 判断是否命中 `FlagStress`，用于压测场景下隔离统计、跳过副作用等处理 |
| 判断是否为影子流量 | `Flag.IsShadow` 判断是否命中 `FlagShadow`，用于流量录制与回放场景，避免影子流量产生真实副作用 |
| 通用标志位组合判断 | `Flag.Is` / `Flag.HasAll` 判断当前标志是否包含指定的标志位组合，适合一次性校验多个标志同时存在的情况 |
| 字符串解析为标志位（宽松） | `ParseFlag("debug\|shadow")` 把 `\|` 分隔的字符串解析成 `Flag`，大小写不敏感、未知项忽略，替代调用方手写 switch |
| 字符串解析为标志位（严格） | `ParseFlagStrict("debug\|shadow")` 遇到未知 token 返回 error，用于 HTTP/gRPC 入站边界，防止 shadow/stress 拼写错误后静默降级成真实流量 |
| 标志位转字符串 | `Flag.String()` 返回 `debug\|shadow` 形式的可读字符串，无标志返回 `none`；带本版本不认识的未知位时追加 `0xNN`（如 `debug\|0x80`），不会把未知位误判为"无标志" |

## API 速查

| 符号 | 说明 |
| --- | --- |
| `type Flag uint8` | 流量模式标志位的位掩码类型（`uint8`，8 个可用标志位，无符号位溢出） |
| `FlagNone` / `FlagDebug` / `FlagStress` / `FlagShadow` | 预定义标志：无标志、调试模式、压测模式、影子模式（流量录制与回放） |
| `(Flag) Is(x Flag) bool` | 判断 f 是否包含 x 中的全部位；`Is(FlagNone)` 返回 `false`（FlagNone 不是真正的标志） |
| `(Flag) HasAll(flags Flag) bool` | 判断 f 是否包含 flags 中的全部位，语义同 `Is` 的别名 |
| `(Flag) IsDebug() bool` | 判断是否设置了 `FlagDebug` |
| `(Flag) IsStress() bool` | 判断是否设置了 `FlagStress` |
| `(Flag) IsShadow() bool` | 判断是否设置了 `FlagShadow` |
| `(Flag) String() string` | 返回 `debug\|shadow` 形式的可读字符串，无标志返回 `none`；未知位追加 `0xNN`（如 `debug\|0x80`） |
| `(Flag) GoString() string` | `%#v` 调试输出形式，如 `Flag(debug\|shadow)` |
| `(Flag) Names() []string` | 返回所有已置位标志的名字切片（`flagTable` 顺序），无标志返回 `nil` |
| `ParseFlag(s string) Flag` | 把 `\|` 分隔字符串宽松解析成 `Flag`，大小写不敏感、未知项忽略，空串返回 `FlagNone` |
| `ParseFlagStrict(s string) (Flag, error)` | 严格解析：遇到未知 token 返回 error（命名首个未知项），用于入站边界防静默降级 |
| `FromContext(ctx context.Context) Flag` | 从 context 读取流量标志，未设置或 `nil` ctx 时返回 `FlagNone` |
| `WithFlag(ctx context.Context, flag Flag) context.Context` | 将标志位写入 context（覆盖式，非累加） |
| `WithFlags(ctx context.Context, flags ...Flag) context.Context` | 将多个标志位按位或组合后写入 context（覆盖式） |
| `EnsureRequestID(ctx) (context.Context, string)` | 保证 context 带请求 ID：已有则原样返回，否则生成并写回，返回新 context 与 ID |
| `RequestIDFromCtx(ctx context.Context) string` | 读取请求 ID，未设置时自动生成并返回一个新 ID（不写回 context） |
| `RequestIDFromCtxOr(ctx context.Context) string` | 读取请求 ID，未设置时返回空字符串 |
| `WithRequestID(ctx context.Context, id string) context.Context` | 将请求 ID 写入 context；空 id 为 no-op |
| `NewRequestID() string` | 生成一个新的去掉短横线的 UUID（32 个十六进制字符），供无 context 场景复用 |

引入路径：`import "github.com/tenz-io/gokit/tracer/v3"`

## 与 v2 的行为差异

v3 不保证与 v2 兼容，以下是显式的行为差异：

| 差异点 | v2 | v3 |
| --- | --- | --- |
| `Flag.Is(FlagNone)` | 恒真（`f&0==0`） | 返回 `false`：`FlagNone` 不是真正的标志，要判断"无标志"请用 `f == FlagNone` |
| 请求 ID 命名 | `WithRequestId` / `RequestIdFromCtx` / `RequestIdFromCtxOr` | `WithRequestID` / `RequestIDFromCtx` / `RequestIDFromCtxOr`（`ID` 大写，符合 Go 命名规范并与 `logger/v3` 一致） |
| `Flag` 底层类型 | `int` | `uint8`（无符号，8 个可用标志位，避免 `1<<7` 符号位溢出） |
| context key | 字符串常量（`"_flag_ctx_key"` 等） | 类型化空结构体（零碰撞、无字符串比较） |
| 字符串互转 | 无（调用方各自手写 switch，如 `ginext/utils.go` 的 `getTracerFlag`） | `ParseFlag`（宽松）/ `ParseFlagStrict`（严格，报错） / `Flag.String()` 收口到本模块 |
| 入口 ID 生成 | 调用方自己在入口生成并 `WithRequestId` 写回 | 新增 `EnsureRequestID` 一步完成（生成并写回，返回新 context） |
| 预留位渲染 | 不适用 | 未知位（`allFlagBits` 之外）在 `String()` 中以 `0xNN` 后缀保留，不被误判为 `none` |

调用方迁移速查（**不在本次范围，留待后续逐步改**）：

| 调用方代码（v2） | v3 等价 |
| --- | --- |
| `tracer.FromContext(ctx).IsDebug()` | 不变 |
| `tracer.WithFlag(ctx, flag)` | 不变 |
| `tracer.RequestIdFromCtx(ctx)` | `tracer.RequestIDFromCtx(ctx)` |
| `tracer.WithRequestId(ctx, id)` | `tracer.WithRequestID(ctx, id)`，或在入口用 `EnsureRequestID(ctx)` 一步生成并写回 |
| `tracer.RequestIdFromCtxOr(ctx)` | `tracer.RequestIDFromCtxOr(ctx)` |
| ginext `getTracerFlag(string)` | `tracer.ParseFlag(string)`（宽松）或 `tracer.ParseFlagStrict(string)`（入站严格，拼写错报错） |
