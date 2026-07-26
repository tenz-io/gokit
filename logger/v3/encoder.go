package logger

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// encodeTime stamps each entry with an RFC3339-ish, ms-precision timestamp
// that carries the local offset, e.g. 2006-01-02T15:04:05.000+08:00.
func encodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.000Z0700"))
}

// encoderConfig returns the shared field layout for both console and JSON
// encoders. Both encoders use the same keys, so a switch from console to
// JSON only changes the wire format, not the field names.
func encoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:          "@t",
		LevelKey:         "lvl",
		CallerKey:        "caller",
		MessageKey:       "msg",
		EncodeTime:       encodeTime,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeDuration:   zapcore.NanosDurationEncoder,
		ConsoleSeparator: " ",
	}
}

// newEncoder returns a console or JSON encoder based on enc.
func newEncoder(enc Encoding) zapcore.Encoder {
	cfg := encoderConfig()
	if enc == JSONEncoding {
		return zapcore.NewJSONEncoder(cfg)
	}
	return zapcore.NewConsoleEncoder(cfg)
}
