package cache

import (
	"context"
	"encoding/json"
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

func (rc *redisCache) active() bool {
	return rc != nil && rc.client != nil
}

func (rc *redisCache) Get(ctx context.Context, key string) (raw string, err error) {
	if !rc.active() {
		return "", ErrInActive
	}
	raw, err = rc.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotFound
		}
		return "", err
	}

	return raw, nil
}

func (rc *redisCache) Set(ctx context.Context, key string, raw string, expire time.Duration) (err error) {
	if !rc.active() {
		return ErrInActive
	}

	err = rc.client.Set(ctx, key, raw, expire).Err()
	return
}

func (rc *redisCache) SetNx(ctx context.Context, key string, raw string, expire time.Duration) (existing bool, err error) {
	if !rc.active() {
		return false, ErrInActive
	}

	existing, err = rc.client.SetNX(ctx, key, raw, expire).Result()
	return
}

func (rc *redisCache) GetBlob(ctx context.Context, key string, output any) (err error) {
	if !rc.active() {
		return ErrInActive
	}

	bs, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrNotFound
		}
		return err
	}

	if err = json.Unmarshal(bs, output); err != nil {
		return fmt.Errorf("decode error: %w", err)
	}
	return nil
}

func (rc *redisCache) SetBlob(ctx context.Context, key string, val any, expire time.Duration) (err error) {
	if !rc.active() {
		return ErrInActive
	}

	bs, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	// expire is 0, then set no expire
	// expire is -1, then set default expire
	if err = rc.client.Set(ctx, key, bs, expire).Err(); err != nil {
		return fmt.Errorf("set error: %w", err)
	}
	return nil

}

func (rc *redisCache) Del(ctx context.Context, key string) (err error) {
	if !rc.active() {
		return ErrInActive
	}

	err = rc.client.Del(ctx, key).Err()
	return
}

func (rc *redisCache) Expire(ctx context.Context, key string, expire time.Duration) (err error) {
	if !rc.active() {
		return ErrInActive
	}

	err = rc.client.Expire(ctx, key, expire).Err()
	return
}

func (rc *redisCache) Eval(ctx context.Context, script string, keys []string, args ...any) (val any, err error) {
	if !rc.active() {
		return nil, ErrInActive
	}

	val, err = rc.client.Eval(ctx, script, keys, args...).Result()
	return
}
