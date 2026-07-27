package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// buildCore 从 console 与 file sink 组装主日志 core。
//
// 每个 sink 自带独立的 LevelEnabler,从而按严重级别拆分文件输出:
// debug.log 接收 Debug 及以上,info.log 接收 Info 及以上,
// warn.log 接收 Warn 及以上,error.log 接收 Error 及以上。console 输出将
// Info/Debug 发往 stdout、Warn/Error 发往 stderr,各自带有相应阈值。全局
// AtomicLevel(以 *level 传入)统一管控所有 sink,因此调高它即可在所有
// sink 上一致地抑制更低严重级别的输出。
func buildCore(cfg Config, enc zapcore.Encoder, level *zap.AtomicLevel) zapcore.Core {
	sinks := buildSinks(cfg)
	if len(sinks) == 0 {
		// 尊重显式禁用的 console。未配置任何输出目标的 logger 有意保持
		// 静默,而不是泄露到 stdout。
		return zapcore.NewNopCore()
	}

	cores := make([]zapcore.Core, 0, len(sinks))
	for _, s := range sinks {
		s := s // 按次迭代捕获(兼容 Go1.22 之前的循环变量语义)
		// 实际阈值为 sink 自身 enabler 与全局 AtomicLevel 二者中更严格者,
		// 这样 SetLevel 仍能在一个本会接受所有级别的 sink 上限制输出。
		enabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
			return level.Enabled(l) && s.enabler.Enabled(l)
		})
		cores = append(cores, zapcore.NewCore(enc, zapcore.AddSync(s.ws), enabler))
	}
	return zapcore.NewTee(cores...)
}

// sink 将输出目标与其接受的最低级别配对。
type sink struct {
	ws      zapcore.WriteSyncer
	enabler zapcore.LevelEnabler
}

// buildSinks 组装 cfg 所描述的 console 与 file sink。
func buildSinks(cfg Config) []sink {
	var sinks []sink

	if cfg.Console {
		// stdout:Debug 与 Info(低于 Warn 的所有级别)。
		sinks = append(sinks, sink{
			ws:      os.Stdout,
			enabler: zap.LevelEnablerFunc(func(l zapcore.Level) bool { return l < zapcore.WarnLevel }),
		})
		// stderr:Warn 与 Error。
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

// fileSinks 按严重级别返回一个由 lumberjack 支撑的 sink,每个 sink 拥有
// 独立的 LevelEnabler。文件是累积的:debug.log 接收所有级别,
// 而 error.log 仅接收 Error 条目。
func fileSinks(cfg Config) []sink {
	type fileSpec struct {
		name string
		min  zapcore.Level
	}
	specs := []fileSpec{
		{"debug", zapcore.DebugLevel}, // Debug+(全部)
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

// validateConfig 为所有未设置的数值字段填充默认值。
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
