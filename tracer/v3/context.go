package tracer

import "context"

// flagCtxKey is the unexported, typed context key under which a Flag is
// stashed. A typed empty struct key is collision-free (no other package can
// produce this type) and avoids string-key comparison on the value path.
type flagCtxKey struct{}

// WithFlag returns a derived context carrying flag. Later WithFlag calls
// overwrite (replace) rather than OR-combine — use [WithFlags] to set
// several at once, or OR the bits yourself: WithFlag(ctx, FlagDebug|FlagShadow).
// A nil ctx is returned unchanged.
func WithFlag(ctx context.Context, flag Flag) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, flagCtxKey{}, flag)
}

// WithFlags OR-combines the given flags and stores the result on ctx,
// replacing any flag set previously. With no flags (or only FlagNone) it
// stores FlagNone. A nil ctx is returned unchanged.
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

// FromContext returns the Flag stored on ctx by [WithFlag] or [WithFlags],
// or FlagNone when none is stored (including when ctx is nil). It never
// panics.
func FromContext(ctx context.Context) Flag {
	if ctx == nil {
		return FlagNone
	}
	if f, ok := ctx.Value(flagCtxKey{}).(Flag); ok {
		return f
	}
	return FlagNone
}
