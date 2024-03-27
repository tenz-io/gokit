package logger

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

const (
	defaultReqFieldName  = "request"
	defaultRespFieldName = "response"
	defaultPairFieldName = "__pair_id"
	defaultDataLevelName = "DATA"
	defaultFieldOccupied = "-"
)

var (
	// defaultTrafficLogConfig is used for defaultTrafficLogger below only
	defaultTrafficLogConfig = TrafficLogConfig{}

	// defaultTrafficLogger is the default dataLogger instance that should be used to log
	// It's assigned a default value here for tests (which do not call log.ConfigureTrafficLog())
	defaultTrafficLogger = newTrafficLogger(defaultTrafficLogConfig, os.Stdout)
)

// TrafficLogConfig for traffic logging
type TrafficLogConfig struct {

	// FileLoggingEnabled makes the framework log to a file
	// the fields below can be skipped if this value is false!
	FileLoggingEnabled bool
	// ConsoleLoggingEnabled makes the framework log to console
	ConsoleLoggingEnabled bool
	// LoggingDirectory to log to to when filelogging is enabled
	LoggingDirectory string
	// Filename is the name of the logfile which will be placed inside the directory
	Filename string
	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int
	// MaxBackups the max number of rolled files to keep
	MaxBackups int
	// MaxAge the max age in days to keep a logfile
	MaxAge int
	// ConsoleStream
	ConsoleStream *os.File
}

// Data Log a request
func Data(tc *Traffic) {
	DataWith(tc, nil)
}

// DataWith Log a request with fields
func DataWith(tc *Traffic, fields Fields) {
	defaultTrafficLogger.DataWith(tc, fields)
}

func WithTrafficFields(ctx context.Context, fields Fields) TrafficEntry {
	return TrafficEntryFromContext(ctx).WithFields(fields)
}

func WithTrafficTracing(ctx context.Context, requestId string) TrafficEntry {
	return TrafficEntryFromContext(ctx).WithTracing(requestId)
}

func WithTrafficIgnores(ctx context.Context, ignores ...string) TrafficEntry {
	return TrafficEntryFromContext(ctx).WithIgnores(ignores...)
}

// TrafficEntryFromContext get traffic dataLogger from context, allows us to pass dataLogger between functions
func TrafficEntryFromContext(ctx context.Context) TrafficEntry {
	data := ctx.Value(trafficLogCtxKey)
	if data == nil {
		return defaultTrafficLogger.clone() // prevent the user from accidentally not setting the dataLogger
	}
	te, ok := data.(*LogTrafficEntry)
	if !ok {
		return &emptyTrafficEntry{}
	}
	return te
}

// WithTrafficEntry set given LogTrafficEntry to context by using trafficLogCtxKey
func WithTrafficEntry(ctx context.Context, te TrafficEntry) context.Context {
	if ctx == nil || te == nil {
		return ctx
	}
	return context.WithValue(ctx, trafficLogCtxKey, te)
}

// StartTrafficRec starts a new traffic log entry
func StartTrafficRec(ctx context.Context, req *TrafficReq, fields Fields) *TrafficRec {
	return TrafficEntryFromContext(ctx).Start(req, fields)
}

// CopyTrafficToContext copies the traffic logger from the current context to the new context
func CopyTrafficToContext(srcCtx context.Context, dstCtx context.Context) context.Context {
	if srcCtx == nil || dstCtx == nil {
		return dstCtx
	}
	dstCtx = WithTrafficEntry(dstCtx, TrafficEntryFromContext(srcCtx))
	return dstCtx
}

// ConfigureTrafficLog sets up traffic logging
func ConfigureTrafficLog(config TrafficLogConfig) {
	var writers []zapcore.WriteSyncer

	if config.FileLoggingEnabled {
		trafficLog := newRollingFile(config.LoggingDirectory, config.Filename, config.MaxSize, config.MaxAge, config.MaxBackups)
		writers = append(writers, trafficLog)
	} else {
		config.ConsoleLoggingEnabled = true
	}

	if config.ConsoleLoggingEnabled {
		if config.ConsoleStream != nil {
			writers = append(writers, config.ConsoleStream)
		} else {
			writers = append(writers, os.Stdout)
		}
	}

	defaultTrafficLogger = newTrafficLogger(config, zapcore.NewMultiWriteSyncer(writers...))
}

func newTrafficLogger(config TrafficLogConfig, logOutput zapcore.WriteSyncer) *LogTrafficEntry {
	encCfg := zapcore.EncoderConfig{
		TimeKey:          "@t",
		MessageKey:       "msg",
		ConsoleSeparator: defaultSeparator,
		EncodeTime:       longTimeEncoder,
		EncodeDuration:   zapcore.NanosDurationEncoder,
	}
	encoder := zapcore.NewConsoleEncoder(encCfg)

	trafficEntry := &LogTrafficEntry{
		dataLogger: zap.New(zapcore.NewCore(encoder, logOutput, zapcore.Level(InfoLevel))),
		sep:        defaultSeparator,
		allow:      true, // default allow log print
	}

	return trafficEntry
}
