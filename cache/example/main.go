package main

import (
	"context"
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

	interceptor := cache.NewInterceptorWithOpts(
		cache.WithEnableMetrics(true),
		cache.WithEnableTraffic(true),
	)

	interceptor.Apply(cli)
	cacheMgr := cache.NewRedis(cli)
	key := "some_key_xxx"
	raw, err := cacheMgr.Get(context.Background(), key)
	logger.WithFields(logger.Fields{
		"key": key,
		"raw": raw,
		"err": err,
	}).Infof("1st time get from cache")

	err = cacheMgr.Set(context.Background(), key, "new_value", 0)
	if err != nil {
		logger.WithFields(logger.Fields{
			"key": key,
			"err": err,
		}).Warnf("set cache failed")
	}

	raw, err = cacheMgr.Get(context.Background(), key)
	logger.WithFields(logger.Fields{
		"key": key,
		"raw": raw,
		"err": err,
	}).Infof("2nd time get from cache")

	/////////////// test blob ///////////////

	type info struct {
		Name string
		Age  int
	}

	key2 := "some_key_yyy"
	val := info{
		Name: "tenz",
		Age:  18,
	}
	err = cacheMgr.GetBlob(context.Background(), key2, val)
	logger.WithFields(logger.Fields{
		"key": key2,
		"err": err,
	}).Infof("1st time get blob")

	err = cacheMgr.SetBlob(context.Background(), key2, val, 5*time.Second)
	logger.WithFields(logger.Fields{
		"key": key2,
		"err": err,
	}).Infof("set blob")

	newVal := info{}
	err = cacheMgr.GetBlob(context.Background(), key2, &newVal)
	logger.WithFields(logger.Fields{
		"key": key2,
		"val": newVal,
		"err": err,
	}).Infof("2nd time get blob")

	time.Sleep(100 * time.Millisecond)

}
