package logger

var (
	// defaultTrafficConfig is used for defaultTrafficLogger below only
	defaultTrafficConfig = TrafficConfig{
		Enabled:    true,
		Directory:  "log",
		Filename:   "data.log",
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     7,
		StrLimit:   defaultStrLimit,
		ArrLimit:   defaultArrLimit,
		DeepLimit:  defaultDeepLimit,
		Ignores:    []string{},
		Stream:     nil,
	}
)

// TrafficConfig for traffic logging
type TrafficConfig struct {
	// Enabled makes the framework log traffic
	Enabled bool
	// Directory to log to when Enabled is enabled
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
	// Stream is the stream to log to
	Stream WriterSyncer
}

type TrafficConfigOption func(*TrafficConfig)

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

func WithTrafficStream(stream WriterSyncer) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.Stream = stream
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

func WithTrafficEnabled(enabled bool) TrafficConfigOption {
	return func(c *TrafficConfig) {
		c.Enabled = enabled
	}
}
