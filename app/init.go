package app

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	syslog "log"
	"net/http"
	"os"
	"strconv"
)

func PrepareConfig(c *Context, confPtr any) error {
	if !c.IsSet("config") {
		return fmt.Errorf("config file is not set")
	}

	configPath := c.StringValue("config")
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

	if c.BoolValue("verbose") {
		bs, err := json.Marshal(confPtr)
		if err != nil {
			syslog.Println("failed to json marshal config")
			return err
		}
		syslog.Println(string(bs))
	}

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
func InitAdminHTTPServer(_ *Context) (func(), error) {
	rawPort := os.Getenv("PORT")
	if rawPort == "" {
		rawPort = "8081"
	}

	port, err := strconv.ParseInt(rawPort, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s (PORT env variable) as number, err: %w", rawPort, err)
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
