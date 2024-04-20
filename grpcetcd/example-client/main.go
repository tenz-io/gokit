package main

import (
	"context"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/tenz-io/gokit/grpcetcd"
)

func main() {
	cli, err := clientv3.NewFromURL("http://localhost:2379")
	if err != nil {
		panic(err)
	}
	registry := grpcetcd.NewDiscovery(cli, "/services/rank")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cc, err := registry.Dial(ctx)
	if err != nil {
		panic(err)
	}

	// use cc to call grpc service
	// ...
	log.Printf("grpc client connected, cc: %+v", cc)
}
