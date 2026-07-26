module logger-v3-example

go 1.24

require github.com/tenz-io/gokit/logger/v3 v3.0.0

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)

replace github.com/tenz-io/gokit/logger/v3 => ./..
