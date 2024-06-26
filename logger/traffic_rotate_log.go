package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultReqFieldName  = "_req"
	defaultRespFieldName = "_resp"
	defaultDataLevelName = "DATA"
	defaultFieldOccupied = "-"
)

var (
	// defaultTrafficLogger is the default dataLogger used by the package-level functions.
	defaultTrafficLogger TrafficEntry = &emptyTrafficEntry{}
)

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
		return defaultTrafficLogger
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
func StartTrafficRec(ctx context.Context, req *ReqEntity) *TrafficRec {
	return TrafficEntryFromContext(ctx).Start(req)
}

// CopyTrafficToContext copies the traffic logger from the current context to the new context
func CopyTrafficToContext(srcCtx context.Context, dstCtx context.Context) context.Context {
	if srcCtx == nil || dstCtx == nil {
		return dstCtx
	}
	dstCtx = WithTrafficEntry(dstCtx, TrafficEntryFromContext(srcCtx))
	return dstCtx
}

// ConfigureTrafficWithOpts sets up traffic logging with options globally
func ConfigureTrafficWithOpts(opts ...TrafficConfigOption) {
	config := defaultTrafficConfig
	for _, opt := range opts {
		opt(&config)
	}
	ConfigureTraffic(config)
}

// ConfigureTraffic sets up traffic logging globally
func ConfigureTraffic(config TrafficConfig) {
	SetupDefaultTrimmer(
		WithArrLimit(config.ArrLimit),
		WithStrLimit(config.StrLimit),
		WithDeepLimit(config.DeepLimit),
		WithIgnores(config.Ignores...),
	)

	defaultTrafficLogger = NewTrafficEntry(config)
}

// NewTrafficEntryWithOpts creates a new traffic entry with options
func NewTrafficEntryWithOpts(opts ...TrafficConfigOption) TrafficEntry {
	config := defaultTrafficConfig
	for _, opt := range opts {
		opt(&config)
	}
	return NewTrafficEntry(config)
}

func NewTrafficEntry(config TrafficConfig) TrafficEntry {
	var (
		writers []WriterSyncer
	)

	if config.Enabled {
		trafficRolling := newRollingFile(config.Directory, config.Filename,
			config.MaxSize, config.MaxAge, config.MaxBackups)
		writers = append(writers, trafficRolling)
	}

	if config.Stream != nil {
		writers = append(writers, config.Stream)
	}

	if len(writers) == 0 {
		return &emptyTrafficEntry{}
	}

	return newTrafficEntry(zapcore.NewMultiWriteSyncer(writers...))
}

func newTrafficEntry(logOutput zapcore.WriteSyncer) TrafficEntry {
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
