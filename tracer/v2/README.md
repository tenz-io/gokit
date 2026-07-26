# tracer

基于 context 的请求 ID 传递及 debug/stress/shadow 标志位管理。

## 功能特性

- `WithRequestId` / `RequestIdFromCtx`：将请求 ID 写入 context，读取时若未设置会自动生成一个 UUID（去掉短横线）作为兜底
- `RequestIdFromCtxOr`：读取 context 中的请求 ID，未设置时返回空字符串，不做兜底生成，适合判断"是否已存在"的场景
- `WithFlag` / `WithFlags`：将 debug/stress/shadow 等流量标志位写入 context，`WithFlags` 支持一次传入多个标志并按位组合
- `FromContext`：从 context 中读取当前的流量标志位，未设置时返回 `FlagNone`
- `Flag.IsDebug` / `Flag.IsStress` / `Flag.IsShadow`：便捷方法，判断是否命中对应的单一标志位
- `Flag.Is` / `Flag.HasAll`：通用方法，判断是否命中给定的标志位组合（两者语义一致，均要求全部命中）

## 能力清单

| 能力 | 含义 |
| --- | --- |
| 生成并透传请求 ID | `WithRequestId` 将请求 ID 写入 context，供跨函数/跨调用链传递，便于日志关联和链路追踪 |
| 缺省自动兜底生成请求 ID | `RequestIdFromCtx` 在 context 中未设置请求 ID 时，自动生成一个去掉短横线的 UUID 作为兜底值，避免下游拿到空 ID |
| 判断请求 ID 是否已存在 | `RequestIdFromCtxOr` 未设置时返回空字符串而不做兜底生成，适合"仅在已有 ID 时才透传，否则自行处理"的场景 |
| 设置单个流量标志位 | `WithFlag` 将 debug/stress/shadow 中的某一个标志写入 context，用于标记当前请求的特殊处理模式 |
| 组合设置多个流量标志位 | `WithFlags` 一次传入多个标志并按位或组合后写入 context，适合同时开启例如 debug+shadow 的场景 |
| 读取当前流量标志位 | `FromContext` 从 context 中取出当前生效的标志位组合，未设置时返回 `FlagNone`，供后续分支判断使用 |
| 判断是否为调试模式 | `Flag.IsDebug` 判断是否命中 `FlagDebug`，用于决定是否输出更详细的调试信息 |
| 判断是否为压测流量 | `Flag.IsStress` 判断是否命中 `FlagStress`，用于压测场景下隔离统计、跳过副作用等处理 |
| 判断是否为影子流量 | `Flag.IsShadow` 判断是否命中 `FlagShadow`，用于流量录制与回放场景，避免影子流量产生真实副作用 |
| 通用标志位组合判断 | `Flag.Is` / `Flag.HasAll` 判断当前标志是否包含指定的标志位组合，适合一次性校验多个标志同时存在的情况 |

## 快速开始

```go
import "github.com/tenz-io/gokit/tracer/v2"

func handle(ctx context.Context) {
	// 写入请求 ID，供日志、下游调用透传使用
	ctx = tracer.WithRequestId(ctx, "req-123")
	id := tracer.RequestIdFromCtx(ctx) // "req-123"

	// 标记本次请求为 debug + shadow 模式
	ctx = tracer.WithFlag(ctx, tracer.FlagDebug|tracer.FlagShadow)

	flag := tracer.FromContext(ctx)
	if flag.IsDebug() {
		// 输出更详细的调试信息
	}
	if flag.IsShadow() {
		// 走影子流量逻辑（录制/回放）
	}

	// 也可以用多个独立标志组合
	ctx = tracer.WithFlags(ctx, tracer.FlagStress, tracer.FlagShadow)
}
```

## API 速查

| 符号 | 说明 |
| --- | --- |
| `type Flag int` | 流量模式标志位的位掩码类型 |
| `FlagNone` / `FlagDebug` / `FlagStress` / `FlagShadow` | 预定义标志：无标志、调试模式、压测模式、影子模式（流量录制与回放） |
| `(Flag) Is(flagToCheck Flag) bool` | 判断 f 是否包含 flagToCheck 中的全部位 |
| `(Flag) HasAll(flags Flag) bool` | 判断 f 是否包含 flags 中的全部位，语义同 `Is` |
| `(Flag) IsDebug() bool` | 判断是否设置了 `FlagDebug` |
| `(Flag) IsStress() bool` | 判断是否设置了 `FlagStress` |
| `(Flag) IsShadow() bool` | 判断是否设置了 `FlagShadow` |
| `FromContext(ctx context.Context) Flag` | 从 context 读取流量标志，未设置时返回 `FlagNone` |
| `WithFlag(ctx context.Context, flag Flag) context.Context` | 将标志位写入 context |
| `WithFlags(ctx context.Context, flags ...Flag) context.Context` | 组合多个标志位后写入 context |
| `RequestIdFromCtx(ctx context.Context) string` | 读取请求 ID，未设置时自动生成并返回一个新 ID |
| `RequestIdFromCtxOr(ctx context.Context) string` | 读取请求 ID，未设置时返回空字符串 |
| `WithRequestId(ctx context.Context, requestID string) context.Context` | 将请求 ID 写入 context |

引用路径：`import "github.com/tenz-io/gokit/tracer/v2"`
