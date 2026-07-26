# cache

纯本地进程内缓存库：map 缓存与泛型 LRU 缓存，均支持按条目过期（TTL）。零网络、零外部依赖——不需要 Redis，也不需要任何远程组件。

```go
import "github.com/tenz-io/gokit/cache/v3"
```

## 模块介绍

cache 解决两类本地缓存场景：

- **无界 map 缓存**（`NewLocal`）：`map[string]*item` + 读写锁，按条目 TTL 过期，后台协程周期清理已过期 key 避免内存无限增长。
- **容量受限的 LRU 缓存**（`NewLRU`）：基于主包内置的并发安全泛型 LRU 实现容量上限 + TTL，超出容量按最近最少使用淘汰，支持淘汰回调 `onEvict`。

两者统一实现 `Manager` 接口，业务代码换后端时无需改调用方式；`GetBlob`/`SetBlob` 内置 MessagePack 编解码，直接存取结构体，体积与分配均小于 JSON。

V3 相对 V2 的核心变化：

- **纯本地**：删除 Redis 后端、`go-redis` 依赖、Interceptor/Hook、Lua `Eval`、testify/mock。一切在进程内完成；唯一外部依赖是 MessagePack 编解码库（`vmihailenco/msgpack/v5`），用于结构体直存直取。
- **签名瘦身**：去掉 `context.Context` 参数——纯内存缓存无法响应取消，携带它只会误导调用方。接口仅暴露缓存语义本身。
- **统一 TTL 契约**：取消 LRU 的默认 TTL。所有 `Set`/`SetNx`/`Expire` 显式传入 `expire`：**非正值（0 或负）= 永不过期，正值 = 相对 now 的绝对截止时间**。两后端语义完全一致，`NewLocal` ↔ `NewLRU` 互换不会改变数据生命周期。
- **可注入时钟**：`WithNow` 注入时钟，过期逻辑用绝对时间判断，单测可不依赖真实 `time.Sleep`，杜绝 flaky。
- **生命周期可控**：`localCache.Close()` 停止后台清理协程并 `Wait` 至其退出（v2 的清理协程永不退出，存在泄漏；`WaitGroup` 保证 `Close` 返回时协程已退出）；LRU 后端惰性过期，无后台协程。
- **并发正确性**：相对 v2 修复了三类缺陷——
  - 惰性过期改为**同步条件删除**：`Get`/`GetBlob` 发现过期后，释放读锁、取写锁、二次确认仍是同一 item 且仍过期再删，避免延迟删除误删并发 `Set` 刚写入的新值；
  - LRU 的 `SetNx` 改为**原子**（存在性检查与写入在同一次锁内完成），并发 `SetNx` 不会都判“不存在”都写入；
  - 淘汰回调 `onEvicted` 在**锁外**执行，回调里安全地回写同一 cache，不会死锁，慢回调也不阻塞其他请求。
- **`GetBlob` 锁外解码**：锁内只复制 `raw` 字节，MessagePack 解码在解锁后进行，慢解码/自定义解码回写也不阻塞写、不死锁。

## 能力清单

| 能力 | 含义 |
| --- | --- |
| 统一本地缓存接口 | `Manager` 接口抹平 map 与 LRU 两种实现的差异，业务换后端无需改调用方式 |
| 统一 TTL 契约 | 两后端对 `expire` 语义一致：非正值永不过期、正值相对 now 过期；后端互换数据生命周期不变 |
| 结构体直存直取 | `GetBlob`/`SetBlob` 内部用 MessagePack 编解码，业务无需手写 marshal；体积与分配小于 JSON |
| 不存在才写入（原子） | `SetNx` 在 key 不存在（或已过期）时才写入并返回是否已存在；存在性检查与写入在单次加锁内原子完成，可用于幂等写入 |
| 进程内缓存自动过期清理 | `NewLocal` 的 map 缓存启动后台协程按间隔扫描，删除已过期 key，避免内存无限增长 |
| 容量受限的 LRU 缓存 | `NewLRU` 基于泛型 LRU 实现容量上限 + TTL，超出容量按最近最少使用淘汰，支持 `onEvict` 回调（锁外执行） |
| 惰性过期 + 同步条件删除 | map 与 LRU 后端均在读取时检查过期：命中过期 key 即同步二次确认并删除、返回未命中，绝不误删并发写入的新值 |
| 过期后 Expire 不复活 | `Expire` 对已过期/缺失的 key 返回 `ErrNotFound`，不会把逻辑上已不存在的 key“复活” |
| 可注入时钟 | `WithNow` 注入时钟，过期判断基于绝对时间，单测可推进时间不依赖真实 sleep |
| 生命周期管理 | `localCache.Close()` 停止后台清理协程并保证其退出，避免泄漏；可重复调用，幂等 |
| 淘汰回调锁外执行 | LRU 的 `onEvict` 在释放锁后调用，回调可安全回写同一 cache，慢回调不阻塞其他请求 |
| 锁外解码 | `GetBlob` 锁内仅复制字节，MessagePack 解码在解锁后进行，慢/重入的自定义解码不阻塞写、不死锁 |

## 快速开始

```go
package main

import (
	"time"

	"github.com/tenz-io/gokit/cache/v3"
)

func main() {
	// 本地 map 缓存，后台每秒清理一次过期 key
	mgr := cache.NewLocal(cache.WithEvictInterval(time.Second))
	defer func() { _ = mgr.Close() }()

	_ = mgr.Set("key", "value", 5*time.Minute)
	val, _ := mgr.Get("key")

	// 结构体存取
	type User struct {
		Name string
		Age  int
	}
	_ = mgr.SetBlob("user:1", &User{Name: "tom", Age: 18}, time.Minute)
	var u User
	_ = mgr.GetBlob("user:1", &u)

	// LRU 缓存：容量 100，超出按 LRU 淘汰，附带淘汰回调
	var evicted []string
	lru := cache.NewLRU(100, func(key string, _ []byte) {
		evicted = append(evicted, key)
	})
	_ = lru.Set("key", val, time.Minute)

	// SetNx：仅当 key 不存在/已过期时写入，适合幂等
	existing, _ := lru.SetNx("once", "v", 0)
	if existing {
		// already present
	}
}
```

## API 速查

| 符号 | 说明 |
| --- | --- |
| `Manager` | 缓存操作接口：`Get`/`Set`/`SetNx`/`GetBlob`/`SetBlob`/`Del`/`Expire`/`Close`（无 `context.Context`、无 `Eval`） |
| `NewLocal(opts ...Option) Manager` | 创建进程内 map 缓存，自带过期清理协程 |
| `NewLRU(capability int, onEvict func(string, []byte), opts ...Option) Manager` | 创建带容量与 TTL 的 LRU 缓存（无默认 TTL 参数） |
| `Option` | 构造期选项函数 |
| `WithNow(now func() time.Time) Option` | 注入时钟，用于测试驱动过期 |
| `WithEvictInterval(d time.Duration) Option` | 设置 map 缓存后台清理间隔；≤0 关闭后台清理（仍惰性过期） |
| `(*localCache).Close() error` | 停止后台清理协程并等待其退出（幂等，可重复调用） |
| `ErrNotFound` / `ErrInactive` | 预定义错误：key 不存在/已过期 / 实例未初始化或已关闭 |

> 泛型 LRU 底层（`lruCache[K,V]`）为包内未导出类型，仅供 `NewLRU` 内部使用，调用方拿到的永远是 `Manager` 接口。

引入路径：`github.com/tenz-io/gokit/cache/v3`
