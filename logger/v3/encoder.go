package logger

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// encodeTime 为每条 entry 打上一个 RFC3339 风格、毫秒精度、带本地时区
// 偏移的时间戳,例如 2006-01-02T15:04:05.000+08:00。
func encodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.000Z0700"))
}

// encoderConfig 返回 console 与 JSON encoder 共用的字段布局。两种 encoder
// 使用相同的 key,因此从 console 切换到 JSON 只改变传输格式,字段名不变。
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

// newEncoder 根据 enc 返回 console 或 JSON encoder。
func newEncoder(enc Encoding) zapcore.Encoder {
	cfg := encoderConfig()
	if enc == JSONEncoding {
		return zapcore.NewJSONEncoder(cfg)
	}
	return zapcore.NewConsoleEncoder(cfg)
}
