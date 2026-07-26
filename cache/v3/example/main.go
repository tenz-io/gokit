// Command cache-example demonstrates the cache/v3 local and LRU backends:
// string and struct (blob) storage, per-entry TTL expiration, and capacity
// eviction. Everything is in-process — no network, no Redis.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/tenz-io/gokit/cache/v3"
)

func main() {
	ctx := context.Background()

	// 1) Local map cache with a 1s background sweep.
	local := cache.NewLocal(cache.WithEvictInterval(time.Second))
	defer func() { _ = local.Close() }()

	if err := local.Set(ctx, "greeting", "hello", 0); err != nil {
		panic(err)
	}
	if raw, err := local.Get(ctx, "greeting"); err != nil {
		panic(err)
	} else {
		fmt.Println("local get:", raw)
	}

	// 2) Struct (blob) round-trip via JSON.
	type user struct {
		Name string
		Age  int
	}
	if err := local.SetBlob(ctx, "user:1", user{Name: "tom", Age: 18}, 0); err != nil {
		panic(err)
	}
	var u user
	if err := local.GetBlob(ctx, "user:1", &u); err != nil {
		panic(err)
	}
	fmt.Printf("local blob: %+v\n", u)

	// 3) Per-entry expiration.
	if err := local.Set(ctx, "temp", "ephemeral", 50*time.Millisecond); err != nil {
		panic(err)
	}
	if _, err := local.Get(ctx, "temp"); err != nil {
		panic(fmt.Errorf("expected hit before expiry: %w", err))
	}
	time.Sleep(60 * time.Millisecond)
	if _, err := local.Get(ctx, "temp"); err != cache.ErrNotFound {
		fmt.Printf("expected ErrNotFound after expiry, got %v\n", err)
	}

	// 4) LRU cache: capacity eviction with an onEvict callback.
	var evicted []string
	lru := cache.NewLRU(2, func(key string, _ []byte) { evicted = append(evicted, key) }, 0)
	for _, k := range []string{"a", "b", "c"} {
		_ = lru.Set(ctx, k, k, 0)
	}
	fmt.Println("lru evicted on capacity:", evicted) // evicts "a"
	if raw, err := lru.Get(ctx, "b"); err != nil {
		panic(err)
	} else {
		fmt.Println("lru get b:", raw)
	}

	// 5) LRU default TTL applied when Set passes a zero duration.
	ttlLRU := cache.NewLRU(0, nil, 50*time.Millisecond)
	_ = ttlLRU.Set(ctx, "ttl", "v", 0) // 0 → use default 50ms
	time.Sleep(60 * time.Millisecond)
	if _, err := ttlLRU.Get(ctx, "ttl"); err != cache.ErrNotFound {
		fmt.Printf("expected default-TTL miss, got %v\n", err)
	}
}
