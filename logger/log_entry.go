package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

const (
	defaultLogEmpty     = "-"   // defaultLogEmpty of fields of the log record
	defaultErrFieldName = "err" // defaultErrFieldName of error field of the log record
	defaultSeparator    = "|"   // defaultSeparator of fields of the log record
)

type LogEntry struct {
	infoLogger  *zap.Logger
	errLogger   *zap.Logger
	debugLogger *zap.Logger

	requestId string
}

func newLogEntry(le *LogEntry, fields Fields) *LogEntry {
	if !le.validate() {
		return le
	}

	args := toZapFields(fields)

	return &LogEntry{
		infoLogger:  le.infoLogger.With(args...),
		errLogger:   le.errLogger.With(args...),
		debugLogger: le.debugLogger.With(args...),
		requestId:   le.requestId,
	}
}

func getLogEntry(infoLogger, errLogger, debugLogger *zap.Logger) *LogEntry {
	return &LogEntry{
		infoLogger:  infoLogger,
		errLogger:   errLogger,
		debugLogger: debugLogger,
	}
}

// Debug logs a message at DebugLevel.
func (le *LogEntry) Debug(msg string) {
	if !le.Enabled(DebugLevel) {
		return
	}

	le.debugLogger.Debug(le.withHead(msg))
}

// Debugf logs a message at DebugLevel.
func (le *LogEntry) Debugf(format string, args ...any) {
	if !le.Enabled(DebugLevel) {
		return
	}

	le.debugLogger.Debug(le.withHead(fmt.Sprintf(format, args...)))
}

// DebugWith logs a message with fields at DebugLevel.
func (le *LogEntry) DebugWith(msg string, fields Fields) {
	if !le.Enabled(DebugLevel) {
		return
	}
	le.debugLogger.Debug(le.withHead(msg), toZapFields(fields)...)
}

// Info logs a message at InfoLevel.
func (le *LogEntry) Info(msg string) {
	if !le.Enabled(InfoLevel) {
		return
	}
	le.infoLogger.Info(le.withHead(msg))
}

func (le *LogEntry) Infof(format string, args ...any) {
	if !le.Enabled(InfoLevel) {
		return
	}

	le.infoLogger.Info(le.withHead(fmt.Sprintf(format, args...)))
}

// InfoWith logs a message with fields at InfoLevel.
func (le *LogEntry) InfoWith(msg string, fields Fields) {
	if !le.Enabled(InfoLevel) {
		return
	}
	le.infoLogger.Info(le.withHead(msg), toZapFields(fields)...)
}

// Warn logs a message at WarnLevel.
func (le *LogEntry) Warn(msg string) {
	if !le.Enabled(WarnLevel) {
		return
	}
	le.errLogger.Warn(le.withHead(msg))
}

func (le *LogEntry) Warnf(format string, args ...any) {
	if !le.Enabled(WarnLevel) {
		return
	}

	le.errLogger.Warn(le.withHead(fmt.Sprintf(format, args...)))
}

// WarnWith logs a message with fields at WarnLevel.
func (le *LogEntry) WarnWith(msg string, fields Fields) {
	if !le.Enabled(WarnLevel) {
		return
	}
	le.errLogger.Warn(le.withHead(msg), toZapFields(fields)...)
}

// Error logs a message at ErrorLevel.
func (le *LogEntry) Error(msg string) {
	if !le.Enabled(ErrorLevel) {
		return
	}
	le.errLogger.Error(le.withHead(msg))
}

func (le *LogEntry) Errorf(format string, args ...any) {
	if !le.Enabled(ErrorLevel) {
		return
	}

	le.errLogger.Error(le.withHead(fmt.Sprintf(format, args...)))
}

// ErrorWith logs a message with fields at ErrorLevel.
func (le *LogEntry) ErrorWith(msg string, fields Fields) {
	if !le.Enabled(ErrorLevel) {
		return
	}
	le.errLogger.Error(le.withHead(msg), toZapFields(fields)...)
}

// With binds a default field to a log message
func (le *LogEntry) With(data any) Entry {
	return le.WithField(defaultLogEmpty, data)
}

// WithError binds a default error field to a log message
func (le *LogEntry) WithError(err error) Entry {
	return le.WithField(defaultErrFieldName, err)
}

// WithField binds a field to a log message
func (le *LogEntry) WithField(k string, v any) Entry {
	return le.WithFields(Fields{k: v})
}

// WithFields Add a map of fields to the Entry.
func (le *LogEntry) WithFields(fields Fields) Entry {
	return newLogEntry(le, fields)
}

// WithTracing create copy of LogEntry with tracing.Span
func (le *LogEntry) WithTracing(requestId string) Entry {
	if !le.validate() {
		return le
	}
	return &LogEntry{
		infoLogger:  le.infoLogger,
		errLogger:   le.errLogger,
		debugLogger: le.debugLogger,
		requestId:   requestId,
	}
}

func (le *LogEntry) Enabled(level Level) bool {
	if le == nil {
		return false
	}
	switch level {
	case DebugLevel:
		return GetLevel() <= DebugLevel && le.debugLogger != nil
	case InfoLevel:
		return GetLevel() <= InfoLevel && le.infoLogger != nil
	case WarnLevel:
		return GetLevel() <= WarnLevel && le.errLogger != nil
	case ErrorLevel:
		return GetLevel() <= ErrorLevel && le.errLogger != nil
	default:
		return false
	}
}

func (le *LogEntry) withHead(msg string) string {
	if le == nil {
		return strings.Join(append([]string{
			defaultLogEmpty,
			msg,
		}), defaultSeparator)
	}
	if le.requestId == "" {
		return strings.Join(append([]string{
			defaultLogEmpty,
			msg,
		}), defaultSeparator)
	}
	return strings.Join(append([]string{
		le.requestId,
		msg,
	}), defaultSeparator)
}

func (le *LogEntry) validate() bool {
	if le == nil {
		return false
	}

	return true
}

func (le *LogEntry) clone() *LogEntry {
	if le == nil {
		return nil
	}

	return &LogEntry{
		debugLogger: le.debugLogger,
		infoLogger:  le.infoLogger,
		errLogger:   le.errLogger,
		requestId:   le.requestId,
	}
}
