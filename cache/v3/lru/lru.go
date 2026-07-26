// Package lru is a generic, concurrency-safe LRU cache with optional
// per-entry TTL expiration. It is the backing store for cache.NewLRU and may
// be used directly for non-string key/value caches.
//
// Unlike the v2 lru (which was explicitly not safe for concurrent access),
// this package guards every mutation and read with a sync.Mutex, so it can
// be shared across goroutines without external locking.
package lru

import (
	"container/list"
	"sync"
	"time"
)

// Cache is an LRU cache. A zero-value Cache is not ready for use; construct
// one with [New].
type Cache[K comparable, V any] struct {
	capability int
	// onEvicted, when non-nil, is called for every item purged from the cache:
	// capacity eviction, explicit Remove/RemoveOldest, expiration, and Clear.
	onEvicted func(key K, val V)

	ll    *list.List
	cache map[K]*list.Element

	defaultExpiration time.Duration
	nowFunc           func() time.Time

	mu sync.Mutex
}

// entry is one element of the LRU. It holds the key (for eviction callbacks
// and map removal) the value, and an absolute expiration time.
type entry[K comparable, V any] struct {
	key      K
	val      V
	expireAt time.Time
}

// expired reports whether the entry's deadline has passed. A zero expireAt
// means the entry never expires. The boundary is inclusive: at the exact
// deadline the entry is considered expired.
func (e *entry[K, V]) expired(now time.Time) bool {
	if e.expireAt.IsZero() {
		return false
	}
	return !now.Before(e.expireAt)
}

// New creates an LRU cache.
//
//   - capability caps the entry count; a non-positive value means no limit
//     (expiration is then the only reclamation mechanism).
//   - onEvicted is an optional callback invoked when an item is purged.
//   - expires is the default TTL applied when a Set call passes a zero
//     duration. A positive default TTL keeps short-lived data bounded even
//     when callers forget to pass one; leave zero to mean "no default TTL".
//
// The cache is safe for concurrent use.
func New[K comparable, V any](
	capability int,
	onEvicted func(key K, val V),
	expires time.Duration,
) *Cache[K, V] {
	return &Cache[K, V]{
		capability:        capability,
		ll:                list.New(),
		cache:             make(map[K]*list.Element),
		onEvicted:         onEvicted,
		defaultExpiration: expires,
		nowFunc:           time.Now,
	}
}

// WithNow injects a clock used for all expiration checks. Returns the cache
// for chaining. Tests use it to advance time without real sleeps.
func (c *Cache[K, V]) WithNow(now func() time.Time) *Cache[K, V] {
	if now != nil {
		c.nowFunc = now
	}
	return c
}

// getExpireAt computes the absolute deadline for a new entry given a per-call
// duration. duration == 0 falls back to the cache default; duration < 0 means
// the entry never expires; otherwise the deadline is now+duration.
func (c *Cache[K, V]) getExpireAt(now time.Time, duration time.Duration) time.Time {
	if duration == 0 {
		duration = c.defaultExpiration
	}
	if duration > 0 {
		return now.Add(duration)
	}
	return time.Time{} // zero → never expires
}

// Set adds (or updates) val under key. A duration of 0 applies the cache's
// default TTL; a negative duration makes the entry never expire.
func (c *Cache[K, V]) Set(key K, val V, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		c.cache = make(map[K]*list.Element)
		c.ll = list.New()
	}

	now := c.nowFunc()
	expireAt := c.getExpireAt(now, duration)
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		e := ee.Value.(*entry[K, V])
		e.val = val
		e.expireAt = expireAt
		return
	}

	ele := c.ll.PushFront(&entry[K, V]{key: key, val: val, expireAt: expireAt})
	c.cache[key] = ele
	if c.capability > 0 && c.ll.Len() > c.capability {
		c.removeOldestLocked(now)
	}
}

// Get looks up key. On a hit the entry is moved to the front (most-recently
// used). An expired entry is removed and reported as a miss.
func (c *Cache[K, V]) Get(key K) (val V, ok bool) {
	var zero V
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		return zero, false
	}
	ele, hit := c.cache[key]
	if !hit {
		return zero, false
	}
	now := c.nowFunc()
	if ele.Value.(*entry[K, V]).expired(now) {
		c.removeElementLocked(ele, now)
		return zero, false
	}
	c.ll.MoveToFront(ele)
	return ele.Value.(*entry[K, V]).val, true
}

// Expire updates key's expiration. duration == 0 applies the cache default;
// duration < 0 makes the entry never expire. A missing key is a no-op.
func (c *Cache[K, V]) Expire(key K, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		return
	}
	ele, hit := c.cache[key]
	if !hit {
		return
	}
	now := c.nowFunc()
	ele.Value.(*entry[K, V]).expireAt = c.getExpireAt(now, duration)
	c.ll.MoveToFront(ele)
}

// Remove deletes key from the cache. A missing key is a no-op.
func (c *Cache[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElementLocked(ele, c.nowFunc())
	}
}

// RemoveOldest evicts the least-recently-used entry.
func (c *Cache[K, V]) RemoveOldest() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.removeOldestLocked(c.nowFunc())
}

// removeOldestLocked evicts the back element. Caller holds c.mu.
func (c *Cache[K, V]) removeOldestLocked(now time.Time) {
	if c.cache == nil || c.ll == nil {
		return
	}
	if ele := c.ll.Back(); ele != nil {
		c.removeElementLocked(ele, now)
	}
}

// RemoveExpired scans and evicts every expired entry. Useful for periodic
// sweeping when reads alone are too lazy.
func (c *Cache[K, V]) RemoveExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return
	}
	now := c.nowFunc()
	// Snapshot elements to avoid mutating the list while ranging the map.
	for _, ele := range c.cache {
		if ele.Value.(*entry[K, V]).expired(now) {
			c.removeElementLocked(ele, now)
		}
	}
}

// removeElementLocked removes ele from the list and map, invoking onEvicted.
// Caller holds c.mu.
func (c *Cache[K, V]) removeElementLocked(ele *list.Element, now time.Time) {
	c.ll.Remove(ele)
	e := ele.Value.(*entry[K, V])
	delete(c.cache, e.key)
	if c.onEvicted != nil {
		c.onEvicted(e.key, e.val)
	}
}

// Len returns the current entry count (including any not-yet-lazily-expired
// entries).
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// Clear removes all entries, invoking onEvicted for each.
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return
	}
	for _, ele := range c.cache {
		e := ele.Value.(*entry[K, V])
		c.ll.Remove(ele)
		delete(c.cache, e.key)
		if c.onEvicted != nil {
			c.onEvicted(e.key, e.val)
		}
	}
}
