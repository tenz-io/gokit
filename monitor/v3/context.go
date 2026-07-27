package monitor

import "context"

// exporterCtxKeyType 是被注入 Exporter 的 context key 类型。使用一个
// 独立的未导出类型,可避免与任何其他包的字符串 context key 冲突。
type exporterCtxKeyType struct{}

var exporterCtxKey = exporterCtxKeyType{}

// Init 是 single-flight 注入入口:它创建一个绑定到 cmd 的 Exporter
// (或复用 ctx 中已存在的一个),并返回携带该 Exporter 的 context。
// 在请求边缘调用一次,使链路上每个下游 Begin 都共享同一个 Exporter,
// 而不是各自创建。
//
// 若 ctx 已携带一个 Exporter,则原样复用 —— Init 在一次请求内幂等,
// 因此重复包裹永远不会用新 flight 覆盖已有的 flight。
func Init(ctx context.Context, cmd string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if exp := exporterFromContext(ctx); exp != nil {
		return ctx
	}
	return context.WithValue(ctx, exporterCtxKey, NewExporter(cmd))
}

// WithExporter 将 exp 注入到 ctx。在请求边缘优先使用 Init;当你已经
// 持有某个具体 Exporter 想复用时(例如 CopyToContext 将它转发到派生
// context),请使用 WithExporter。
func WithExporter(ctx context.Context, exp Exporter) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, exporterCtxKey, exp)
}

// FromContext 返回 ctx 携带的 Exporter;若未注入,则返回一个非 nil
// 的 no-op Exporter。因此调用方永远可以直接使用返回值而无需 nil
// 检查:在未插桩的 context 中,每个采样点都是廉价的 no-op。
func FromContext(ctx context.Context) Exporter {
	if ctx == nil {
		return noop
	}
	if exp := exporterFromContext(ctx); exp != nil {
		return exp
	}
	return noop
}

// HasExporter 报告 ctx 是否携带一个真实(非 noop)的 Exporter,即
// 是否已在该 context 链上调用过 Init 或 WithExporter。
func HasExporter(ctx context.Context) bool {
	return exporterFromContext(ctx) != nil
}

// CopyToContext 将 Exporter 从 srcCtx 转发到 dstCtx,返回增强后的
// dstCtx。当你创建一个全新 context 却必须保留 single-flight
// Exporter 时(例如 goroutine 或拥有独立 context 根的子请求)使用
// 它。若 srcCtx 无 Exporter,则原样返回 dstCtx。
func CopyToContext(srcCtx, dstCtx context.Context) context.Context {
	if srcCtx == nil || dstCtx == nil {
		return dstCtx
	}
	exp := exporterFromContext(srcCtx)
	if exp == nil {
		return dstCtx
	}
	return WithExporter(dstCtx, exp)
}

// exporterFromContext 返回被注入的 Exporter,不存在时返回 nil。它是
// Init(用于判断是否复用)、FromContext 与 HasExporter 共用的读取器。
func exporterFromContext(ctx context.Context) Exporter {
	if ctx == nil {
		return nil
	}
	v := ctx.Value(exporterCtxKey)
	if v == nil {
		return nil
	}
	exp, ok := v.(Exporter)
	if !ok {
		return nil
	}
	return exp
}
