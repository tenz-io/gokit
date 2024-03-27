package logger

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	t.Run("test rotate log config", func(t *testing.T) {
		logcfg := Config{
			LoggingLevel:          InfoLevel,
			ConsoleLoggingEnabled: false,
			FileLoggingEnabled:    true,
			Directory:             "log",
			CallerEnabled:         true,
			CallerSkip:            1,
			Filename:              "",
			MaxSize:               100,
			MaxBackups:            10,
		}
		Configure(logcfg)

		Info("set up log success")
	})

	t.Run("test traffic log config", func(t *testing.T) {
		ConfigureTrafficLog(TrafficLogConfig{
			ConsoleLoggingEnabled: false,
			FileLoggingEnabled:    true,
			LoggingDirectory:      "log",
			Filename:              "data.log",
			MaxSize:               100,
			MaxBackups:            10,
		})
	})
	Data(&Traffic{
		Typ:  TrafficTypReq,
		Cmd:  "test command",
		Code: 200,
		Msg:  "test message",
		Cost: time.Second,
		Req:  "test request",
		Resp: "test response",
	})
}
