package cache

import (
	"context"

	"github.com/go-redis/redis/v8"

	"github.com/tenz-io/gokit/monitor"
)

type hookCtxKey string

const (
	metricsHookKey hookCtxKey = "_metrics_hook_ctx_key"
)

type metricsHook struct {
}

func (mh *metricsHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	var (
		name = cmd.Name()
	)
	rec := monitor.BeginRecord(ctx, name)
	ctx = context.WithValue(ctx, metricsHookKey, rec)
	return ctx, nil
}

func (mh *metricsHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	var (
		err = cmd.Err()
	)
	if rec, ok := ctx.Value(metricsHookKey).(*monitor.Recorder); ok {
		rec.EndWithError(err)
	}

	return nil
}

func (mh *metricsHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	// not support pipeline, skip
	return ctx, nil
}

func (mh *metricsHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	// not support pipeline, skip
	return nil
}
