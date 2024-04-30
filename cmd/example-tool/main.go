package main

import (
	"github.com/tenz-io/gokit/cmd"
)

var commands = []*cmd.Command{
	getCmd,
	setCmd,
}

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
		Name:    "example-tool",
		Usage:   "example tool",
		ConfPtr: nil,
		Inits:   []cmd.InitFunc{},
	}

	cmd.Run(app, flags, commands...)
}
