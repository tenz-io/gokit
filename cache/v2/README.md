# cache

缓存抽象（Manager 接口），支持内存 map、泛型 LRU、Redis 三种后端。

```go
import "github.com/tenz-io/gokit/cache/v2"
```

## 快速开始

```go
mgr := cache.NewLocal()
mgr.Set(ctx, "key", "value", 5*time.Minute)
val, err := mgr.Get(ctx, "key")
```
