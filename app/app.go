package app

import (
	"context"
	"fmt"
	"log"
)

type CleanFunc func()

// InitFunc is the init function to call
type InitFunc func(c *Context, confPtr any) (CleanFunc, error)

// RunFunc is the run function
type RunFunc func(c *Context, confPtr any, errC chan<- error)

// Config is a config for an app
type Config struct {
	// Name is the app name.
	Name string

	// Usage is the app usage that will be shown in help
	Usage string

	// Conf is the config that will be passed to Run and Inits
	// It should be a pointer type
	Conf any

	// Inits is a list of initializer function that will be called sequentially before Run
	Inits []InitFunc

	// Run is used for server, will block until the server is stopped
	// e.g: such as http server
	Run RunFunc
}

type application struct {
	name     string
	initFns  []InitFunc
	runFn    RunFunc
	cleanFns []CleanFunc
}

func newApplication(
	name string,
	initFns []InitFunc,
	runFn RunFunc,
) *application {
	return &application{
		name:     name,
		initFns:  initFns,
		runFn:    runFn,
		cleanFns: make([]CleanFunc, 0, len(initFns)),
	}
}

// Run creates a new app and run
func Run(cfg Config, flags []Flag) {
	fs, err := parseArgs(cfg, flags)
	if err != nil {
		log.Fatalf("parse args error, err: %v", err)
	}

	appCtx, cancel := context.WithCancel(context.Background())
	c := NewContext(appCtx, fs)

	app := newApplication(cfg.Name, cfg.Inits, cfg.Run)
	err = app.run(c, cfg.Conf, cancel)
	if err != nil {
		log.Fatalf("run error, err: %+v", err)
	}

}

func parseArgs(cfg Config, flags []Flag) (*Flags, error) {
	if flags == nil {
		flags = defaultFlags
	} else {
		flags = append(flags, defaultFlags...)
	}

	fs, err := NewFlags(flags)
	if err != nil {
		return nil, fmt.Errorf("new flags error, err: %w", err)
	}
	// parse args into flags
	err = Parse(cfg.Name, fs)
	if err != nil {
		return nil, fmt.Errorf("parse args error, err: %w", err)
	}
	fs.Print()

	return fs, nil
}
