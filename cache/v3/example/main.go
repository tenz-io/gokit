// Command cache-example demonstrates the cache/v3 local and LRU backends:
// string and struct (blob) storage, per-entry TTL expiration, and capacity
// eviction. Everything is in-process — no network, no Redis, no context.
package main

import (
	"fmt"
	"time"

	"github.com/tenz-io/gokit/cache/v3"
)

func main() {
	// 1) Local map cache with a 1s background sweep.
	local := cache.NewLocal(cache.WithEvictInterval(time.Second))
	defer func() { _ = local.Close() }()

	if err := local.Set("greeting", "hello", 0); err != nil {
		panic(err)
	}
	if raw, err := local.Get("greeting"); err != nil {
		panic(err)
	} else {
		fmt.Println("local get:", raw)
	}

	// 2) Struct (blob) round-trip via JSON.
	type user struct {
		Name string
		Age  int
	}
	if err := local.SetBlob("user:1", user{Name: "tom", Age: 18}, 0); err != nil {
		panic(err)
	}
	var u user
	if err := local.GetBlob("user:1", &u); err != nil {
		panic(err)
	}
	fmt.Printf("local blob: %+v\n", u)

	// 3) Per-entry expiration. There is no default TTL: a non-positive
	// duration means "never expires", so the 50ms TTL below is explicit.
	if err := local.Set("temp", "ephemeral", 50*time.Millisecond); err != nil {
		panic(err)
	}
	if _, err := local.Get("temp"); err != nil {
		panic(fmt.Errorf("expected hit before expiry: %w", err))
	}
	time.Sleep(60 * time.Millisecond)
	if _, err := local.Get("temp"); err != cache.ErrNotFound {
		fmt.Printf("expected ErrNotFound after expiry, got %v\n", err)
	}

	// 4) LRU cache: capacity eviction with an onEvict callback. No default
	// TTL — every entry's lifetime is whatever its Set call specifies.
	var evicted []string
	lru := cache.NewLRU(2, func(key string, _ []byte) { evicted = append(evicted, key) })
	for _, k := range []string{"a", "b", "c"} {
		_ = lru.Set(k, k, 0)
	}
	fmt.Println("lru evicted on capacity:", evicted) // evicts "a"
	if raw, err := lru.Get("b"); err != nil {
		panic(err)
	} else {
		fmt.Println("lru get b:", raw)
	}

	// 5) SetNx: write-through only when absent.
	if existing, _ := lru.SetNx("b", "overwrite", 0); !existing {
		fmt.Println("SetNx should report existing=true for present key 'b'")
	}
	if existing, _ := lru.SetNx("new", "v", 0); existing {
		fmt.Println("SetNx should report existing=false for a new key")
	}
}
