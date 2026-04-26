package grpcext

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var (
	newUnaryServerInterceptors = []newUnaryServerInterceptor{
		newTrackingUnaryServerInterceptor,
		newTrafficUnaryServerInterceptor,
		newMetricsUnaryServerInterceptor,
	}
	newUnaryClientInterceptors = []newUnaryClientInterceptor{
		newTrafficUnaryClientInterceptor,
		newMetricsUnaryClientInterceptor,
	}
)

type (
	newUnaryServerInterceptor  func(grpc.UnaryServerInterceptor, Config) grpc.UnaryServerInterceptor
	newStreamServerInterceptor func(grpc.StreamServerInterceptor, Config) grpc.StreamServerInterceptor
	newUnaryClientInterceptor  func(grpc.UnaryClientInterceptor, Config) grpc.UnaryClientInterceptor
	newStreamClientInterceptor func(grpc.StreamClientInterceptor, Config) grpc.StreamClientInterceptor
)

type Interceptor interface {
	ApplyUnaryServerInterceptor(grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor
	ApplyStreamServerInterceptor(grpc.StreamServerInterceptor) grpc.StreamServerInterceptor
	ApplyUnaryClientInterceptor(grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor
	ApplyStreamClientInterceptor(grpc.StreamClientInterceptor) grpc.StreamClientInterceptor
}

func NewInterceptor(config Config) Interceptor {
	return &interceptor{
		config: config,
	}
}

func NewInterceptorWithOpts(opts ...Option) Interceptor {
	config := defaultConfig
	for _, opt := range opts {
		opt(&config)
	}
	return NewInterceptor(config)

}

type interceptor struct {
	config Config
}

func (i *interceptor) ApplyUnaryServerInterceptor(serverInterceptor grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	if serverInterceptor == nil {
		serverInterceptor = func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	}

	for _, newServerInterceptor := range newUnaryServerInterceptors {
		serverInterceptor = newServerInterceptor(serverInterceptor, i.config)
	}

	return serverInterceptor
}

func (i *interceptor) ApplyStreamServerInterceptor(serverInterceptor grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	if serverInterceptor == nil {
		serverInterceptor = func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return handler(srv, stream)
		}
	}

	return serverInterceptor
}

func (i *interceptor) ApplyUnaryClientInterceptor(clientInterceptor grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor {
	if clientInterceptor == nil {
		clientInterceptor = func(ctx context.Context, method string, req, reply any,
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
	}

	for _, newClientInterceptor := range newUnaryClientInterceptors {
		clientInterceptor = newClientInterceptor(clientInterceptor, i.config)
	}
	return clientInterceptor
}

func (i *interceptor) ApplyStreamClientInterceptor(clientInterceptor grpc.StreamClientInterceptor) grpc.StreamClientInterceptor {
	if clientInterceptor == nil {
		clientInterceptor = func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
			streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return streamer(ctx, desc, cc, method, opts...)
		}
	}
	return clientInterceptor
}

func errCode(err error) string {
	if err == nil {
		return "0"
	}

	return "1"
}

func errMsg(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

func pettyPeer(peer *peer.Peer) string {
	if peer == nil || peer.Addr == nil {
		return ""
	}
	return fmt.Sprintf("%s", peer.Addr)
}
