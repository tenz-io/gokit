package main

import (
	"github.com/tenz-io/gokit/cmd"
)

var commands = []*cmd.Command{
	getCmd,
	setCmd,
}

func main() {
	app := cmd.App{
		Name:    "example-tool",
		Usage:   "example tool",
		ConfPtr: &MyConfig{},
		Inits: []cmd.InitFunc{
			cmd.WithYamlConfig(),
			cmd.WithLogger(false),
		},
	}

	cmd.Run(app, nil, commands...)
}

type MyConfig struct {
	Foo string `yaml:"foo" json:"foo"`
}
