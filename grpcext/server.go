package grpcext

import (
	"context"

	"google.golang.org/grpc"

	"github.com/tenz-io/gokit/logger"
	"github.com/tenz-io/gokit/monitor"
	"github.com/tenz-io/gokit/tracer"
)

func newTrackingUnaryServerInterceptor(serverInterceptor grpc.UnaryServerInterceptor, _ Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		var (
			reqID = tracer.RequestIdFromCtx(ctx)
			cmd   = info.FullMethod
		)
		ctx = tracer.WithRequestId(ctx, reqID)
		ctx = logger.WithLogger(ctx,
			logger.WithTracing(reqID).
				WithFields(logger.Fields{
					"cmd": cmd,
				}))

		return serverInterceptor(ctx, req, info, handler)
	}

}

func newTrafficUnaryServerInterceptor(serverInterceptor grpc.UnaryServerInterceptor, config Config) grpc.UnaryServerInterceptor {
	if !config.EnabledServerTraffic {
		return serverInterceptor
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		var (
			cmd   = info.FullMethod
			reqID = tracer.RequestIdFromCtx(ctx)
		)

		ctx = logger.WithTrafficEntry(
			ctx,
			logger.WithTrafficTracing(ctx, reqID).
				WithFields(logger.Fields{
					"cmd": cmd,
				}).WithIgnores(
				"password",
			),
		)

		rec := logger.StartTrafficRec(ctx, &logger.ReqEntity{
			Typ: logger.TrafficTypRecv,
			Cmd: cmd,
			Req: req,
		})

		defer func() {
			rec.EndWithIgnores(&logger.RespEntity{
				Code: errCode(err),
				Msg:  errMsg(err),
				Resp: resp,
			}, logger.Fields{},
				"state", "sizeCache", "unknownFields",
			)
		}()

		return serverInterceptor(ctx, req, info, handler)

	}
}

func newMetricsUnaryServerInterceptor(serverInterceptor grpc.UnaryServerInterceptor, config Config) grpc.UnaryServerInterceptor {
	if !config.EnabledServerMetrics {
		return serverInterceptor
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = monitor.InitSingleFlight(ctx, info.FullMethod)
		rec := monitor.BeginRecord(ctx, "total")

		defer func() {
			rec.EndWithError(err)
		}()

		return serverInterceptor(ctx, req, info, handler)
	}
}
