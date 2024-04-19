package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

func NewRedis(
	client *redis.Client,
) Manager {
	return &redisCache{
		client: client,
	}
}

type redisCache struct {
	client *redis.Client
}

func (m *redisCache) active() bool {
	return m != nil && m.client != nil
}

func (m *redisCache) Get(ctx context.Context, key string) (raw string, err error) {
	if !m.active() {
		return "", ErrInActive
	}
	raw, err = m.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotFound
		}
		return "", err
	}

	return raw, nil
}

func (m *redisCache) Set(ctx context.Context, key string, raw string, expire time.Duration) (err error) {
	if !m.active() {
		return ErrInActive
	}

	err = m.client.Set(ctx, key, raw, expire).Err()
	return
}

func (m *redisCache) SetNx(ctx context.Context, key string, raw string, expire time.Duration) (existing bool, err error) {
	if !m.active() {
		return false, ErrInActive
	}

	existing, err = m.client.SetNX(ctx, key, raw, expire).Result()
	return
}

func (m *redisCache) GetBlob(ctx context.Context, key string, output any) (err error) {
	if !m.active() {
		return ErrInActive
	}

	bs, err := m.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrNotFound
		}
		return err
	}

	r := bytes.NewReader(bs)
	decoder := gob.NewDecoder(r)
	if err = decoder.Decode(output); err != nil {
		return fmt.Errorf("decode error: %w", err)
	}
	return nil
}

func (m *redisCache) SetBlob(ctx context.Context, key string, val any, expire time.Duration) (err error) {
	if !m.active() {
		return ErrInActive
	}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err = encoder.Encode(val); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	// expire is 0, then set no expire
	// expire is -1, then set default expire
	if err = m.client.Set(ctx, key, buf.Bytes(), expire).Err(); err != nil {
		return fmt.Errorf("set error: %w", err)
	}
	return nil

}

func (m *redisCache) Del(ctx context.Context, key string) (err error) {
	if !m.active() {
		return ErrInActive
	}

	err = m.client.Del(ctx, key).Err()
	return
}

func (m *redisCache) Expire(ctx context.Context, key string, expire time.Duration) (err error) {
	if !m.active() {
		return ErrInActive
	}

	err = m.client.Expire(ctx, key, expire).Err()
	return
}

func (m *redisCache) Eval(ctx context.Context, script string, keys []string, args ...any) (val any, err error) {
	if !m.active() {
		return nil, ErrInActive
	}

	val, err = m.client.Eval(ctx, script, keys, args...).Result()
	return
}
