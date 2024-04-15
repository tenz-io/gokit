package app

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
	return func(c *Context, confPtr any) (CleanFunc, error) {
		var (
			cleanFn = func() {}
		)
		configPath, err := c.GetFlags().String("config")
		if err != nil {
			return nil, fmt.Errorf("get config from flags error, err: %w", err)
		}
		if configPath == "" {
			return nil, fmt.Errorf("config file is empty")
		}

		err = ReadConfig(configPath, confPtr, yaml.Unmarshal)
		if err != nil {
			return cleanFn, err
		}

		if v, err := c.GetFlags().Bool("verbose"); v && err == nil {
			log.Println("config: ", PrettyString(confPtr))
		}

		return cleanFn, nil
	}
}

// WithJsonConfig will read the json config file and unmarshal it to the confPtr
func WithJsonConfig() InitFunc {
	return func(c *Context, confPtr any) (CleanFunc, error) {
		var (
			cleanFn = func() {}
		)

		configPath, err := c.GetFlags().String("config")
		if err != nil {
			return nil, fmt.Errorf("get config from flags error, err: %w", err)
		}
		if configPath == "" {
			return nil, fmt.Errorf("config file is empty")
		}

		err = ReadConfig(configPath, confPtr, json.Unmarshal)
		if err != nil {
			return cleanFn, err
		}

		if v, err := c.GetFlags().Bool("verbose"); v && err == nil {
			log.Println("config: ", PrettyString(confPtr))
		}

		return cleanFn, nil
	}
}

// WithDotEnvConfig will read the .env file and set the environment variables
func WithDotEnvConfig(filenames ...string) InitFunc {
	return func(c *Context, _ any) (CleanFunc, error) {
		var (
			cleanFn = func() {}
		)

		err := godotenv.Load(filenames...)
		if err != nil {
			return cleanFn, fmt.Errorf("error loading .env file: %w", err)
		}

		return cleanFn, nil
	}

}

// WithLogger will configure the logger with the given options
func WithLogger(trafficEnabled bool) InitFunc {
	return func(c *Context, _ any) (CleanFunc, error) {
		var (
			logDir         = "log"
			verbose        = false
			lvl            = logger.InfoLevel
			loggingFile    = true
			loggingConsole = false
			cleanFn        = func() {}
		)

		if lp, err := c.GetFlags().String(FlagNameLog); err == nil && lp != "" {
			logDir = lp
		}

		if v, err := c.GetFlags().Bool(FlagNameVerbose); err == nil {
			verbose = v
		}

		if v, err := c.GetFlags().Bool(FlagNameLoggingFile); err == nil {
			loggingFile = v
		}

		if v, err := c.GetFlags().Bool(FlagNameLoggingConsole); err == nil {
			loggingConsole = v
		}

		lvl = If(verbose, logger.DebugLevel, logger.InfoLevel)

		logger.ConfigureWithOpts(
			logger.WithLoggerLevel(lvl),
			logger.WithDirectory(logDir),
			logger.WithFileEnabled(loggingFile),
			logger.WithConsoleEnabled(loggingConsole),
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
	return func(c *Context, confPtr any) (CleanFunc, error) {
		var (
			cleanFn = func() {}
		)

		adminPort, err := c.GetFlags().Int(FlagNameAdminPort)
		if err != nil {
			return cleanFn, fmt.Errorf("cannot get admin port, err: %w", err)
		}

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

		return func() {
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
