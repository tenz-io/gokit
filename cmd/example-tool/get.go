package main

import (
	"fmt"
	"log"

	"github.com/tenz-io/gokit/cmd"
)

var getCmd = &cmd.Command{
	Name:  "get",
	Usage: "demonstrate get command",
	Flags: []cmd.Flag{
		&cmd.StringFlag{
			Name:    "key",
			Aliases: []string{"k"},
			Usage:   "key",
			Value:   "",
		},
	},
	Action: get,
}

func get(c *cmd.Context) error {
	key := c.String("key")
	if key == "" {
		return fmt.Errorf("key is empty")
	}
	log.Println("get key:", key)
	return nil
}
