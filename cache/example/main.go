package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/tenz-io/gokit/cache"
	"github.com/tenz-io/gokit/logger"
)

func init() {
	logger.ConfigureWithOpts(
		logger.WithLoggerLevel(logger.DebugLevel),
		logger.WithConsoleEnabled(true),
		logger.WithCallerEnabled(true),
		logger.WithCallerSkip(1),
	)

	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficEnabled(true),
	)
}

func main() {
	cli := redis.NewClient(
		&redis.Options{
			Addr: "localhost:6379",
		},
	)

	cli.Conn(context.Background()).Set(context.Background(), "some_key", "some_value", 0)

	interceptor := cache.NewInterceptorWithOpts(
		cache.WithEnableMetrics(true),
		cache.WithEnableTraffic(true),
	)

	interceptor.Apply(cli)

	cli.Set(context.Background(), "some_key", "some_value", 15*time.Second)

	cmd := cli.Get(context.Background(), "some_key")
	log.Println(cmd.Result())
	_ = cli.Get(context.Background(), "some_key_not_exist")

	time.Sleep(100 * time.Millisecond)

}
