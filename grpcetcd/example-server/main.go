package main

import (
	"context"
	"log"
	"net"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/tenz-io/gokit/grpcetcd"
	"github.com/tenz-io/gokit/grpcext"
	pb "github.com/tenz-io/gokit/grpcext/proto"
	"github.com/tenz-io/gokit/logger"
)

const (
	etcdAddr = "http://127.0.0.1:2379"
	path     = "/services/echo"
	port     = ":50051"
	addr     = "127.0.0.0" + port
)

func init() {
	logger.ConfigureWithOpts(
		logger.WithConsoleEnabled(true),
		logger.WithFileEnabled(false),
		logger.WithLoggerLevel(logger.DebugLevel),
		logger.WithSetAsDefaultLvl(true),
		logger.WithCallerEnabled(true),
		logger.WithCallerSkip(1),
	)
	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficEnabled(true),
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

	go register(addr)

	// Register your services here
	pb.RegisterEchoServiceServer(srv, &server{})
	log.Printf("Starting gRPC server on %s", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func register(addr string) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdAddr},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	registry := grpcetcd.NewRegistry(cli, path, 15, logger.WithFields(logger.Fields{}))

	_, err = registry.Register(context.Background(), addr)
	if err != nil {
		return
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
