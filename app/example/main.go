package main

import (
	"context"
	"fmt"
	"time"

	"github.com/tenz-io/gokit/app"
	"github.com/tenz-io/gokit/logger"
)

var flags = []app.Flag{
	&app.StringFlag{
		Name:  "env",
		Value: "test",
		Usage: "Environment",
	},
}

func main() {
	cfg := app.Config{
		Name:  "sample",
		Usage: "Sample Server",
		Conf:  &MyConfig{},
		Inits: []app.InitFunc{
			app.WithYamlConfig(),
			app.WithLogger(),
			app.WithAdminHTTPServer(),
		},
		Run: run,
	}
	app.Run(cfg, flags)
}

func run(c *app.Context, confPtr any, errC chan<- error) {
	logger.Infof("run application")

	env, err := c.GetFlags().String("env")
	if err != nil {
		logger.Warnf("failed to get env: %v", err)
	}

	logger.WithFields(logger.Fields{
		"env": env,
	}).Info("get env")

	mycnf, ok := confPtr.(*MyConfig)
	if !ok {
		errC <- fmt.Errorf("invalid config type: %T", confPtr)
	}

	logger.WithFields(logger.Fields{
		"config": mycnf,
	}).Debug("debug config")

	logTraffic(context.Background())
}

func logTraffic(ctx context.Context) {
	rec := logger.StartTrafficRec(ctx, &logger.ReqEntity{
		Typ: logger.TrafficTypRecv,
		Cmd: "/",
		Req: []byte("hello"),
		Fields: logger.Fields{
			"method": "GET",
		},
	})

	defer func() {
		rec.End(&logger.RespEntity{
			Code: "200",
			Msg:  "Success",
			Resp: []byte("world"),
		}, logger.Fields{
			"tag": "traffic",
		})
	}()

	logger.Infof("hello world")
	time.Sleep(200 * time.Millisecond)

}

type MyConfig struct {
	Foo string `yaml:"foo" json:"foo"`
}
