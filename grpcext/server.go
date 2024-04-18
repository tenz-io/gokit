package grpcext

import (
	"context"
	"log"

	"google.golang.org/grpc"
)

// UnaryServerInterceptor logs details of the unary RPC calls
func UnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// Log the incoming request
	log.Printf("Request - Method: %s, Request: %+v", info.FullMethod, req)

	// Handle the request
	resp, err := handler(ctx, req)

	// Log the response
	if err != nil {
		log.Printf("Response - Method: %s, Error: %v", info.FullMethod, err)
	} else {
		log.Printf("Response - Method: %s, Response: %+v", info.FullMethod, resp)
	}

	return resp, err
}

// StreamServerInterceptor logs the initiation of stream RPC calls
func StreamServerInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Log the stream initiation
	log.Printf("Stream Call - Method: %s, IsClientStream: %v, IsServerStream: %v", info.FullMethod, info.IsClientStream, info.IsServerStream)

	// Handle the stream
	err := handler(srv, ss)
	if err != nil {
		log.Printf("Stream Error - Method: %s, Error: %v", info.FullMethod, err)
	}

	return err
}
