package logger

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type loggerCtxKeyType string

const (
	logCtxKey        = loggerCtxKeyType("_log_ctx_key")
	trafficLogCtxKey = loggerCtxKeyType("_traffic_log_ctx_key")
)

var (
	loglv        zap.AtomicLevel
	defaultLevel = InfoLevel // default log level
)

var (
	// defaultLogger is the default logger
	defaultLogger = newEntry(defaultConfig, os.Stdout, os.Stderr, os.Stdout)
)

// Debug Log a message at the debug defaultLevel
func Debug(msg string) {
	if !Enabled(DebugLevel) {
		return
	}
	msg = withHead(msg)
	defaultLogger.infoLogger.Debug(msg)
}

func Debugf(format string, args ...any) {
	if !Enabled(DebugLevel) {
		return
	}
	msg := withHead(fmt.Sprintf(format, args...))
	defaultLogger.debugLogger.Debug(msg)
}

// DebugWith Log a message with fields at the debug defaultLevel
func DebugWith(msg string, fields Fields) {
	if !Enabled(DebugLevel) {
		return
	}
	msg = withHead(msg)
	if len(fields) > 0 {
		defaultLogger.infoLogger.Debug(msg, toZapFields(fields)...)
	} else {
		defaultLogger.infoLogger.Debug(msg)
	}
}

// Info Log a message at the info defaultLevel
func Info(msg string) {
	if !Enabled(InfoLevel) {
		return
	}
	msg = withHead(msg)
	defaultLogger.infoLogger.Info(msg)
}

func Infof(format string, args ...any) {
	if !Enabled(InfoLevel) {
		return
	}
	msg := withHead(fmt.Sprintf(format, args...))
	defaultLogger.infoLogger.Info(msg)
}

// InfoWith Log a message with fields at the info defaultLevel
func InfoWith(msg string, fields Fields) {
	if !Enabled(InfoLevel) {
		return
	}
	msg = withHead(msg)
	if len(fields) > 0 {
		defaultLogger.infoLogger.Info(msg, toZapFields(fields)...)
	} else {
		defaultLogger.infoLogger.Info(msg)
	}
}

// Warn Log a message at the warn defaultLevel
func Warn(msg string) {
	if !Enabled(WarnLevel) {
		return
	}
	msg = withHead(msg)
	defaultLogger.errLogger.Warn(msg)
}

func Warnf(format string, args ...any) {
	if !Enabled(WarnLevel) {
		return
	}
	msg := withHead(fmt.Sprintf(format, args...))
	defaultLogger.errLogger.Warn(msg)
}

// WarnWith Log a message with fields at the warn defaultLevel
func WarnWith(msg string, fields Fields) {
	if !Enabled(WarnLevel) {
		return
	}
	msg = withHead(msg)
	if len(fields) > 0 {
		defaultLogger.errLogger.Warn(msg, toZapFields(fields)...)
	} else {
		defaultLogger.errLogger.Warn(msg)
	}
}

// Error Log a message at the error defaultLevel
func Error(msg string) {
	if !Enabled(ErrorLevel) {
		return
	}
	msg = withHead(msg)
	defaultLogger.errLogger.Error(msg)
}

func Errorf(format string, args ...any) {
	if !Enabled(ErrorLevel) {
		return
	}
	msg := withHead(fmt.Sprintf(format, args...))
	defaultLogger.errLogger.Error(msg)
}

// ErrorWith Log a message with fields at the error defaultLevel
func ErrorWith(msg string, fields Fields) {
	if !Enabled(ErrorLevel) {
		return
	}
	msg = withHead(msg)
	if len(fields) > 0 {
		defaultLogger.errLogger.Error(msg, toZapFields(fields)...)
	} else {
		defaultLogger.errLogger.Error(msg)
	}
}

// WithFields binds a set of fields to a log message
func WithFields(fields Fields) Entry {
	return newLogEntry(defaultLogger, fields)
}

// WithField binds a field to a log message
func WithField(k string, v any) Entry {
	return WithFields(Fields{k: v})
}

// With binds a default field to a log message
func With(data any) Entry {
	return WithField(defaultLogEmpty, data)
}

// WithError binds an error to a log message
func WithError(err error) Entry {
	return WithField(defaultErrFieldName, err)
}

// WithTracing create copy of LogEntry with tracing.Span
func WithTracing(requestId string) Entry {
	return defaultLogger.WithTracing(requestId)
}

func withHead(msg string) string {
	if defaultLogger == nil {
		return strings.Join(append([]string{
			defaultLogEmpty,
			msg,
		}), defaultSeparator)
	}
	if defaultLogger.requestId == "" {
		return strings.Join(append([]string{
			defaultLogEmpty,
			msg,
		}), defaultSeparator)
	}

	return strings.Join(append([]string{
		defaultLogger.requestId,
		msg,
	}), defaultSeparator)
}

func ConfigureWithOpts(opts ...ConfigOption) {
	config := defaultConfig
	for _, opt := range opts {
		opt(&config)
	}
	Configure(config)
}

// Configure sets up the defaultLogger
func Configure(config Config) {
	defaultLogger = NewEntry(config).(*LogEntry)
}

// NewEntryWithOpts create a new LogEntry with options
func NewEntryWithOpts(opts ...ConfigOption) Entry {
	config := defaultConfig
	for _, opt := range opts {
		opt(&config)
	}
	return NewEntry(config)
}

// NewEntry create a new LogEntry instead of override defaultzaplogger
func NewEntry(config Config) Entry {
	var (
		errWriters   []WriterSyncer
		infoWriters  []WriterSyncer
		debugWriters []WriterSyncer
	)

	if config.FileEnabled {
		errRolling := newRollingFile(config.Directory, getNameByLogLevel(config.Filename, ErrorLevel),
			config.MaxSize, config.MaxAge, config.MaxBackups)
		errWriters = append(errWriters, errRolling)

		infoRolling := newRollingFile(config.Directory, getNameByLogLevel(config.Filename, InfoLevel),
			config.MaxSize, config.MaxAge, config.MaxBackups)
		infoWriters = append(infoWriters, infoRolling)

		debugRolling := newRollingFile(config.Directory, getNameByLogLevel(config.Filename, DebugLevel),
			config.MaxSize, config.MaxAge, config.MaxBackups)
		debugWriters = append(debugWriters, debugRolling)
	} else {
		// if file logging is disabled, enable console logging
		config.ConsoleEnabled = true
	}

	if config.ConsoleEnabled {
		errWriters = append(errWriters, os.Stderr)
		infoWriters = append(infoWriters, os.Stdout)
		debugWriters = append(debugWriters, os.Stdout)
	}

	if config.Stream != nil {
		errWriters = append(errWriters, config.Stream)
		infoWriters = append(infoWriters, config.Stream)
		debugWriters = append(debugWriters, config.Stream)
	}

	logEntry := newEntry(
		config,
		zapcore.NewMultiWriteSyncer(infoWriters...),
		zapcore.NewMultiWriteSyncer(errWriters...),
		zapcore.NewMultiWriteSyncer(debugWriters...),
	)

	declareLogger(config, logEntry.InfoWith)
	declareLogger(config, logEntry.ErrorWith)
	declareLogger(config, logEntry.DebugWith)
	return logEntry
}

func declareLogger(config Config, logv func(msg string, fields Fields)) {
	logv("logger configured", Fields{"config": config})
}

func SetLevel(l Level) {
	if !l.validate() {
		return
	}
	loglv.SetLevel(zapcore.Level(l))
	defaultLevel = l
}

func GetLevel() Level {
	return defaultLevel
}

func Enabled(level Level) bool {
	return defaultLogger.Enabled(level)
}

func newRollingFile(dir, filename string, maxSize, maxAge, maxBackups int) zapcore.WriteSyncer {
	if err := os.MkdirAll(dir, 0744); err != nil {
		WithFields(Fields{
			"error": err,
			"path":  dir,
		}).Error("failed create log directory")
		return nil
	}

	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   path.Join(dir, filename),
		MaxSize:    maxSize,    //megabytes
		MaxAge:     maxAge,     //days
		MaxBackups: maxBackups, //files
		Compress:   true,
		LocalTime:  true,
	})
}

func getNameByLogLevel(filename string, level Level) string {
	var name string
	if filename != "" {
		filename = strings.Replace(filename, ".log", "", -1)
		name = filename + "_"
	}
	switch level {
	case WarnLevel, ErrorLevel:
		name += "error.log"
	case DebugLevel:
		name += "debug.log"
	default:
		name += "info.log"
	}
	return name
}

func newEntry(
	config Config,
	infoOutput, errOutput, debugOutput zapcore.WriteSyncer,
) *LogEntry {
	encCfg := zapcore.EncoderConfig{
		TimeKey:          "@t",
		LevelKey:         "lvl",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		ConsoleSeparator: config.Separator,
		EncodeDuration:   zapcore.NanosDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       longTimeEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(encCfg)

	// level setting
	localLoglv := zap.NewAtomicLevelAt(zapcore.Level(config.LoggerLevel))
	if config.SetAsDefaultLvl {
		loglv = localLoglv
		defaultLevel = config.LoggerLevel
	}

	if config.CallerEnabled {
		return getLogEntry(
			zap.New(zapcore.NewCore(encoder, infoOutput, localLoglv), zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip)),
			zap.New(zapcore.NewCore(encoder, errOutput, localLoglv), zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip)),
			zap.New(zapcore.NewCore(encoder, debugOutput, localLoglv), zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip)),
		)
	}
	return getLogEntry(
		zap.New(zapcore.NewCore(encoder, infoOutput, localLoglv)),
		zap.New(zapcore.NewCore(encoder, errOutput, localLoglv)),
		zap.New(zapcore.NewCore(encoder, debugOutput, localLoglv)),
	)
}

// FromContext get Entry from context, if not found, return default logger
func FromContext(ctx context.Context) Entry {
	data := ctx.Value(logCtxKey)
	if data == nil {
		return defaultLogger.clone()
	}
	entry, ok := data.(Entry)
	if !ok {
		return &empty{}
	}
	return entry
}

// WithLogger set given LogEntry to context and return new context, if ctx or entry is nil, return ctx
func WithLogger(ctx context.Context, entry Entry) context.Context {
	if ctx == nil || entry == nil {
		return ctx
	}

	return context.WithValue(ctx, logCtxKey, entry)
}

// CopyToContext copy logger from srcCtx to dstCtx
func CopyToContext(srcCtx, dstCtx context.Context) context.Context {
	if srcCtx == nil || dstCtx == nil {
		return dstCtx
	}

	dstCtx = WithLogger(dstCtx, FromContext(srcCtx))
	return dstCtx
}
