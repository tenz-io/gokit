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
	Command = cli.Command
	Context = cli.Context

	InitFunc  func(c *Context, confPtr any) (CleanFunc, error)
	RunFunc   func(c *Context, confPtr any, errC chan<- error)
	CleanFunc func(c *Context)
)

type App struct {
	Name    string
	Usage   string
	ConfPtr any
	Run     RunFunc
	Inits   []InitFunc
	cleans  []CleanFunc
}

// Run creates a new tool and run
func Run(app App, extraFlags []Flag, extraCommands ...*Command) error {
	var (
		flags = append(commonFlags, extraFlags...)
	)

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
