package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"

	"github.com/tenz-io/gokit/annotation/v3"
	"github.com/tenz-io/gokit/logger/v3"
)

// UnmarshalFunc decodes raw config bytes into the config value. yaml.Unmarshal
// and json.Unmarshal satisfy it.
type UnmarshalFunc func([]byte, any) error

// WithYamlConfig decodes the YAML config file named by the `config` flag into
// conf, applying annotation defaults first and validating after. Returns no
// cleanup (file I/O is one-shot).
func WithYamlConfig() InitFunc { return decodeConfig(yaml.Unmarshal) }

// WithJsonConfig decodes the JSON config file named by the `config` flag into
// conf, applying annotation defaults first and validating after.
func WithJsonConfig() InitFunc { return decodeConfig(json.Unmarshal) }

func decodeConfig(unmarshal UnmarshalFunc) InitFunc {
	return func(c *Context, conf any) (CleanFunc, error) {
		path := c.Flags().String(FlagNameConfig)
		if path == "" {
			return nil, fmt.Errorf("config file is empty")
		}
		if err := ReadConfig(path, conf, unmarshal); err != nil {
			return nil, fmt.Errorf("read config %s: %w", path, err)
		}
		if c.Flags().Bool(FlagNameVerbose) {
			logger.Debugf("config: %s", PrettyString(conf))
		}
		return nil, nil
	}
}

// WithDotEnvConfig loads the given .env files into the process environment.
// With no filenames it loads the default ".env".
func WithDotEnvConfig(filenames ...string) InitFunc {
	return func(_ *Context, _ any) (CleanFunc, error) {
		if len(filenames) == 0 {
			filenames = []string{".env"}
		}
		if err := godotenv.Load(filenames...); err != nil {
			return nil, fmt.Errorf("load .env %v: %w", filenames, err)
		}
		return nil, nil
	}
}

// WithLogger reconfigures the global logger from flags, optionally enabling the
// traffic logger. It is a thin wrapper over logger/v3 so callers that want
// traffic logging can opt in with one line; the package already configures a
// basic logger in Run, so this is only needed for the Traffic flag.
func WithLogger(trafficEnabled bool) InitFunc {
	return func(c *Context, _ any) (CleanFunc, error) {
		logDir := c.Flags().String(FlagNameLog)
		if logDir == "" {
			logDir = "log"
		}
		verbose := c.Flags().Bool(FlagNameVerbose)
		loggingFile := c.Flags().Bool(FlagNameLoggingFile)
		loggingConsole := c.Flags().Bool(FlagNameLoggingConsole)

		lvl := logger.InfoLevel
		if verbose {
			lvl = logger.DebugLevel
		}

		opts := []logger.ConfigOption{
			logger.WithLevel(lvl),
			logger.WithConsole(loggingConsole),
			logger.WithCaller(true),
			logger.WithTraffic(trafficEnabled),
		}
		if loggingFile {
			opts = append(opts, logger.WithFilePath(logDir))
		}
		logger.ConfigureWithOpts(opts...)
		return nil, nil
	}
}

// WithAdminHTTPServer starts the admin HTTP server on the `admin-port` flag and
// mounts /debug/pprof, /metrics and /ping on its own ServeMux (never the global
// DefaultServeMux, which v2 polluted). The returned CleanFunc performs a
// graceful Shutdown with a timeout.
func WithAdminHTTPServer() InitFunc {
	return func(c *Context, _ any) (CleanFunc, error) {
		port := c.Flags().Int(FlagNameAdminPort)
		if port <= 0 || port > 65535 {
			return nil, fmt.Errorf("invalid admin port: %d", port)
		}

		mux := http.NewServeMux()
		AddProfilingHandler(mux)
		AddPingHandler(mux)
		AddPrometheusHandler(mux)

		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
			// Conservative timeouts so a stuck admin request can't hold shutdown.
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		}

		errC := make(chan error, 1)
		go func() {
			logger.Infow("starting admin http server", "addr", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errC <- err
			}
		}()

		clean := func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			logger.Infow("shutting down admin http server", "addr", srv.Addr)
			if err := srv.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("admin http shutdown: %w", err)
			}
			return nil
		}

		// Surface a listen error that arrived before Run wired up errC.
		select {
		case err := <-errC:
			_ = clean(context.Background())
			return nil, fmt.Errorf("admin http listen: %w", err)
		default:
			return clean, nil
		}
	}
}

// ReadConfig reads the config file at path, applies annotation defaults,
// expands ${VAR} placeholders against the process environment, then
// unmarshals the bytes into conf, then validates. The order is:
//
//	ApplyDefaults  — struct-tag defaults on conf (set field zero values)
//	Expand         — replace ${VAR} / ${VAR:-x} / ${VAR:?m} in the raw bytes
//	                using os.LookupEnv (so .env values loaded by
//	                WithDotEnvConfig are visible; place that Init first)
//	unmarshal      — bytes → conf
//	Validate       — run annotation rules
//
// A placeholder referencing an unset variable with no :-default or :?error is
// an error, so a missing sensitive value fails startup instead of leaking the
// literal "${VAR}" into the decoded config.
func ReadConfig(path string, conf any, unmarshal UnmarshalFunc) error {
	bs, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := annotation.ApplyDefaults(conf); err != nil {
		return fmt.Errorf("apply defaults: %w", err)
	}
	if bs, err = Expand(bs, os.LookupEnv); err != nil {
		return fmt.Errorf("expand placeholders: %w", err)
	}
	if err := unmarshal(bs, conf); err != nil {
		return fmt.Errorf("unmarshal %s: %w", path, err)
	}
	if err := annotation.Validate(conf); err != nil {
		return fmt.Errorf("validate %s: %w", path, err)
	}
	return nil
}
