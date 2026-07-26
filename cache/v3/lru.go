package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// lruManager adapts the generic [lruCache] (K=string, V=[]byte) to the
// [Manager] interface. Values are stored as []byte (JSON for blobs, raw bytes
// for Set/Get). There is no default TTL: a non-positive expire passed to any
// method means "never expires", matching the local cache.
type lruManager struct {
	c *lruCache[string, []byte]
}

// NewLRU creates an LRU-bounded cache implementing [Manager].
//
//   - capability caps the entry count; a non-positive value defaults to 120.
//   - onEvict is an optional callback invoked (outside the cache's lock) when
//     an entry is evicted by capacity pressure, expiration, or removal.
//
// Options: [WithNow] (injectable clock). [WithEvictInterval] is ignored —
// the LRU expires entries lazily on access.
func NewLRU(
	capability int,
	onEvict func(key string, val []byte),
	opts ...Option,
) Manager {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	if capability <= 0 {
		capability = 120
	}
	c := newLRU[string, []byte](capability, onEvict)
	c.withNow(o.nowFunc)
	return &lruManager{c: c}
}

func (m *lruManager) active() bool {
	return m != nil && m.c != nil
}

func (m *lruManager) Get(key string) (string, error) {
	if !m.active() {
		return "", ErrInactive
	}
	bs, ok := m.c.get(key)
	if !ok {
		return "", ErrNotFound
	}
	return string(bs), nil
}

func (m *lruManager) Set(key string, raw string, expire time.Duration) error {
	if !m.active() {
		return ErrInactive
	}
	m.c.set(key, []byte(raw), expire)
	return nil
}

func (m *lruManager) SetNx(key string, raw string, expire time.Duration) (bool, error) {
	if !m.active() {
		return false, ErrInactive
	}
	// lruCache.setNx does the existence check and the write under one lock,
	// so concurrent SetNx callers cannot both observe "absent" and both write.
	return m.c.setNx(key, []byte(raw), expire), nil
}

func (m *lruManager) GetBlob(key string, output any) error {
	if !m.active() {
		return ErrInactive
	}
	bs, ok := m.c.get(key)
	if !ok {
		return ErrNotFound
	}
	if err := decodeBlob(bs, output); err != nil {
		return err
	}
	return nil
}

func (m *lruManager) SetBlob(key string, val any, expire time.Duration) error {
	if !m.active() {
		return ErrInactive
	}
	bs, err := encodeBlob(val)
	if err != nil {
		return fmt.Errorf("cache: encode error: %w", err)
	}
	m.c.set(key, bs, expire)
	return nil
}

func (m *lruManager) Del(key string) error {
	if !m.active() {
		return ErrInactive
	}
	m.c.remove(key)
	return nil
}

func (m *lruManager) Expire(key string, expire time.Duration) error {
	if !m.active() {
		return ErrInactive
	}
	// lruCache.expire checks present-and-unexpired under one lock and returns
	// ok=false for a missing or expired key, so we never resurrect an expired
	// entry nor race a concurrent writer.
	if ok := m.c.expire(key, expire); !ok {
		return ErrNotFound
	}
	return nil
}

// Close is a no-op for the LRU backend, which has no background resources to
// release. It satisfies the [Manager] lifecycle contract.
func (m *lruManager) Close() error { return nil }

// lruCache is a generic, concurrency-safe LRU cache with optional per-entry
// TTL expiration. It backs [lruManager]. A zero-value lruCache is not ready
// for use; construct one with [newLRU].
//
// Every mutation and read is guarded by a sync.Mutex, so it can be shared
// across goroutines without external locking (unlike the v2 lru, which was
// explicitly not safe for concurrent access).
//
// Eviction callbacks (onEvict) run *outside* the cache's lock: a method reaps
// any purged entries under the lock, releases the lock, and only then invokes
// the callback. This lets a callback safely call back into the same cache
// (e.g. to read or count) without deadlocking, and keeps a slow callback from
// blocking unrelated readers.
type lruCache[K comparable, V any] struct {
	capability int
	// onEvict, when non-nil, is called for every item purged from the cache:
	// capacity eviction, explicit remove/removeOldest, expiration, and clear.
	// It is invoked after the lock is released (see the type doc).
	onEvict func(key K, val V)

	ll      *list.List
	cache   map[K]*list.Element
	nowFunc func() time.Time

	mu sync.Mutex
}

// lruEntry is one element of the LRU. It holds the key (for eviction callbacks
// and map removal), the value, and an absolute expiration time.
type lruEntry[K comparable, V any] struct {
	key      K
	val      V
	expireAt time.Time
}

// expired reports whether the entry's deadline has passed. A zero expireAt
// means the entry never expires. The boundary is inclusive: at the exact
// deadline the entry is considered expired.
func (e *lruEntry[K, V]) expired(now time.Time) bool {
	if e.expireAt.IsZero() {
		return false
	}
	return !now.Before(e.expireAt)
}

// newLRU creates an LRU cache.
//
//   - capability caps the entry count; a non-positive value means no limit
//     (expiration is then the only reclamation mechanism).
//   - onEvict is an optional callback invoked (outside the lock) when an item
//     is purged.
//
// There is no default TTL: every set/setNx takes an explicit duration, where a
// non-positive value means "never expires". The cache is safe for concurrent
// use.
func newLRU[K comparable, V any](
	capability int,
	onEvict func(key K, val V),
) *lruCache[K, V] {
	return &lruCache[K, V]{
		capability: capability,
		ll:         list.New(),
		cache:      make(map[K]*list.Element),
		onEvict:    onEvict,
		nowFunc:    time.Now,
	}
}

// withNow injects a clock used for all expiration checks. Returns the cache
// for chaining. Tests use it to advance time without real sleeps.
func (c *lruCache[K, V]) withNow(now func() time.Time) *lruCache[K, V] {
	if now != nil {
		c.nowFunc = now
	}
	return c
}

// deadlineFor returns the absolute expiration time for a new entry given a
// per-call duration. A non-positive duration means the entry never expires
// (the zero time.Time); a positive duration gives now+duration.
func deadlineFor(now time.Time, duration time.Duration) time.Time {
	if duration <= 0 {
		return time.Time{}
	}
	return now.Add(duration)
}

// set adds (or updates) val under key with the given expiration. A
// non-positive duration means the entry never expires.
func (c *lruCache[K, V]) set(key K, val V, duration time.Duration) {
	now := c.nowFunc()
	expireAt := deadlineFor(now, duration)

	c.mu.Lock()
	evicted := c.setLocked(key, val, expireAt, now)
	c.mu.Unlock()

	c.fireOnEvict(evicted)
}

// setLocked inserts or updates key and returns any entry purged by capacity
// pressure so the caller can fire callbacks after unlocking. Caller holds c.mu.
func (c *lruCache[K, V]) setLocked(key K, val V, expireAt time.Time, now time.Time) []lruEntry[K, V] {
	if c.cache == nil {
		c.cache = make(map[K]*list.Element)
		c.ll = list.New()
	}
	var evicted []lruEntry[K, V]
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		e := ee.Value.(*lruEntry[K, V])
		e.val = val
		e.expireAt = expireAt
		return nil
	}
	ele := c.ll.PushFront(&lruEntry[K, V]{key: key, val: val, expireAt: expireAt})
	c.cache[key] = ele
	if c.capability > 0 && c.ll.Len() > c.capability {
		if e := c.removeOldestLocked(now); e != nil {
			evicted = append(evicted, *e)
		}
	}
	return evicted
}

// setNx atomically writes val under key only when key is absent or has
// expired. It returns existing=true when the key was present and unexpired
// (in which case no write happens), and existing=false when the write went
// through. Unlike a get-then-set sequence, the existence check and the write
// happen under a single lock, so concurrent setNx callers cannot both see
// "absent" and both write. A capacity eviction triggered by the insert fires
// its onEvict callback after the lock is released.
func (c *lruCache[K, V]) setNx(key K, val V, duration time.Duration) (existing bool) {
	now := c.nowFunc()
	expireAt := deadlineFor(now, duration)

	c.mu.Lock()
	if c.cache == nil {
		c.cache = make(map[K]*list.Element)
		c.ll = list.New()
	}
	var evicted []lruEntry[K, V]
	if ee, ok := c.cache[key]; ok {
		e := ee.Value.(*lruEntry[K, V])
		if !e.expired(now) {
			// Present and unexpired: do not overwrite.
			c.mu.Unlock()
			return true
		}
		// Expired: treat as absent. Overwrite in place (no eviction callback
		// — same as set's in-place update).
		c.ll.MoveToFront(ee)
		e.val = val
		e.expireAt = expireAt
		c.mu.Unlock()
		return false
	}
	ele := c.ll.PushFront(&lruEntry[K, V]{key: key, val: val, expireAt: expireAt})
	c.cache[key] = ele
	if c.capability > 0 && c.ll.Len() > c.capability {
		if e := c.removeOldestLocked(now); e != nil {
			evicted = append(evicted, *e)
		}
	}
	c.mu.Unlock()
	c.fireOnEvict(evicted)
	return false
}

// get looks up key. On a hit the entry is moved to the front (most-recently
// used). An expired entry is removed (and its onEvict fired after unlock) and
// reported as a miss.
func (c *lruCache[K, V]) get(key K) (val V, ok bool) {
	var zero V
	now := c.nowFunc()

	c.mu.Lock()
	if c.cache == nil {
		c.mu.Unlock()
		return zero, false
	}
	ele, hit := c.cache[key]
	if !hit {
		c.mu.Unlock()
		return zero, false
	}
	if ele.Value.(*lruEntry[K, V]).expired(now) {
		evicted := []lruEntry[K, V]{*ele.Value.(*lruEntry[K, V])}
		c.removeElementLocked(ele)
		c.mu.Unlock()
		c.fireOnEvict(evicted)
		return zero, false
	}
	c.ll.MoveToFront(ele)
	result := ele.Value.(*lruEntry[K, V]).val
	c.mu.Unlock()
	return result, true
}

// expire updates key's expiration and reports whether the update applied. A
// non-positive duration makes the entry never expire. A missing or expired
// key is left untouched and returns ok=false — expire cannot resurrect a key
// that has logically expired. The check and the update happen under one lock,
// so no concurrent writer can slip between them.
func (c *lruCache[K, V]) expire(key K, duration time.Duration) (ok bool) {
	now := c.nowFunc()
	expireAt := deadlineFor(now, duration)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		return false
	}
	ele, hit := c.cache[key]
	if !hit {
		return false
	}
	e := ele.Value.(*lruEntry[K, V])
	if e.expired(now) {
		// Do not resurrect an expired entry; the caller should set it again.
		return false
	}
	e.expireAt = expireAt
	c.ll.MoveToFront(ele)
	return true
}

// remove deletes key from the cache. A missing key is a no-op. The eviction
// callback (if any) fires after the lock is released.
func (c *lruCache[K, V]) remove(key K) {
	c.mu.Lock()
	if c.cache == nil {
		c.mu.Unlock()
		return
	}
	var evicted []lruEntry[K, V]
	if ele, hit := c.cache[key]; hit {
		evicted = []lruEntry[K, V]{*ele.Value.(*lruEntry[K, V])}
		c.removeElementLocked(ele)
	}
	c.mu.Unlock()
	c.fireOnEvict(evicted)
}

// removeOldest evicts the least-recently-used entry; the eviction callback
// fires after unlock.
func (c *lruCache[K, V]) removeOldest() {
	now := c.nowFunc()
	c.mu.Lock()
	e := c.removeOldestLocked(now)
	c.mu.Unlock()
	if e != nil {
		c.fireOnEvict([]lruEntry[K, V]{*e})
	}
}

// removeOldestLocked evicts the back element and returns its entry (without
// firing the callback). Caller holds c.mu.
func (c *lruCache[K, V]) removeOldestLocked(now time.Time) *lruEntry[K, V] {
	if c.cache == nil || c.ll == nil {
		return nil
	}
	ele := c.ll.Back()
	if ele == nil {
		return nil
	}
	e := ele.Value.(*lruEntry[K, V])
	c.removeElementLocked(ele)
	return e
}

// removeExpired scans and evicts every expired entry; callbacks fire after
// unlock.
func (c *lruCache[K, V]) removeExpired() {
	now := c.nowFunc()
	c.mu.Lock()
	if c.cache == nil {
		c.mu.Unlock()
		return
	}
	var evicted []lruEntry[K, V]
	for _, ele := range c.cache {
		e := ele.Value.(*lruEntry[K, V])
		if e.expired(now) {
			evicted = append(evicted, *e)
			c.removeElementLocked(ele)
		}
	}
	c.mu.Unlock()
	c.fireOnEvict(evicted)
}

// removeElementLocked unlinks ele from the list and map without firing the
// callback. Caller holds c.mu.
func (c *lruCache[K, V]) removeElementLocked(ele *list.Element) {
	c.ll.Remove(ele)
	delete(c.cache, ele.Value.(*lruEntry[K, V]).key)
}

// len returns the current entry count (including any not-yet-lazily-expired
// entries).
func (c *lruCache[K, V]) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// clear removes all entries, invoking onEvict for each after unlock.
func (c *lruCache[K, V]) clear() {
	c.mu.Lock()
	if c.cache == nil {
		c.mu.Unlock()
		return
	}
	evicted := make([]lruEntry[K, V], 0, len(c.cache))
	for _, ele := range c.cache {
		e := ele.Value.(*lruEntry[K, V])
		evicted = append(evicted, *e)
		c.ll.Remove(ele)
		delete(c.cache, e.key)
	}
	c.mu.Unlock()
	c.fireOnEvict(evicted)
}

// fireOnEvict invokes the onEvict callback for each entry. It is a no-op when
// there is no callback or nothing to report. Callers must have released c.mu
// before calling, since a callback may re-enter the cache.
func (c *lruCache[K, V]) fireOnEvict(entries []lruEntry[K, V]) {
	if c.onEvict == nil || len(entries) == 0 {
		return
	}
	for _, e := range entries {
		c.onEvict(e.key, e.val)
	}
}
