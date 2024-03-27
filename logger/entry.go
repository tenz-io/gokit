package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

type Fields map[string]any

type Level zapcore.Level

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

type Entry interface {
	// Debug logs a message at DebugLevel.
	Debug(msg string)
	// Debugf logs a message at DebugLevel.
	Debugf(format string, args ...any)
	// DebugWith logs a message with fields at DebugLevel.
	DebugWith(msg string, fields Fields)
	// Info logs a message at InfoLevel.
	Info(msg string)
	// Infof logs a message at InfoLevel.
	Infof(format string, args ...any)
	// InfoWith logs a message with fields at InfoLevel.
	InfoWith(msg string, fields Fields)
	// Warn logs a message at WarnLevel.
	Warn(msg string)
	// Warnf logs a message at WarnLevel.
	Warnf(format string, args ...any)
	// WarnWith logs a message with fields at WarnLevel.
	WarnWith(msg string, fields Fields)
	// Error logs a message at ErrorLevel.
	Error(msg string)
	// Errorf logs a message at ErrorLevel.
	Errorf(format string, args ...any)
	// ErrorWith logs a message with fields at ErrorLevel.
	ErrorWith(msg string, fields Fields)

	// WithFields returns a new entry with after adding fields
	WithFields(fields Fields) Entry
	// WithField returns a new entry with after adding field
	WithField(k string, v any) Entry
	// With returns a new entry with after adding data with default field name
	With(data any) Entry
	// WithError returns a new entry with after adding error
	WithError(err error) Entry
	// WithTracing returns a new entry with after adding requestId
	WithTracing(requestId string) Entry

	// Enabled is entry enabled at level
	Enabled(level Level) bool
}

// validate checks if the given level is valid, only support DebugLevel, InfoLevel, WarnLevel, ErrorLevel
func (l Level) validate() bool {
	switch l {
	case DebugLevel, InfoLevel, WarnLevel, ErrorLevel:
		return true
	default:
		return false
	}
}

// toZapFields converts the fields to zapcore.Field
func toZapFields(fields Fields, ignores ...string) []zapcore.Field {
	if fields == nil {
		return []zapcore.Field{}
	}
	zapFields := make([]zapcore.Field, 0, len(fields))
	for k, v := range fields {
		f := zap.Any(k, v)
		switch typ := f.Type; typ {
		//case zapcore.StringType, zapcore.StringerType:
		//	zapFields = append(zapFields, zap.String(k, utils.StringLimit(fmt.Sprintf("%s", v), maxStringFieldSize)))
		case zapcore.StringType:
			zapFields = append(zapFields, f)
		case zapcore.StringerType,
			zapcore.BinaryType,
			zapcore.ArrayMarshalerType,
			zapcore.ObjectMarshalerType,
			zapcore.ReflectType:
			zapFields = append(zapFields, zap.Any(k, TrimObjectWithOpts(v, WithIgnores(ignores...))))
		default:
			zapFields = append(zapFields, f)
		}
	}
	return zapFields
}

// shortTimeEncoder serializes a time.Time to an short-formatted string
func shortTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.000"))
}

// longTimeEncoder serializes a time.Time to an short-formatted string
func longTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	//enc.AppendString(t.Format(time.RFC3339))
	enc.AppendString(t.Format("2006-01-02T15:04:05.000Z0700"))
}

type empty struct {
}

func (e *empty) Debug(msg string) {
}

func (e *empty) Debugf(format string, args ...any) {
}

func (e *empty) DebugWith(msg string, fields Fields) {
}

func (e *empty) Info(msg string) {
}

func (e *empty) Infof(format string, args ...any) {
}

func (e *empty) InfoWith(msg string, fields Fields) {
}

func (e *empty) Warn(msg string) {
}

func (e *empty) Warnf(format string, args ...any) {
}

func (e *empty) WarnWith(msg string, fields Fields) {
}

func (e *empty) Error(msg string) {
}

func (e *empty) Errorf(format string, args ...any) {
}

func (e *empty) ErrorWith(msg string, fields Fields) {
}

func (e *empty) WithFields(fields Fields) Entry {
	return e
}

func (e *empty) WithField(k string, v any) Entry {
	return e
}

func (e *empty) With(data any) Entry {
	return e
}

func (e *empty) WithError(err error) Entry {
	return e
}

func (e *empty) WithTracing(requestId string) Entry {
	return e
}

func (e *empty) Enabled(level Level) bool {
	return false
}
