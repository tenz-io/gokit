# cache

纯本地进程内缓存库：map 缓存与泛型 LRU 缓存，均支持按条目过期（TTL）。零网络、零外部依赖——不需要 Redis，也不需要任何远程组件。

```go
import "github.com/tenz-io/gokit/cache/v3"
```

## 模块介绍

cache 解决两类本地缓存场景：

- **无界 map 缓存**（`NewLocal`）：`map[string]*item` + 读写锁，按条目 TTL 过期，后台协程周期清理已过期 key 避免内存无限增长。
- **容量受限的 LRU 缓存**（`NewLRU`）：基于并发安全的泛型 LRU（`cache/v3/lru` 子包）实现容量上限 + TTL，超出容量按最近最少使用淘汰，支持淘汰回调 `onEvict`。

两者统一实现 `Manager` 接口（`Get`/`Set`/`SetNx`/`GetBlob`/`SetBlob`/`Del`/`Expire`），业务代码换后端时无需改调用方式；`GetBlob`/`SetBlob` 内置 JSON 编解码，直接存取结构体。

V3 相对 V2 的核心变化：

- **纯本地**：删除 Redis 后端、`go-redis` 依赖、Interceptor/Hook、Lua `Eval`。一切在进程内完成，零外部依赖，与 `async/v3` 看齐。
- **可注入时钟**：`WithNow` 注入时钟，过期逻辑用绝对时间判断，单测可不依赖真实 `time.Sleep`，杜绝 flaky。
- **生命周期可控**：`localCache` 新增 `Close()` 停止后台清理协程（v2 的清理协程永不退出，存在泄漏）；LRU 后端惰性过期，无需后台协程。
- **并发安全 LRU**：v2 的 `lru` 子包显式声明"not safe for concurrent access"，v3 加 `sync.Mutex` 修复，可直接跨 goroutine 共享。

## 能力清单

| 能力 | 含义 |
| --- | --- |
| 统一本地缓存接口 | `Manager` 接口抹平 map 与 LRU 两种实现的差异，业务换后端无需改调用方式 |
| 结构体直存直取 | `GetBlob`/`SetBlob` 内部用 JSON 编解码，业务无需手写 `Marshal`/`Unmarshal` 即可缓存结构体 |
| 不存在才写入 | `SetNx` 在 key 不存在（或已过期）时才写入并返回是否已存在，可用于幂等写入等场景 |
| 进程内缓存自动过期清理 | `NewLocal` 的 map 缓存启动后台协程按间隔扫描，删除已过期 key，避免内存无限增长 |
| 容量受限的 LRU 缓存 | `NewLRU` 基于泛型 LRU 实现容量上限 + TTL，超出容量按最近最少使用淘汰，支持 `onEvict` 回调 |
| 惰性过期 | map 与 LRU 后端均在读取时检查过期：命中过期 key 即删除并返回未命中，无需等待后台扫描 |
| 可注入时钟 | `WithNow` 注入时钟，过期判断基于绝对时间，单测可推进时间不依赖真实 sleep |
| 生命周期管理 | `localCache.Close()` 停止后台清理协程，避免泄漏；可重复调用，幂等 |

## 快速开始

```go
package main

import (
	"context"
	"time"

	"github.com/tenz-io/gokit/cache/v3"
)

func main() {
	ctx := context.Background()

	// 本地 map 缓存，后台每秒清理一次过期 key
	mgr := cache.NewLocal(cache.WithEvictInterval(time.Second))
	defer func() { _ = mgr.Close() }()

	_ = mgr.Set(ctx, "key", "value", 5*time.Minute)
	val, _ := mgr.Get(ctx, "key")

	// 结构体存取
	type User struct {
		Name string
		Age  int
	}
	_ = mgr.SetBlob(ctx, "user:1", &User{Name: "tom", Age: 18}, time.Minute)
	var u User
	_ = mgr.GetBlob(ctx, "user:1", &u)

	// LRU 缓存：容量 100，超出按 LRU 淘汰，附带淘汰回调
	var evicted []string
	lru := cache.NewLRU(100, func(key string, _ []byte) {
		evicted = append(evicted, key)
	}, 0)
	_ = lru.Set(ctx, "key", val, time.Minute)

	// LRU 默认 TTL：Set 传 0 时套用默认 5 分钟
	ttlLRU := cache.NewLRU(0, nil, 5*time.Minute)
	_ = ttlLRU.Set(ctx, "key", val, 0) // 0 → 默认 5 分钟
}
```

## API 速查

| 符号 | 说明 |
| --- | --- |
| `Manager` | 缓存操作接口：`Get`/`Set`/`SetNx`/`GetBlob`/`SetBlob`/`Del`/`Expire` |
| `NewLocal(opts ...Option) Manager` | 创建进程内 map 缓存，自带过期清理协程 |
| `NewLRU(capability int, onEvict func(string, []byte), expire time.Duration, opts ...Option) Manager` | 创建带容量与（可选默认）TTL 的 LRU 缓存 |
| `Option` | 构造期选项函数 |
| `WithNow(now func() time.Time) Option` | 注入时钟，用于测试驱动过期 |
| `WithEvictInterval(d time.Duration) Option` | 设置 map 缓存后台清理间隔；≤0 关闭后台清理（仍惰性过期） |
| `(*localCache).Close() error` | 停止后台清理协程（幂等，可重复调用） |
| `ErrNotFound` / `ErrInActive` | 预定义错误：key 不存在/已过期 / 实例未初始化 |
| `lru.New[K,V](capability int, onEvict func(K,V), expires time.Duration) *Cache[K,V]` | 泛型并发安全 LRU 构造函数（`cache/v3/lru` 子包） |
| `(*lru.Cache).WithNow(func() time.Time) *Cache` | 给 LRU 注入时钟（链式） |
| `(*lru.Cache).Set/Get/Expire/Remove/RemoveOldest/RemoveExpired/Len/Clear` | 泛型 LRU 操作 |

引入路径：`github.com/tenz-io/gokit/cache/v3`
