package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Entry is the primary logging interface.
//
// Every With* method returns a new Entry that inherits the parent's fields
// without mutating the parent, so an Entry is safe to share across
// goroutines and to derive per-request children from.
type Entry interface {
	// Plain methods are print-style, f methods are printf-style, and w methods
	// accept alternating structured key-value fields.
	Debug(args ...any)
	Debugf(format string, args ...any)
	Debugw(msg string, fields ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Infow(msg string, fields ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Warnw(msg string, fields ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Errorw(msg string, fields ...any)

	// With returns a child Entry with the given structured fields appended.
	// Fields are alternating key-value pairs: With("user", "bob", "age", 42).
	With(args ...any) Entry

	// WithError adds an "error" field with err.Error(). Returns the same
	// entry when err is nil.
	WithError(err error) Entry

	// WithRequestID sets the request_id field. WithTracing is an alias.
	WithRequestID(id string) Entry
	WithTracing(id string) Entry

	// WithFields adds multiple structured fields from a Fields map.
	WithFields(fields Fields) Entry
	// WithField adds a single structured field.
	WithField(k string, v any) Entry

	// StartTraffic begins a traffic span for request/response logging.
	// Returns nil if traffic logging is not configured.
	StartTraffic(cmd string) *TrafficRec

	// Enabled reports whether the entry will log at the given level.
	Enabled(lvl Level) bool

	// SetLevel adjusts the log level of this entry's core at runtime.
	SetLevel(lvl Level)

	// GetLevel returns the current log level of this entry's core.
	GetLevel() Level

	// Logger returns the underlying *zap.SugaredLogger.
	Logger() *zap.SugaredLogger
}

// logEntry implements Entry using a single *zap.SugaredLogger backed by a
// core whose level is governed by a zap.AtomicLevel (so SetLevel works).
type logEntry struct {
	base    *zap.SugaredLogger
	level   zap.AtomicLevel
	traffic *zap.SugaredLogger // nil when traffic is disabled
	trimmer *OutputTrimmer
}

// --- Entry methods ---

func (e *logEntry) Logger() *zap.SugaredLogger { return e.base }

func (e *logEntry) Debug(args ...any)            { e.base.Debug(args...) }
func (e *logEntry) Debugf(f string, args ...any) { e.base.Debugf(f, args...) }
func (e *logEntry) Debugw(msg string, fields ...any) {
	e.base.Debugw(msg, e.trimmer.TrimFields(fields)...)
}
func (e *logEntry) Info(args ...any)            { e.base.Info(args...) }
func (e *logEntry) Infof(f string, args ...any) { e.base.Infof(f, args...) }
func (e *logEntry) Infow(msg string, fields ...any) {
	e.base.Infow(msg, e.trimmer.TrimFields(fields)...)
}
func (e *logEntry) Warn(args ...any)            { e.base.Warn(args...) }
func (e *logEntry) Warnf(f string, args ...any) { e.base.Warnf(f, args...) }
func (e *logEntry) Warnw(msg string, fields ...any) {
	e.base.Warnw(msg, e.trimmer.TrimFields(fields)...)
}
func (e *logEntry) Error(args ...any)            { e.base.Error(args...) }
func (e *logEntry) Errorf(f string, args ...any) { e.base.Errorf(f, args...) }
func (e *logEntry) Errorw(msg string, fields ...any) {
	e.base.Errorw(msg, e.trimmer.TrimFields(fields)...)
}

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
	return e.With("error", err.Error())
}

func (e *logEntry) WithRequestID(id string) Entry { return e.WithTracing(id) }

func (e *logEntry) WithTracing(id string) Entry {
	if id == "" {
		return e
	}
	return e.With("request_id", id)
}

func (e *logEntry) WithFields(fields Fields) Entry {
	if len(fields) == 0 {
		return e
	}
	return e.With(fields.toArgs()...)
}

func (e *logEntry) WithField(k string, v any) Entry { return e.With(k, v) }

func (e *logEntry) StartTraffic(cmd string) *TrafficRec {
	if e.traffic == nil {
		return nil
	}
	return startTrafficRec(e.traffic, cmd, e.trimmer)
}

func (e *logEntry) Enabled(lvl Level) bool {
	return e.level.Enabled(zapcore.Level(lvl))
}

func (e *logEntry) SetLevel(lvl Level) {
	e.level.SetLevel(zapcore.Level(lvl))
}

func (e *logEntry) GetLevel() Level {
	return Level(e.level.Level())
}

// newEntry builds a logEntry from cfg. The AtomicLevel is wired into the
// core (as the LevelEnabler) so that SetLevel/GetLevel actually take effect,
// unlike v2 where a detached AtomicLevel made SetLevel a no-op.
func newEntry(cfg Config) *logEntry {
	validateConfig(&cfg)

	enc := newEncoder(cfg.Encoding)
	level := zap.NewAtomicLevelAt(zapcore.Level(cfg.Level))

	core := buildCore(cfg, enc, &level)

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
