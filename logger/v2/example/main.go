package main

import (
	"context"
	"errors"
	"time"

	"github.com/tenz-io/gokit/logger/v2"
)

func main() {
	// 1. Configure the global logger
	logger.Configure(logger.Config{
		Level:   logger.InfoLevel,
		Console: true,
		FilePath: "log",
	})

	// 2. Basic logging (structured key-value pairs via SugaredLogger)
	logger.Info("server started", "port", 8080, "env", "production")
	logger.Debug("this won't appear at InfoLevel")
	logger.Warn("low disk space", "disk", "/dev/sda1", "percent", 85)
	logger.Error("connection refused", "addr", "db.internal:5432")

	// 3. Create a child logger with persistent fields
	userLog := logger.With("user_id", "usr_123", "session", "abc")
	userLog.Info("user logged in")
	userLog.Info("user updated profile", "fields_changed", 3)

	// 4. With error
	logger.WithError(errors.New("timeout")).Warn("retrying request")

	// 5. With request ID
	reqLog := logger.WithRequestID("req_abc123")
	reqLog.Info("processing request")
	reqLog.Infof("request took %v", 42*time.Millisecond)

	// 6. Traffic logging
	rec := logger.StartTraffic("getUser")
	if rec != nil {
		time.Sleep(10 * time.Millisecond)
		rec.End(map[string]any{"name": "bob"}, "200")
	}

	// 7. Context propagation
	ctx := logger.WithLogger(context.Background(), userLog)
	logger.FromContext(ctx).Info("log from context")
}
