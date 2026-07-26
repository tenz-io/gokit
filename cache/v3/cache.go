// Package cache provides purely in-process caching primitives: an unbounded
// map cache and a capacity-bounded LRU cache, both with optional per-entry
// TTL expiration. There is no network backend, no Redis, no Lua — everything
// lives in the local process.
//
// V3 is a clean rewrite of cache/v2 with the network surface removed and the
// concurrency contract fixed:
//   - the [Manager] interface keeps Get/Set/SetNx/GetBlob/SetBlob/Del/Expire
//     plus a Close for lifecycle management, but drops Eval (Lua scripts need
//     a remote store) and drops the context.Context argument (a pure-memory
//     cache cannot honor cancellation, so carrying it was misleading);
//   - [NewLocal] returns a map-backed cache with a background sweep goroutine
//     that reaps expired keys; Close stops it and waits for it to exit (v2
//     leaked the goroutine);
//   - [NewLRU] returns an LRU-bounded cache backed by an in-package
//     concurrency-safe generic LRU; eviction callbacks run outside the lock,
//     so a callback may safely re-enter the cache;
//   - there is no default TTL: every Set/SetNx/Expire takes an explicit
//     duration, where a non-positive value means "never expires" and a
//     positive value sets an absolute deadline. Both backends honor the
//     same contract, so swapping them never changes data lifetime;
//   - an injectable clock ([WithNow]) lets tests drive expiration without
//     real sleeps.
package cache

import (
	"errors"
	"time"
)

// Cache operation errors.
var (
	// ErrNotFound is returned when a key is absent or has expired.
	ErrNotFound = errors.New("cache: key not found")
	// ErrInactive is returned when the cache was never initialized (nil
	// receiver) or has been closed.
	ErrInactive = errors.New("cache: inactive")
)

// Manager is the unified cache interface across both backends. Business code
// can swap between [NewLocal] and [NewLRU] without changing call sites; the
// two implementations honor the same expiration contract.
//
// Expiration: every Set/SetNx/Expire takes an explicit expire duration.
// A non-positive expire means the key never expires; a positive expire sets
// an absolute deadline relative to now. This holds for both backends — there
// is no default TTL to surprise a caller who swaps backends.
//
// Expiry of a read key is lazy: Get/GetBlob return ErrNotFound for an expired
// key and remove it. An Expire call on an expired key is a no-op that returns
// ErrNotFound — Expire cannot resurrect a key that has logically expired.
type Manager interface {
	// Get returns the raw string value for key, or ErrNotFound when the key is
	// absent or expired. An expired key is removed lazily on read.
	Get(key string) (raw string, err error)
	// Set stores raw under key with the given expiration. A non-positive
	// expire means the key never expires.
	Set(key string, raw string, expire time.Duration) (err error)
	// SetNx stores raw under key only when the key does not exist (or has
	// expired); it returns existing=true when the key was already present and
	// unexpired. The existence check and the write are atomic with respect to
	// other SetNx/Set calls. A non-positive expire means the key never expires.
	SetNx(key string, raw string, expire time.Duration) (existing bool, err error)
	// GetBlob fetches the JSON-encoded value for key and decodes it into
	// output (which must be a pointer). Returns ErrNotFound on miss or expiry.
	// JSON decoding happens outside the cache's lock.
	GetBlob(key string, output any) (err error)
	// SetBlob JSON-encodes val and stores it under key with the given
	// expiration. A non-positive expire means the key never expires.
	SetBlob(key string, val any, expire time.Duration) (err error)
	// Del removes key. It is a no-op when the key is absent.
	Del(key string) (err error)
	// Expire resets key's expiration. A non-positive expire makes the key
	// never expire. Returns ErrNotFound when the key is absent or has already
	// expired (it will not resurrect an expired key).
	Expire(key string, expire time.Duration) (err error)
	// Close releases any background resources (e.g. the sweep goroutine in
	// [NewLocal]). It is idempotent and safe to call multiple times; backends
	// without resources (the LRU) treat it as a no-op. After Close the cache
	// remains usable for reads/writes; only background reaping stops, and the
	// sweep goroutine is guaranteed to have exited when Close returns.
	Close() error
}

// Option configures a cache at construction time. It is shared by [NewLocal]
// and [NewLRU]; irrelevant options are ignored by a given backend.
type Option func(*options)

type options struct {
	nowFunc       func() time.Time
	evictInterval time.Duration
}

func defaultOptions() options {
	return options{
		nowFunc:       time.Now,
		evictInterval: 5 * time.Minute,
	}
}

// WithNow injects the clock used for all expiration decisions. Tests pass a
// controllable time source so expiration can be advanced without real sleeps;
// production leaves it as time.Now.
func WithNow(now func() time.Time) Option {
	return func(o *options) {
		if now != nil {
			o.nowFunc = now
		}
	}
}

// WithEvictInterval sets how often the local cache's background sweep runs.
// It only affects [NewLocal]; [NewLRU] expires lazily on access. A
// non-positive value disables the background sweep entirely (expiration is
// still enforced lazily on read).
func WithEvictInterval(d time.Duration) Option {
	return func(o *options) {
		o.evictInterval = d
	}
}
