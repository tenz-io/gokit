package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type lru struct {
	impl *expirable.LRU[string, []byte]
}

func NewURL(
	capability int,
	onEvict func(key string, val []byte),
	expire time.Duration,
) Manager {
	l := &lru{}
	if capability <= 0 {
		capability = 150
	}
	if expire <= 0 {
		expire = 15 * time.Minute
	}
	l.impl = expirable.NewLRU[string, []byte](capability, onEvict, expire)
	return l
}

func (l *lru) active() bool {
	return l != nil && l.impl != nil
}

func (l *lru) Get(ctx context.Context, key string) (raw string, err error) {
	if !l.active() {
		return "", ErrInActive
	}

	bs, ok := l.impl.Get(key)
	if !ok {
		return "", ErrNotFound
	}

	return string(bs), nil

}

func (l *lru) Set(ctx context.Context, key string, raw string, expire time.Duration) (err error) {
	if !l.active() {
		return ErrInActive
	}

	l.impl.Add(key, []byte(raw))
	return nil
}

func (l *lru) SetNx(ctx context.Context, key string, raw string, expire time.Duration) (existing bool, err error) {
	if !l.active() {
		return false, ErrInActive
	}

	if _, ok := l.impl.Get(key); ok {
		return true, nil
	}

	l.impl.Add(key, []byte(raw))
	return false, nil
}

func (l *lru) GetBlob(ctx context.Context, key string, output any) (err error) {
	if !l.active() {
		return ErrInActive
	}

	bs, ok := l.impl.Get(key)
	if !ok {
		return ErrNotFound
	}

	decoder := gob.NewDecoder(bytes.NewReader(bs))
	if err = decoder.Decode(output); err != nil {
		return fmt.Errorf("decode error: %w", err)
	}
	return nil
}

func (l *lru) SetBlob(ctx context.Context, key string, val any, expire time.Duration) (err error) {
	if !l.active() {
		return ErrInActive
	}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err = encoder.Encode(val); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	l.impl.Add(key, buf.Bytes())
	return nil
}

func (l *lru) Del(ctx context.Context, key string) (err error) {
	if !l.active() {
		return ErrInActive
	}

	l.impl.Remove(key)
	return nil
}

func (l *lru) Expire(ctx context.Context, key string, expire time.Duration) (err error) {
	// skip
	return nil
}

func (l *lru) Eval(ctx context.Context, script string, keys []string, args ...any) (val any, err error) {
	// not supported
	return nil, ErrNotSupported
}
