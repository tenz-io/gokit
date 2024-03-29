package app

import (
	"context"
	"fmt"
	syslog "log"
)

// WaitFunc is a function that will block until it's time to quit
type WaitFunc func(...<-chan error)

// RunFunc is the run function
type RunFunc func(c *Context, confPtr any, waitFunc WaitFunc) error

// InitFunc is the init function to call
type InitFunc func(c *Context) (func(), error)

type Prepare func(c *Context, confPtr any) error

// Config is a config for an app
type Config struct {
	// Name is the app name.
	Name string

	// Usage is the app usage that will be shown in help
	Usage string

	// Config is the basic YAML config which will be parsed passed to Run.
	// It should be a pointer type
	Config any

	// Run is the main function that will be called
	Run RunFunc

	// Preparations is the function that will be called before Run
	Preparations []Prepare

	// Inits is a list of initializer function that will be called sequentially before Run
	Inits []InitFunc
}

type application struct {
	name          string
	preparations  []Prepare
	initFunctions []InitFunc
	runFunction   RunFunc
}

func newApplication(
	name string,
	preparations []Prepare,
	initFunctions []InitFunc,
	runFunction RunFunc,
) *application {
	return &application{
		name:          name,
		preparations:  preparations,
		initFunctions: initFunctions,
		runFunction:   runFunction,
	}
}

// Run creates a new app and run
func Run(cfg Config, flags []Flag) {
	fs, err := parseArgs(cfg, flags)
	if err != nil {
		syslog.Fatalf("parse args error, err: %v", err)
	}

	appCtx, cancel := context.WithCancel(context.Background())
	c := NewContext(appCtx, fs)

	app := newApplication(cfg.Name, cfg.Preparations, cfg.Inits, cfg.Run)
	err = app.run(c, cfg.Config, cancel)
	if err != nil {
		syslog.Fatalf("run error, err: %+v", err)
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
