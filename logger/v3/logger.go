// Package logger provides a structured, leveled logging API built on
// go.uber.org/zap.
//
// v3 is a clean rewrite with no backwards-compatibility shims. It supports:
//   - Four log levels: Debug, Info, Warn, Error
//   - Console output (default) and per-level file output with rotation
//   - Structured fields via With() / WithError() / WithRequestID() chaining
//   - Context propagation: attach an Entry to a context.Context and pull it
//     back out across call chains
//   - Traffic logging: start/end spans that record cmd, cost, code and resp
//     into a separate traffic.log
//   - Output trimming for large strings/slices/maps/structs
//   - Runtime level changes via SetLevel/GetLevel
//
// Quick start:
//
//	logger.Configure(logger.Config{
//	    Level:   logger.InfoLevel,
//	    Console: true,
//	})
//	logger.Infow("server started", "port", 8080)
package logger

import (
	"sync/atomic"

	"go.uber.org/zap/zapcore"
)

// Fields is a map of structured key-value pairs attached to log entries.
type Fields map[string]any

func (f Fields) toArgs() []any {
	if len(f) == 0 {
		return nil
	}
	args := make([]any, 0, len(f)*2)
	for k, v := range f {
		args = append(args, k, v)
	}
	return args
}

// Level represents a log severity level.
type Level zapcore.Level

const (
	DebugLevel Level = Level(zapcore.DebugLevel)
	InfoLevel  Level = Level(zapcore.InfoLevel)
	WarnLevel  Level = Level(zapcore.WarnLevel)
	ErrorLevel Level = Level(zapcore.ErrorLevel)
)

// Encoding selects the output format of the logger.
type Encoding string

const (
	// ConsoleEncoding (default) writes human-friendly "LEVEL time msg key=value" lines.
	ConsoleEncoding Encoding = "console"
	// JSONEncoding writes one JSON object per line, suitable for log aggregation.
	JSONEncoding Encoding = "json"
)

// WriterSyncer is an alias for zapcore.WriteSyncer.
type WriterSyncer = zapcore.WriteSyncer

// LevelEnabler is an alias for zapcore.LevelEnabler.
type LevelEnabler = zapcore.LevelEnabler

// Config configures a logger instance. The initial global logger and the
// functional-option constructors use the documented defaults. With the
// struct constructors, boolean fields are explicit: Console:false means no
// console output.
type Config struct {
	// Level sets the minimum log level. Default: InfoLevel.
	Level Level

	// Encoding selects between console and JSON output. Default: console.
	Encoding Encoding

	// Console enables logging to stdout (Info/Debug) and stderr (Warn/Error).
	// Default: true.
	Console bool

	// FilePath enables file logging to the given directory. When set, log
	// files are created under FilePath/<level>.log, split by severity:
	// debug.log (Debug+), info.log (Info+), warn.log (Warn+), error.log (Error+).
	// Default: "" (disabled).
	FilePath string

	// MaxSize is the max size in MB before a log file is rotated. Default: 100.
	MaxSize int
	// MaxAge is the max age in days to keep a rotated log file. Default: 7.
	MaxAge int
	// MaxBackups is the max number of rotated files to keep. Default: 10.
	MaxBackups int

	// Caller adds the caller's file and line number to each log entry.
	Caller bool
	// CallerSkip increases the number of callers skipped (default 0).
	CallerSkip int

	// Traffic enables the traffic logger component, which writes a separate
	// traffic.log recording request/response spans.
	Traffic bool
	// TrafficPath overrides the directory for traffic.log. Falls back to
	// FilePath, then "log" if empty.
	TrafficPath string
	// TrafficMaxSize/MaxAge/MaxBackups override rotation settings for the
	// traffic log. Zero falls back to the main MaxSize/MaxAge/MaxBackups.
	TrafficMaxSize    int
	TrafficMaxAge     int
	TrafficMaxBackups int

	// Trimmer configures output truncation for large fields. If nil, sensible
	// defaults are applied (arr=3, str=128, depth=10).
	Trimmer *TrimConfig
}

// TrimConfig controls output truncation.
type TrimConfig struct {
	ArrLimit  int      // max elements kept from a slice/array (default 3)
	StrLimit  int      // max bytes kept from a string (default 128)
	DeepLimit int      // max nesting depth for structs/maps (default 10)
	Ignores   []string // field names to drop entirely
}

// defaultConfig is the configuration applied when none is provided.
var defaultConfig = Config{
	Level:             InfoLevel,
	Encoding:          ConsoleEncoding,
	Console:           true,
	MaxSize:           100,
	MaxAge:            7,
	MaxBackups:        10,
	TrafficMaxSize:    100,
	TrafficMaxAge:     7,
	TrafficMaxBackups: 10,
}

// ConfigOption is a functional option for Configure / NewEntry.
type ConfigOption func(*Config)

// --- Config options ---

func WithLevel(lvl Level) ConfigOption           { return func(c *Config) { c.Level = lvl } }
func WithEncoding(enc Encoding) ConfigOption     { return func(c *Config) { c.Encoding = enc } }
func WithConsole(on bool) ConfigOption           { return func(c *Config) { c.Console = on } }
func WithFilePath(dir string) ConfigOption       { return func(c *Config) { c.FilePath = dir } }
func WithMaxSize(mb int) ConfigOption            { return func(c *Config) { c.MaxSize = mb } }
func WithMaxAge(days int) ConfigOption           { return func(c *Config) { c.MaxAge = days } }
func WithMaxBackups(n int) ConfigOption          { return func(c *Config) { c.MaxBackups = n } }
func WithCaller(on bool) ConfigOption            { return func(c *Config) { c.Caller = on } }
func WithCallerSkip(skip int) ConfigOption       { return func(c *Config) { c.CallerSkip = skip } }
func WithTraffic(on bool) ConfigOption           { return func(c *Config) { c.Traffic = on } }
func WithTrafficPath(dir string) ConfigOption    { return func(c *Config) { c.TrafficPath = dir } }
func WithTrafficMaxSize(mb int) ConfigOption     { return func(c *Config) { c.TrafficMaxSize = mb } }
func WithTrafficMaxAge(days int) ConfigOption    { return func(c *Config) { c.TrafficMaxAge = days } }
func WithTrafficMaxBackups(n int) ConfigOption   { return func(c *Config) { c.TrafficMaxBackups = n } }
func WithTrimConfig(tc *TrimConfig) ConfigOption { return func(c *Config) { c.Trimmer = tc } }

// --- Construction ---

// Configure initializes the global logger.
func Configure(config Config) {
	global.Store(newEntry(config))
}

// ConfigureWithOpts initializes the global logger using functional options.
func ConfigureWithOpts(opts ...ConfigOption) {
	cfg := defaultConfig
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	Configure(cfg)
}

// NewEntry creates a standalone Entry without affecting the global logger.
func NewEntry(config Config) Entry { return newEntry(config) }

// NewEntryWithOpts creates a standalone Entry using functional options.
func NewEntryWithOpts(opts ...ConfigOption) Entry {
	cfg := defaultConfig
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	return NewEntry(cfg)
}

// --- Global state ---

var global atomic.Pointer[logEntry]

func init() {
	global.Store(newEntry(defaultConfig))
}

func current() *logEntry { return global.Load() }

// L returns the global logger entry.
func L() Entry { return current() }

// SetLevel adjusts the global log level at runtime. Unlike the v2
// implementation, the level is wired into the core, so this takes effect
// immediately for all subsequent calls.
func SetLevel(lvl Level) {
	current().SetLevel(lvl)
}

// GetLevel returns the current global log level.
func GetLevel() Level {
	return current().GetLevel()
}

// --- Package-level convenience functions ---

// The package-level logging functions call the underlying SugaredLogger
// directly. This keeps caller reporting correct: there is one wrapper frame
// here, just as there is one wrapper frame in the corresponding Entry method.
func Debug(args ...any)                 { current().base.Debug(args...) }
func Debugf(format string, args ...any) { current().base.Debugf(format, args...) }
func Debugw(msg string, fields ...any) {
	e := current()
	e.base.Debugw(msg, e.trimmer.TrimFields(fields)...)
}
func Info(args ...any)                 { current().base.Info(args...) }
func Infof(format string, args ...any) { current().base.Infof(format, args...) }
func Infow(msg string, fields ...any) {
	e := current()
	e.base.Infow(msg, e.trimmer.TrimFields(fields)...)
}
func Warn(args ...any)                 { current().base.Warn(args...) }
func Warnf(format string, args ...any) { current().base.Warnf(format, args...) }
func Warnw(msg string, fields ...any) {
	e := current()
	e.base.Warnw(msg, e.trimmer.TrimFields(fields)...)
}
func Error(args ...any)                 { current().base.Error(args...) }
func Errorf(format string, args ...any) { current().base.Errorf(format, args...) }
func Errorw(msg string, fields ...any) {
	e := current()
	e.base.Errorw(msg, e.trimmer.TrimFields(fields)...)
}

func With(args ...any) Entry              { return current().With(args...) }
func WithFields(fields Fields) Entry      { return current().WithFields(fields) }
func WithField(k string, v any) Entry     { return current().WithField(k, v) }
func WithError(err error) Entry           { return current().WithError(err) }
func WithRequestID(id string) Entry       { return current().WithRequestID(id) }
func WithTracing(id string) Entry         { return current().WithTracing(id) }
func Enabled(lvl Level) bool              { return current().Enabled(lvl) }
func StartTraffic(cmd string) *TrafficRec { return current().StartTraffic(cmd) }
