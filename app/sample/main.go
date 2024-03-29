package main

import (
	"github.com/tenz-io/gokit/app"
	"github.com/tenz-io/gokit/logger"
)

var flags = []app.Flag{
	&app.StringFlag{
		Name:  "env",
		Value: "test",
		Usage: "Environment",
	},
}

func main() {
	cfg := app.Config{
		Name:   "sample",
		Usage:  "Sample Server",
		Config: &MyConfig{},
		Preparations: []app.Prepare{
			app.PrepareConfig,
			app.PrepareLogger,
		},
		Inits: []app.InitFunc{
			app.InitDefaultHandler,
			app.InitAdminHTTPServer,
		},
		Run: run,
	}
	app.Run(cfg, flags)
}

func run(c *app.Context, confPtr any, waitFunc app.WaitFunc) error {
	errC := make(chan error)
	waitFunc(errC)

	env, err := c.GetFlags().String("env")
	if err != nil {
		logger.Warnf("failed to get env: %v", err)
	}
	logger.WithFields(logger.Fields{
		"env": env,
	}).Info("get env")

	logger.WithFields(logger.Fields{
		"config": confPtr,
	}).Debug("debug config")

	return nil
}

type MyConfig struct {
	Foo string `yaml:"foo" json:"foo"`
}
