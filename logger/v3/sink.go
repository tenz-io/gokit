package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// buildCore assembles the main logging core from console and file sinks.
//
// Each sink carries its own LevelEnabler so that file output is split by
// severity: debug.log receives Debug and above, info.log Info and above,
// warn.log Warn and above, error.log Error and above. Console output sends
// Info/Debug to stdout and Warn/Error to stderr, each with the appropriate
// threshold. The global AtomicLevel (passed as *level) gates everything,
// so raising it suppresses lower-severity output across all sinks uniformly.
func buildCore(cfg Config, enc zapcore.Encoder, level *zap.AtomicLevel) zapcore.Core {
	sinks := buildSinks(cfg)
	if len(sinks) == 0 {
		// Respect an explicitly disabled console. A logger with no configured
		// destinations is intentionally silent instead of leaking to stdout.
		return zapcore.NewNopCore()
	}

	cores := make([]zapcore.Core, 0, len(sinks))
	for _, s := range sinks {
		s := s // capture per-iteration (pre-Go1.22 loop var semantics)
		// The effective threshold is the stricter of the sink's own enabler
		// and the global AtomicLevel, so SetLevel can still throttle output
		// on a sink that would otherwise accept everything.
		enabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
			return level.Enabled(l) && s.enabler.Enabled(l)
		})
		cores = append(cores, zapcore.NewCore(enc, zapcore.AddSync(s.ws), enabler))
	}
	return zapcore.NewTee(cores...)
}

// sink pairs a destination with the minimum level it accepts.
type sink struct {
	ws      zapcore.WriteSyncer
	enabler zapcore.LevelEnabler
}

// buildSinks assembles the console and file sinks described by cfg.
func buildSinks(cfg Config) []sink {
	var sinks []sink

	if cfg.Console {
		// stdout: Debug and Info (anything below Warn).
		sinks = append(sinks, sink{
			ws:      os.Stdout,
			enabler: zap.LevelEnablerFunc(func(l zapcore.Level) bool { return l < zapcore.WarnLevel }),
		})
		// stderr: Warn and Error.
		sinks = append(sinks, sink{
			ws:      os.Stderr,
			enabler: zap.LevelEnablerFunc(func(l zapcore.Level) bool { return l >= zapcore.WarnLevel }),
		})
	}

	if cfg.FilePath != "" {
		if err := os.MkdirAll(cfg.FilePath, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "logger: failed to create log directory %s: %v\n", cfg.FilePath, err)
		} else {
			sinks = append(sinks, fileSinks(cfg)...)
		}
	}

	return sinks
}

// fileSinks returns one lumberjack-backed sink per severity, each with a
// distinct LevelEnabler. Files are cumulative: debug.log receives every
// level, while error.log receives only Error entries.
func fileSinks(cfg Config) []sink {
	type fileSpec struct {
		name string
		min  zapcore.Level
	}
	specs := []fileSpec{
		{"debug", zapcore.DebugLevel}, // Debug+ (everything)
		{"info", zapcore.InfoLevel},   // Info+
		{"warn", zapcore.WarnLevel},   // Warn+
		{"error", zapcore.ErrorLevel}, // Error+
	}
	sinks := make([]sink, 0, len(specs))
	for _, s := range specs {
		name, min := s.name, s.min
		sinks = append(sinks, sink{
			ws: zapcore.AddSync(&lumberjack.Logger{
				Filename:   filepath.Join(cfg.FilePath, name+".log"),
				MaxSize:    cfg.MaxSize,
				MaxAge:     cfg.MaxAge,
				MaxBackups: cfg.MaxBackups,
				Compress:   true,
				LocalTime:  true,
			}),
			enabler: zap.LevelEnablerFunc(func(l zapcore.Level) bool { return l >= min }),
		})
	}
	return sinks
}

// validateConfig fills in defaults for any unset numeric fields.
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
	if cfg.Encoding == "" {
		cfg.Encoding = ConsoleEncoding
	}
	if cfg.CallerSkip < 0 {
		cfg.CallerSkip = 0
	}
}
