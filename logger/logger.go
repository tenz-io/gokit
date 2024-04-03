package logger

import "go.uber.org/zap/zapcore"

type Fields map[string]any

type Level zapcore.Level

// WriterSyncer is an interface that groups the Write and Sync methods.
type WriterSyncer = zapcore.WriteSyncer

const (
	// DebugLevel enum -1: logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel = Level(zapcore.DebugLevel)

	// InfoLevel enum 0: is the default logging priority.
	InfoLevel = Level(zapcore.InfoLevel)

	// WarnLevel enum 1: logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = Level(zapcore.WarnLevel)

	// ErrorLevel enum 2: logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-defaultLevel logs.
	ErrorLevel = Level(zapcore.ErrorLevel)
)
