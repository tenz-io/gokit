package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

type (
	Flag = cli.Flag

	StringFlag = cli.StringFlag
	BoolFlag   = cli.BoolFlag
	IntFlag    = cli.IntFlag
)

const (
	FlagNameConfig  = "config"
	FlagNameEnv     = "env"
	FlagNameLog     = "log"
	FlagNameVerbose = "verbose"
	FlagNameConsole = "console"
	FlagNameAdmin   = "admin"
	FlagNameHelp    = "help"
)

const (
	FlagAliasConfig  = "c"
	FlagAliasEnv     = "e"
	FlagAliasLog     = "l"
	FlagAliasVerbose = "v"
	FlagAliasConsole = "s"
	FlagAliasAdmin   = "a"
	FlagAliasHelp    = "h"
)

var (
	helpFlag = &BoolFlag{
		Name:    FlagNameHelp,
		Aliases: []string{FlagAliasHelp},
		Usage:   "show help",
		Value:   false,
	}
	verboseFlag = &BoolFlag{
		Name:    FlagNameVerbose,
		Aliases: []string{FlagAliasVerbose},
		Usage:   "verbose mode",
		Value:   false,
	}
	yamlConfigFlag = &StringFlag{
		Name:    FlagNameConfig,
		Aliases: []string{FlagAliasConfig},
		Usage:   "config file path",
		Value:   "config/app.yaml",
	}
	jsonConfigFlag = &StringFlag{
		Name:    FlagNameConfig,
		Aliases: []string{FlagAliasConfig},
		Usage:   "config file path",
		Value:   "config/app.json",
	}
	dotEnvFlag = &StringFlag{
		Name:    FlagNameEnv,
		Aliases: []string{FlagAliasEnv},
		Usage:   "env file path",
		Value:   "config/.env",
	}
	logFlag = &StringFlag{
		Name:    FlagNameLog,
		Aliases: []string{FlagAliasLog},
		Usage:   "log output directory",
		Value:   "log",
	}
	consoleFlag = &BoolFlag{
		Name:    FlagNameConsole,
		Aliases: []string{FlagAliasConsole},
		Usage:   "print log to console",
		Value:   false,
	}
	adminFlag = &IntFlag{
		Name:    FlagNameAdmin,
		Aliases: []string{FlagAliasAdmin},
		Usage:   "admin port",
		Value:   8085,
	}
)

// commonFlags app common flags
var (
	commonFlags = []Flag{
		helpFlag,
		verboseFlag,
	}
)

func GetConfig[Ptr any](c *Context) (Ptr, error) {
	var (
		zeroPtr Ptr
	)
	cnf, ok := c.App.Metadata[FlagNameConfig]
	if !ok {
		return zeroPtr, fmt.Errorf("config not found")
	}

	if cnf == nil {
		return zeroPtr, nil
	}

	if v, ok := cnf.(Ptr); ok {
		return v, nil
	}

	return zeroPtr, fmt.Errorf("invalid config type")
}
