package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

// newTestLRU builds an lruCache Manager with a fake clock so expiration tests
// need no real sleeps. The sweep option is irrelevant for the LRU backend
// (it expires lazily) and is omitted.
func newTestLRU(t *testing.T, capability int, expire time.Duration) (Manager, *fakeClock) {
	t.Helper()
	clk := newFakeClock()
	mgr := NewLRU(capability, nil, expire, WithNow(clk.Now))
	return mgr, clk
}

func TestLRU_Get_Set(t *testing.T) {
	mgr, _ := newTestLRU(t, 10, 0)
	ctx := context.Background()

	if _, err := mgr.Get(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get missing = %v, want ErrNotFound", err)
	}
	if err := mgr.Set(ctx, "k", "v", 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := mgr.Get(ctx, "k")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "v" {
		t.Fatalf("Get = %q, want %q", got, "v")
	}
}

func TestLRU_EvictOnCapacity(t *testing.T) {
	var evicted []string
	clk := newFakeClock()
	mgr := NewLRU(2, func(key string, _ []byte) { evicted = append(evicted, key) }, 0, WithNow(clk.Now))
	ctx := context.Background()

	if err := mgr.Set(ctx, "a", "1", 0); err != nil {
		t.Fatalf("Set a: %v", err)
	}
	if err := mgr.Set(ctx, "b", "2", 0); err != nil {
		t.Fatalf("Set b: %v", err)
	}
	if err := mgr.Set(ctx, "c", "3", 0); err != nil { // evicts "a"
		t.Fatalf("Set c: %v", err)
	}
	if len(evicted) != 1 || evicted[0] != "a" {
		t.Fatalf("evicted = %v, want [a]", evicted)
	}
	if _, err := mgr.Get(ctx, "a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get a after eviction = %v, want ErrNotFound", err)
	}
}

func TestLRU_Expire(t *testing.T) {
	mgr, clk := newTestLRU(t, 0, 0)
	ctx := context.Background()

	if err := mgr.Set(ctx, "k", "v", time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Second)
	if _, err := mgr.Get(ctx, "k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after expiry = %v, want ErrNotFound", err)
	}
}

func TestLRU_DefaultExpiration(t *testing.T) {
	// Default TTL of 5s applies when a Set call passes duration 0.
	clk := newFakeClock()
	mgr := NewLRU(0, nil, 5*time.Second, WithNow(clk.Now))
	ctx := context.Background()

	if err := mgr.Set(ctx, "k", "v", 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(4 * time.Second)
	if got, err := mgr.Get(ctx, "k"); err != nil || got != "v" {
		t.Fatalf("Get at 4s = (%q,%v), want (v,nil)", got, err)
	}
	clk.advance(2 * time.Second)
	if _, err := mgr.Get(ctx, "k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get at 6s = %v, want ErrNotFound (past 5s default)", err)
	}
}

func TestLRU_SetNx(t *testing.T) {
	mgr, _ := newTestLRU(t, 10, 0)
	ctx := context.Background()

	if existing, err := mgr.SetNx(ctx, "k", "v1", 0); err != nil || existing {
		t.Fatalf("first SetNx = (%v,%v), want (false,nil)", existing, err)
	}
	if existing, err := mgr.SetNx(ctx, "k", "v2", 0); err != nil || !existing {
		t.Fatalf("second SetNx = (%v,%v), want (true,nil)", existing, err)
	}
	if got, _ := mgr.Get(ctx, "k"); got != "v1" {
		t.Fatalf("value = %q, SetNx must not overwrite an existing key", got)
	}
}

func TestLRU_Blob(t *testing.T) {
	mgr, _ := newTestLRU(t, 10, 0)
	ctx := context.Background()

	type item struct{ Name string }
	if err := mgr.SetBlob(ctx, "i", item{Name: "tom"}, 0); err != nil {
		t.Fatalf("SetBlob: %v", err)
	}
	var got item
	if err := mgr.GetBlob(ctx, "i", &got); err != nil {
		t.Fatalf("GetBlob: %v", err)
	}
	if got.Name != "tom" {
		t.Fatalf("decoded = %+v, want {tom}", got)
	}
}

func TestLRU_Del(t *testing.T) {
	mgr, _ := newTestLRU(t, 10, 0)
	ctx := context.Background()

	if err := mgr.Set(ctx, "k", "v", 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := mgr.Del(ctx, "k"); err != nil {
		t.Fatalf("Del: %v", err)
	}
	if _, err := mgr.Get(ctx, "k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after Del = %v, want ErrNotFound", err)
	}
}

func TestLRU_Expire_Method(t *testing.T) {
	mgr, clk := newTestLRU(t, 10, 0)
	ctx := context.Background()

	if err := mgr.Set(ctx, "k", "v", 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := mgr.Expire(ctx, "k", time.Second); err != nil {
		t.Fatalf("Expire: %v", err)
	}
	clk.advance(time.Second)
	if _, err := mgr.Get(ctx, "k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after reset expiry = %v, want ErrNotFound", err)
	}

	// Expire on a missing key returns ErrNotFound.
	if err := mgr.Expire(ctx, "nope", time.Second); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expire missing = %v, want ErrNotFound", err)
	}
}

func TestLRU_Inactive(t *testing.T) {
	var nilMgr Manager = (*lruCache)(nil)
	ctx := context.Background()
	if _, err := nilMgr.Get(ctx, "k"); !errors.Is(err, ErrInActive) {
		t.Fatalf("nil Get = %v, want ErrInActive", err)
	}
}
