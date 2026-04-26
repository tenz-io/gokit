package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tenz-io/gokit/cache/lru"
)

type lruCache struct {
	c *lru.Cache[string, []byte]
}

func NewLRU(
	capability int,
	onEvict func(key string, val []byte),
	expire time.Duration,
) Manager {
	l := &lruCache{}
	if capability <= 0 {
		capability = 120
	}
	l.c = lru.New(capability, onEvict, expire)
	return l
}

func (lc *lruCache) active() bool {
	return lc != nil && lc.c != nil
}

func (lc *lruCache) Get(ctx context.Context, key string) (raw string, err error) {
	if !lc.active() {
		return "", ErrInActive
	}

	bs, ok := lc.c.Get(key)
	if !ok {
		return "", ErrNotFound
	}

	return string(bs), nil

}

func (lc *lruCache) Set(ctx context.Context, key string, raw string, expire time.Duration) (err error) {
	if !lc.active() {
		return ErrInActive
	}

	lc.c.Set(key, []byte(raw), expire)
	return nil
}

func (lc *lruCache) SetNx(ctx context.Context, key string, raw string, expire time.Duration) (existing bool, err error) {
	if !lc.active() {
		return false, ErrInActive
	}

	if _, ok := lc.c.Get(key); ok {
		return true, nil
	}

	lc.c.Set(key, []byte(raw), expire)
	return false, nil
}

func (lc *lruCache) GetBlob(ctx context.Context, key string, output any) (err error) {
	if !lc.active() {
		return ErrInActive
	}

	bs, ok := lc.c.Get(key)
	if !ok {
		return ErrNotFound
	}

	if err = json.Unmarshal(bs, output); err != nil {
		return fmt.Errorf("decode error: %w", err)
	}
	return nil
}

func (lc *lruCache) SetBlob(ctx context.Context, key string, val any, expire time.Duration) (err error) {
	if !lc.active() {
		return ErrInActive
	}

	bs, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	lc.c.Set(key, bs, expire)
	return nil
}

func (lc *lruCache) Del(ctx context.Context, key string) (err error) {
	if !lc.active() {
		return ErrInActive
	}

	lc.c.Remove(key)
	return nil
}

func (lc *lruCache) Expire(ctx context.Context, key string, expire time.Duration) (err error) {
	if !lc.active() {
		return ErrInActive
	}

	lc.c.Expire(key, expire)
	return nil
}

func (lc *lruCache) Eval(ctx context.Context, script string, keys []string, args ...any) (val any, err error) {
	// not supported
	return nil, ErrNotSupported
}
