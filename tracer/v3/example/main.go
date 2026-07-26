// Example: tracer/v3 — request ID propagation and traffic-mode flags.
//
// Demonstrates: setting/reading a request ID on the context with fallback
// generation, combining flags, parsing a flag string off a header, and
// reading everything back from the context.
package main

import (
	"context"
	"fmt"

	"github.com/tenz-io/gokit/tracer/v3"
)

func main() {
	ctx := context.Background()

	// 1. Request ID: set explicitly, then read it back.
	ctx = tracer.WithRequestID(ctx, "req-123")
	fmt.Println("request id:", tracer.RequestIDFromCtx(ctx)) // req-123

	// 2. Reading an unset context auto-generates a fresh ID.
	fresh := tracer.RequestIDFromCtx(context.Background())
	fmt.Println("generated id:", fresh, "len:", len(fresh))

	// 3. Combine flags with bitwise OR and stash on the context.
	ctx = tracer.WithFlag(ctx, tracer.FlagDebug|tracer.FlagShadow)
	flag := tracer.FromContext(ctx)
	fmt.Println("flag debug?", flag.IsDebug())   // true
	fmt.Println("flag shadow?", flag.IsShadow()) // true
	fmt.Println("flag stress?", flag.IsStress()) // false
	fmt.Println("flag string:", flag)            // debug|shadow

	// 4. Parse a header value ("debug|stress|shadow") into a Flag in one call.
	header := "debug|stress|shadow"
	parsed := tracer.ParseFlag(header)
	fmt.Printf("parse %q -> %s (debug=%v stress=%v shadow=%v)\n",
		header, parsed, parsed.IsDebug(), parsed.IsStress(), parsed.IsShadow())

	// 5. WithFlags OR-combines its args (within a single call) and stores.
	ctx2 := tracer.WithFlags(context.Background(), tracer.FlagDebug, tracer.FlagShadow)
	fmt.Println("with flags:", tracer.FromContext(ctx2)) // debug|shadow

	// 6. End-to-end: header -> ParseFlag -> WithFlag -> read -> String.
	ctx3 := tracer.WithFlag(context.Background(), tracer.ParseFlag("debug|shadow"))
	fmt.Println("round-trip:", tracer.FromContext(ctx3).String()) // debug|shadow

	// 7. RequestIDFromCtxOr: no auto-generation, for "is one set?" checks.
	fmt.Println("has id?", tracer.RequestIDFromCtxOr(ctx3) == "") // true (no id on ctx3)
}
