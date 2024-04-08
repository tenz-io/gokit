package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sync"
	"time"
)

type item struct {
	raw    []byte
	expire int64
}

type local struct {
	m       map[string]*item
	nowFunc func() time.Time
	lock    sync.RWMutex
}

func NewLocal() Manager {
	lm := &local{
		m:       make(map[string]*item),
		nowFunc: time.Now,
	}

	lm.startEvict(5 * time.Minute)

	return lm
}

func (l *local) active() bool {
	if l == nil || l.m == nil {
		return false
	}
	return true
}

// startEvict evict expired with interval
func (l *local) startEvict(interval time.Duration) {
	if !l.active() {
		return
	}

	go func() {
		for {
			l.evict()
			time.Sleep(interval)
		}
	}()
}

// evict expired items
func (l *local) evict() {
	if !l.active() {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	now := l.nowFunc().Unix()
	for k, v := range l.m {
		if v.expire != 0 && now > v.expire {
			delete(l.m, k)
		}
	}
}

func (l *local) Get(ctx context.Context, key string) (raw string, err error) {
	if !l.active() {
		return "", ErrInActive
	}

	l.lock.RLock()

	it, found := l.m[key]
	if !found {
		defer l.lock.RUnlock()
		return "", ErrNotFound
	}

	if it == nil {
		l.lock.RUnlock()

		l.lock.Lock()
		defer l.lock.Unlock()
		delete(l.m, key)
		return "", ErrNotFound
	}

	if it.expire == 0 || l.nowFunc().Unix() < it.expire {
		defer l.lock.RUnlock()
		return string(it.raw), nil
	} else {
		l.lock.RUnlock()

		l.lock.Lock()
		defer l.lock.Unlock()
		delete(l.m, key)
		return "", ErrNotFound
	}

}

func (l *local) Set(ctx context.Context, key string, raw string, expire time.Duration) (err error) {
	if !l.active() {
		return ErrInActive
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	l.m[key] = &item{
		raw:    []byte(raw),
		expire: l.expireAt(expire),
	}
	return nil
}

func (l *local) SetNx(ctx context.Context, key string, raw string, expire time.Duration) (existing bool, err error) {
	if !l.active() {
		return false, ErrInActive
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if _, ok := l.m[key]; ok {
		return true, nil
	} else {
		l.m[key] = &item{
			raw:    []byte(raw),
			expire: l.expireAt(expire),
		}
		return false, nil
	}
}

func (l *local) GetBlob(ctx context.Context, key string, output any) (err error) {
	if !l.active() {
		return ErrInActive
	}

	l.lock.RLock()
	it, found := l.m[key]
	if !found {
		defer l.lock.RUnlock()
		return ErrNotFound
	}

	if it == nil {
		l.lock.RUnlock()

		l.lock.Lock()
		defer l.lock.Unlock()
		delete(l.m, key)
		return ErrNotFound
	}

	if it.expire == 0 || l.nowFunc().Unix() < it.expire {
		defer l.lock.RUnlock()

		r := bytes.NewReader(it.raw)
		decoder := gob.NewDecoder(r)
		if err = decoder.Decode(output); err != nil {
			return fmt.Errorf("decode error: %w", err)
		}
		return nil
	} else {
		l.lock.RUnlock()

		l.lock.Lock()
		defer l.lock.Unlock()
		delete(l.m, key)
		return ErrNotFound
	}

}

func (l *local) SetBlob(ctx context.Context, key string, val any, expire time.Duration) (err error) {
	if !l.active() {
		return ErrInActive
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err = encoder.Encode(val); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	l.m[key] = &item{
		raw:    buf.Bytes(),
		expire: l.expireAt(expire),
	}
	return nil

}

func (l *local) Del(ctx context.Context, key string) (err error) {
	if !l.active() {
		return ErrInActive
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if _, ok := l.m[key]; ok {
		delete(l.m, key)
	}
	return nil
}

func (l *local) Expire(ctx context.Context, key string, expire time.Duration) (err error) {
	if !l.active() {
		return ErrInActive
	}

	l.lock.Lock()
	defer l.lock.Unlock()
	if it, ok := l.m[key]; ok && it != nil {
		it.expire = l.expireAt(expire)
		return nil
	} else {
		return ErrNotFound
	}

}

func (l *local) Eval(ctx context.Context, script string, keys []string, args ...any) (val any, err error) {
	// ignore
	return nil, fmt.Errorf("not support")
}

func (l *local) expireAt(expire time.Duration) int64 {
	if expire == 0 {
		return 0
	} else {
		return l.nowFunc().Add(expire).Unix()
	}
}
