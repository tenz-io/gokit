package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/tenz-io/gokit/grpcext"
	pb "github.com/tenz-io/gokit/grpcext/proto"
	"github.com/tenz-io/gokit/logger"
)

func init() {
	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficEnabled(true),
	)
	logger.ConfigureWithOpts(
		logger.WithConsoleEnabled(true),
		logger.WithFileEnabled(false),
		logger.WithCallerEnabled(true),
		logger.WithCallerSkip(1),
	)
}

func main() {
	interceptor := grpcext.NewInterceptorWithOpts(
		grpcext.WithServerTraffic(true),
		grpcext.WithServerMetrics(true),
	)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.ApplyUnaryServerInterceptor(nil)),
		grpc.StreamInterceptor(interceptor.ApplyStreamServerInterceptor(nil)),
	)

	// Register your services here
	pb.RegisterEchoServiceServer(srv, &server{})
	addr := ":50051"
	log.Printf("Starting gRPC server on %s", addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type server struct {
	pb.UnimplementedEchoServiceServer
}

func (s *server) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{
		Msg: "welcome " + req.GetMsg(),
	}, nil
}
