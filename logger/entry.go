package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Entry is the primary logging interface.
type Entry interface {
	Debug(args ...any)
	Debugf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)

	// With returns a child Entry with the given structured fields appended.
	// Fields are passed as alternating key-value pairs: With("user", "bob", "age", 42).
	With(args ...any) Entry

	// WithError adds an error field ("error", err).
	WithError(err error) Entry

	// WithRequestID sets the request ID field.
	WithRequestID(id string) Entry

	// WithTracing is an alias for WithRequestID.
	WithTracing(id string) Entry

	// WithFields adds multiple structured fields from a Fields map.
	WithFields(fields Fields) Entry

	// WithField adds a single field.
	WithField(k string, v any) Entry

	// StartTraffic begins a traffic span for request/response logging.
	// Returns nil if traffic logging is not configured.
	StartTraffic(cmd string) *TrafficRec

	// Enabled reports whether the entry will log at the given level.
	Enabled(lvl Level) bool

	// Logger returns the underlying *zap.SugaredLogger.
	Logger() *zap.SugaredLogger
}

// logEntry implements Entry using a single *zap.SugaredLogger.
type logEntry struct {
	base     *zap.SugaredLogger
	level    zap.AtomicLevel
	traffic  *zap.SugaredLogger // nil when traffic is disabled
	trimmer  *OutputTrimmer
}

// --- Entry methods ---

func (e *logEntry) Logger() *zap.SugaredLogger { return e.base }

func (e *logEntry) Debug(args ...any)            { e.base.Debug(args...) }
func (e *logEntry) Debugf(fmt string, args ...any) { e.base.Debugf(fmt, args...) }
func (e *logEntry) Info(args ...any)             { e.base.Info(args...) }
func (e *logEntry) Infof(fmt string, args ...any)  { e.base.Infof(fmt, args...) }
func (e *logEntry) Warn(args ...any)             { e.base.Warn(args...) }
func (e *logEntry) Warnf(fmt string, args ...any)  { e.base.Warnf(fmt, args...) }
func (e *logEntry) Error(args ...any)            { e.base.Error(args...) }
func (e *logEntry) Errorf(fmt string, args ...any)  { e.base.Errorf(fmt, args...) }

func (e *logEntry) With(args ...any) Entry {
	if len(args) == 0 {
		return e
	}
	trimmed := e.trimmer.TrimArgs(args)
	return &logEntry{
		base:    e.base.With(trimmed...),
		level:   e.level,
		traffic: e.traffic,
		trimmer: e.trimmer,
	}
}

func (e *logEntry) WithError(err error) Entry {
	if err == nil {
		return e
	}
	return &logEntry{
		base:    e.base.With("error", err.Error()),
		level:   e.level,
		traffic: e.traffic,
		trimmer: e.trimmer,
	}
}

func (e *logEntry) WithRequestID(id string) Entry {
	return e.WithTracing(id)
}

func (e *logEntry) WithTracing(id string) Entry {
	if id == "" {
		return e
	}
	return &logEntry{
		base:    e.base.With("request_id", id),
		level:   e.level,
		traffic: e.traffic,
		trimmer: e.trimmer,
	}
}

func (e *logEntry) WithFields(fields Fields) Entry {
	if len(fields) == 0 {
		return e
	}
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return e.With(args...)
}

func (e *logEntry) WithField(k string, v any) Entry {
	return e.WithFields(Fields{k: v})
}

func (e *logEntry) StartTraffic(cmd string) *TrafficRec {
	if e.traffic == nil {
		return nil
	}
	return startTrafficRec(e.traffic, cmd, e.trimmer)
}

func (e *logEntry) Enabled(lvl Level) bool {
	return e.level.Enabled(zapcore.Level(lvl))
}

// --- Construction ---

// Configure initializes the global logger.
func Configure(config Config) {
	global = newEntry(config)
}

// ConfigureWithOpts initializes the global logger using functional options.
func ConfigureWithOpts(opts ...ConfigOption) {
	cfg := defaultConfig
	for _, o := range opts {
		o(&cfg)
	}
	Configure(cfg)
}

// NewEntry creates a standalone Entry without affecting the global logger.
func NewEntry(config Config) Entry {
	return newEntry(config)
}

// NewEntryWithOpts creates a standalone Entry using functional options.
func NewEntryWithOpts(opts ...ConfigOption) Entry {
	cfg := defaultConfig
	for _, o := range opts {
		o(&cfg)
	}
	return NewEntry(cfg)
}

func newEntry(cfg Config) *logEntry {
	validateConfig(&cfg)

	enc := encoderConfig()
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(enc),
		zapcore.NewMultiWriteSyncer(writers(cfg)...),
		zapcore.Level(cfg.Level),
	)
	level := zap.NewAtomicLevelAt(zapcore.Level(cfg.Level))

	var logger *zap.Logger
	if cfg.Caller {
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(cfg.CallerSkip+1))
	} else {
		logger = zap.New(core)
	}

	trimmer := newTrimmer(cfg.Trimmer)

	var traffic *zap.SugaredLogger
	if cfg.Traffic {
		traffic = newTrafficCore(cfg)
	}

	return &logEntry{
		base:    logger.Sugar(),
		level:   level,
		traffic: traffic,
		trimmer: trimmer,
	}
}

func writers(cfg Config) []zapcore.WriteSyncer {
	var ws []zapcore.WriteSyncer

	if cfg.Console {
		ws = append(ws, os.Stdout)
	}

	if cfg.FilePath != "" {
		if err := os.MkdirAll(cfg.FilePath, 0744); err != nil {
			fmt.Fprintf(os.Stderr, "logger: failed to create log directory %s: %v\n", cfg.FilePath, err)
		} else {
			for _, lvl := range []struct {
				name string
			}{
				{"info"}, {"error"}, {"debug"},
			} {
				name := lvl.name
				ws = append(ws, zapcore.AddSync(&lumberjack.Logger{
					Filename:   cfg.FilePath + "/" + name + ".log",
					MaxSize:    cfg.MaxSize,
					MaxAge:     cfg.MaxAge,
					MaxBackups: cfg.MaxBackups,
					Compress:   true,
					LocalTime:  true,
				}))
			}
		}
	}

	if len(ws) == 0 {
		// Always have at least stdout
		ws = append(ws, os.Stdout)
	}

	return ws
}

func newTrafficCore(cfg Config) *zap.SugaredLogger {
	trafficDir := cfg.TrafficPath
	if trafficDir == "" {
		trafficDir = cfg.FilePath
	}
	if trafficDir == "" {
		trafficDir = "log"
	}

	if err := os.MkdirAll(trafficDir, 0744); err != nil {
		return nil
	}

	enc := zapcore.EncoderConfig{
		TimeKey:        "@t",
		MessageKey:     "msg",
		EncodeTime:     func(t time.Time, e zapcore.PrimitiveArrayEncoder) { e.AppendString(t.Format("2006-01-02T15:04:05.000Z0700")) },
		EncodeDuration: zapcore.NanosDurationEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(enc),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   trafficDir + "/traffic.log",
			MaxSize:    cfg.TrafficMaxSize,
			MaxAge:     cfg.TrafficMaxAge,
			MaxBackups: cfg.TrafficMaxBackups,
			Compress:   true,
			LocalTime:  true,
		}),
		zapcore.InfoLevel,
	)

	return zap.New(core).Sugar()
}

func encoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:       "@t",
		LevelKey:      "lvl",
		CallerKey:     "caller",
		MessageKey:    "msg",
		EncodeTime:    func(t time.Time, e zapcore.PrimitiveArrayEncoder) { e.AppendString(t.Format("2006-01-02T15:04:05.000Z0700")) },
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
	}
}

func validateConfig(cfg *Config) {
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 100
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 10
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = 7
	}
	if cfg.TrafficMaxSize <= 0 {
		cfg.TrafficMaxSize = cfg.MaxSize
	}
	if cfg.TrafficMaxBackups <= 0 {
		cfg.TrafficMaxBackups = cfg.MaxBackups
	}
	if cfg.TrafficMaxAge <= 0 {
		cfg.TrafficMaxAge = cfg.MaxAge
	}
}

// --- Global state ---

var global *logEntry

func init() {
	global = newEntry(defaultConfig)
}

// L returns the global logger entry.
func L() Entry { return global }

// SetLevel adjusts the global log level at runtime.
func SetLevel(lvl Level) {
	global.level.SetLevel(zapcore.Level(lvl))
}

// GetLevel returns the current global log level.
func GetLevel() Level {
	return Level(global.level.Level())
}

// --- Package-level convenience functions ---

func Debug(args ...any)                 { global.Debug(args...) }
func Debugf(fmt string, args ...any)    { global.Debugf(fmt, args...) }
func Info(args ...any)                  { global.Info(args...) }
func Infof(fmt string, args ...any)     { global.Infof(fmt, args...) }
func Warn(args ...any)                  { global.Warn(args...) }
func Warnf(fmt string, args ...any)     { global.Warnf(fmt, args...) }
func Error(args ...any)                 { global.Error(args...) }
func Errorf(fmt string, args ...any)    { global.Errorf(fmt, args...) }
func With(args ...any) Entry              { return global.With(args...) }
func WithFields(fields Fields) Entry       { return global.WithFields(fields) }
func WithField(k string, v any) Entry      { return global.WithField(k, v) }
func WithTracing(id string) Entry          { return global.WithRequestID(id) }
func WithError(err error) Entry            { return global.WithError(err) }
func WithRequestID(id string) Entry        { return global.WithRequestID(id) }
func Enabled(lvl Level) bool               { return global.Enabled(lvl) }
func StartTraffic(cmd string) *TrafficRec  { return global.StartTraffic(cmd) }

// StartTrafficRec is the compat function for old API consumers.
func StartTrafficRec(ctx any, req *ReqEntity) *TrafficRec {
	return global.StartTraffic(req.Cmd)
}

// WithTrafficEntry is a compat passthrough. The new API stores the traffic logger
// on the main logEntry, so this is a no-op.
func WithTrafficEntry(ctx context.Context, _ any) context.Context { return ctx }

// WithTrafficTracing is a compat function that returns a compatTraffic wrapper.
func WithTrafficTracing(ctx context.Context, reqID string) compatTraffic {
	return compatTraffic{ctx: ctx, reqID: reqID}
}

type compatTraffic struct {
	ctx    context.Context
	reqID  string
	fields Fields
}

func (c compatTraffic) WithFields(flds Fields) compatTraffic {
	c.fields = flds
	return c
}
func (c compatTraffic) WithIgnores(ignores ...string) compatTraffic { return c }

// --- Context propagation ---

type ctxKey string

const logCtxKey ctxKey = "_log_ctx_key"

// FromContext returns the logger from the context, or the global logger.
func FromContext(ctx context.Context) Entry {
	if ctx == nil {
		return global
	}
	if e, ok := ctx.Value(logCtxKey).(Entry); ok {
		return e
	}
	return global
}

// WithLogger attaches a logger to the context.
func WithLogger(ctx context.Context, e Entry) context.Context {
	if ctx == nil || e == nil {
		return ctx
	}
	return context.WithValue(ctx, logCtxKey, e)
}

// CopyToContext copies the logger from srcCtx to dstCtx.
func CopyToContext(srcCtx, dstCtx context.Context) context.Context {
	if srcCtx == nil || dstCtx == nil {
		return dstCtx
	}
	return WithLogger(dstCtx, FromContext(srcCtx))
}

