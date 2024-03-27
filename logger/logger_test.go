package logger

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {

	t.Run("test rotate log config", func(t *testing.T) {
		le := NewEntryWithOpts(
			WithLoggerLevel(DebugLevel),
			WithConsoleEnabled(true),
		)

		le.Info("set up log success")
	})

	t.Run("test traffic log config", func(t *testing.T) {
		ConfigureTraffic(TrafficConfig{
			ConsoleEnabled: false,
			FileEnabled:    true,
			Directory:      "log",
			Filename:       "data.log",
			MaxSize:        100,
			MaxBackups:     10,
		})
	})
	Data(&Traffic{
		Typ:  TrafficTypRecv,
		Cmd:  "test command",
		Code: "200",
		Msg:  "test message",
		Cost: time.Second,
		Req:  "test request",
		Resp: "test response",
	})

}
