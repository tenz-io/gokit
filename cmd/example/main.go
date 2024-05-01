package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tenz-io/gokit/cmd"
	"github.com/tenz-io/gokit/logger"
)

var flags = []cmd.Flag{
	&cmd.StringFlag{
		Name:    "env",
		Aliases: []string{"e"},
		Value:   "test",
		Usage:   "Environment",
	},
}

func main() {
	app := cmd.App{
		Name:    "example",
		Usage:   "this is ann example",
		ConfPtr: &MyConfig{},
		Inits: []cmd.InitFunc{
			cmd.WithYamlConfig(),
			cmd.WithLogger(true),
			cmd.WithAdminHTTPServer(),
		},
		Run: run(),
	}
	err := cmd.Run(app, flags)
	if err != nil {
		log.Fatal(err)
	}
}

func run() cmd.RunFunc {
	return func(c *cmd.Context, confPtr any, errC chan<- error) {
		logger.Infof("run application")

		env := c.String("env")

		logger.WithFields(logger.Fields{
			"env": env,
		}).Info("get env")

		mycnf, ok := confPtr.(*MyConfig)
		if !ok {
			errC <- fmt.Errorf("invalid config type: %T", confPtr)
		}

		mycnf2, err := cmd.GetConfig[*MyConfig](c)
		if err != nil {
			errC <- fmt.Errorf("get config error: %w", err)
		}

		logger.WithFields(logger.Fields{
			"config": mycnf,
			"mycnf2": mycnf2,
			"FOO":    os.Getenv("FOO"),
		}).Debug("debug config")

	}
}

type MyConfig struct {
	Foo string `yaml:"foo" json:"foo"`
}
