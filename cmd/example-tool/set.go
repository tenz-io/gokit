package main

import (
	"fmt"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/tenz-io/gokit/cmd"
)

var setCmd = &cmd.Command{
	Name:  "set",
	Usage: "demonstrate set command",
	Flags: []cmd.Flag{
		&cli.StringFlag{
			Name:    "key",
			Aliases: []string{"k"},
			Usage:   "key",
			Value:   "",
		},
		&cli.StringFlag{
			Name:  "val",
			Usage: "val",
			Value: "",
		},
	},
	Action: set,
}

func set(c *cmd.Context) error {
	var (
		key = c.String("key")
		val = c.String("val")
	)
	if key == "" {
		return fmt.Errorf("key is empty")
	}

	if val == "" {
		return fmt.Errorf("val is empty")
	}

	log.Printf("set key: %s, val: %s\n", key, val)
	return nil
}
