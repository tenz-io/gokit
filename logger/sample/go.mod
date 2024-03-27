module logger-sample

go 1.20

require github.com/tenz-io/trackingo/logger v1.0.0

require (
	github.com/google/uuid v1.6.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)

replace github.com/tenz-io/trackingo/logger => ./..
