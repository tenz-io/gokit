# protoc-gen-go-gin

protoc 插件。从 proto 服务定义生成 Gin HTTP handler。

```go
import "github.com/tenz-io/gokit/protoc-gen-go-gin/v2"
```

## 快速开始

```bash
protoc --go-gin_out=. --go-gin_opt=paths=source_relative service.proto
```
