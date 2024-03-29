package app

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	syslog "log"
	"net/http"
	"os"
	"strconv"

	"github.com/tenz-io/gokit/logger"
)

func PrepareConfig(c *Context, confPtr any) error {
	if !c.GetFlags().IsSet("config") {
		return fmt.Errorf("config file is not set")
	}

	configPath, err := c.GetFlags().String("config")
	if err != nil {
		return fmt.Errorf("get config file error, err: %w", err)
	}
	if configPath == "" {
		return fmt.Errorf("config file is empty")
	}

	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config file fail, err: %w", err)
	}

	err = yaml.Unmarshal(yamlFile, confPtr)
	if err != nil {
		return fmt.Errorf("unmarshal config file fail, err: %w", err)
	}

	if v, err := c.GetFlags().Bool("verbose"); v && err == nil {
		bs, err := json.Marshal(confPtr)
		if err != nil {
			syslog.Println("failed to json marshal config")
			return err
		}
		syslog.Println(string(bs))
	}

	return nil
}

func PrepareLogger(c *Context, _ any) error {
	var (
		logDir  = "log"
		verbose = false
		lvl     = logger.InfoLevel
	)

	if lp, err := c.GetFlags().String("log"); err == nil && lp != "" {
		logDir = lp
	}

	if v, err := c.GetFlags().Bool("verbose"); err == nil {
		verbose = v
	}

	lvl = If(verbose, logger.DebugLevel, logger.InfoLevel)

	logger.ConfigureWithOpts(
		logger.WithLoggerLevel(lvl),
		logger.WithDirectory(logDir),
		logger.WithFileEnabled(true),
		logger.WithSetAsDefaultLvl(true),
		logger.WithCallerEnabled(true),
	)

	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficDirectory(logDir),
		logger.WithTrafficFileEnabled(true),
	)

	return nil
}

// InitDefaultHandler will register profiling, ping, and prometheus metric
// handler to http.DefaultServeMux. Don't forget to run http.ListenAndServe on
// the main run function or use InitAdminHTTPServer
func InitDefaultHandler(_ *Context) (func(), error) {
	AddProfilingHandler(http.DefaultServeMux)
	AddPingHandler(http.DefaultServeMux)
	AddPrometheusHandler(http.DefaultServeMux)

	return func() {
		syslog.Println("cleanup default handler")
	}, nil
}

// InitAdminHTTPServer will run the default http.DefaultServeMux with port from
// env. If PORT environment variable is not set, the HTTP server will not be run.
// Use this if your service don't server any other HTTP traffic
func InitAdminHTTPServer(c *Context) (func(), error) {
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

	syslog.Printf("port source: %s, port: %s\n", portSrc, rawPort)
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
