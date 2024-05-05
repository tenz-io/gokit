package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/joho/godotenv"

	"github.com/tenz-io/gokit/logger"
)

// WithYamlConfig will read the yaml config file and unmarshal it to the confPtr
func WithYamlConfig() InitFunc {
	commonFlags = append(commonFlags, yamlConfigFlag)

	return func(c *Context, confPtr any) (CleanFunc, error) {
		var (
			cleanFn = func(_ *Context) {}
		)

		configPath := c.String(FlagNameConfig)
		if configPath == "" {
			return cleanFn, fmt.Errorf("config file is empty")
		}

		err := ReadConfig(configPath, confPtr, yaml.Unmarshal)
		if err != nil {
			return cleanFn, err
		}

		return cleanFn, nil
	}
}

// WithJsonConfig will read the json config file and unmarshal it to the confPtr
func WithJsonConfig() InitFunc {
	commonFlags = append(commonFlags, jsonConfigFlag)

	return func(c *Context, confPtr any) (CleanFunc, error) {
		var (
			cleanFn = func(_ *Context) {}
		)

		configPath := c.String(FlagNameConfig)
		if configPath == "" {
			return cleanFn, fmt.Errorf("config file is empty")
		}

		err := ReadConfig(configPath, confPtr, json.Unmarshal)
		if err != nil {
			return cleanFn, err
		}

		return cleanFn, nil
	}
}

// WithDotEnvConfig will read the .env file and set the environment variables
func WithDotEnvConfig() InitFunc {
	commonFlags = append(commonFlags, dotEnvFlag)

	return func(c *Context, _ any) (CleanFunc, error) {
		var (
			cleanFn = func(_ *Context) {}
		)

		err := godotenv.Load(c.String(FlagNameEnv))
		if err != nil {
			return cleanFn, fmt.Errorf("error loading .env file: %w", err)
		}

		return cleanFn, nil
	}

}

// WithLogger will configure the logger with the given options
func WithLogger(trafficEnabled bool) InitFunc {
	commonFlags = append(commonFlags, logFlag, consoleFlag)

	return func(c *Context, _ any) (CleanFunc, error) {
		var (
			logDir      = c.String(FlagNameLog)
			verbose     = c.Bool(FlagNameVerbose)
			console     = c.Bool(FlagNameConsole)
			lvl         = logger.InfoLevel
			loggingFile = true
			cleanFn     = func(_ *Context) {}
		)

		if verbose {
			lvl = logger.DebugLevel
		}
		if console {
			loggingFile = false
		}

		logger.ConfigureWithOpts(
			logger.WithLoggerLevel(lvl),
			logger.WithDirectory(logDir),
			logger.WithFileEnabled(loggingFile),
			logger.WithConsoleEnabled(console),
			logger.WithSetAsDefaultLvl(true),
			logger.WithCallerEnabled(true),
		)

		logger.ConfigureTrafficWithOpts(
			logger.WithTrafficEnabled(trafficEnabled),
			logger.WithTrafficDirectory(logDir),
		)

		return cleanFn, nil
	}
}

// WithAdminHTTPServer will run the default http.DefaultServeMux with port from
// env. If PORT environment variable is not set, the HTTP server will not be run.
// Use this if the service don't server any other HTTP traffic
func WithAdminHTTPServer() InitFunc {
	commonFlags = append(commonFlags, adminFlag)

	return func(c *Context, confPtr any) (CleanFunc, error) {
		var (
			cleanFn = func(_ *Context) {}
		)

		adminPort := c.Int(FlagNameAdmin)
		if adminPort <= 0 || adminPort > 65535 {
			return cleanFn, fmt.Errorf("invalid admin port: %d", adminPort)
		}

		initDefaultHandler(c, confPtr)

		go func() {
			listenOn := fmt.Sprintf(":%d", adminPort)
			log.Printf("Starting admin HTTP server at %s\n", listenOn)
			if err := http.ListenAndServe(listenOn, nil); err != nil {
				log.Fatalf("Admin HTTP server error, err: %+v", err)
			}
		}()

		return func(_ *Context) {
			log.Println("cleanup admin http server")
		}, nil
	}
}

// ReadConfig will read the config file and unmarshal it to the confPtr
func ReadConfig(confPath string, confPtr any, unmarshalFn func([]byte, any) error) error {
	bs, err := os.ReadFile(confPath)
	if err != nil {
		return fmt.Errorf("read config file %s fail, err: %w", confPath, err)
	}

	err = unmarshalFn(bs, confPtr)
	if err != nil {
		return fmt.Errorf("unmarshal config file fail, err: %w", err)
	}
	return nil
}
