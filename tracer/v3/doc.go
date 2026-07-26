// Package tracer provides context-based request ID propagation and traffic
// mode (debug/stress/shadow) flag management.
//
// v3 is a clean rewrite with no backwards-compatibility shims. It is designed
// around a data-driven flag registry so that parsing a flag string
// ("debug|shadow"), rendering a flag back to its string form, and iterating
// the known flags all read from a single table — callers no longer need to
// hand-roll a switch per transport.
//
// Quick start:
//
//	ctx := tracer.WithRequestID(context.Background(), "req-123")
//	ctx = tracer.WithFlag(ctx, tracer.FlagDebug|tracer.FlagShadow)
//
//	if tracer.FromContext(ctx).IsDebug() {
//	    // verbose path
//	}
//
//	// Parse a header value into a flag set, then stash it on the context.
//	f := tracer.ParseFlag("debug|shadow")
//	ctx = tracer.WithFlag(ctx, f)
//
//	fmt.Println(tracer.FromContext(ctx)) // "debug|shadow"
//
// ID conventions:
//   - WithRequestID / RequestIDFromCtx: set/read the request ID; reading an
//     unset context auto-generates one (UUID with dashes stripped) so
//     downstream never observes an empty ID.
//   - RequestIDFromCtxOr: read without auto-generation; returns "" when
//     absent, for "is one already present?" checks.
//
// Behavior notes (differ from v2):
//   - Flag.Is(FlagNone) returns false (v2 returned true). FlagNone is not a
//     real flag; test for it with f == FlagNone instead.
//   - Request-id symbols are spelled RequestID (capital ID) to match the Go
//     naming convention and logger/v3.
//   - Flag is int8 (a small bitmask is all that is needed) and context keys
//     are typed empty structs (zero collision, no string comparison).
package tracer
