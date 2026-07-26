package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tenz-io/gokit/logger/v3"
)

// wait blocks until one of: a signal is received, the app context is done, or
// the Run goroutine reports on errC. hook runs the graceful cleanup once and
// is invoked before returning in every branch. The returned ExitCode reflects
// the cause: ExitRunError on a non-nil errC, ExitOK on a nil errC (clean
// completion), ExitSignal on an interrupt.
//
// Unlike v2's WaitSignal, this never calls os.Exit — Run returns the code and
// the caller decides.
func wait(ctx context.Context, errC <-chan error, hook func()) ExitCode {
	// SIGINT (Ctrl-C) and SIGTERM (container/k8s shutdown). os.Kill is not
	// catchable, so it is deliberately omitted.
	signC := make(chan os.Signal, 1)
	signal.Notify(signC, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signC)

	for {
		select {
		case <-signC:
			logger.Infow("received interrupt signal")
			hook()
			return ExitSignal
		case <-ctx.Done():
			logger.Infow("context done")
			hook()
			return ExitOK
		case err := <-errC:
			if err != nil {
				logger.Errorf("run error: %+v", err)
				hook()
				return ExitRunError
			}
			logger.Infow("run completed")
			hook()
			return ExitOK
		}
	}
}

// PrettyString renders v as a compact JSON string, falling back to %+v. Useful
// for printing decoded config in verbose mode.
func PrettyString(v any) string {
	if v == nil {
		return "nil"
	}
	if j, err := json.Marshal(v); err == nil {
		return string(j)
	}
	return fmt.Sprintf("%+v", v)
}
