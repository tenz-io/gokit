package cache

import (
	"fmt"
	"sync"
	"time"
)

// item is a single local-cache entry. raw is the stored bytes; expireAt is
// the absolute deadline, with the zero value meaning "never expires". A
// time.Time (not a unix-second int) keeps sub-second TTLs accurate.
type item struct {
	raw      []byte
	expireAt time.Time
}

// localCache is a process-local map cache with optional per-entry TTL and a
// background sweep goroutine that reaps expired keys so memory does not grow
// unbounded when reads are sparse.
//
// It implements [Manager]. Construct with [NewLocal].
//
// Concurrency: every mutation and read is guarded by an RWMutex. Lazy expiry
// on read is *synchronous and conditional*: a Get that finds an expired key
// drops the read lock, takes the write lock, re-checks that the same item is
// still present and still expired, and only then deletes it. This avoids the
// race where an async delete removes a value just written by a concurrent
// Set to the same key. GetBlob additionally copies the raw bytes under the
// lock and decodes JSON after unlocking, so a slow or reentrant Unmarshal
// cannot block writers or deadlock.
type localCache struct {
	m       map[string]*item
	nowFunc func() time.Time
	lock    sync.RWMutex

	// sweepWg tracks the background sweep goroutine so Close can wait for it
	// to exit. stopCh, when non-nil, is closed to signal the sweep to stop.
	stopCh  chan struct{}
	sweepWg sync.WaitGroup
	startMu sync.Mutex
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

// Close stops the background sweep goroutine and blocks until it has exited.
// It is safe to call multiple times and on a cache whose sweep was disabled.
// The cache contents remain usable after Close; only the reaping loop stops.
func (lc *localCache) Close() error {
	if lc == nil {
		return nil
	}
	lc.startMu.Lock()
	if lc.started && lc.stopCh != nil {
		select {
		case <-lc.stopCh:
			// already closed
		default:
			close(lc.stopCh)
		}
		lc.started = false
	}
	lc.startMu.Unlock()
	// Wait outside startMu so the sweep goroutine (which does not touch
	// startMu on exit) can actually terminate.
	lc.sweepWg.Wait()
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
	lc.startMu.Lock()
	defer lc.startMu.Unlock()
	if lc.started {
		return
	}
	lc.stopCh = make(chan struct{})
	lc.started = true
	lc.sweepWg.Add(1)
	go lc.evictLoop(interval)
}

func (lc *localCache) evictLoop(interval time.Duration) {
	defer lc.sweepWg.Done()
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
// duration means "never expires" (the zero time.Time).
func (lc *localCache) expireAt(expire time.Duration) time.Time {
	if expire <= 0 {
		return time.Time{}
	}
	return lc.nowFunc().Add(expire)
}

// expired reports whether it is past the entry's deadline. A zero expireAt
// means the entry never expires.
func (lc *localCache) expired(it *item) bool {
	if it.expireAt.IsZero() {
		return false
	}
	return !lc.nowFunc().Before(it.expireAt)
}

// deleteIfExpired removes key only when the cache still holds the exact item
// (by pointer) and that item is still expired. The caller has just observed
// the item under a read lock; re-confirming under the write lock closes the
// window in which a concurrent Set could have replaced the value — if it
// did, we leave the new value alone. Always called with the read lock
// already released.
func (lc *localCache) deleteIfExpired(key string, stale *item) {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	cur, ok := lc.m[key]
	if !ok || cur != stale {
		// Replaced or removed concurrently — do not touch the new value.
		return
	}
	if !lc.expired(cur) {
		// Time rolled back or was re-Expire'd to a live deadline.
		return
	}
	delete(lc.m, key)
}

func (lc *localCache) Get(key string) (raw string, err error) {
	if !lc.active() {
		return "", ErrInactive
	}

	lc.lock.RLock()
	it, found := lc.m[key]
	if !found || it == nil {
		lc.lock.RUnlock()
		return "", ErrNotFound
	}
	expired := lc.expired(it)
	val := string(it.raw)
	lc.lock.RUnlock()

	if expired {
		lc.deleteIfExpired(key, it)
		return "", ErrNotFound
	}
	return val, nil
}

func (lc *localCache) Set(key string, raw string, expire time.Duration) error {
	if !lc.active() {
		return ErrInactive
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	lc.m[key] = &item{raw: []byte(raw), expireAt: lc.expireAt(expire)}
	return nil
}

func (lc *localCache) SetNx(key string, raw string, expire time.Duration) (existing bool, err error) {
	if !lc.active() {
		return false, ErrInactive
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()

	if it, ok := lc.m[key]; ok && it != nil {
		// Treat an expired key as absent so SetNx writes through.
		if !lc.expired(it) {
			return true, nil
		}
	}
	lc.m[key] = &item{raw: []byte(raw), expireAt: lc.expireAt(expire)}
	return false, nil
}

func (lc *localCache) GetBlob(key string, output any) error {
	if !lc.active() {
		return ErrInactive
	}

	lc.lock.RLock()
	it, found := lc.m[key]
	if !found || it == nil {
		lc.lock.RUnlock()
		return ErrNotFound
	}
	expired := lc.expired(it)
	// Copy the raw bytes under the lock; decode after unlocking so a slow or
	// reentrant Unmarshal cannot block writers or deadlock.
	raw := make([]byte, len(it.raw))
	copy(raw, it.raw)
	lc.lock.RUnlock()

	if expired {
		lc.deleteIfExpired(key, it)
		return ErrNotFound
	}
	if err := decodeBlob(raw, output); err != nil {
		return err
	}
	return nil
}

func (lc *localCache) SetBlob(key string, val any, expire time.Duration) error {
	if !lc.active() {
		return ErrInactive
	}
	bs, err := encodeBlob(val)
	if err != nil {
		return fmt.Errorf("cache: encode error: %w", err)
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	lc.m[key] = &item{raw: bs, expireAt: lc.expireAt(expire)}
	return nil
}

func (lc *localCache) Del(key string) error {
	if !lc.active() {
		return ErrInactive
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	delete(lc.m, key)
	return nil
}

func (lc *localCache) Expire(key string, expire time.Duration) error {
	if !lc.active() {
		return ErrInactive
	}
	lc.lock.Lock()
	defer lc.lock.Unlock()
	it, ok := lc.m[key]
	if !ok || it == nil {
		return ErrNotFound
	}
	// Do not resurrect an expired key.
	if lc.expired(it) {
		return ErrNotFound
	}
	it.expireAt = lc.expireAt(expire)
	return nil
}
