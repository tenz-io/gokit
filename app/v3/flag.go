// Package app 提供一个轻量的应用生命周期框架:flag 解析、
// 带 LIFO cleanup 的有序 init、可选的 admin HTTP server
// (pprof / metrics / ping),以及 graceful、signal 驱动的 shutdown。
//
// v3 是在 logger/v3 与 annotation/v3 之上的全新重写,修复了
// v2 的启动期风险:
//   - Flag 为不可变 value spec;解析结果快照到一个 Flags
//     值中,而非修改调用方定义的 Flag 结构体(v2 通过
//     &f.Value 给调用方字段起别名)。
//   - Parse 返回 error 而非调用 os.Exit,因此错误的参数与
//     -h/--help 可被测试,且永远不会中止进程。
//   - Cleanup 以 LIFO 顺序执行,即便后续 init 失败也会触发,
//     因此已获取的资源不会泄漏。
//   - admin server 使用专属的 *http.ServeMux(绝不使用包级
//     DefaultServeMux),并在退出时通过 Shutdown 执行 graceful shutdown。
//   - Run 返回 exit code 而非内部调用 os.Exit;由
//     调用方决定是否 os.Exit。
package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

// Flag name constants 为内置 flag 与 With* initializer 使用的 flag 名常量。
const (
	FlagNameConfig         = "config"          // config 文件路径
	FlagNamePort           = "port"            // 服务 HTTP 端口
	FlagNameAdminPort      = "admin-port"      // admin HTTP 端口
	FlagNameLog            = "log"             // 日志输出目录
	FlagNameLoggingFile    = "logging-file"    // 启用文件日志
	FlagNameLoggingConsole = "logging-console" // 启用控制台日志
	FlagNameVerbose        = "verbose"         // verbose(debug 级别)日志
)

// FlagKind 标识 flag 的值类型,使 Parse 能将解析出的 token 路由到正确的
// setter,并让 Flags 返回有类型的值。
type FlagKind int

const (
	FlagKindString FlagKind = iota
	FlagKindInt
	FlagKindBool
	FlagKindDuration
)

// FlagSpec 将一个命令行 flag 描述为不可变 value:flag 的
// 名字、kind、默认值、usage 文本,以及一个可选的 "env" 钩子用于从环境变量取默认值。
// 与 v2 的 *Flag 结构体不同,FlagSpec 不携带可变目的端 ——
// 解析结果写入一个全新的 Flags 快照,因此同一份 FlagSpec 切片可安全地跨调用和并发解析复用。
type FlagSpec struct {
	Name    string
	Kind    FlagKind
	Default any
	Usage   string
	// Env 非空时,若 flag 未出现在 argv 中,则从该名字的环境变量取默认值。
	// 优先级: argv > env > Default。
	Env string
}

// StringFlag、IntFlag、BoolFlag 与 DurationFlag 是便捷构造器,
// 让调用方无需显式写出 FlagKind 即可构造 spec。
func StringFlag(name, def, usage string) FlagSpec {
	return FlagSpec{Name: name, Kind: FlagKindString, Default: def, Usage: usage}
}

func IntFlag(name string, def int, usage string) FlagSpec {
	return FlagSpec{Name: name, Kind: FlagKindInt, Default: def, Usage: usage}
}

func BoolFlag(name string, def bool, usage string) FlagSpec {
	return FlagSpec{Name: name, Kind: FlagKindBool, Default: def, Usage: usage}
}

func DurationFlag(name string, def time.Duration, usage string) FlagSpec {
	return FlagSpec{Name: name, Kind: FlagKindDuration, Default: def, Usage: usage}
}

// DefaultFlags 是每个 App 默认注册的内置 flag 集合,除非调用方
// 传入覆盖切片。With* initializer 读取这些名字。
var DefaultFlags = []FlagSpec{
	StringFlag(FlagNameConfig, "config/app.yaml", "Conf file"),
	IntFlag(FlagNamePort, 8080, "HTTP port"),
	IntFlag(FlagNameAdminPort, 8085, "Admin HTTP port"),
	StringFlag(FlagNameLog, "./log", "Log output directory"),
	BoolFlag(FlagNameLoggingFile, true, "Log to file(true/false)"),
	BoolFlag(FlagNameLoggingConsole, false, "Log to console(true/false)"),
	BoolFlag(FlagNameVerbose, false, "Verbose mode(true/false)"),
}

// Flags 是 ParseFlags 返回的不可变解析快照。每次查找
// 返回解析后的值以及一个表示该 flag 是否已注册的 bool;
// 忽略该 bool 的调用方对于未知名字会拿到零值,这是最常见的"直接读"路径。
// Flags 可安全并发读取。
type Flags struct {
	values map[string]flagValue
}

type flagValue struct {
	kind FlagKind
	str  string
	num  int64
	b    bool
	dur  time.Duration
}

// Lookup 返回原始 flagValue 以及该 flag 是否已注册。多数
// 调用方改用类型化的 String/Int/Bool/Duration 访问器。
func (fs *Flags) Lookup(name string) (flagValue, bool) {
	v, ok := fs.values[name]
	return v, ok
}

// String 返回 name 的 string 值,flag 不存在时返回 ""。
func (fs *Flags) String(name string) string { v, _ := fs.values[name]; return v.str }

// Int 返回 name 的 int 值,flag 不存在时返回 0。
func (fs *Flags) Int(name string) int { v, _ := fs.values[name]; return int(v.num) }

// Bool 返回 name 的 bool 值,flag 不存在时返回 false。
func (fs *Flags) Bool(name string) bool { v, _ := fs.values[name]; return v.b }

// Duration 返回 name 的 duration 值,flag 不存在时返回零值。
func (fs *Flags) Duration(name string) time.Duration {
	v, _ := fs.values[name]
	return v.dur
}

// IsSet 报告 name 是否为已注册的 flag。
func (fs *Flags) IsSet(name string) bool { _, ok := fs.values[name]; return ok }

// Print 将解析后的 flag 值以紧凑的 "name: value" 块写入 w。
// 输出至调用方提供的 io.Writer,便于测试捕获;Run
// 会将其经由 logger 输出。
func (fs *Flags) Print(w io.Writer) {
	fmt.Fprintln(w, "args: ==================")
	for _, f := range DefaultFlags {
		v, ok := fs.values[f.Name]
		if !ok {
			continue
		}
		fmt.Fprintf(w, "%s: %s\n", f.Name, v.display())
	}
	fmt.Fprintln(w, "==================")
}

func (v flagValue) display() string {
	switch v.kind {
	case FlagKindString:
		return v.str
	case FlagKindInt:
		return strconv.FormatInt(v.num, 10)
	case FlagKindBool:
		return strconv.FormatBool(v.b)
	case FlagKindDuration:
		return v.dur.String()
	}
	return v.str
}

// ParseFlags 将 specs 与 argv(默认为 os.Args[1:])对照解析,
// 返回不可变的 Flags 快照。它从不调用 os.Exit:解析错误、
// 未知 flag 或 -h/--help 都以 error 形式返回(help 时为 flag.ErrHelp),以便
// 调用方打印 usage 并决定 exit code。
//
// specs 为 nil 时与 DefaultFlags 合并;否则调用方的 specs
// 扩展 DefaultFlags(名字冲突时以调用方 specs 为准)。
func ParseFlags(name string, specs []FlagSpec, args []string) (*Flags, error) {
	if specs == nil {
		specs = DefaultFlags
	} else {
		// 在 mergeSpecs 将 override 折叠为单条目之前(会掩盖重复)拒绝
		// 调用方侧的重复。一个 spec 可以与内置 default 冲突(override),
		// 但两条同名的调用方 spec 视为错误。
		if dup := firstDuplicate(specs); dup != "" {
			return nil, fmt.Errorf("app: duplicate flag %q", dup)
		}
		specs = mergeSpecs(specs)
	}
	if args == nil {
		args = os.Args[1:]
	}

	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	// 静默默认错误输出;ParseFlags 以值形式返回错误,
	// 由调用方控制格式化与 exit code。
	fs.SetOutput(io.Discard)

	holders := make(map[string]*flagValue, len(specs))
	for _, s := range specs {
		if s.Name == "" {
			return nil, fmt.Errorf("app: flag with empty name")
		}
		if _, dup := holders[s.Name]; dup {
			return nil, fmt.Errorf("app: duplicate flag %q", s.Name)
		}
		v := &flagValue{kind: s.Kind}
		applySpec(fs, s, v)
		holders[s.Name] = v
	}

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("app: parse flags: %w", err)
	}

	out := make(map[string]flagValue, len(holders))
	for n, h := range holders {
		out[n] = *h
	}
	return &Flags{values: out}, nil
}

// mergeSpecs 返回 DefaultFlags 经 specs 增强后的结果;specs 中
// 名字已存在于 DefaultFlags 的条目将替换它(调用方 override),否则
// 追加。返回的切片为全新拷贝,因此调用方切片不会被修改。
// 调用方必须已对自身 specs 去重。
func mergeSpecs(specs []FlagSpec) []FlagSpec {
	out := make([]FlagSpec, 0, len(DefaultFlags)+len(specs))
	out = append(out, DefaultFlags...)

	seen := make(map[string]int, len(out))
	for i, s := range out {
		seen[s.Name] = i
	}
	for _, s := range specs {
		if idx, ok := seen[s.Name]; ok {
			out[idx] = s
		} else {
			out = append(out, s)
			seen[s.Name] = len(out) - 1
		}
	}
	return out
}

// firstDuplicate 返回 specs 中第一个出现两次的 flag 名,
// 若所有名字互不相同则返回 ""。供 ParseFlags 在 mergeSpecs
// 折叠 override 之前拒绝调用方侧重复使用。
func firstDuplicate(specs []FlagSpec) string {
	seen := make(map[string]struct{}, len(specs))
	for _, s := range specs {
		if s.Name == "" {
			continue
		}
		if _, ok := seen[s.Name]; ok {
			return s.Name
		}
		seen[s.Name] = struct{}{}
	}
	return ""
}

// applySpec 在 fs 上注册一个 flag,将解析出的默认值(env > spec
// default)写入 v 的字段,并将该 flag 的指针绑定到 v,使 Parse 能
// 填入 argv 的值。
func applySpec(fs *flag.FlagSet, s FlagSpec, v *flagValue) {
	def := s.Default
	if s.Env != "" {
		if e, ok := os.LookupEnv(s.Env); ok {
			def = e
		}
	}
	switch s.Kind {
	case FlagKindString:
		str, _ := def.(string)
		v.str = str
		fs.Func(s.Name, s.Usage, func(val string) error {
			v.str = val
			return nil
		})
	case FlagKindInt:
		num := toInt(def)
		v.num = num
		fs.Func(s.Name, s.Usage, func(val string) error {
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			v.num = n
			return nil
		})
	case FlagKindBool:
		b := toBool(def)
		v.b = b
		fs.BoolFunc(s.Name, s.Usage, func(val string) error {
			b, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			v.b = b
			return nil
		})
	case FlagKindDuration:
		d := toDuration(def)
		v.dur = d
		fs.Func(s.Name, s.Usage, func(val string) error {
			d, err := time.ParseDuration(val)
			if err != nil {
				return err
			}
			v.dur = d
			return nil
		})
	}
}

func toInt(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int64:
		return x
	case string:
		if n, err := strconv.ParseInt(x, 10, 64); err == nil {
			return n
		}
	}
	return 0
}

func toBool(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		if b, err := strconv.ParseBool(x); err == nil {
			return b
		}
	}
	return false
}

func toDuration(v any) time.Duration {
	switch x := v.(type) {
	case time.Duration:
		return x
	case string:
		if d, err := time.ParseDuration(x); err == nil {
			return d
		}
	}
	return 0
}
