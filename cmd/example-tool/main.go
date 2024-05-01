package main

import (
	"github.com/tenz-io/gokit/cmd"
	"log"
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

	err := cmd.Run(app, flags, commands...)
	if err != nil {
		log.Fatal(err)
	}
}
