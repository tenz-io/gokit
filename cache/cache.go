package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("cache: key not found")
	ErrInActive = errors.New("cache: inactive")
)

//go:generate mockery --name Manager --filename Manager_mock.go --inpackage
type Manager interface {
	// Get returns the value associated with the given key.
	Get(ctx context.Context, key string) (raw string, err error)
	// Set stores the given value with the given key.
	// if expire is 0, then the key will not expire.
	Set(ctx context.Context, key string, raw string, expire time.Duration) (err error)
	// SetNx stores the given value with the given key if the key does not exist.
	// if expire is 0, then the key will not expire.
	SetNx(ctx context.Context, key string, raw string, expire time.Duration) (existing bool, err error)
	// GetBlob returns the value associated with the given key.
	GetBlob(ctx context.Context, key string, output any) (err error)
	// SetBlob stores the given value with the given key.
	// if expire is 0, then the key will not expire.
	SetBlob(ctx context.Context, key string, val any, expire time.Duration) (err error)
	// Del deletes the given key.
	Del(ctx context.Context, key string) (err error)
	// Expire sets the expiration for the given key.
	// if expire is 0, then the key will not expire.
	Expire(ctx context.Context, key string, expire time.Duration) (err error)
	// Eval evaluates the given script with the given keys and arguments.
	Eval(ctx context.Context, script string, keys []string, args ...any) (val any, err error)
}
