// Example: logger/v3 — levels, chained fields, context propagation and
// traffic logging.
package main

import (
	"context"
	"errors"
	"time"

	"github.com/tenz-io/gokit/logger/v3"
)

func main() {
	// 1. Configure the global logger: console + per-level files + traffic.
	logger.Configure(logger.Config{
		Level:    logger.InfoLevel,
		Console:  true,
		FilePath: "log",
		Traffic:  true,
	})

	// 2. Basic logging (structured key-value pairs via SugaredLogger).
	logger.Infow("server started", "port", 8080, "env", "production")
	logger.Debug("this won't appear at InfoLevel")
	logger.Warnw("low disk space", "disk", "/dev/sda1", "percent", 85)
	logger.Errorw("connection refused", "addr", "db.internal:5432")

	// 3. Create a child logger with persistent fields.
	userLog := logger.With("user_id", "usr_123", "session", "abc")
	userLog.Info("user logged in")
	userLog.Infow("user updated profile", "fields_changed", 3)

	// 4. With error.
	logger.WithError(errors.New("timeout")).Warn("retrying request")

	// 5. With request ID.
	reqLog := logger.WithRequestID("req_abc123")
	reqLog.Info("processing request")
	reqLog.Infof("request took %v", 42*time.Millisecond)

	// 6. Traffic logging: start a span, end it with the response and code.
	rec := logger.StartTraffic("getUser")
	if rec != nil {
		time.Sleep(10 * time.Millisecond)
		rec.End(map[string]any{"name": "bob"}, "200")
	}

	// 6b. A downstream call recorded as "send" direction.
	sendRec := logger.StartTraffic("callUserService").WithTyp(logger.TrafficTypSend)
	if sendRec != nil {
		time.Sleep(5 * time.Millisecond)
		sendRec.EndWithError(errors.New("upstream timeout"))
	}

	// 7. Context propagation: stash the per-request logger on ctx so every
	// downstream function recovers it with FromContext, carrying the
	// request_id field automatically.
	ctx := logger.WithLogger(context.Background(), reqLog)
	handle(ctx)
}

func handle(ctx context.Context) {
	logger.FromContext(ctx).Info("handling request")
	logger.FromContext(ctx).Warn("low balance", "balance", 3)
}
