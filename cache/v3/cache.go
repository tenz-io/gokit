// Package cache provides purely in-process caching primitives: an unbounded
// map cache and a capacity-bounded LRU cache, both with optional per-entry
// TTL expiration. There is no network backend, no Redis, no Lua — everything
// lives in the local process.
//
// V3 is a clean rewrite of cache/v2 with the network surface removed:
//   - the [Manager] interface keeps Get/Set/SetNx/GetBlob/SetBlob/Del/Expire
//     but drops Eval (Lua scripts need a remote store);
//   - [NewLocal] returns a map-backed cache with a background sweep goroutine
//     that reaps expired keys, plus a Close method to stop it (v2 leaked it);
//   - [NewLRU] returns an LRU-bounded cache backed by a concurrency-safe
//     generic LRU (cache/v3/lru), also TTL-aware;
//   - an injectable clock ([WithNow]) lets tests drive expiration without
//     real sleeps, and a configurable sweep interval ([WithEvictInterval]).
package cache

import (
	"context"
	"errors"
	"time"
)

// Cache operation errors.
var (
	// ErrNotFound is returned when a key is absent or has expired.
	ErrNotFound = errors.New("cache: key not found")
	// ErrInActive is returned when the cache was never initialized (nil receiver).
	ErrInActive = errors.New("cache: inactive")
)

// Manager is the unified cache interface across both backends. Business code
// can swap between [NewLocal] and [NewLRU] without changing call sites.
//
// A zero-valued expire means the key never expires; a positive expire sets an
// absolute deadline relative to now.
type Manager interface {
	// Get returns the raw string value for key, or ErrNotFound when the key is
	// absent or expired. An expired key is removed lazily on read.
	Get(ctx context.Context, key string) (raw string, err error)
	// Set stores raw under key with the given expiration. A zero expire means
	// the key never expires.
	Set(ctx context.Context, key string, raw string, expire time.Duration) (err error)
	// SetNx stores raw under key only when the key does not exist (or has
	// expired); it returns existing=true when the key was already present.
	// A zero expire means the key never expires.
	SetNx(ctx context.Context, key string, raw string, expire time.Duration) (existing bool, err error)
	// GetBlob fetches the JSON-encoded value for key and decodes it into
	// output (which must be a pointer). Returns ErrNotFound on miss.
	GetBlob(ctx context.Context, key string, output any) (err error)
	// SetBlob JSON-encodes val and stores it under key with the given
	// expiration. A zero expire means the key never expires.
	SetBlob(ctx context.Context, key string, val any, expire time.Duration) (err error)
	// Del removes key. It is a no-op when the key is absent.
	Del(ctx context.Context, key string) (err error)
	// Expire resets key's expiration. A zero expire makes the key never
	// expire. Returns ErrNotFound when the key is absent.
	Expire(ctx context.Context, key string, expire time.Duration) (err error)
	// Close releases any background resources (e.g. the sweep goroutine in
	// [NewLocal]). It is idempotent and safe to call multiple times; backends
	// without resources (the LRU) treat it as a no-op. After Close the cache
	// remains usable for reads/writes; only background reaping stops.
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
// It only affects [NewLocal]; [NewLRU] expires lazily on access. A negative
// or zero value disables the background sweep entirely.
func WithEvictInterval(d time.Duration) Option {
	return func(o *options) {
		o.evictInterval = d
	}
}
