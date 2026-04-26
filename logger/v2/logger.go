// Package logger provides a structured, leveled logging API built on go.uber.org/zap.
//
// Features:
//   - Four log levels: Debug, Info, Warn, Error
//   - File rotation via lumberjack (size/age/backup limits)
//   - Structured fields via With() / WithError() / WithRequestID()
//   - Traffic logging with start/end duration tracking
//   - Context propagation
//   - Output trimming for large structs/slices/maps
//
// Quick start:
//
//	logger.Configure(logger.Config{
//	    Level:   logger.InfoLevel,
//	    Console: true,
//	})
//	logger.Info("server started", "port", 8080)
package logger

import (
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

// WriterSyncer is an alias for zapcore.WriteSyncer.
type WriterSyncer = zapcore.WriteSyncer

// Config configures a logger instance.
type Config struct {
	// Level sets the minimum log level. Default: InfoLevel.
	Level Level

	// Console enables logging to stdout (Info/Debug) and stderr (Warn/Error).
	Console bool

	// FilePath enables file logging to the given directory.
	// When set, log files are created under FilePath/<level>.log.
	// Default: "" (disabled).
	FilePath string

	// MaxSize is the max size in MB before a log file is rotated. Default: 100.
	MaxSize int

	// MaxAge is the max age in days to keep a log file. Default: 7.
	MaxAge int

	// MaxBackups is the max number of rotated files to keep. Default: 10.
	MaxBackups int

	// Caller adds the caller's file and line number to each log entry.
	Caller bool

	// CallerSkip increases the number of callers skipped (default 1).
	CallerSkip int

	// Traffic enables the traffic logger component.
	Traffic bool

	// TrafficPath overrides the directory for traffic log files.
	// Falls back to FilePath if empty.
	TrafficPath string

	// TrafficMaxSize/MaxAge/MaxBackups override rotation settings for traffic logs.
	TrafficMaxSize    int
	TrafficMaxAge     int
	TrafficMaxBackups int

	// Trimmer configures output truncation for large fields.
	// If nil, default truncation is applied (arr=3, str=128, depth=10).
	Trimmer *TrimConfig
}

// TrimConfig controls output truncation.
type TrimConfig struct {
	ArrLimit  int
	StrLimit  int
	DeepLimit int
	Ignores   []string
}

var defaultConfig = Config{
	Level:          InfoLevel,
	Console:        true,
	MaxSize:        100,
	MaxAge:         7,
	MaxBackups:     10,
	TrafficMaxSize: 100,
	TrafficMaxAge:  7,
}

// ConfigOption is a functional option for Configure.
type ConfigOption func(*Config)

// Deprecated config option aliases for backward compatibility.
func WithLoggerLevel(lvl Level) ConfigOption   { return WithLevel(lvl) }
func WithDirectory(dir string) ConfigOption    { return WithFilePath(dir) }
func WithFileEnabled(on bool) ConfigOption     { return func(c *Config) { if on { c.FilePath = "log" } } }
func WithConsoleEnabled(on bool) ConfigOption  { return WithConsole(on) }
func WithCallerEnabled(on bool) ConfigOption   { return WithCaller(on) }
func WithSetAsDefaultLvl(on bool) ConfigOption { return func(c *Config) {} }
func WithTrafficEnabled(on bool) ConfigOption  { return WithTraffic(on) }
func WithTrafficStream(s WriterSyncer) ConfigOption { return func(c *Config) {} }

// --- Config options ---

func WithLevel(lvl Level) ConfigOption           { return func(c *Config) { c.Level = lvl } }
func WithConsole(on bool) ConfigOption            { return func(c *Config) { c.Console = on } }
func WithFilePath(dir string) ConfigOption        { return func(c *Config) { c.FilePath = dir } }
func WithMaxSize(mb int) ConfigOption             { return func(c *Config) { c.MaxSize = mb } }
func WithMaxAge(days int) ConfigOption            { return func(c *Config) { c.MaxAge = days } }
func WithMaxBackups(n int) ConfigOption           { return func(c *Config) { c.MaxBackups = n } }
func WithCaller(on bool) ConfigOption             { return func(c *Config) { c.Caller = on } }
func WithCallerSkip(skip int) ConfigOption        { return func(c *Config) { c.CallerSkip = skip } }
func WithTraffic(on bool) ConfigOption            { return func(c *Config) { c.Traffic = on } }
func WithTrafficPath(dir string) ConfigOption     { return func(c *Config) { c.TrafficPath = dir } }
func WithTrafficMaxSize(mb int) ConfigOption      { return func(c *Config) { c.TrafficMaxSize = mb } }
func WithTrafficMaxAge(days int) ConfigOption     { return func(c *Config) { c.TrafficMaxAge = days } }
func WithTrafficMaxBackups(n int) ConfigOption    { return func(c *Config) { c.TrafficMaxBackups = n } }
func WithTrimConfig(tc *TrimConfig) ConfigOption  { return func(c *Config) { c.Trimmer = tc } }

// --- Deprecated traffic config (old API compat) ---

// TrafficConfig is the old traffic logging config struct. Kept for API compat.
type TrafficConfig struct {
	Enabled    bool
	Directory  string
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	StrLimit   int
	ArrLimit   int
	DeepLimit  int
	Ignores    []string
	Stream     WriterSyncer
}

// TrafficConfigOption is a functional option for old traffic config.
type TrafficConfigOption func(*TrafficConfig)

// ConfigureTrafficWithOpts applies old-style traffic config options.
func ConfigureTrafficWithOpts(opts ...TrafficConfigOption) {
	tc := TrafficConfig{
		Enabled: false, MaxSize: 100, MaxBackups: 10, MaxAge: 7,
		StrLimit: 128, ArrLimit: 3, DeepLimit: 10,
	}
	for _, o := range opts {
		o(&tc)
	}
	if tc.Enabled {
		cfg := defaultConfig
		cfg.Traffic = true
		if tc.Directory != "" {
			cfg.TrafficPath = tc.Directory
		}
		cfg.TrafficMaxSize = tc.MaxSize
		cfg.TrafficMaxBackups = tc.MaxBackups
		cfg.TrafficMaxAge = tc.MaxAge
		global = newEntry(cfg)
	}
}

// Old traffic config option constructors.
func WithTrafficDirectory(dir string) TrafficConfigOption { return func(tc *TrafficConfig) { tc.Directory = dir } }
func WithTrafficFilename(name string) TrafficConfigOption { return func(tc *TrafficConfig) { tc.Filename = name } }
func WithTrafficMaxSizeOld(mb int) TrafficConfigOption    { return func(tc *TrafficConfig) { tc.MaxSize = mb } }
func WithTrafficMaxBackupsOld(n int) TrafficConfigOption  { return func(tc *TrafficConfig) { tc.MaxBackups = n } }
func WithTrafficMaxAgeOld(days int) TrafficConfigOption   { return func(tc *TrafficConfig) { tc.MaxAge = days } }
func WithTrafficStrLimitOld(limit int) TrafficConfigOption { return func(tc *TrafficConfig) { tc.StrLimit = limit } }
func WithTrafficArrLimitOld(limit int) TrafficConfigOption { return func(tc *TrafficConfig) { tc.ArrLimit = limit } }
func WithTrafficDeepLimitOld(limit int) TrafficConfigOption { return func(tc *TrafficConfig) { tc.DeepLimit = limit } }
func WithTrafficIgnoresOld(ignores ...string) TrafficConfigOption { return func(tc *TrafficConfig) { tc.Ignores = ignores } }
func WithTrafficEnabledOld(on bool) TrafficConfigOption { return func(tc *TrafficConfig) { tc.Enabled = on } }
func WithTrafficStreamOld(s WriterSyncer) TrafficConfigOption { return func(tc *TrafficConfig) { tc.Stream = s } }
