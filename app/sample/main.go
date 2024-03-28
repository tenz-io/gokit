package main

import (
	"encoding/json"
	"github.com/tenz-io/gokit/app"
	syslog "log"
)

var flags = []app.Flag{
	&app.StringFlag{
		Name:  "log",
		Value: "log",
		Usage: "Log output directory",
	},
	&app.StringFlag{
		Name:  "config",
		Value: "test.yaml",
		Usage: "Config file",
	},
	&app.BoolFlag{
		Name:  "verbose",
		Value: false,
		Usage: "Verbose mode",
	},
	&app.IntFlag{
		Name:  "port",
		Value: 8080,
		Usage: "Admin HTTP port",
	},
}

func main() {
	cfg := app.Config{
		Name:    "sample",
		Usage:   "Sample Server",
		Config:  &MyConfig{},
		Prepare: app.PrepareConfig,
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

	verbose := c.StringValue("verbose")
	syslog.Printf("verbose: %s\n", verbose)

	bs, err := json.Marshal(confPtr)
	if err != nil {
		syslog.Println("failed to marshal config")
		return err
	}
	syslog.Println(string(bs))
	return nil
}

type MyConfig struct {
	Foo string `yaml:"foo"`
}
