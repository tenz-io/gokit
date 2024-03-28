package app

import (
	"context"
	"flag"
	"fmt"
	"os"
)

type Context struct {
	context.Context
	flags map[string]Flag
}

// NewContext creates a new context. For use in when invoking an App or Command action.
func NewContext(ctx context.Context) *Context {
	return &Context{
		Context: ctx,
		flags:   make(map[string]Flag),
	}
}

// LoadFlags loads the flags into the context
func (c *Context) LoadFlags(name string, flags []Flag) error {
	fs := flag.NewFlagSet(name, flag.ExitOnError)

	for _, f := range flags {
		f.Apply(fs)
	}

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	fmt.Println("args: ==================")
	for _, f := range flags {
		c.Set(f.GetName(), f)
		fmt.Println(f.GetName(), ":", f.GetValue())
	}
	fmt.Println("==================")

	return nil

}

// Set sets a context flag to a value.
func (c *Context) Set(name string, value Flag) {
	c.flags[name] = value
}

// IsSet determines if the flag was actually set
func (c *Context) IsSet(name string) bool {
	_, ok := c.flags[name]
	return ok
}

// StringValue retrieves the string value of a flag
func (c *Context) StringValue(name string) string {
	if f, ok := c.flags[name]; ok {
		return f.(*StringFlag).Value
	}
	return ""
}

// IntValue retrieves the int value of a flag
func (c *Context) IntValue(name string) int {
	if f, ok := c.flags[name]; ok {
		return f.(*IntFlag).Value
	}
	return 0
}

// BoolValue retrieves the bool value of a flag
func (c *Context) BoolValue(name string) bool {
	if f, ok := c.flags[name]; ok {
		return f.(*BoolFlag).Value
	}
	return false
}
