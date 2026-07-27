package app

import "os"

// RunOption 配置 Run 的可选输入:扩展 flag 集与注入参数。生产调用方通常
// 不传任何 option(直接 app.Run(cfg)),测试通过 WithArgs 注入 argv。
type RunOption func(*runOptions)

type runOptions struct {
	extraFlags []FlagSpec
	args       []string
	argsSet    bool
}

func defaultRunOptions() *runOptions {
	return &runOptions{}
}

// WithExtraFlags 向内置 DefaultFlags 追加/覆盖 flag 规格。同名 spec 覆盖
// 内置项,与 ParseFlags 的 merge 语义一致。无扩展则不调用本 option。
func WithExtraFlags(flags ...FlagSpec) RunOption {
	return func(o *runOptions) {
		o.extraFlags = append(o.extraFlags, flags...)
	}
}

// WithArgs 注入命令行参数(默认为 os.Args[1:]),供测试在不改动进程
// 实参的情况下驱动 flag 解析。生产调用方不应使用本 option。
func WithArgs(args []string) RunOption {
	return func(o *runOptions) {
		o.args = args
		o.argsSet = true
	}
}

// resolveArgs 返回最终 argv:显式注入优先,否则 os.Args[1:]。
func (o *runOptions) resolveArgs() []string {
	if o.argsSet {
		return o.args
	}
	return os.Args[1:]
}
