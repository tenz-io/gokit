// Package logger 提供基于 go.uber.org/zap 构建的、结构化且分级别的日志 API。
//
// v3 是一次全新重写,不带任何向后兼容垫片。它支持:
//   - 四个 log level:Debug、Info、Warn、Error
//   - console 输出(默认)以及按级别拆分并带 rotation 的文件输出
//   - 通过 With()/WithError()/WithRequestID() 链式追加结构化 field
//   - context 传播:将一个 Entry 绑定到 context.Context,并在调用链中
//     取回
//   - traffic 日志:记录 start/end span,将 cmd、cost、code 与 resp 写入
//     独立的 traffic.log
//   - 对大字符串/slice/map/struct 的输出裁剪
//   - 通过 SetLevel/GetLevel 在运行时修改级别
//
// 快速上手:
//
//	logger.Configure(logger.Config{
//	    Level:   logger.InfoLevel,
//	    Console: true,
//	})
//	logger.Infow("server started", "port", 8080)
package logger

import (
	"sync/atomic"

	"go.uber.org/zap/zapcore"
)

// Fields 是附加到 log entry 上的结构化键值对 map。
type Fields map[string]any

func (f Fields) toArgs() []any {
	if len(f) == 0 {
		return nil
	}
	args := make([]any, 0, len(f)*2)
	for k, v := range f {
		args = append(args, k, v)
	}
	return args
}

// Level 表示一个日志严重级别。
type Level zapcore.Level

const (
	DebugLevel Level = Level(zapcore.DebugLevel)
	InfoLevel  Level = Level(zapcore.InfoLevel)
	WarnLevel  Level = Level(zapcore.WarnLevel)
	ErrorLevel Level = Level(zapcore.ErrorLevel)
)

// Encoding 选择 logger 的输出格式。
type Encoding string

const (
	// ConsoleEncoding (默认)写入对人类友好的 "LEVEL time msg key=value" 行。
	ConsoleEncoding Encoding = "console"
	// JSONEncoding 每行写入一个 JSON 对象,适合日志聚合。
	JSONEncoding Encoding = "json"
)

// WriterSyncer 是 zapcore.WriteSyncer 的别名。
type WriterSyncer = zapcore.WriteSyncer

// LevelEnabler 是 zapcore.LevelEnabler 的别名。
type LevelEnabler = zapcore.LevelEnabler

// Config 配置一个 logger 实例。初始全局 logger 与 functional-option 构造函数
// 使用文档中记载的默认值。使用 struct 构造函数时,布尔字段需显式指定:
// Console:false 表示不输出到 console。
type Config struct {
	// Level 设置最低 log level。默认:InfoLevel。
	Level Level

	// Encoding 在 console 与 JSON 输出之间选择。默认:console。
	Encoding Encoding

	// Console 启用向 stdout(Info/Debug)与 stderr(Warn/Error)的日志输出。
	// 默认:true。
	Console bool

	// FilePath 启用向指定目录的文件日志。设置后,日志文件会创建在
	// FilePath/<level>.log 下,按严重级别拆分:
	// debug.log(Debug+)、info.log(Info+)、warn.log(Warn+)、error.log(Error+)。
	// 默认:""(禁用)。
	FilePath string

	// MaxSize 是日志文件被 rotation 前的最大体积(MB)。默认:100。
	MaxSize int
	// MaxAge 是已 rotation 日志文件的最大保留天数。默认:7。
	MaxAge int
	// MaxBackups 是保留的已 rotation 文件最大数量。默认:10。
	MaxBackups int

	// Caller 在每条 log entry 中加入调用方的文件与行号。
	Caller bool
	// CallerSkip 增加跳过的调用者层数(默认 0)。
	CallerSkip int

	// Traffic 启用 traffic logger 组件,它会写入独立的
	// traffic.log 来记录请求/响应 span。
	Traffic bool
	// TrafficPath 覆盖 traffic.log 的目录。为空时回退到
	// FilePath,仍为空则回退到 "log"。
	TrafficPath string
	// TrafficMaxSize/MaxAge/MaxBackups 覆盖 traffic 日志的 rotation 设置。
	// 为零时回退到主 MaxSize/MaxAge/MaxBackups。
	TrafficMaxSize    int
	TrafficMaxAge     int
	TrafficMaxBackups int

	// Trimmer 配置对大型 field 的输出裁剪。为 nil 时,应用合理默认值
	// (arr=3、str=128、depth=10)。
	Trimmer *TrimConfig
}

// TrimConfig 控制输出裁剪。
type TrimConfig struct {
	ArrLimit  int      // 从 slice/array 保留的最大元素数(默认 3)
	StrLimit  int      // 从 string 保留的最大字节数(默认 128)
	DeepLimit int      // struct/map 的最大嵌套深度(默认 10)
	Ignores   []string // 完全丢弃的 field 名
}

// defaultConfig 是未提供配置时应用的配置。
var defaultConfig = Config{
	Level:             InfoLevel,
	Encoding:          ConsoleEncoding,
	Console:           true,
	MaxSize:           100,
	MaxAge:            7,
	MaxBackups:        10,
	TrafficMaxSize:    100,
	TrafficMaxAge:     7,
	TrafficMaxBackups: 10,
}

// ConfigOption 是用于 Configure / NewEntry 的 functional option。
type ConfigOption func(*Config)

// --- 配置项 ---

func WithLevel(lvl Level) ConfigOption           { return func(c *Config) { c.Level = lvl } }
func WithEncoding(enc Encoding) ConfigOption     { return func(c *Config) { c.Encoding = enc } }
func WithConsole(on bool) ConfigOption           { return func(c *Config) { c.Console = on } }
func WithFilePath(dir string) ConfigOption       { return func(c *Config) { c.FilePath = dir } }
func WithMaxSize(mb int) ConfigOption            { return func(c *Config) { c.MaxSize = mb } }
func WithMaxAge(days int) ConfigOption           { return func(c *Config) { c.MaxAge = days } }
func WithMaxBackups(n int) ConfigOption          { return func(c *Config) { c.MaxBackups = n } }
func WithCaller(on bool) ConfigOption            { return func(c *Config) { c.Caller = on } }
func WithCallerSkip(skip int) ConfigOption       { return func(c *Config) { c.CallerSkip = skip } }
func WithTraffic(on bool) ConfigOption           { return func(c *Config) { c.Traffic = on } }
func WithTrafficPath(dir string) ConfigOption    { return func(c *Config) { c.TrafficPath = dir } }
func WithTrafficMaxSize(mb int) ConfigOption     { return func(c *Config) { c.TrafficMaxSize = mb } }
func WithTrafficMaxAge(days int) ConfigOption    { return func(c *Config) { c.TrafficMaxAge = days } }
func WithTrafficMaxBackups(n int) ConfigOption   { return func(c *Config) { c.TrafficMaxBackups = n } }
func WithTrimConfig(tc *TrimConfig) ConfigOption { return func(c *Config) { c.Trimmer = tc } }

// --- 构造 ---

// Configure 初始化全局 logger。
func Configure(config Config) {
	global.Store(newEntry(config))
}

// ConfigureWithOpts 使用 functional option 初始化全局 logger。
func ConfigureWithOpts(opts ...ConfigOption) {
	cfg := defaultConfig
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	Configure(cfg)
}

// NewEntry 创建一个不影响全局 logger 的独立 Entry。
func NewEntry(config Config) Entry { return newEntry(config) }

// NewEntryWithOpts 使用 functional option 创建一个独立 Entry。
func NewEntryWithOpts(opts ...ConfigOption) Entry {
	cfg := defaultConfig
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	return NewEntry(cfg)
}

// --- 全局状态 ---

var global atomic.Pointer[logEntry]

func init() {
	global.Store(newEntry(defaultConfig))
}

func current() *logEntry { return global.Load() }

// L 返回全局 logger entry。
func L() Entry { return current() }

// SetLevel 在运行时调整全局 log level。与 v2 实现不同,该 level 直接
// 接入 core,因此对所有后续调用立即生效。
func SetLevel(lvl Level) {
	current().SetLevel(lvl)
}

// GetLevel 返回当前全局 log level。
func GetLevel() Level {
	return current().GetLevel()
}

// --- 包级便捷函数 ---

// 包级别的日志函数直接调用底层 SugaredLogger。这保证了调用者信息正确:
// 这里有一个包装帧,与对应 Entry 方法中的那一个包装帧保持一致。
func Debug(args ...any)                 { current().base.Debug(args...) }
func Debugf(format string, args ...any) { current().base.Debugf(format, args...) }
func Debugw(msg string, fields ...any) {
	e := current()
	e.base.Debugw(msg, e.trimmer.TrimFields(fields)...)
}
func Info(args ...any)                 { current().base.Info(args...) }
func Infof(format string, args ...any) { current().base.Infof(format, args...) }
func Infow(msg string, fields ...any) {
	e := current()
	e.base.Infow(msg, e.trimmer.TrimFields(fields)...)
}
func Warn(args ...any)                 { current().base.Warn(args...) }
func Warnf(format string, args ...any) { current().base.Warnf(format, args...) }
func Warnw(msg string, fields ...any) {
	e := current()
	e.base.Warnw(msg, e.trimmer.TrimFields(fields)...)
}
func Error(args ...any)                 { current().base.Error(args...) }
func Errorf(format string, args ...any) { current().base.Errorf(format, args...) }
func Errorw(msg string, fields ...any) {
	e := current()
	e.base.Errorw(msg, e.trimmer.TrimFields(fields)...)
}

func With(args ...any) Entry              { return current().With(args...) }
func WithFields(fields Fields) Entry      { return current().WithFields(fields) }
func WithField(k string, v any) Entry     { return current().WithField(k, v) }
func WithError(err error) Entry           { return current().WithError(err) }
func WithRequestID(id string) Entry       { return current().WithRequestID(id) }
func WithTracing(id string) Entry         { return current().WithTracing(id) }
func Enabled(lvl Level) bool              { return current().Enabled(lvl) }
func StartTraffic(cmd string) *TrafficRec { return current().StartTraffic(cmd) }
