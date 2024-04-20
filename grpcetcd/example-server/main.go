package main

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/tenz-io/gokit/grpcetcd"
	"github.com/tenz-io/gokit/logger"
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
}

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	le := logger.WithFields(logger.Fields{})
	registry := grpcetcd.NewRegistry(cli, "/services/rank", 15, le)

	revoke, err := registry.Register(context.Background(), "localhost:50051", "{}")
	if err != nil {
		return
	}
	defer revoke(context.Background())

	// wait interrupt signal
	select {}

}
