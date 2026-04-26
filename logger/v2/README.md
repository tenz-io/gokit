# logger

基于 zap 的结构化日志。四级日志、文件轮转、context 传递。

```go
import "github.com/tenz-io/gokit/logger/v2"
```

## 快速开始

```go
logger.Configure(logger.Config{Level: logger.InfoLevel, Console: true})
logger.Info("server started", "port", 8080)
userLog := logger.With("user_id", "u123")
userLog.Info("login")
```
