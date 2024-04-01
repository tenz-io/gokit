package logger

import "os"

var (
	// defaultTrafficConfig is used for defaultTrafficLogger below only
	defaultTrafficConfig = TrafficConfig{
		FileEnabled: false,
		Directory:   "log",
		Filename:    "data.log",
		MaxSize:     100,
		MaxBackups:  10,
		MaxAge:      7,
		StrLimit:    defaultStrLimit,
		ArrLimit:    defaultArrLimit,
		DeepLimit:   defaultDeepLimit,
		Ignores:     []string{},
	}
)

// TrafficConfig for traffic logging
type TrafficConfig struct {
	// FileEnabled makes the framework log to a file
	// the fields below can be skipped if this value is false!
	FileEnabled bool
	// ConsoleEnabled makes the framework log to console
	ConsoleEnabled bool
	// Directory to log to when FileEnabled is enabled
	Directory string
	// Filename is the name of the logfile which will be placed inside the directory
	Filename string
	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int
	// MaxBackups the max number of rolled files to keep
	MaxBackups int
	// MaxAge the max age in days to keep a logfile
	MaxAge int
	// StrLimit the max length of string to log
	StrLimit int
	// ArrLimit the max length of array to log
	ArrLimit int
	// DeepLimit the max depth of struct to log
	DeepLimit int
	// Ignores is a list of fields to ignore when logging
	Ignores []string
	// ConsoleStream
	ConsoleStream *os.File
}

type TrafficConfigOption func(*TrafficConfig)

func WithTrafficFileEnabled(enabled bool) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.FileEnabled = enabled
	}
}

func WithTrafficConsoleEnabled(enabled bool) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.ConsoleEnabled = enabled
	}
}

func WithTrafficDirectory(directory string) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.Directory = directory
	}
}

func WithTrafficFilename(filename string) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.Filename = filename
	}
}

func WithTrafficMaxSize(maxSize int) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.MaxSize = maxSize
	}
}

func WithTrafficMaxBackups(maxBackups int) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.MaxBackups = maxBackups
	}
}

func WithTrafficMaxAge(maxAge int) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.MaxAge = maxAge
	}
}

func WithTrafficConsoleStream(stream *os.File) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.ConsoleStream = stream
	}
}

func WithTrafficStrLimit(strLimit int) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.StrLimit = strLimit
	}
}

func WithTrafficArrLimit(arrLimit int) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.ArrLimit = arrLimit
	}
}

func WithTrafficDeepLimit(deepLimit int) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.DeepLimit = deepLimit
	}
}

func WithTrafficIgnoresOpt(ignores ...string) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.Ignores = ignores
	}
}
