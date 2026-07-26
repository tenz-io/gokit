package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// item is a single local-cache entry. raw is the stored bytes; expireAt is the
// absolute deadline, with the zero value meaning "never expires". Storing a
// time.Time (rather than a unix-second int) keeps sub-second TTLs accurate;
// earlier versions truncated to seconds, which dropped millisecond TTLs early.
type item struct {
	raw      []byte
	expireAt time.Time
}

// localCache is a process-local map cache with optional per-entry TTL and a
// background sweep goroutine that reaps expired keys so memory does not grow
// unbounded when reads are sparse.
//
// It implements [Manager]. Construct with [NewLocal].
type localCache struct {
	m       map[string]*item
	nowFunc func() time.Time
	lock    sync.RWMutex

	// eviction loop bookkeeping. stopCh closes to terminate the sweep; the
	// started flag (guarded by lock) prevents double-start.
	stopCh  chan struct{}
	started bool
}

// NewLocal creates a process-local map cache. By default it starts a
// background goroutine that sweeps expired keys every 5 minutes; pass
// [WithEvictInterval](0) to disable it (expiration is still enforced lazily
// on read). The returned cache must be Close'd when done to stop the sweep.
//
// Options: [WithNow] (injectable clock), [WithEvictInterval].
func NewLocal(opts ...Option) Manager {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	lc := &localCache{
		m:       make(map[string]*item),
		nowFunc: o.nowFunc,
	}
	lc.startEvict(o.evictInterval)
	return lc
}

// Close stops the background sweep goroutine. It is safe to call multiple
// times and to call on a cache whose sweep was disabled. The cache contents
// remain usable after Close; only the reaping loop is stopped.
func (lc *localCache) Close() error {
	if lc == nil {
		return nil
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	if lc.started && lc.stopCh != nil {
		select {
		case <-lc.stopCh:
			// already closed
		default:
			close(lc.stopCh)
		}
		lc.started = false
	}
	return nil
}

func (lc *localCache) active() bool {
	return lc != nil && lc.m != nil
}

// startEvict launches the background sweep at the given interval. A
// non-positive interval disables sweeping.
func (lc *localCache) startEvict(interval time.Duration) {
	if !lc.active() || interval <= 0 {
		return
	}
	lc.lock.Lock()
	if lc.started {
		lc.lock.Unlock()
		return
	}
	lc.stopCh = make(chan struct{})
	lc.started = true
	lc.lock.Unlock()

	go lc.evictLoop(interval)
}

func (lc *localCache) evictLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-lc.stopCh:
			return
		case <-ticker.C:
			lc.evict()
		}
	}
}

// evict removes all expired entries under a write lock.
func (lc *localCache) evict() {
	if !lc.active() {
		return
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()

	now := lc.nowFunc()
	for k, v := range lc.m {
		if !v.expireAt.IsZero() && !now.Before(v.expireAt) {
			delete(lc.m, k)
		}
	}
}

// expireAt returns the absolute deadline for a new entry. A non-positive
// duration means "never expires" (stored as the zero time.Time).
func (lc *localCache) expireAt(expire time.Duration) time.Time {
	if expire <= 0 {
		return time.Time{}
	}
	return lc.nowFunc().Add(expire)
}

func (lc *localCache) Get(ctx context.Context, key string) (raw string, err error) {
	if !lc.active() {
		return "", ErrInActive
	}

	lc.lock.RLock()
	defer lc.lock.RUnlock()

	it, found := lc.m[key]
	if !found || it == nil {
		return "", ErrNotFound
	}
	if !it.expireAt.IsZero() && !lc.nowFunc().Before(it.expireAt) {
		// Expired: best-effort lazy deletion on a fresh goroutine so we can
		// stay under the read lock (Del takes the write lock).
		go lc.Del(ctx, key)
		return "", ErrNotFound
	}
	return string(it.raw), nil
}

func (lc *localCache) Set(_ context.Context, key string, raw string, expire time.Duration) error {
	if !lc.active() {
		return ErrInActive
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	lc.m[key] = &item{raw: []byte(raw), expireAt: lc.expireAt(expire)}
	return nil
}

func (lc *localCache) SetNx(_ context.Context, key string, raw string, expire time.Duration) (existing bool, err error) {
	if !lc.active() {
		return false, ErrInActive
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()

	if it, ok := lc.m[key]; ok && it != nil {
		// Treat an expired key as absent so SetNx writes through.
		if it.expireAt.IsZero() || lc.nowFunc().Before(it.expireAt) {
			return true, nil
		}
	}
	lc.m[key] = &item{raw: []byte(raw), expireAt: lc.expireAt(expire)}
	return false, nil
}

func (lc *localCache) GetBlob(ctx context.Context, key string, output any) error {
	if !lc.active() {
		return ErrInActive
	}

	lc.lock.RLock()
	defer lc.lock.RUnlock()

	it, found := lc.m[key]
	if !found || it == nil {
		return ErrNotFound
	}
	if !it.expireAt.IsZero() && !lc.nowFunc().Before(it.expireAt) {
		go lc.Del(ctx, key)
		return ErrNotFound
	}
	if err := json.Unmarshal(it.raw, output); err != nil {
		return fmt.Errorf("cache: decode error: %w", err)
	}
	return nil
}

func (lc *localCache) SetBlob(_ context.Context, key string, val any, expire time.Duration) error {
	if !lc.active() {
		return ErrInActive
	}
	bs, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("cache: encode error: %w", err)
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	lc.m[key] = &item{raw: bs, expireAt: lc.expireAt(expire)}
	return nil
}

func (lc *localCache) Del(_ context.Context, key string) error {
	if !lc.active() {
		return ErrInActive
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	delete(lc.m, key)
	return nil
}

func (lc *localCache) Expire(_ context.Context, key string, expire time.Duration) error {
	if !lc.active() {
		return ErrInActive
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	it, ok := lc.m[key]
	if !ok || it == nil {
		return ErrNotFound
	}
	it.expireAt = lc.expireAt(expire)
	return nil
}
