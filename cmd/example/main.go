package main

import (
	"fmt"
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
	cmd.Run(app, flags)
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

		logger.WithFields(logger.Fields{
			"config": mycnf,
			"FOO":    os.Getenv("FOO"),
		}).Debug("debug config")

	}
}

type MyConfig struct {
	Foo string `yaml:"foo" json:"foo"`
}
