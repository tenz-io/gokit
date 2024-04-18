package main

import (
	"log"

	"google.golang.org/grpc"

	"github.com/tenz-io/gokit/grpcext"
)

func main() {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpcext.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpcext.StreamServerInterceptor),
	)

	// Register your services here
	log.Printf("Starting gRPC server, %+v", server)

}
