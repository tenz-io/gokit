package main

import (
	"context"
	"log"
	"time"

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
		grpcext.WithClientMetrics(true),
		grpcext.WithClientTraffic(true),
	)

	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(interceptor.ApplyUnaryClientInterceptor(nil)),
		grpc.WithStreamInterceptor(interceptor.ApplyStreamClientInterceptor(nil)),
		grpc.WithInsecure(),
	}

	// Create a client connection
	conn, err := grpc.Dial("localhost:50051", opts...)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	resp, err := pb.NewEchoServiceClient(conn).Echo(context.Background(), &pb.EchoRequest{Msg: "Alice"})
	if err != nil {
		log.Fatalf("Failed to call Echo: %v", err)
		return
	}
	log.Printf("Response: %s", resp.GetMsg())
	time.Sleep(100 * time.Millisecond)

}
