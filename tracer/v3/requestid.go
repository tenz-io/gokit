package tracer

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// requestIDCtxKey 是用于存入 request ID 字符串的带类型 context key。
// 带类型(非字符串),因此不会与其它包挂在同一 context 上的 key 碰撞。
type requestIDCtxKey struct{}

// WithRequestID 返回一个以 id 作为 request ID 的派生 context。存入空 id
// 是 no-op(context 原样返回),因此调用方可直接 WithRequestID(ctx,
// maybeEmpty) 链式调用而无需守卫。nil ctx 原样返回。
func WithRequestID(ctx context.Context, id string) context.Context {
	if ctx == nil || id == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDCtxKey{}, id)
}

// RequestIDFromCtx 返回由 [WithRequestID] 存入 ctx 的 request ID。当未存入
// (或 ctx 为 nil)时,它通过 [NewRequestID] 自动生成一个新 ID,使下游
// 永不观察到空 ID —— 用于 inbound 处理的读路径。
//
// 注意:生成的 ID 不会写回 ctx,因此对同一无 ID 的 context 反复读取会
// 得到不同 id。要为整个请求固定一个 id,请在 inbound 边界调用一次
// [EnsureRequestID] 并把返回的 context 向下游透传。
func RequestIDFromCtx(ctx context.Context) string {
	if ctx != nil {
		if id, ok := ctx.Value(requestIDCtxKey{}).(string); ok && id != "" {
			return id
		}
	}
	return NewRequestID()
}

// EnsureRequestID 保证 ctx 携带一个 request ID。若已存入,则返回已存 id
// 且 ctx 原样返回(不重新派生);否则生成一个新 ID,存入派生 context,
// 并一并返回。在 inbound 边界(middleware)调用一次,使整个请求生命周期
// 内日志、响应 header 与下游调用都观察到同一 id —— 不同于
// [RequestIDFromCtx],后者在未设置时每次调用都生成新 id。
//
// nil ctx 返回一个生成的 id 与一个 nil context。
func EnsureRequestID(ctx context.Context) (context.Context, string) {
	if ctx != nil {
		if id, ok := ctx.Value(requestIDCtxKey{}).(string); ok && id != "" {
			return ctx, id // 已存在 —— 不重新派生
		}
	}
	id := NewRequestID()
	return WithRequestID(ctx, id), id
}

// RequestIDFromCtxOr 返回存入 ctx 的 request ID,缺失时返回 ""。不同于
// [RequestIDFromCtx],它不会自动生成,适用于 "仅当已存在时才转发" 的
// 判断。
func RequestIDFromCtxOr(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(requestIDCtxKey{}).(string); ok {
		return id
	}
	return ""
}

// NewRequestID 生成一个新的、去除连字符的 UUIDv4 字符串(32 个 hex 字符)。
// 它是 [RequestIDFromCtx] 使用的回退实现,被导出以便在尚无 context 时
// 构建 ID 的调用方复用同一生成器。
func NewRequestID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
