package logger

import (
	"os"
)

// Configure configures the default logger
var defaultConfig = Config{
	LoggerLevel:   InfoLevel,
	CallerEnabled: false,
	CallerSkip:    1,
	Directory:     "log",
	Filename:      "",
	MaxSize:       100,
	MaxAge:        7,
	MaxBackups:    20,
}

// Config for logging
type Config struct {
	// LoggerLevel set log defaultLevel
	LoggerLevel Level
	// FileEnabled makes the framework log to a file
	// the fields below can be skipped if this value is false!
	FileEnabled bool
	// ConsoleEnabled makes the framework log to console
	ConsoleEnabled bool
	// CallerEnabled makes the caller log to a file
	CallerEnabled bool
	// CallerSkip increases the number of callers skipped by caller
	CallerSkip int
	// Directory to log to to when FileEnabled is enabled
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

// ConfigOption is a function that configures the logger
type ConfigOption func(*Config)

func WithLoggerLevel(level Level) ConfigOption {
	return func(c *Config) {
		c.LoggerLevel = level
	}
}

func WithFileEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.FileEnabled = enabled
	}
}

func WithConsoleEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.ConsoleEnabled = enabled
	}
}

func WithCallerEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.CallerEnabled = enabled
	}
}

func WithCallerSkip(skip int) ConfigOption {
	return func(c *Config) {
		c.CallerSkip = skip
	}
}

func WithDirectory(dir string) ConfigOption {
	return func(c *Config) {
		c.Directory = dir
	}
}

func WithFilename(filename string) ConfigOption {
	return func(c *Config) {
		c.Filename = filename
	}
}

func WithMaxSize(size int) ConfigOption {
	return func(c *Config) {
		c.MaxSize = size
	}
}

func WithMaxBackups(backups int) ConfigOption {
	return func(c *Config) {
		c.MaxBackups = backups
	}
}

func WithMaxAge(age int) ConfigOption {
	return func(c *Config) {
		c.MaxAge = age
	}
}

func WithConsoleInfoStream(stream *os.File) ConfigOption {
	return func(c *Config) {
		c.ConsoleInfoStream = stream
	}
}

func WithConsoleErrorStream(stream *os.File) ConfigOption {
	return func(c *Config) {
		c.ConsoleErrorStream = stream
	}
}

func WithConsoleDebugStream(stream *os.File) ConfigOption {
	return func(c *Config) {
		c.ConsoleDebugStream = stream
	}
}
