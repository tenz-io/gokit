// Example: app/v3 — flag parsing, ordered init, admin server and graceful
// shutdown, with logger/v3 traffic logging and tracer/v3 request IDs.
package main

import (
	"context"
	"os"
	"time"

	"github.com/tenz-io/gokit/app/v3"
	"github.com/tenz-io/gokit/logger/v3"
	"github.com/tenz-io/gokit/tracer/v3"
)

// extraFlags shows how an app extends the built-in flag set without touching it.
var extraFlags = []app.FlagSpec{
	app.StringFlag("env", "test", "Environment"),
}

func main() {
	cfg := app.Config{
		Name:  "sample",
		Usage: "Sample Server",
		Conf:  &MyConfig{},
		Inits: []app.InitFunc{
			app.WithDotEnvConfig(),
			app.WithYamlConfig(),
			app.WithTraffic(), // force-enable traffic logging (equivalent to -traffic flag)
			app.WithAdminHTTPServer(),
		},
		Run: app.AdaptRun(run),
	}
	os.Exit(int(app.Run(cfg, app.WithExtraFlags(extraFlags...))))
}

// run is the main service loop. AdaptRun handles the *MyConfig assertion and
// the errC plumbing: returning nil triggers clean shutdown, returning an error
// is treated as fatal. No boilerplate here.
func run(c *app.Context, conf *MyConfig) error {
	env := c.Flags().String("env")

	// Pin a request id on the context for the whole request; the logger
	// attached below carries it onto every log line.
	ctx, reqID := tracer.EnsureRequestID(c.Context)
	reqLog := logger.WithRequestID(reqID)
	ctx = logger.WithLogger(ctx, reqLog)

	logger.FromContext(ctx).Infow("service started",
		"name", conf.Name, "env", env,
		"db_password_set", conf.DBPassword != "", // value itself isn't logged
	)

	logTraffic(ctx)

	// Block until shutdown (signal or context cancel).
	<-c.Done()
	logger.FromContext(ctx).Info("service stopped")
	return nil
}

func logTraffic(ctx context.Context) {
	// A recv-direction traffic span: a fake inbound request.
	rec := logger.FromContext(ctx).StartTraffic("ping")
	if rec != nil {
		defer rec.End(map[string]any{"ok": true}, "200", "env", "example")
		time.Sleep(50 * time.Millisecond)
	}
}

type MyConfig struct {
	Name       string `yaml:"name" json:"name" default:"sample"`
	PageSize   int    `yaml:"page_size" json:"page_size" default:"10" validate:"required,gt=0,lte=100"`
	DBPassword string `yaml:"db_password" json:"db_password"` // from ${DB_PASSWORD} env, never committed
}
