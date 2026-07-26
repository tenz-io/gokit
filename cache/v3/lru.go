package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tenz-io/gokit/cache/v3/lru"
)

// lruCache adapts a concurrency-safe generic LRU ([lru.Cache]) to the
// [Manager] interface. Values are stored as []byte (JSON for blobs, raw
// bytes for Set/Get).
type lruCache struct {
	c *lru.Cache[string, []byte]
}

// NewLRU creates an LRU-bounded cache implementing [Manager].
//
//   - capability caps the entry count; a non-positive value defaults to 120.
//   - onEvict is an optional callback invoked when an entry is evicted by
//     capacity pressure, expiration, or explicit removal.
//   - expire is the default TTL applied when a Set/SetBlob/SetNx call passes
//     a zero duration; leave zero for "no default TTL".
//
// Options: [WithNow] (injectable clock). [WithEvictInterval] is ignored — the
// LRU expires entries lazily on access; call [lru.Cache.RemoveExpired] on the
// underlying store directly if periodic sweeping is wanted.
func NewLRU(
	capability int,
	onEvict func(key string, val []byte),
	expire time.Duration,
	opts ...Option,
) Manager {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	if capability <= 0 {
		capability = 120
	}
	c := lru.New[string, []byte](capability, onEvict, expire)
	c.WithNow(o.nowFunc)
	return &lruCache{c: c}
}

func (lc *lruCache) active() bool {
	return lc != nil && lc.c != nil
}

func (lc *lruCache) Get(_ context.Context, key string) (string, error) {
	if !lc.active() {
		return "", ErrInActive
	}
	bs, ok := lc.c.Get(key)
	if !ok {
		return "", ErrNotFound
	}
	return string(bs), nil
}

func (lc *lruCache) Set(_ context.Context, key string, raw string, expire time.Duration) error {
	if !lc.active() {
		return ErrInActive
	}
	lc.c.Set(key, []byte(raw), expire)
	return nil
}

func (lc *lruCache) SetNx(_ context.Context, key string, raw string, expire time.Duration) (bool, error) {
	if !lc.active() {
		return false, ErrInActive
	}
	if _, ok := lc.c.Get(key); ok {
		return true, nil
	}
	lc.c.Set(key, []byte(raw), expire)
	return false, nil
}

func (lc *lruCache) GetBlob(_ context.Context, key string, output any) error {
	if !lc.active() {
		return ErrInActive
	}
	bs, ok := lc.c.Get(key)
	if !ok {
		return ErrNotFound
	}
	if err := json.Unmarshal(bs, output); err != nil {
		return fmt.Errorf("cache: decode error: %w", err)
	}
	return nil
}

func (lc *lruCache) SetBlob(_ context.Context, key string, val any, expire time.Duration) error {
	if !lc.active() {
		return ErrInActive
	}
	bs, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("cache: encode error: %w", err)
	}
	lc.c.Set(key, bs, expire)
	return nil
}

func (lc *lruCache) Del(_ context.Context, key string) error {
	if !lc.active() {
		return ErrInActive
	}
	lc.c.Remove(key)
	return nil
}

func (lc *lruCache) Expire(_ context.Context, key string, expire time.Duration) error {
	if !lc.active() {
		return ErrInActive
	}
	// lru.Cache.Expire is a no-op on a missing key, but Manager.Expire
	// promises ErrNotFound. Detect the miss by reading first.
	if _, ok := lc.c.Get(key); !ok {
		return ErrNotFound
	}
	lc.c.Expire(key, expire)
	return nil
}

// Close is a no-op for the LRU backend, which has no background resources to
// release. It satisfies the [Manager] lifecycle contract.
func (lc *lruCache) Close() error { return nil }
