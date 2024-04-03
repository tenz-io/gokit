package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/tenz-io/gokit/logger"
)

func WithYamlConfig() InitFunc {
	return func(c *Context, confPtr any) (CleanFunc, error) {
		var (
			cleanFn = func() {}
		)
		bs, err := readConfigFile(c)
		if err != nil {
			return cleanFn, err
		}

		if err = yaml.Unmarshal(bs, confPtr); err != nil {
			return cleanFn, fmt.Errorf("yaml unmarshal config file fail, err: %w", err)
		}

		if v, err := c.GetFlags().Bool("verbose"); v && err == nil {
			log.Println("config: ", PrettyString(confPtr))
		}

		return cleanFn, nil
	}
}

func WithJsonConfig() InitFunc {
	return func(c *Context, confPtr any) (CleanFunc, error) {
		var (
			cleanFn = func() {}
		)
		bs, err := readConfigFile(c)
		if err != nil {
			return cleanFn, err
		}

		if err = json.Unmarshal(bs, confPtr); err != nil {
			return cleanFn, fmt.Errorf("json unmarshal config file fail, err: %w", err)
		}

		if v, err := c.GetFlags().Bool("verbose"); v && err == nil {
			log.Println("config: ", PrettyString(confPtr))
		}

		return cleanFn, nil
	}
}

func WithLogger() InitFunc {
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

func readConfigFile(c *Context) ([]byte, error) {
	configPath, err := c.GetFlags().String("config")
	if err != nil {
		return nil, fmt.Errorf("get config file error, err: %w", err)
	}
	if configPath == "" {
		return nil, fmt.Errorf("config file is empty")
	}

	bs, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file fail, err: %w", err)
	}

	return bs, nil
}
