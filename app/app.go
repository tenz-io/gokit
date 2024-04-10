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

	// Command is used for run once tools, exist after the command is done
	// e.g: such as command line tools
	Command RunFunc

	// Run is used for server, will block until the server is stopped
	// e.g: such as http server
	Run RunFunc
}

type application struct {
	name     string
	initFs   []InitFunc
	runF     RunFunc // for server
	commandF RunFunc // for command line tools
	cleanFs  []CleanFunc
}

func newApplication(
	name string,
	initFs []InitFunc,
	runF RunFunc,
	commandF RunFunc,
) *application {
	return &application{
		name:     name,
		initFs:   initFs,
		runF:     runF,
		commandF: commandF,
		cleanFs:  make([]CleanFunc, 0, len(initFs)),
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

	app := newApplication(cfg.Name, cfg.Inits, cfg.Run, cfg.Command)
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
