package logger

import (
	"context"
	"reflect"
)

// ctxKey 是用于将 Entry 存入 context 的未导出 context key。
type ctxKey struct{}

// FromContext 返回绑定到 ctx 的 Entry,当未绑定时(包括 ctx 为 nil 时)
// 返回全局 logger。这是 context 传播链的读取端:调用方用 WithLogger 把
// 一个携带 request_id、url 等的每请求 Entry 绑定到 ctx,随后调用链下方
// 的任意函数都能用 FromContext 取回它,使共享 field 自动盖到每行日志,
// 而无需反复传递。
func FromContext(ctx context.Context) Entry {
	if ctx == nil {
		return current()
	}
	if e, ok := ctx.Value(ctxKey{}).(Entry); ok && !isNilEntry(e) {
		return e
	}
	return current()
}

// WithLogger 将一个 Entry 绑定到 ctx,返回派生的 context。传入 nil Entry
// 时为 no-op,因此调用方可以链式调用 WithLogger(ctx, logger.With...)
// 而无需做守卫。
func WithLogger(ctx context.Context, e Entry) context.Context {
	if ctx == nil || isNilEntry(e) {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, e)
}

// CopyToContext 将显式绑定在 srcCtx 上的 Entry 复制到 dstCtx,返回派生
// 的 dstCtx。当 srcCtx 未绑定 Entry 时,原样返回 dstCtx;不会捕获全局
// 回退项。这在跨越 context 边界(例如启动一个 goroutine 或进入一个
// 构造全新 context 的 transport)时保留请求 logger 很有用。
func CopyToContext(srcCtx, dstCtx context.Context) context.Context {
	if srcCtx == nil || dstCtx == nil {
		return dstCtx
	}
	e, ok := srcCtx.Value(ctxKey{}).(Entry)
	if !ok || isNilEntry(e) {
		return dstCtx
	}
	return WithLogger(dstCtx, e)
}

func isNilEntry(e Entry) bool {
	if e == nil {
		return true
	}
	v := reflect.ValueOf(e)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
