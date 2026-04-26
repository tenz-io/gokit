# bloom

概率型布隆过滤器，Murmur3 哈希，可调节误判率。

```go
import "github.com/tenz-io/gokit/bloom/v2"
```

## 快速开始

```go
bf := bloom.NewFilter(100_000, 0.01)
bf.AddString("user-42")
bf.ExistsString("user-42") // true
```
