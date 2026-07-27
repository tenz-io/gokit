package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Entry 是主要的日志接口。
//
// 每个 With* 方法都返回一个新的 Entry,它继承父级的 field 而不修改父级,
// 因此 Entry 可安全跨 goroutine 共享,也适合派生每请求的子 entry。
type Entry interface {
	// 无后缀方法为 print 风格,f 方法为 printf 风格,w 方法
	// 接受交替的结构化键值 field。
	Debug(args ...any)
	Debugf(format string, args ...any)
	Debugw(msg string, fields ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Infow(msg string, fields ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Warnw(msg string, fields ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Errorw(msg string, fields ...any)

	// With 返回一个追加了给定结构化 field 的子 Entry。
	// field 为交替的键值对:With("user", "bob", "age", 42)。
	With(args ...any) Entry

	// WithError 追加一个取值为 err.Error() 的 "error" field。当 err 为 nil 时
	// 返回同一个 entry。
	WithError(err error) Entry

	// WithRequestID 设置 request_id field。WithTracing 是其别名。
	WithRequestID(id string) Entry
	WithTracing(id string) Entry

	// WithFields 从 Fields map 追加多个结构化 field。
	WithFields(fields Fields) Entry
	// WithField 追加单个结构化 field。
	WithField(k string, v any) Entry

	// StartTraffic 为请求/响应日志开启一个 traffic span。
	// 未配置 traffic 日志时返回 nil。
	StartTraffic(cmd string) *TrafficRec

	// Enabled 报告该 entry 是否会在给定级别下产生日志。
	Enabled(lvl Level) bool

	// SetLevel 在运行时调整该 entry 所属 core 的 log level。
	SetLevel(lvl Level)

	// GetLevel 返回该 entry 所属 core 的当前 log level。
	GetLevel() Level

	// Logger 返回底层 *zap.SugaredLogger。
	Logger() *zap.SugaredLogger
}

// logEntry 基于 *zap.SugaredLogger 实现 Entry,其底层 core 的 level 由
// zap.AtomicLevel 管控(从而使 SetLevel 生效)。
type logEntry struct {
	base    *zap.SugaredLogger
	level   zap.AtomicLevel
	traffic *zap.SugaredLogger // traffic 禁用时为 nil
	trimmer *OutputTrimmer
}

// --- Entry 方法 ---

func (e *logEntry) Logger() *zap.SugaredLogger { return e.base }

func (e *logEntry) Debug(args ...any)            { e.base.Debug(args...) }
func (e *logEntry) Debugf(f string, args ...any) { e.base.Debugf(f, args...) }
func (e *logEntry) Debugw(msg string, fields ...any) {
	e.base.Debugw(msg, e.trimmer.TrimFields(fields)...)
}
func (e *logEntry) Info(args ...any)            { e.base.Info(args...) }
func (e *logEntry) Infof(f string, args ...any) { e.base.Infof(f, args...) }
func (e *logEntry) Infow(msg string, fields ...any) {
	e.base.Infow(msg, e.trimmer.TrimFields(fields)...)
}
func (e *logEntry) Warn(args ...any)            { e.base.Warn(args...) }
func (e *logEntry) Warnf(f string, args ...any) { e.base.Warnf(f, args...) }
func (e *logEntry) Warnw(msg string, fields ...any) {
	e.base.Warnw(msg, e.trimmer.TrimFields(fields)...)
}
func (e *logEntry) Error(args ...any)            { e.base.Error(args...) }
func (e *logEntry) Errorf(f string, args ...any) { e.base.Errorf(f, args...) }
func (e *logEntry) Errorw(msg string, fields ...any) {
	e.base.Errorw(msg, e.trimmer.TrimFields(fields)...)
}

func (e *logEntry) With(args ...any) Entry {
	if len(args) == 0 {
		return e
	}
	trimmed := e.trimmer.TrimArgs(args)
	return &logEntry{
		base:    e.base.With(trimmed...),
		level:   e.level,
		traffic: e.traffic,
		trimmer: e.trimmer,
	}
}

func (e *logEntry) WithError(err error) Entry {
	if err == nil {
		return e
	}
	return e.With("error", err.Error())
}

func (e *logEntry) WithRequestID(id string) Entry { return e.WithTracing(id) }

func (e *logEntry) WithTracing(id string) Entry {
	if id == "" {
		return e
	}
	return e.With("request_id", id)
}

func (e *logEntry) WithFields(fields Fields) Entry {
	if len(fields) == 0 {
		return e
	}
	return e.With(fields.toArgs()...)
}

func (e *logEntry) WithField(k string, v any) Entry { return e.With(k, v) }

func (e *logEntry) StartTraffic(cmd string) *TrafficRec {
	if e.traffic == nil {
		return nil
	}
	return startTrafficRec(e.traffic, cmd, e.trimmer)
}

func (e *logEntry) Enabled(lvl Level) bool {
	return e.level.Enabled(zapcore.Level(lvl))
}

func (e *logEntry) SetLevel(lvl Level) {
	e.level.SetLevel(zapcore.Level(lvl))
}

func (e *logEntry) GetLevel() Level {
	return Level(e.level.Level())
}

// newEntry 从 cfg 构建一个 logEntry。AtomicLevel 被接入 core(作为
// LevelEnabler),从而 SetLevel/GetLevel 真正生效,而非 v2 中那样
// AtomicLevel 脱离连接使 SetLevel 成为 no-op。
func newEntry(cfg Config) *logEntry {
	validateConfig(&cfg)

	enc := newEncoder(cfg.Encoding)
	level := zap.NewAtomicLevelAt(zapcore.Level(cfg.Level))

	core := buildCore(cfg, enc, &level)

	var logger *zap.Logger
	if cfg.Caller {
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(cfg.CallerSkip+1))
	} else {
		logger = zap.New(core)
	}

	trimmer := newTrimmer(cfg.Trimmer)

	var traffic *zap.SugaredLogger
	if cfg.Traffic {
		traffic = newTrafficCore(cfg)
	}

	return &logEntry{
		base:    logger.Sugar(),
		level:   level,
		traffic: traffic,
		trimmer: trimmer,
	}
}
