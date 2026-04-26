package app

import (
	"context"
)

type Context struct {
	context.Context
	flags *Flags
}

// NewContext creates a new context. For use in when invoking an App or Command action.
func NewContext(ctx context.Context, flags *Flags) *Context {
	return &Context{
		Context: ctx,
		flags:   flags,
	}
}

// GetFlags retrieves the flags
func (c *Context) GetFlags() *Flags {
	return c.flags
}
