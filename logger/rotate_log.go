package logger

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path"
	"strings"
)

const (
	maxWholeSize = 4096
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

// Config for logging
type Config struct {
	// LoggingLevel set log defaultLevel
	LoggingLevel Level
	// FileLoggingEnabled makes the framework log to a file
	// the fields below can be skipped if this value is false!
	FileLoggingEnabled bool
	// ConsoleLoggingEnabled makes the framework log to console
	ConsoleLoggingEnabled bool
	// CallerEnabled makes the caller log to a file
	CallerEnabled bool
	// CallerSkip increases the number of callers skipped by caller
	CallerSkip int
	// Directory to log to to when filelogging is enabled
	Directory string
	// Filename is the name of the logfile which will be placed inside the directory
	Filename string
	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int
	// MaxBackups the max number of rolled files to keep
	MaxBackups int
	// MaxAge the max age in days to keep a logfile
	MaxAge int
	// ConsoleInfoStream
	ConsoleInfoStream *os.File
	// ConsoleErrorStream
	ConsoleErrorStream *os.File
	// ConsoleDebugStream
	ConsoleDebugStream *os.File
}

// Configure configures the default logger
var defaultConfig = Config{
	LoggingLevel:  InfoLevel,
	CallerEnabled: false,
	CallerSkip:    1,
}

// defaultLogger is the default logger
var defaultLogger = newEntry(defaultConfig, os.Stdout, os.Stderr, os.Stdout, true)

// Debug Log a message at the debug defaultLevel
func Debug(msg string) {
	if !Enabled(DebugLevel) {
		return
	}
	msg = withTrace(msg)
	defaultLogger.infoLogger.Debug(msg)
}

func Debugf(format string, args ...any) {
	if !Enabled(DebugLevel) {
		return
	}
	msg := withTrace(fmt.Sprintf(format, args...))
	defaultLogger.debugLogger.Debug(msg)
}

// DebugWith Log a message with fields at the debug defaultLevel
func DebugWith(msg string, fields Fields) {
	if !Enabled(DebugLevel) {
		return
	}
	msg = withTrace(msg)
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
	msg = withTrace(msg)
	defaultLogger.infoLogger.Info(msg)
}

func Infof(format string, args ...any) {
	if !Enabled(InfoLevel) {
		return
	}
	msg := withTrace(fmt.Sprintf(format, args...))
	defaultLogger.infoLogger.Info(msg)
}

// InfoWith Log a message with fields at the info defaultLevel
func InfoWith(msg string, fields Fields) {
	if !Enabled(InfoLevel) {
		return
	}
	msg = withTrace(msg)
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
	msg = withTrace(msg)
	defaultLogger.errLogger.Warn(msg)
}

func Warnf(format string, args ...any) {
	if !Enabled(WarnLevel) {
		return
	}
	msg := withTrace(fmt.Sprintf(format, args...))
	defaultLogger.errLogger.Warn(msg)
}

// WarnWith Log a message with fields at the warn defaultLevel
func WarnWith(msg string, fields Fields) {
	if !Enabled(WarnLevel) {
		return
	}
	msg = withTrace(msg)
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
	msg = withTrace(msg)
	defaultLogger.errLogger.Error(msg)
}

func Errorf(format string, args ...any) {
	if !Enabled(ErrorLevel) {
		return
	}
	msg := withTrace(fmt.Sprintf(format, args...))
	defaultLogger.errLogger.Error(msg)
}

// ErrorWith Log a message with fields at the error defaultLevel
func ErrorWith(msg string, fields Fields) {
	if !Enabled(ErrorLevel) {
		return
	}
	msg = withTrace(msg)
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
	return WithField(defaultFieldName, data)
}

// WithError binds an error to a log message
func WithError(err error) Entry {
	return WithField(defaultErrFieldName, err)
}

// WithTracing create copy of LogEntry with tracing.Span
func WithTracing(requestId string) Entry {
	return defaultLogger.WithTracing(requestId)
}

func withTrace(msg string) string {
	if defaultLogger == nil {
		return strings.Join(append([]string{
			defaultTraceOccupy,
			msg,
		}), defaultSeparator)
	}
	if defaultLogger.requestId == "" {
		return strings.Join(append([]string{
			defaultTraceOccupy,
			msg,
		}), defaultSeparator)
	}
	return strings.Join(append([]string{
		defaultLogger.requestId,
		msg,
	}), defaultSeparator)
}

// Configure sets up the defaultLogger
func Configure(config Config) {
	var infoWriters []zapcore.WriteSyncer
	var errWriters []zapcore.WriteSyncer
	var debugWriters []zapcore.WriteSyncer

	if config.FileLoggingEnabled {
		infoLog := newRollingFile(config.Directory, getNameByLogLevel(config.Filename, InfoLevel), config.MaxSize, config.MaxAge, config.MaxBackups)
		errLog := newRollingFile(config.Directory, getNameByLogLevel(config.Filename, ErrorLevel), config.MaxSize, config.MaxAge, config.MaxBackups)
		debugLog := newRollingFile(config.Directory, getNameByLogLevel(config.Filename, DebugLevel), config.MaxSize, config.MaxAge, config.MaxBackups)
		infoWriters = append(infoWriters, infoLog)
		errWriters = append(errWriters, errLog)
		debugWriters = append(debugWriters, debugLog)
	} else {
		config.ConsoleLoggingEnabled = true
	}

	if config.ConsoleLoggingEnabled {
		if config.ConsoleInfoStream != nil {
			infoWriters = append(infoWriters, config.ConsoleInfoStream)
		} else {
			infoWriters = append(infoWriters, os.Stdout)
		}
		if config.ConsoleErrorStream != nil {
			errWriters = append(errWriters, config.ConsoleErrorStream)
		} else {
			errWriters = append(errWriters, os.Stderr)
		}
		if config.ConsoleDebugStream != nil {
			debugWriters = append(debugWriters, config.ConsoleDebugStream)
		} else {
			debugWriters = append(debugWriters, os.Stdout)
		}
	}

	defaultLogger = newEntry(
		config,
		zapcore.NewMultiWriteSyncer(infoWriters...),
		zapcore.NewMultiWriteSyncer(errWriters...),
		zapcore.NewMultiWriteSyncer(debugWriters...),
		true,
	)

	declareLogger(config, InfoWith)
	declareLogger(config, ErrorWith)
	declareLogger(config, DebugWith)

}

// NewEntry create a new LogEntry instead of override defaultzaplogger
func NewEntry(config Config) Entry {
	var infoWriters []zapcore.WriteSyncer
	var errWriters []zapcore.WriteSyncer
	var debugWriters []zapcore.WriteSyncer

	if config.FileLoggingEnabled {
		infoLog := newRollingFile(config.Directory, getNameByLogLevel(config.Filename, InfoLevel), config.MaxSize, config.MaxAge, config.MaxBackups)
		errLog := newRollingFile(config.Directory, getNameByLogLevel(config.Filename, ErrorLevel), config.MaxSize, config.MaxAge, config.MaxBackups)
		debugLog := newRollingFile(config.Directory, getNameByLogLevel(config.Filename, DebugLevel), config.MaxSize, config.MaxAge, config.MaxBackups)
		infoWriters = append(infoWriters, infoLog)
		errWriters = append(errWriters, errLog)
		debugWriters = append(debugWriters, debugLog)
	} else {
		config.ConsoleLoggingEnabled = true
		infoWriters = append(infoWriters, os.Stdout)
		errWriters = append(errWriters, os.Stderr)
		debugWriters = append(debugWriters, os.Stdout)
	}

	logEntry := newEntry(
		config,
		zapcore.NewMultiWriteSyncer(infoWriters...),
		zapcore.NewMultiWriteSyncer(errWriters...),
		zapcore.NewMultiWriteSyncer(debugWriters...),
		true)

	declareLogger(config, logEntry.InfoWith)
	declareLogger(config, logEntry.ErrorWith)
	declareLogger(config, logEntry.DebugWith)
	return logEntry
}

func declareLogger(config Config, logv func(msg string, fields Fields)) {
	logv("logging configured", Fields{"config": config})
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

func newEntry(config Config, infoOutput, errOutput, debugOutput zapcore.WriteSyncer, isDefaultLogger bool) *LogEntry {
	encCfg := zapcore.EncoderConfig{
		TimeKey:          "@t",
		LevelKey:         "lvl",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		ConsoleSeparator: defaultSeparator,
		EncodeDuration:   zapcore.NanosDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       longTimeEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(encCfg)

	// level setting
	localLoglv := zap.NewAtomicLevelAt(zapcore.Level(config.LoggingLevel))
	if isDefaultLogger {
		loglv = localLoglv
		defaultLevel = config.LoggingLevel
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
