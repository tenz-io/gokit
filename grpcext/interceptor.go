package grpcext

import (
	"context"

	"google.golang.org/grpc"

	"github.com/tenz-io/gokit/logger"
	"github.com/tenz-io/gokit/monitor"
)

type Interceptor interface {
	UnaryServerInterceptor() grpc.UnaryServerInterceptor
	StreamServerInterceptor() grpc.StreamServerInterceptor
	UnaryClientInterceptor() grpc.UnaryClientInterceptor
	StreamClientInterceptor() grpc.StreamClientInterceptor
}

type interceptor struct {
	unaryServerInterceptor  grpc.UnaryServerInterceptor
	streamServerInterceptor grpc.StreamServerInterceptor
	unaryClientInterceptor  grpc.UnaryClientInterceptor
	streamClientInterceptor grpc.StreamClientInterceptor
}

type trafficUnaryServerInterceptor struct {
	unaryServerInterceptor grpc.UnaryServerInterceptor
	enabled                bool
}

func newTrafficUnaryServerInterceptor(unaryServerInterceptor grpc.UnaryServerInterceptor, config Config) grpc.UnaryServerInterceptor {
	if unaryServerInterceptor == nil {
		unaryServerInterceptor = func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	}

	if !config.EnabledServerTraffic {
		return unaryServerInterceptor
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {

		ctx = logger.WithTrafficEntry(
			ctx,
			logger.WithTrafficTracing(ctx, reqID).
				WithFields(logger.Fields{
					"url": url,
				}).WithIgnores(
				"password",
				//"Authorization",
			),
		)

		rec := logger.StartTrafficRec(ctx, &logger.ReqEntity{
			Typ: logger.TrafficTypRecv,
			Cmd: info.FullMethod,
			Req: req,
		})

		defer func() {
			rec.End(&logger.RespEntity{
				Code: "0",
				Msg:  err.Error(),
				Resp: resp,
			}, logger.Fields{})
		}()

		return unaryServerInterceptor(ctx, req, info, handler)

	}
}

func newMetricsUnaryServerInterceptor(unaryServerInterceptor grpc.UnaryServerInterceptor, config Config) grpc.UnaryServerInterceptor {
	if unaryServerInterceptor == nil {
		unaryServerInterceptor = func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	}

	if !config.EnabledServerMetrics {
		return unaryServerInterceptor
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		var (
			rec = monitor.BeginRecord(ctx, info.FullMethod)
		)

		defer func() {
			rec.EndWithError(err)
		}()

		return unaryServerInterceptor(ctx, req, info, handler)
	}
}
