package logger

import (
	"os"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	defer func() {
		time.Sleep(100 * time.Millisecond)
	}()

	t.Run("test rotate log config", func(t *testing.T) {
		le := NewEntryWithOpts(
			WithLoggerLevel(DebugLevel),
			WithConsoleEnabled(true),
		)

		le.Debug("debug message")
		le.Info("set up log success")
		le.Warn("warning message")
		le.Error("error message")
	})

	t.Run("test traffic log config", func(t *testing.T) {
		ConfigureTrafficWithOpts(
			WithTrafficEnabled(false),
			WithTrafficStream(os.Stdout),
		)

		Data(&Traffic{
			Typ:  TrafficTypRecv,
			Cmd:  "test command",
			Code: "200",
			Msg:  "test message",
			Cost: 53 * time.Second,
			Req:  "test request",
			Resp: "test response",
		})
		Data(&Traffic{
			Typ:  TrafficTypRecv,
			Cmd:  "test command2",
			Code: "212",
			Msg:  "test message2",
			Cost: 23 * time.Second,
			Req:  "test request2",
			Resp: "test response2",
		})
	})

}
