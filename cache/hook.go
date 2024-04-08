package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"

	"github.com/tenz-io/gokit/logger"
	"github.com/tenz-io/gokit/monitor"
)

var (
	newHooks = []func(conf Config) redis.Hook{
		newMetricsHook,
		newTrafficHook,
	}
)

type hookCtxKey string

const (
	metricsHookKey hookCtxKey = "_metrics_hook_ctx_key"
	trafficHookKey hookCtxKey = "_traffic_hook_ctx_key"
)

type metricsHook struct {
	enable bool
}

func newMetricsHook(conf Config) redis.Hook {
	return &metricsHook{
		enable: conf.EnableMetrics,
	}
}

func (mh *metricsHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	if !mh.enable {
		return ctx, nil
	}

	rec := monitor.BeginRecord(ctx, cmd.Name())
	ctx = context.WithValue(ctx, metricsHookKey, rec)
	return ctx, nil
}

func (mh *metricsHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if !mh.enable {
		return nil
	}

	if rec, ok := ctx.Value(metricsHookKey).(*monitor.Recorder); ok {
		var (
			err = cmd.Err()
			opt string
		)

		// if cmd is non-get command
		if !strings.EqualFold(cmd.Name(), "get") {
			rec.EndWithError(err)
			return nil
		}

		// get command
		if err != nil {
			// if error is redis.Nil, it means key not found, ignore it
			if errors.Is(err, redis.Nil) {
				opt = "miss"
			} else {
				opt = "error"
			}
		} else {
			opt = "hit"
		}

		rec.EndWithErrorOpt(err, opt)
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

type trafficHook struct {
	enable bool
}

func newTrafficHook(conf Config) redis.Hook {
	return &trafficHook{
		enable: conf.EnableTraffic,
	}
}

func (th *trafficHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	if !th.enable {
		return ctx, nil
	}

	rec := logger.StartTrafficRec(ctx, &logger.ReqEntity{
		Typ:    logger.TrafficTypSend,
		Cmd:    cmd.Name(),
		Req:    cmd.String(),
		Fields: logger.Fields{},
	})

	ctx = context.WithValue(ctx, trafficHookKey, rec)
	return ctx, nil
}

func (th *trafficHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if !th.enable {
		return nil
	}

	if rec, ok := ctx.Value(trafficHookKey).(*logger.TrafficRec); ok {
		// dump response entity, how?
		rec.End(&logger.RespEntity{
			Code: fmt.Sprintf("%d", errCode(cmd.Err())),
			Msg:  errorMsg(cmd.Err()),
			Resp: func() string {
				if cmd.Err() != nil {
					return fmt.Sprintf("err: %v", cmd.Err())
				}
				return cmd.String()
			}(),
		}, logger.Fields{})
	}

	return nil

}

func (th *trafficHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	// not support pipeline, skip
	return ctx, nil
}

func (th *trafficHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	// not support pipeline, skip
	return nil
}
