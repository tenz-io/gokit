package monitor

import "context"

// exporterCtxKeyType is the context key type for the injected Exporter. A
// distinct unexported type avoids collisions with any other package's string
// context keys.
type exporterCtxKeyType struct{}

var exporterCtxKey = exporterCtxKeyType{}

// Init is the single-flight injection entry point: it creates an Exporter
// bound to cmd (or reuses one already present in ctx) and returns a context
// carrying it. Call this once at the request edge so every downstream Begin
// along the chain shares the same Exporter instead of each creating its own.
//
// If ctx already carries an Exporter it is reused unchanged — Init is
// idempotent within a request, so wrapping twice never overwrites the
// existing flight with a fresh one.
func Init(ctx context.Context, cmd string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if exp := exporterFromContext(ctx); exp != nil {
		return ctx
	}
	return context.WithValue(ctx, exporterCtxKey, NewExporter(cmd))
}

// WithExporter injects exp into ctx. Prefer Init at the request edge; use
// WithExporter when you already hold a specific Exporter you want to reuse
// (e.g. CopyToContext forwarding it to a derived context).
func WithExporter(ctx context.Context, exp Exporter) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, exporterCtxKey, exp)
}

// FromContext returns the Exporter carried by ctx, or a non-nil no-op
// Exporter when none was injected. Callers can therefore always use the
// returned value without nil-checking: in un-instrumented contexts every
// point is a cheap no-op.
func FromContext(ctx context.Context) Exporter {
	if ctx == nil {
		return noop
	}
	if exp := exporterFromContext(ctx); exp != nil {
		return exp
	}
	return noop
}

// HasExporter reports whether ctx carries a real (non-noop) Exporter, i.e.
// whether Init or WithExporter has been called on this context chain.
func HasExporter(ctx context.Context) bool {
	return exporterFromContext(ctx) != nil
}

// CopyToContext forwards the Exporter from srcCtx into dstCtx, returning the
// augmented dstCtx. Use this when spawning a fresh context that must keep the
// single-flight Exporter (e.g. a goroutine or a sub-request with its own
// context root). If srcCtx has no Exporter, dstCtx is returned unchanged.
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

// exporterFromContext returns the injected Exporter or nil if absent. It is the
// shared reader used by Init (to decide reuse), FromContext, and HasExporter.
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
