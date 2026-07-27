package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/tenz-io/gokit/logger/v3"
)

// CleanFunc 释放由 InitFunc 获取的资源。它始终按
// LIFO 顺序调用,即便后续 InitFunc 失败也会触发,因此部分启动不会
// 泄漏。即使其 InitFunc 返回了部分初始化的资源,CleanFunc 也必须可安全调用。
type CleanFunc func(context.Context) error

// InitFunc 初始化一个应用关注点,从
// Context 读取 config 与 flag,以及解码后的 conf。它返回(可能为 nil)
// CleanFunc,App 在 shutdown 时调用。非 nil error 会中止启动;
// App 在返回前会运行已收集的 cleanup。
type InitFunc func(c *Context, conf any) (CleanFunc, error)

// RunFunc 是主服务循环。它必须阻塞,直到 application
// context 被取消(c.Done()),或通过向 errC 发送以报告致命错误。
// errC 上的 nil error 表示干净完成。
type RunFunc func(c *Context, conf any, errC chan<- error)

// Config 描述一个应用。
type Config struct {
	// Name 为 app 名,在 usage 与日志中作为 flag-set 名使用。
	Name string

	// Usage 是在 -h/--help 输出中显示的简短说明。
	Usage string

	// Conf 是应用的 config 值,由
	// With* initializer 从 config 文件解码。它(若调用方使用
	// 指针则按指针)会传给每个 Init 与 Run。
	Conf any

	// Inits 在 Run 之前顺序执行。每个可返回一个 CleanFunc。
	Inits []InitFunc

	// Run 是主服务循环,在所有 Inits 成功后启动。
	Run RunFunc
}

// ExitCode 是 Run 报告的进程 exit code。0 表示成功。
type ExitCode int

const (
	ExitOK       ExitCode = 0
	ExitSetup    ExitCode = 1 // flag 解析或 init 失败
	ExitRunError ExitCode = 2 // Run 报告了 error
	ExitSignal   ExitCode = 3 // 被 signal 中断
)

// String 返回 ExitCode 的可读名,使日志与退出码输出显示
// "ExitSignal" 而非裸数字 3。
func (c ExitCode) String() string {
	switch c {
	case ExitOK:
		return "ExitOK"
	case ExitSetup:
		return "ExitSetup"
	case ExitRunError:
		return "ExitRunError"
	case ExitSignal:
		return "ExitSignal"
	}
	return fmt.Sprintf("ExitCode(%d)", int(c))
}

// Run 构建并启动由 cfg 描述的应用,阻塞至
// shutdown。它返回 ExitCode 而非调用 os.Exit,以便调用方
// (与测试)决定如何处理;main 包装器通常执行
// `os.Exit(int(app.Run(cfg)))`。
//
// opts 配置可选输入:WithExtraFlags 扩展/覆盖内置 flag,
// WithArgs 注入命令行参数(测试使用;默认 os.Args[1:])。
// 生产入口通常直接 app.Run(cfg),无需任何 option。
func Run(cfg Config, opts ...RunOption) ExitCode {
	// 前置校验:缺失 Run 会让 Run goroutine nil-func panic,
	// 空 Name 会让 flag-set 与日志难以辨识来源。两者均以
	// setup 失败返回,而非让进程在启动期崩溃。
	if cfg.Name == "" {
		fmt.Fprintln(os.Stderr, "app: Config.Name is empty")
		return ExitSetup
	}
	if cfg.Run == nil {
		fmt.Fprintf(os.Stderr, "%s: Config.Run is nil\n", cfg.Name)
		return ExitSetup
	}

	o := defaultRunOptions()
	for _, opt := range opts {
		opt(o)
	}
	fs, err := ParseFlags(cfg.Name, o.extraFlags, o.resolveArgs())
	if err != nil {
		// Help 与 parse 错误都落到此处。输出到 stderr 以便可见;
		// 绝不在包内部 os.Exit。
		fmt.Fprintf(os.Stderr, "%s: %v\n", cfg.Name, err)
		return ExitSetup
	}
	fs.Print(flagOutput)

	// 将 logger 配置为应用最先使用的部件,以便后续
	// 启动日志为结构化。WithLogger 也可作为 Init 供
	// 希望默认开启 traffic 日志的调用方使用;在此处运行
	// 保证即便在 Inits 运行之前也有可用的全局 logger。
	configureLogger(cfg.Name, fs)

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := NewContext(appCtx, fs)

	app := newApplication(cfg.Name, cfg.Inits, cfg.Run)
	return app.run(c, cfg.Conf, cancel)
}

// configureLogger 从解析后的 flag 装配全局 logger,一次完成
// level/console/file/traffic。它在第一个 InitFunc 之前无条件
// 运行,因此后续 init 已有可用 logger。WithTraffic /
// WithLogger 仅在希望运行期覆盖 traffic 时才需要,且不再重复装配基础 logger。
func configureLogger(name string, fs *Flags) {
	logDir := fs.String(FlagNameLog)
	if logDir == "" {
		logDir = "log"
	}
	verbose := fs.Bool(FlagNameVerbose)
	loggingFile := fs.Bool(FlagNameLoggingFile)
	loggingConsole := fs.Bool(FlagNameLoggingConsole)
	traffic := fs.Bool(FlagNameTraffic)

	lvl := logger.InfoLevel
	if verbose {
		lvl = logger.DebugLevel
	}

	opts := []logger.ConfigOption{
		logger.WithLevel(lvl),
		logger.WithConsole(loggingConsole),
		logger.WithCaller(true),
		logger.WithTraffic(traffic),
	}
	if loggingFile {
		opts = append(opts, logger.WithFilePath(logDir))
	}
	logger.ConfigureWithOpts(opts...)
	logger.Infow("application starting", "name", name, "level", fmt.Sprint(int(lvl)))
}

// flagOutput 是 Run 输出解析后 flag 值的去处。生产环境写至
// os.Stdout;测试将其替换为 io.Discard 以保持输出干净。不与
// 替换并发 —— App 生命周期在启动期为单线程。
var flagOutput io.Writer = os.Stdout

// application 是由 Config 构建的 runnable。
type application struct {
	name     string
	initFns  []InitFunc
	runFn    RunFunc
	cleanFns []CleanFunc
}

func newApplication(name string, inits []InitFunc, run RunFunc) *application {
	return &application{
		name:     name,
		initFns:  inits,
		runFn:    run,
		cleanFns: make([]CleanFunc, 0, len(inits)),
	}
}

// run 执行 init/cleanup/run/wait 生命周期。它从不调用 os.Exit;
// 唯一由 signal 驱动的副作用是取消 appCtx 并运行 cleanup,
// 之后返回 ExitCode。
func (a *application) run(c *Context, conf any, cancelApp context.CancelFunc) ExitCode {
	// 顺序 init。任何失败时,按 LIFO 顺序运行已收集的 cleanup
	// 并返回 setup exit code。
	for _, init := range a.initFns {
		clean, err := init(c, conf)
		if err != nil {
			logger.Errorf("init failed: %+v", err)
			_ = a.runCleanup(c.Context)
			return ExitSetup
		}
		if clean != nil {
			a.cleanFns = append(a.cleanFns, clean)
		}
	}

	// errC 为 buffered,使"报告即退出"的 Run 不会因
	// 无人接收而阻塞(v2 的无缓冲 channel 在 Run
	// 于 WaitSignal select 之前 panic 时可能挂起)。
	errC := make(chan error, 1)

	go a.runFn(c, conf, errC)

	// 等待:来自 Run 的致命错误(-> run error)、signal(-> graceful
	// cleanup 后 signal 退出),或 Run 干净完成(errC<-nil)。
	code := wait(c, errC, func() {
		logger.Infow("shutting down", "name", a.name)
		cancelApp()
		_ = a.runCleanup(c.Context)
	})
	return code
}

// runCleanup 以 LIFO 顺序调用已收集的 cleanup。错误仅记录但不
// 中断链;一个 cleanup 失败不得跳过后续的。
func (a *application) runCleanup(ctx context.Context) error {
	var errs []error
	for i := len(a.cleanFns) - 1; i >= 0; i-- {
		fn := a.cleanFns[i]
		if fn == nil {
			continue
		}
		if err := fn(ctx); err != nil {
			logger.Errorf("cleanup error: %+v", err)
			errs = append(errs, err)
		}
	}
	a.cleanFns = nil
	return errors.Join(errs...)
}
