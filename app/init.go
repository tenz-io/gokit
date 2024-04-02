package app

import (
	"encoding/json"
	"fmt"
	syslog "log"
	"net/http"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/tenz-io/gokit/logger"
)

func InitYamlConfig(c *Context, confPtr any) (CleanFunc, error) {
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
		syslog.Println("config: ", PrettyString(confPtr))
	}

	return cleanFn, nil
}

func InitJsonConfig(c *Context, confPtr any) (CleanFunc, error) {
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
		syslog.Println("config: ", PrettyString(confPtr))
	}

	return cleanFn, nil
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

func InitLogger(c *Context, _ any) (CleanFunc, error) {
	var (
		logDir         = "log"
		verbose        = false
		lvl            = logger.InfoLevel
		loggingFile    = true
		loggingConsole = false
		cleanFn        = func() {}
	)

	if lp, err := c.GetFlags().String(flagNameLog); err == nil && lp != "" {
		logDir = lp
	}

	if v, err := c.GetFlags().Bool(flagNameVerbose); err == nil {
		verbose = v
	}

	if v, err := c.GetFlags().Bool(flagNameLoggingFile); err == nil {
		loggingFile = v
	}

	if v, err := c.GetFlags().Bool(flagNameLoggingConsole); err == nil {
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
		logger.WithTrafficFileEnabled(loggingFile),
		logger.WithTrafficConsoleEnabled(loggingConsole),
	)

	return cleanFn, nil
}

// InitDefaultHandler will register profiling, ping, and prometheus metric
// handler to http.DefaultServeMux. Don't forget to run http.ListenAndServe on
// the main run function or use InitAdminHTTPServer
func InitDefaultHandler(_ *Context, _ any) (CleanFunc, error) {
	AddProfilingHandler(http.DefaultServeMux)
	AddPingHandler(http.DefaultServeMux)
	AddPrometheusHandler(http.DefaultServeMux)

	return func() {
		syslog.Println("cleanup default handler")
	}, nil
}

// InitAdminHTTPServer will run the default http.DefaultServeMux with port from
// env. If PORT environment variable is not set, the HTTP server will not be run.
// Use this if the service don't server any other HTTP traffic
func InitAdminHTTPServer(c *Context, _ any) (CleanFunc, error) {
	var (
		rawPort = "8081"
		portSrc = ""
	)

	if envPort := os.Getenv("ADMIN_PORT"); envPort != "" {
		rawPort = envPort
		portSrc = "ADMIN_PORT env variable"
	}

	if argPort, err := c.GetFlags().Int("admin-port"); err == nil && argPort > 0 {
		rawPort = strconv.Itoa(argPort)
		portSrc = "admin-port command line argument"
	}

	syslog.Printf("source: (%s), admin-port: %s\n", portSrc, rawPort)
	port, err := strconv.ParseInt(rawPort, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s (%s) as number, err: %w", portSrc, rawPort, err)
	}

	go func() {
		listenOn := fmt.Sprintf(":%d", port)
		syslog.Printf("Starting admin HTTP server at %s\n", listenOn)
		if err := http.ListenAndServe(listenOn, nil); err != nil {
			syslog.Fatalf("Admin HTTP server error, err: %+v", err)
		}
	}()

	return func() {
		syslog.Println("cleanup admin http server")
	}, nil
}
