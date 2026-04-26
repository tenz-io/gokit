package main

import (
	"log"

	"github.com/tenz-io/gokit/cmd"
)

var commands = []*cmd.Command{
	getCmd,
	setCmd,
}

var flags []cmd.Flag

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
