# tracer

基于 context 的轻量级请求追踪库。提供请求 ID 生成与传递、以及 debug/stress/shadow 标志位管理。零内部依赖，仅需 `google/uuid`。

## 快速开始

```go
import "github.com/tenz-io/gokit/tracer"

// 请求 ID：有则取，无则自动生成
ctx := context.Background()
id := tracer.RequestIdFromCtx(ctx)       // 自动生成 UUID
ctx = tracer.WithRequestId(ctx, "my-id") // 手动设置
id = tracer.RequestIdFromCtx(ctx)        // "my-id"

// 仅查询，不自动生成
raw := tracer.RequestIdFromCtxOr(ctx)    // "my-id" 或 ""

// 标志位：在 context 中传递运行模式
ctx = tracer.WithFlag(ctx, tracer.FlagDebug)
ctx = tracer.WithFlags(ctx, tracer.FlagDebug, tracer.FlagShadow)
```

## API 参考

### 请求 ID

| 函数 | 签名 |
|------|------|
| `RequestIdFromCtx` | `func RequestIdFromCtx(ctx context.Context) string` |
| `RequestIdFromCtxOr` | `func RequestIdFromCtxOr(ctx context.Context) string` |
| `WithRequestId` | `func WithRequestId(ctx context.Context, id string) context.Context` |

- `RequestIdFromCtx` — 从 context 获取请求 ID。若不存在则自动生成（UUID v7，无连接符）。**总是返回非空字符串**。
- `RequestIdFromCtxOr` — 从 context 获取请求 ID。若不存在返回空字符串。适用于无需自动生成的场景。
- `WithRequestId` — 将请求 ID 写入 context，返回新 context。

### 标志位

| 类型/函数 | 签名 |
|-----------|------|
| `Flag` | `type Flag int` |
| `FlagNone` | `Flag(0)` |
| `FlagDebug` | `1 << 0` |
| `FlagStress` | `1 << 1` |
| `FlagShadow` | `1 << 2` |
| `FromContext` | `func FromContext(ctx context.Context) Flag` |
| `WithFlag` | `func WithFlag(ctx context.Context, flag Flag) context.Context` |
| `WithFlags` | `func WithFlags(ctx context.Context, flags ...Flag) context.Context` |

Flag 方法：

| 方法 | 说明 |
|------|------|
| `Is(flag Flag) bool` | 检查指定标志位是否全部置位 |
| `HasAll(flags Flag) bool` | 检查多个标志位是否全部置位 |
| `IsDebug() bool` | 是否为 Debug 模式 |
| `IsStress() bool` | 是否为压测模式 |
| `IsShadow() bool` | 是否为影子流量模式 |

## 最佳实践

### 请求 ID 贯穿调用链

在每个入口（HTTP handler / gRPC interceptor）设置一次请求 ID，后续通过 context 自动传递：

```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ginext 等框架已自动处理此步骤
        ctx := tracer.WithRequestId(r.Context(), r.Header.Get("X-Request-Id"))
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 标志位组合

标志位可以自由组合，使用 `HasAll` 检查多个条件：

```go
ctx = tracer.WithFlags(ctx, tracer.FlagDebug, tracer.FlagShadow)
flag := tracer.FromContext(ctx)

if flag.HasAll(tracer.FlagDebug | tracer.FlagShadow) {
    // debug shadow 流量特殊处理
}
```

### nil context 安全

`FromContext` 和 `RequestIdFromCtx` 均接受 nil context，不会 panic：

```go
tracer.FromContext(nil)      // FlagNone
tracer.RequestIdFromCtx(nil) // 自动生成新 ID
```

## 与 v1 的区别

| 变更项 | v1 | v2 |
|--------|----|----|
| `RequestIdFromCtxOr` | 不存在 | 新增（仅查询，不生成） |
| `HasAll` | 不存在 | 新增（多标志位同时检查） |
| 实现 | 复杂 span 追踪系统 | 纯 context 键值存储 |
| 依赖 | `log`、`util`（损坏的导入） | 仅 `google/uuid` |
