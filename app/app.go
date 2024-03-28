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

	// Prepare is the function that will be called before Run
	Prepare Prepare

	// Inits is a list of initializer function that will be called sequentially before Run
	Inits []InitFunc
}

type application struct {
	name          string
	prepare       Prepare
	runFunction   RunFunc
	initFunctions []InitFunc
}

func newApplication(
	name string,
	prepare Prepare,
	initFunctions []InitFunc,
	runFunction RunFunc,
) *application {
	return &application{
		name:          name,
		prepare:       prepare,
		initFunctions: initFunctions,
		runFunction:   runFunction,
	}
}

func (a *application) setupLog(c *Context) error {
	if !c.IsSet("log") {
		return fmt.Errorf("log directory is not set")
	}

	logDir := c.StringValue("log")
	if logDir == "" {
		logDir = "log"
	}

	return nil
}

// Run creates a new app and run
func Run(cfg Config, flags []Flag) {
	appCtx, cancel := context.WithCancel(context.Background())
	c := NewContext(appCtx)
	err := c.LoadFlags(cfg.Name, flags)
	if err != nil {
		syslog.Fatalf("load flags error, err: %v", err)
	}

	app := newApplication(cfg.Name, cfg.Prepare, cfg.Inits, cfg.Run)
	err = app.run(c, cfg.Config, cancel)
	if err != nil {
		syslog.Fatalf("run error, err: %v", err)
	}

}
