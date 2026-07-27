package tracer

import "context"

// flagCtxKey 是用于存入 Flag 的未导出、带类型 context key。带类型的空
// struct key 不会碰撞(其他包无法产生此类型),且在取值路径上避免了字符串
// key 比较。
type flagCtxKey struct{}

// WithFlag 返回一个携带 flag 的派生 context。后续的 WithFlag 调用会
// 覆盖(替换)而非 OR 组合 —— 用 [WithFlags] 一次设置多个,或自行 OR
// 各位:WithFlag(ctx, FlagDebug|FlagShadow)。nil ctx 原样返回。
func WithFlag(ctx context.Context, flag Flag) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, flagCtxKey{}, flag)
}

// WithFlags 将给定 flag 做 OR 组合并把结果存入 ctx,替换此前已设置的
// 任何 flag 集。无 flag(或只有 FlagNone)时存入 FlagNone。nil ctx 原样
// 返回。
func WithFlags(ctx context.Context, flags ...Flag) context.Context {
	if ctx == nil {
		return nil
	}
	var f Flag
	for _, fl := range flags {
		f |= fl
	}
	return context.WithValue(ctx, flagCtxKey{}, f)
}

// FromContext 返回由 [WithFlag] 或 [WithFlags] 存入 ctx 的 Flag,若未
// 存入(含 ctx 为 nil)则返回 FlagNone。它永不 panic。
func FromContext(ctx context.Context) Flag {
	if ctx == nil {
		return FlagNone
	}
	if f, ok := ctx.Value(flagCtxKey{}).(Flag); ok {
		return f
	}
	return FlagNone
}
