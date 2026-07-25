# cache

缓存抽象（`Manager` 接口），支持可插拔后端：内存 map、泛型 LRU（带 TTL）、Redis（含可观测性钩子）。

## 功能特性

- `Manager` 接口统一 `Get`/`Set`/`SetNx`/`Del`/`Expire`/`Eval` 等操作，业务代码可跨后端复用
- `GetBlob`/`SetBlob` 内置 JSON 编解码，直接存取结构体，无需手动 Marshal
- `NewLocal` 提供进程内 map 缓存，后台协程定期清理过期 key，零依赖、无需外部组件
- `NewLRU` 基于泛型 LRU（`cache/v2/lru` 子包）实现容量限制 + TTL 过期，支持淘汰回调 `onEvict`
- `NewRedis` 包装 `go-redis` 客户端，`Eval` 支持 Lua 脚本，`SetNx` 用于分布式锁等场景
- `NewInterceptor`/`NewInterceptorWithOpts` 为 Redis 客户端挂载指标（`monitor`）和流量日志（`logger`/`tracer`）钩子，通过 `WithEnableTraffic`/`WithEnableMetrics` 控制开关
- `MockManager`（`manager_mock.go`）基于 testify/mock 生成，便于业务层单测替身

## 快速开始

```go
import "github.com/tenz-io/gokit/cache/v2"
```

```go
ctx := context.Background()

// 使用内存缓存
mgr := cache.NewLocal()
_ = mgr.Set(ctx, "key", "value", 5*time.Minute)
val, err := mgr.Get(ctx, "key")

// 结构体存取
type User struct {
    Name string
}
_ = mgr.SetBlob(ctx, "user:1", &User{Name: "tom"}, time.Minute)
var u User
_ = mgr.GetBlob(ctx, "user:1", &u)
```

Redis 后端 + 可观测性拦截器：

```go
client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

interceptor := cache.NewInterceptorWithOpts(
    cache.WithEnableMetrics(true),
    cache.WithEnableTraffic(true),
)
interceptor.Apply(client)

mgr := cache.NewRedis(client)
_ = mgr.Set(ctx, "key", "value", time.Minute)
```

## API 速查

| 符号 | 说明 |
| --- | --- |
| `Manager` | 缓存操作接口：`Get`/`Set`/`SetNx`/`GetBlob`/`SetBlob`/`Del`/`Expire`/`Eval` |
| `NewLocal() Manager` | 创建进程内 map 缓存，自带过期清理协程 |
| `NewLRU(capability int, onEvict func(key string, val []byte), expire time.Duration) Manager` | 创建带容量与 TTL 的 LRU 缓存 |
| `NewRedis(client *redis.Client) Manager` | 用现有 `go-redis` 客户端创建 Redis 缓存 |
| `Interceptor` | Redis 客户端拦截器接口，`Apply(client *redis.Client)` 挂载钩子 |
| `NewInterceptor(config Config) Interceptor` | 用指定 `Config` 创建拦截器 |
| `NewInterceptorWithOpts(opts ...ConfigOption) Interceptor` | 用 `ConfigOption` 函数式选项创建拦截器 |
| `Config` / `ConfigOption` | 拦截器配置及其选项类型 |
| `WithEnableTraffic(enable bool) ConfigOption` | 开关 Redis 命令流量日志钩子 |
| `WithEnableMetrics(enable bool) ConfigOption` | 开关 Redis 命令指标采集钩子 |
| `MockManager` / `NewMockManager(t) *MockManager` | `Manager` 的测试替身（testify/mock 生成） |
| `ErrNotFound` / `ErrInActive` / `ErrNotSupported` | 预定义错误：key 不存在 / 实例未初始化 / 操作不支持 |
| `lru.New[K, V](capability int, onEvict func(K, V), expire time.Duration) *Cache[K, V]` | 泛型 LRU 缓存构造函数（`cache/v2/lru` 子包） |

引入路径：`github.com/tenz-io/gokit/cache/v2`
