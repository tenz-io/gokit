package lru

import (
	"container/list"
	"time"
)

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache[K comparable, V any] struct {
	// capability is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	capability int

	// onEvicted optionally specifies a callback function to be
	// executed when an item is purged from the cache.
	onEvicted func(key K, val V)

	ll    *list.List
	cache map[K]*list.Element

	defaultExpiration time.Duration
}

type item[K comparable, V any] struct {
	key      K
	val      V
	expireAt time.Time
}

func (i *item[K, V]) expired() bool {
	if i.expireAt.IsZero() {
		return false
	}
	return i.expireAt.Before(time.Now())
}

// New creates a new Cache with a default expire duration
// If capability is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func New[K comparable, V any](
	capability int,
	onEvicted func(key K, val V),
	expires time.Duration,
) *Cache[K, V] {
	if expires == 0 {
		expires = -1
	}
	return &Cache[K, V]{
		capability:        capability,
		ll:                list.New(),
		cache:             make(map[K]*list.Element),
		onEvicted:         onEvicted,
		defaultExpiration: expires,
	}
}

// Expired Returns true if the item has expired.
func (c *Cache[K, V]) Expired(key K) bool {
	ele, hit := c.cache[key]
	if !hit {
		return false
	}
	return ele.Value.(*item[K, V]).expired()
}

// getExpireAt returns the expireAt time for a newly added item.
// If the duration is 0, the cache's default expireAt time is used.
// If it is -1, the item never expires.
func (c *Cache[K, V]) getExpireAt(expires time.Duration) time.Time {
	var expireAt time.Time
	if expires == 0 {
		expires = c.defaultExpiration
	}
	if expires > 0 {
		expireAt = time.Now().Add(expires)
	}
	return expireAt
}

// Set adds a val to the cache. If the duration is 0,
// the cache's default expireAt time is used. If it is -1, the item never
// expires.
func (c *Cache[K, V]) Set(key K, val V, expires time.Duration) {
	if c.cache == nil {
		c.cache = make(map[K]*list.Element)
		c.ll = list.New()
	}

	e := c.getExpireAt(expires)
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*item[K, V]).val = val
		ee.Value.(*item[K, V]).expireAt = e
		return
	}

	ele := c.ll.PushFront(&item[K, V]{key, val, e})
	c.cache[key] = ele
	if c.capability != 0 && c.ll.Len() > c.capability {
		c.RemoveOldest()
	}
}

// Get looks up a key's val from the cache.
func (c *Cache[K, V]) Get(key K) (val V, existing bool) {
	var (
		zeroVal V
	)
	if c.cache == nil {
		return zeroVal, false
	}
	ele, hit := c.cache[key]
	if !hit {
		return zeroVal, false
	}
	if ele.Value.(*item[K, V]).expired() {
		c.removeElement(ele)
		return zeroVal, false
	}
	c.ll.MoveToFront(ele)
	return ele.Value.(*item[K, V]).val, true
}

// Expire updates the expiration of a key.
// If the key does not exist, this function does nothing.
// If the duration is 0, the cache's default expiration time is used.
// If it is -1, the item never expires.
func (c *Cache[K, V]) Expire(key K, expires time.Duration) {
	if c.cache == nil {
		return
	}

	ele, hit := c.cache[key]
	if !hit {
		return
	}

	expireAt := c.getExpireAt(expires)
	ele.Value.(*item[K, V]).expireAt = expireAt
	c.ll.MoveToFront(ele)
}

// Remove removes the provided key from the cache.
func (c *Cache[K, V]) Remove(key K) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache[K, V]) RemoveOldest() {
	if c.cache == nil {
		return
	}
	if ele := c.ll.Back(); ele != nil {
		c.removeElement(ele)
	}
}

// RemoveExpired removes all expired items from the cache.
func (c *Cache[K, V]) RemoveExpired() {
	if c.cache == nil {
		return
	}
	for _, ele := range c.cache {
		if ele.Value.(*item[K, V]).expired() {
			c.removeElement(ele)
		}
	}
}

func (c *Cache[K, V]) removeElement(elem *list.Element) {
	c.ll.Remove(elem)
	kv := elem.Value.(*item[K, V])
	delete(c.cache, kv.key)
	if c.onEvicted != nil {
		c.onEvicted(kv.key, kv.val)
	}
}

// Len returns the number of items in the cache.
func (c *Cache[K, V]) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// Clear purges all stored items from the cache.
func (c *Cache[K, V]) Clear() {
	if c.onEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*item[K, V])
			c.onEvicted(kv.key, kv.val)
		}
	}
	c.ll = list.New()
	c.cache = make(map[K]*list.Element)
}
