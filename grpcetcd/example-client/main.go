package main

import (
	"context"
	"fmt"
	"log"
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
)

func init() {
	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficEnabled(true),
	)
	logger.ConfigureWithOpts(
		logger.WithConsoleEnabled(true),
		logger.WithFileEnabled(false),
		logger.WithLoggerLevel(logger.DebugLevel),
		logger.WithSetAsDefaultLvl(true),
		logger.WithCallerEnabled(true),
		logger.WithCallerSkip(1),
	)
}

func main() {
	// Create a client connection
	conn, err := getConnClient()
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

func getConnClient() (*grpc.ClientConn, error) {
	interceptor := grpcext.NewInterceptorWithOpts(
		grpcext.WithClientMetrics(true),
		grpcext.WithClientTraffic(true),
	)

	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(interceptor.ApplyUnaryClientInterceptor(nil)),
		grpc.WithInsecure(),
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdAddr},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	le := logger.WithFields(logger.Fields{})
	discovery := grpcetcd.NewDiscovery(cli, path, le, opts...)

	// Create a client connection
	diaCtx, diaCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer diaCancel()

	conn, err := discovery.Dial(diaCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}

	return conn, nil
}
