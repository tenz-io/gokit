package main

import (
	"context"
	"github.com/tenz-io/trackingo/logger"
	"time"
)

func main() {
	// log
	logger.ConfigureWithOpts(
		logger.WithConsoleEnabled(true),
		logger.WithLoggerLevel(logger.DebugLevel),
	)

	logger.Debugf("set up log success")

	// log with context
	ctx := context.Background()
	le := logger.FromContext(ctx).WithTracing("abc:123:456")
	le.WithFields(logger.Fields{
		"tag": "test",
	}).Debugf("set up log with context success")

	/////////////////////////
	// traffic log
	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficConsoleEnabled(true),
	)

	logger.DataWith(&logger.Traffic{
		Typ:  logger.TrafficTypRecv,
		Cmd:  "test_command",
		Cost: 30 * time.Millisecond,
		Code: "200",
		Msg:  "test message",
		Req:  "test request",
		Resp: "test response",
	}, logger.Fields{
		"tag": "test",
	})

	ctx = context.Background()
	te := logger.TrafficEntryFromContext(ctx).WithTracing("abc:123:456")
	te.DataWith(&logger.Traffic{
		Typ:  logger.TrafficTypRecv,
		Cmd:  "test_command",
		Cost: 40 * time.Millisecond,
		Code: "200",
		Msg:  "test message",
		Req:  "test request",
		Resp: "test response",
	}, logger.Fields{
		"tag": "test with context",
	})

	time.Sleep(100 * time.Millisecond)

}
