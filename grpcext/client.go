package grpcext

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"github.com/tenz-io/gokit/logger"
	"github.com/tenz-io/gokit/monitor"
)

func newTrafficUnaryClientInterceptor(clientInterceptor grpc.UnaryClientInterceptor, config Config) grpc.UnaryClientInterceptor {
	if !config.EnabledClientTraffic {
		return clientInterceptor
	}

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		var (
			grpcPeer = &peer.Peer{}
		)
		rec := logger.StartTrafficRec(ctx, &logger.ReqEntity{
			Typ: logger.TrafficTypSend,
			Cmd: method,
			Req: req,
		})

		defer func() {
			rec.End(&logger.RespEntity{
				Code: errCode(err),
				Msg:  errMsg(err),
				Resp: reply,
			}, logger.Fields{
				"peer": pettyPeer(grpcPeer),
			})
		}()

		moreOpts := append(opts, grpc.Peer(grpcPeer))
		return clientInterceptor(ctx, method, req, reply, cc, invoker, moreOpts...)
	}
}

func newMetricsUnaryClientInterceptor(clientInterceptor grpc.UnaryClientInterceptor, config Config) grpc.UnaryClientInterceptor {
	if !config.EnabledClientMetrics {
		return clientInterceptor
	}

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		rec := monitor.BeginRecord(ctx, method)

		defer func() {
			rec.EndWithError(err)
		}()

		return clientInterceptor(ctx, method, req, reply, cc, invoker, opts...)
	}
}
