package cache

import (
	"context"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type lru struct {
	impl *expirable.LRU[string, []byte]
}

func NewURL(
	capability int,
) Manager {
	l := &lru{}
	l.impl = expirable.NewLRU[string, []byte](capability, l.onEvict, 5*time.Minute)
	return l
}

func (l *lru) active() bool {
	return l != nil && l.impl != nil
}

func (l *lru) onEvict(key string, val []byte) {

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
	panic("implement me")
}

func (l *lru) SetNx(ctx context.Context, key string, raw string, expire time.Duration) (existing bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (l *lru) GetBlob(ctx context.Context, key string, output any) (err error) {
	//TODO implement me
	panic("implement me")
}

func (l *lru) SetBlob(ctx context.Context, key string, val any, expire time.Duration) (err error) {
	//TODO implement me
	panic("implement me")
}

func (l *lru) Del(ctx context.Context, key string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (l *lru) Expire(ctx context.Context, key string, expire time.Duration) (err error) {
	//TODO implement me
	panic("implement me")
}

func (l *lru) Eval(ctx context.Context, script string, keys []string, args ...any) (val any, err error) {
	//TODO implement me
	panic("implement me")
}
