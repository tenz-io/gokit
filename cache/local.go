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

type localCache struct {
	m       map[string]*item
	nowFunc func() time.Time
	lock    sync.RWMutex
}

func NewLocal() Manager {
	lm := &localCache{
		m:       make(map[string]*item),
		nowFunc: time.Now,
	}

	lm.startEvict(5 * time.Minute)

	return lm
}

func (lc *localCache) active() bool {
	if lc == nil || lc.m == nil {
		return false
	}
	return true
}

// startEvict evict expired with interval
func (lc *localCache) startEvict(interval time.Duration) {
	if !lc.active() {
		return
	}

	go func() {
		for {
			lc.evict()
			time.Sleep(interval)
		}
	}()
}

// evict expired items
func (lc *localCache) evict() {
	if !lc.active() {
		return
	}

	lc.lock.Lock()
	defer lc.lock.Unlock()

	now := lc.nowFunc().Unix()
	for k, v := range lc.m {
		if v.expire != 0 && now > v.expire {
			delete(lc.m, k)
		}
	}
}

func (lc *localCache) Get(ctx context.Context, key string) (raw string, err error) {
	if !lc.active() {
		return "", ErrInActive
	}

	var needDel bool
	lc.lock.RLock()
	defer func() {
		lc.lock.RUnlock()
		if needDel {
			go lc.Del(ctx, key)
		}
	}()

	it, found := lc.m[key]
	if !found {
		return "", ErrNotFound
	}

	if it == nil {
		needDel = true
		return "", ErrNotFound
	}

	if it.expire == 0 || lc.nowFunc().Unix() < it.expire {
		return string(it.raw), nil
	} else {
		needDel = true
		return "", ErrNotFound
	}

}

func (lc *localCache) Set(_ context.Context, key string, raw string, expire time.Duration) (err error) {
	if !lc.active() {
		return ErrInActive
	}

	lc.lock.Lock()
	defer lc.lock.Unlock()

	lc.m[key] = &item{
		raw:    []byte(raw),
		expire: lc.expireAt(expire),
	}
	return nil
}

func (lc *localCache) SetNx(_ context.Context, key string, raw string, expire time.Duration) (existing bool, err error) {
	if !lc.active() {
		return false, ErrInActive
	}

	lc.lock.Lock()
	defer lc.lock.Unlock()

	if _, ok := lc.m[key]; ok {
		return true, nil
	} else {
		lc.m[key] = &item{
			raw:    []byte(raw),
			expire: lc.expireAt(expire),
		}
		return false, nil
	}
}

func (lc *localCache) GetBlob(ctx context.Context, key string, output any) (err error) {
	if !lc.active() {
		return ErrInActive
	}

	var needDel bool
	lc.lock.RLock()
	defer func() {
		lc.lock.RUnlock()
		if needDel {
			go lc.Del(ctx, key)
		}
	}()

	it, found := lc.m[key]
	if !found {
		return ErrNotFound
	}

	if it == nil {
		// invalid item
		needDel = true
		return ErrNotFound
	}

	if it.expire == 0 || lc.nowFunc().Unix() < it.expire {
		r := bytes.NewReader(it.raw)
		decoder := gob.NewDecoder(r)
		if err = decoder.Decode(output); err != nil {
			return fmt.Errorf("decode error: %w", err)
		}
		return nil
	} else {
		// expired
		needDel = true
		return ErrNotFound
	}

}

func (lc *localCache) SetBlob(_ context.Context, key string, val any, expire time.Duration) (err error) {
	if !lc.active() {
		return ErrInActive
	}

	lc.lock.Lock()
	defer lc.lock.Unlock()

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err = encoder.Encode(val); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	lc.m[key] = &item{
		raw:    buf.Bytes(),
		expire: lc.expireAt(expire),
	}
	return nil

}

func (lc *localCache) Del(_ context.Context, key string) (err error) {
	if !lc.active() {
		return ErrInActive
	}

	lc.lock.Lock()
	defer lc.lock.Unlock()

	if _, ok := lc.m[key]; ok {
		delete(lc.m, key)
	}
	return nil
}

func (lc *localCache) Expire(_ context.Context, key string, expire time.Duration) (err error) {
	if !lc.active() {
		return ErrInActive
	}

	lc.lock.Lock()
	defer lc.lock.Unlock()
	if it, ok := lc.m[key]; ok && it != nil {
		it.expire = lc.expireAt(expire)
		return nil
	} else {
		return ErrNotFound
	}

}

func (lc *localCache) Eval(_ context.Context, script string, keys []string, args ...any) (val any, err error) {
	// ignore
	return nil, ErrNotSupported
}

func (lc *localCache) expireAt(expire time.Duration) int64 {
	if expire == 0 {
		return 0
	} else {
		return lc.nowFunc().Add(expire).Unix()
	}
}
