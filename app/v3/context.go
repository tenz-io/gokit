package app

import (
	"context"
)

// Context 将应用的 cancellation context 与解析后的 flag
// 传递给每个 Init 与 Run 回调。它由 Run 构造;调用方不
// 直接构造它。内嵌的 context.Context 为 application context,
// 在 shutdown 时取消,使 Run 回调可监听 <-c.Done()。
type Context struct {
	context.Context
	flags *Flags
}

// NewContext 由 application context 与解析后的 flag 构造一个 Context。
// 导出供测试在不经过 Run 的情况下构造 Context 使用。
func NewContext(ctx context.Context, flags *Flags) *Context {
	return &Context{Context: ctx, flags: flags}
}

// Flags 返回解析后的命令行 flag。
func (c *Context) Flags() *Flags { return c.flags }
