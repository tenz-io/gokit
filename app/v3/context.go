package app

import (
	"context"
)

// Context threads the application's cancellation context and the parsed flags
// to every Init and Run callback. It is constructed by Run; callers do not
// build it directly. The embedded context.Context is the application context,
// cancelled on shutdown so Run callbacks can watch <-c.Done().
type Context struct {
	context.Context
	flags *Flags
}

// NewContext builds a Context from an application context and parsed flags.
// Exported for tests that construct a Context without going through Run.
func NewContext(ctx context.Context, flags *Flags) *Context {
	return &Context{Context: ctx, flags: flags}
}

// Flags returns the parsed command-line flags.
func (c *Context) Flags() *Flags { return c.flags }
