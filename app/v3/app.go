package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/tenz-io/gokit/logger/v3"
)

// CleanFunc releases resources acquired by an InitFunc. It is always called in
// LIFO order, including when a later InitFunc fails, so partial startup never
// leaks. CleanFunc must be safe to call even when its InitFunc returned a
// partially-initialized resource.
type CleanFunc func(context.Context) error

// InitFunc initializes one application concern, reading config and flags from
// the Context and the decoded config at conf. It returns a (possibly nil)
// CleanFunc the App invokes at shutdown. A non-nil error aborts startup; the
// App runs the cleanups collected so far before returning.
type InitFunc func(c *Context, conf any) (CleanFunc, error)

// RunFunc is the main service loop. It must block until the application
// context is cancelled (c.Done()) or report a fatal error by sending to errC.
// A nil error on errC signals clean completion.
type RunFunc func(c *Context, conf any, errC chan<- error)

// Config describes an application.
type Config struct {
	// Name is the app name, used as the flag-set name in usage and logs.
	Name string

	// Usage is a short description shown in -h/--help output.
	Usage string

	// Conf is the application's config value, decoded from the config file by
	// the With* initializers. It is passed (by pointer if the caller used a
	// pointer) to every Init and Run.
	Conf any

	// Inits runs sequentially before Run. Each may return a CleanFunc.
	Inits []InitFunc

	// Run is the main service loop, started after all Inits succeed.
	Run RunFunc
}

// ExitCode is the process exit code Run reports. Zero is success.
type ExitCode int

const (
	ExitOK       ExitCode = 0
	ExitSetup    ExitCode = 1 // flag parse or init failure
	ExitRunError ExitCode = 2 // Run reported an error
	ExitSignal   ExitCode = 3 // interrupted by signal
)

// Run builds and starts the application described by cfg, blocking until
// shutdown. It returns an ExitCode rather than calling os.Exit so callers
// (and tests) can decide what to do; main wrappers typically do
// `os.Exit(int(app.Run(cfg)))`.
//
// flags overrides the built-in DefaultFlags; pass nil for the defaults. argv
// defaults to os.Args[1:]; pass a slice to inject args (used by tests).
func Run(cfg Config, flags []FlagSpec, argv ...[]string) ExitCode {
	var args []string
	if len(argv) > 0 {
		args = argv[0]
	}
	fs, err := ParseFlags(cfg.Name, flags, args)
	if err != nil {
		// Help and parse errors both land here. Print to stderr for visibility;
		// never os.Exit from inside the package.
		fmt.Fprintf(os.Stderr, "%s: %v\n", cfg.Name, err)
		return ExitSetup
	}
	fs.Print(flagOutput)

	// Configure the logger as the first thing the app uses, so subsequent
	// startup logging is structured. WithLogger is also available as an Init
	// for callers that want traffic logging on by default; running it here
	// guarantees a working global logger even before Inits run.
	configureLogger(cfg.Name, fs)

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := NewContext(appCtx, fs)

	app := newApplication(cfg.Name, cfg.Inits, cfg.Run)
	return app.run(c, cfg.Conf, cancel)
}

// configureLogger wires the global logger from resolved flags. It mirrors the
// WithLogger initializer but runs unconditionally so the package has a working
// logger before the first InitFunc.
func configureLogger(name string, fs *Flags) {
	logDir := fs.String(FlagNameLog)
	if logDir == "" {
		logDir = "log"
	}
	verbose := fs.Bool(FlagNameVerbose)
	loggingFile := fs.Bool(FlagNameLoggingFile)
	loggingConsole := fs.Bool(FlagNameLoggingConsole)

	lvl := logger.InfoLevel
	if verbose {
		lvl = logger.DebugLevel
	}

	opts := []logger.ConfigOption{
		logger.WithLevel(lvl),
		logger.WithConsole(loggingConsole),
		logger.WithCaller(true),
	}
	if loggingFile {
		opts = append(opts, logger.WithFilePath(logDir))
	}
	logger.ConfigureWithOpts(opts...)
	logger.Infow("application starting", "name", name, "level", fmt.Sprint(int(lvl)))
}

// flagOutput is where Run prints resolved flag values. Production writes to
// os.Stdout; tests swap it to io.Discard to keep output clean. Not concurrent
// with swaps — the App lifecycle is single-threaded at startup.
var flagOutput io.Writer = os.Stdout

// application is the runnable built from a Config.
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

// run executes the init/cleanup/run/wait lifecycle. It never calls os.Exit;
// the only signal-driven side effect is cancelling appCtx and running cleanup,
// after which it returns an ExitCode.
func (a *application) run(c *Context, conf any, cancelApp context.CancelFunc) ExitCode {
	// Sequential init. On any failure, run collected cleanups in LIFO order
	// and report the setup exit code.
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

	// errC is buffered so a Run that reports-and-exits never blocks on a
	// nobody-listening receiver (v2's unbuffered channel could hang if Run
	// panicked before WaitSignal selected).
	errC := make(chan error, 1)

	go a.runFn(c, conf, errC)

	// Wait for: a fatal error from Run (-> run error), a signal (-> signal
	// exit after graceful cleanup), or Run completing cleanly (errC<-nil).
	code := wait(c, errC, func() {
		logger.Infow("shutting down", "name", a.name)
		cancelApp()
		_ = a.runCleanup(c.Context)
	})
	return code
}

// runCleanup invokes collected cleanups in LIFO order. Errors are logged but do
// not abort the chain; one cleanup failing must not skip later ones.
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
