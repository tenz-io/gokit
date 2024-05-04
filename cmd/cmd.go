package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"

	"github.com/urfave/cli/v2"
)

type (
	Flag    = cli.Flag
	Command = cli.Command
	Context = cli.Context

	StringFlag = cli.StringFlag
	BoolFlag   = cli.BoolFlag
	IntFlag    = cli.IntFlag

	InitFunc  func(c *Context, confPtr any) (CleanFunc, error)
	RunFunc   func(c *Context, confPtr any, errC chan<- error)
	CleanFunc func(c *Context)
)

const (
	FlagNameConfig  = "config"
	FlagNameLog     = "log"
	FlagNameVerbose = "verbose"
	FlagNameConsole = "console"
	FlagNameAdmin   = "admin"
	FlagNameHelp    = "help"
)

// commonFlags app common flags
var (
	basicFlags = []Flag{
		&StringFlag{
			Name:    FlagNameConfig,
			Aliases: []string{"c"},
			Usage:   "config file path",
			Value:   "config/app.yaml",
		},
		&BoolFlag{
			Name:    FlagNameHelp,
			Aliases: []string{"h"},
			Usage:   "print help message",
			Value:   false,
		},
	}
	commonFlags = []Flag{
		&StringFlag{
			Name:    FlagNameLog,
			Aliases: []string{"l"},
			Usage:   "log directory",
			Value:   "log",
		},
		&BoolFlag{
			Name:    FlagNameVerbose,
			Aliases: []string{"vv"},
			Usage:   "verbose mode",
			Value:   true,
		},
		&BoolFlag{
			Name:    FlagNameConsole,
			Aliases: []string{"s"},
			Usage:   "print log to console",
			Value:   false,
		},
		&IntFlag{
			Name:    FlagNameAdmin,
			Aliases: []string{"a"},
			Usage:   "admin port",
			Value:   8085,
		},
	}
)

type App struct {
	Name        string
	Usage       string
	CommonFlags bool
	ConfPtr     any
	Run         RunFunc
	Inits       []InitFunc
	cleans      []CleanFunc
}

// Run creates a new tool and run
func Run(app App, extraFlags []Flag, extraCommands ...*Command) error {
	var (
		flags = append(basicFlags, extraFlags...)
	)
	if app.CommonFlags {
		flags = append(flags, commonFlags...)
	}
	sort.Sort(cli.FlagsByName(flags))

	cliApp := cli.App{
		Name:        app.Name,
		Usage:       app.Usage,
		HideHelp:    true,
		HideVersion: true,
		Flags:       flags,
		Commands:    extraCommands,
		Metadata: map[string]any{
			FlagNameConfig: app.ConfPtr,
		},
		Before: func(c *Context) error {
			for _, init := range app.Inits {
				if clean, err := init(c, app.ConfPtr); err != nil {
					return err
				} else {
					app.cleans = append(app.cleans, clean)
				}
			}
			return nil
		},
		After: func(c *Context) error {
			for _, clean := range app.cleans {
				clean(c)
			}
			return nil
		},
		Action: run(app.Run, app.ConfPtr),
	}

	if err := cliApp.Run(os.Args); err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func run(runFunc RunFunc, confPtr any) func(c *Context) error {
	return func(c *Context) error {
		if c.Bool(FlagNameHelp) {
			return cli.ShowAppHelp(c)
		}

		if c.Bool(FlagNameVerbose) {
			printArgs(c)
			printConfig(confPtr)
		}

		if runFunc != nil {
			errC := make(chan error)
			runFunc(c, confPtr, errC)
			waitSignal(c.Context, errC, func() {
				log.Println("start cleanup")
			})
		}

		return nil

	}
}

// printConfig returns the config interface
func printConfig(confPtr any) {
	j, err := json.Marshal(confPtr)
	if err != nil {
		fmt.Printf("config: %+v\n", confPtr)
		return
	}
	fmt.Printf("config: %s\n", string(j))
}

// printArgs returns the command line arguments
func printArgs(c *Context) {
	fmt.Println("=== command line arguments ===")
	for _, flagName := range c.FlagNames() {
		fmt.Printf("%s: %v\n", flagName, c.Generic(flagName))
	}
	fmt.Println("================================")
}

func waitSignal(ctx context.Context, errC <-chan error, hook func()) {
	signC := make(chan os.Signal, 1)
	signal.Notify(signC, os.Interrupt, os.Kill)
	select {
	case <-signC:
		log.Println("received interrupt signal")
		hook()
		os.Exit(0)
	case <-ctx.Done():
		log.Println("context done")
		hook()
		os.Exit(0)
	case err := <-errC:
		if err != nil {
			log.Printf("run error: %+v", err)
			hook()
			os.Exit(1)
		} else {
			log.Println("run successfully")
			hook()
			os.Exit(0)
		}
	}
}

func GetConfig[Ptr any](c *Context) (Ptr, error) {
	var (
		zeroPtr Ptr
	)
	cnf, ok := c.App.Metadata[FlagNameConfig]
	if !ok {
		return zeroPtr, fmt.Errorf("config not found")
	}

	if cnf == nil {
		return zeroPtr, nil
	}

	if v, ok := cnf.(Ptr); ok {
		return v, nil
	}

	return zeroPtr, fmt.Errorf("invalid config type")
}
