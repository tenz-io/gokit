# tracer

基于 context 的请求 ID 传递及 debug/stress/shadow 标志位管理。

## 功能特性

- `WithRequestId` / `RequestIdFromCtx`：将请求 ID 写入 context，读取时若未设置会自动生成一个 UUID（去掉短横线）作为兜底
- `RequestIdFromCtxOr`：读取 context 中的请求 ID，未设置时返回空字符串，不做兜底生成，适合判断"是否已存在"的场景
- `WithFlag` / `WithFlags`：将 debug/stress/shadow 等流量标志位写入 context，`WithFlags` 支持一次传入多个标志并按位组合
- `FromContext`：从 context 中读取当前的流量标志位，未设置时返回 `FlagNone`
- `Flag.IsDebug` / `Flag.IsStress` / `Flag.IsShadow`：便捷方法，判断是否命中对应的单一标志位
- `Flag.Is` / `Flag.HasAll`：通用方法，判断是否命中给定的标志位组合（两者语义一致，均要求全部命中）

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
