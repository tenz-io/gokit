# logger

基于 `go.uber.org/zap` 的结构化日志库。支持四级日志、文件轮转、结构化字段、请求追踪、流量日志和 context 传递。

## 快速开始

```go
import "github.com/tenz-io/gokit/logger"

// 1. 配置（通常只在启动时调用一次）
logger.Configure(logger.Config{
    Level:    logger.InfoLevel,
    Console:  true,
    FilePath: "log",         // 启用文件日志
    Traffic:  true,          // 启用流量日志
    Caller:   true,          // 显示调用位置
})

// 2. 日志（支持 zap sugared 风格，交替 key=value 对）
logger.Info("server started", "port", 8080, "env", "prod")
logger.Infof("listening on :%d", 8080)
logger.Debug("cache miss", "key", "user:42")

// 3. 持久化字段（返回子 logger）
userLog := logger.With("user_id", "u123", "session", "abc")
userLog.Info("login")
userLog.Info("update profile", "fields", 3)

// 4. 错误和请求 ID
logger.WithError(err).Warn("retry")
logger.WithRequestID("req-abc").Info("processing")

// 5. 流量日志
rec := logger.StartTraffic("getUser")
defer rec.End(user, "200")
```

## API 参考

### 全局配置

```go
type Config struct {
    Level       Level       // 最低日志级别。默认 InfoLevel
    Console     bool        // 输出到控制台。默认 true
    FilePath    string      // 文件日志目录。空=不启用。默认 ""
    MaxSize     int         // 单文件最大 MB。默认 100
    MaxAge      int         // 文件保留天数。默认 7
    MaxBackups  int         // 轮转文件保留数。默认 10
    Caller      bool        // 记录调用位置。默认 false
    CallerSkip  int         // 调用栈跳过层数。默认 1
    Traffic     bool        // 启用流量日志。默认 false
    TrafficPath string      // 流量日志目录。空=使用 FilePath
    TrafficMaxSize    int   // 流量文件最大 MB
    TrafficMaxAge     int   // 流量文件保留天数
    TrafficMaxBackups int   // 流量轮转文件保留数
    Trimmer     *TrimConfig // 输出裁剪配置
}

type TrimConfig struct {
    ArrLimit  int      // 数组/切片最大展示元素数。默认 3
    StrLimit  int      // 字符串最大长度。默认 128
    DeepLimit int      // 结构体最大嵌套深度。默认 10
    Ignores   []string // 忽略的字段名列表
}
```

### Entry 接口

| 方法 | 说明 |
|------|------|
| `Debug(args ...any)` | Debug 级别日志 |
| `Debugf(format string, args ...any)` | Debug 格式化 |
| `Info(args ...any)` | Info 级别日志 |
| `Infof(format string, args ...any)` | Info 格式化 |
| `Warn(args ...any)` | Warn 级别日志 |
| `Warnf(format string, args ...any)` | Warn 格式化 |
| `Error(args ...any)` | Error 级别日志 |
| `Errorf(format string, args ...any)` | Error 格式化 |
| `With(args ...any) Entry` | 创建带持久化字段的子 logger（偶数个参数，交替 key=value） |
| `WithFields(fields Fields) Entry` | 通过 Fields map 添加字段 |
| `WithField(k string, v any) Entry` | 添加单个字段 |
| `WithError(err error) Entry` | 添加 error 字段 |
| `WithRequestID(id string) Entry` | 添加 request_id 字段 |
| `WithTracing(id string) Entry` | WithRequestID 别名 |
| `StartTraffic(cmd string) *TrafficRec` | 开始流量记录 |
| `Enabled(level Level) bool` | 判断级别是否启用 |
| `Logger() *zap.SugaredLogger` | 获取底层 zap logger |

### 流量记录

```go
type TrafficRec struct{ ... }

// 开始记录
rec := entry.StartTraffic("getUser")
// 结束记录
rec.End(responseData, "200")            // 或
rec.EndWithError(err)                   // 错误响应
```

### 便捷函数

| 函数 | 说明 |
|------|------|
| `L() Entry` | 获取全局 logger |
| `Info/Debug/Warn/Error/Infof/...` | 包级日志函数 |
| `With/WithFields/WithError/WithRequestID` | 包级持久化字段 |
| `Configure/ConfigureWithOpts` | 全局配置 |
| `SetLevel(level Level)` | 运行时修改级别 |
| `GetLevel() Level` | 获取当前级别 |
| `StartTraffic(cmd string) *TrafficRec` | 包级流量记录 |

### Context 传递

| 函数 | 说明 |
|------|------|
| `FromContext(ctx) Entry` | 从 context 获取 logger（无则返回全局） |
| `WithLogger(ctx, entry)` | 将 logger 放入 context |
| `CopyToContext(src, dst)` | 在 context 间拷贝 logger |

## 最佳实践

### 生产配置

```go
logger.Configure(logger.Config{
    Level:      logger.InfoLevel,   // 生产环境关闭 Debug
    Console:    true,                // 配合容器日志收集
    FilePath:   "/var/log/myapp",    // 持久化日志
    MaxSize:    500,                 // 大文件上限
    MaxAge:     30,                  // 保留一个月
    MaxBackups: 5,
    Caller:     true,                // 记录调用位置辅助排查
    Traffic:    true,                // 启用流量日志
})
```

### 避免热路径创建子 Logger

`With()` 会创建新的 `*zap.SugaredLogger`，涉及内存分配。不要在每请求都创建，应在中间件层创建一次并放入 context：

```go
// 中间件
ctx := logger.WithLogger(r.Context(), 
    logger.WithRequestID(reqID).With("path", r.URL.Path),
)
// 后续处理直接用 FromContext
logger.FromContext(ctx).Info("processing")
```

### 流量日志

只在需要记录完整请求/响应内容时使用：

```go
rec := logger.FromContext(ctx).StartTraffic("api.getUser")
defer func() {
    if err != nil {
        rec.EndWithError(err)
    } else {
        rec.End(result, "200")
    }
}()
```

### 自定义 Writer

```go
type myWriter struct{}
func (m *myWriter) Write(p []byte) (int, error) { ... }
func (m *myWriter) Sync() error { return nil }

logger.Configure(logger.Config{
    Level:  logger.InfoLevel,
    Console: false,
    Stream: &myWriter{},   // 自定义输出
})
```

## 与 v1 的区别

| 变更项 | v1 | v2 |
|--------|----|----|
| 底层架构 | 3 个 zap.Logger（debug/info/error） | 单一 `*zap.SugaredLogger` |
| `*With` 方法 | `DebugWith/InfoWith/...` | 移除，使用 `.With().Info()` |
| 请求 ID | 拼接到消息体前缀 | zap 结构化字段 `request_id` |
| 配置项 | `WithLoggerLevel` 等 | `WithLevel` 等（旧名保留兼容） |
| 流量 goroutine | 每次异步 `go func()` | 同步写入，无 goroutine 泄露 |
| Policy 限流 | `RateLimitPolicy/SamplingPolicy` | 移除（应用层自行控制） |
| `OutputTrimmer` | 全局可变 `SetupDefaultTrimmer` | 内建不可变 trimer |
| 文件命名 | `info.log/error.log/debug.log` | 单文件轮转 |
| 依赖 | `golang.org/x/time` (rate) | 移除，仅 zap+lumberjack |
| context 传递 | 双 key（log + traffic） | 单 key（log） |
