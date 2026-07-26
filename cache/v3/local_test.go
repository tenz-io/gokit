package cache

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"
)

// fakeClock is a controllable time source so expiration tests never depend
// on real time.Sleep. localCache stores deadlines as unix seconds, so the
// clock only needs to advance in those increments.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Unix(1_700_000_000, 0)}
}

func (f *fakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *fakeClock) advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
}

// newTestLocal builds a localCache with the sweep disabled and a fake clock,
// returning both so a test can advance time and Close the cache afterwards.
func newTestLocal(t *testing.T) (Manager, *fakeClock, func()) {
	t.Helper()
	clk := newFakeClock()
	mgr := NewLocal(WithNow(clk.Now), WithEvictInterval(0))
	return mgr, clk, func() {
		if c, ok := mgr.(*localCache); ok {
			_ = c.Close()
		}
	}
}

func TestLocal_Get_Set(t *testing.T) {
	mgr, _, cleanup := newTestLocal(t)
	defer cleanup()
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

func TestLocal_Expire(t *testing.T) {
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()
	ctx := context.Background()

	if err := mgr.Set(ctx, "k", "v", time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Second)

	if _, err := mgr.Get(ctx, "k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after expiry = %v, want ErrNotFound", err)
	}
}

func TestLocal_NeverExpire(t *testing.T) {
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()
	ctx := context.Background()

	if err := mgr.Set(ctx, "k", "v", 0); err != nil { // 0 → never expires
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Hour)
	if got, err := mgr.Get(ctx, "k"); err != nil || got != "v" {
		t.Fatalf("Get after long time = (%q,%v), want (v,nil)", got, err)
	}
}

func TestLocal_Expire_Method(t *testing.T) {
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()
	ctx := context.Background()

	// Set never-expire, then shorten with Expire.
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

func TestLocal_SetNx(t *testing.T) {
	mgr, _, cleanup := newTestLocal(t)
	defer cleanup()
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

func TestLocal_SetNx_WritesThroughExpired(t *testing.T) {
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()
	ctx := context.Background()

	if err := mgr.Set(ctx, "k", "old", time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Second)

	existing, err := mgr.SetNx(ctx, "k", "new", 0)
	if err != nil || existing {
		t.Fatalf("SetNx on expired key = (%v,%v), want (false,nil)", existing, err)
	}
	if got, _ := mgr.Get(ctx, "k"); got != "new" {
		t.Fatalf("value = %q, want 'new' (expired key should be replaceable)", got)
	}
}

func TestLocal_Blob(t *testing.T) {
	mgr, _, cleanup := newTestLocal(t)
	defer cleanup()
	ctx := context.Background()

	type user struct {
		Name string
		Age  int
	}

	// GetBlob before set → ErrNotFound.
	var u user
	if err := mgr.GetBlob(ctx, "u:1", &u); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetBlob missing = %v, want ErrNotFound", err)
	}
	if err := mgr.SetBlob(ctx, "u:1", user{Name: "tom", Age: 18}, 0); err != nil {
		t.Fatalf("SetBlob: %v", err)
	}
	var got user
	if err := mgr.GetBlob(ctx, "u:1", &got); err != nil {
		t.Fatalf("GetBlob: %v", err)
	}
	if got.Name != "tom" || got.Age != 18 {
		t.Fatalf("decoded = %+v, want {tom 18}", got)
	}
}

func TestLocal_Blob_Expiry(t *testing.T) {
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()
	ctx := context.Background()

	type payload struct{ V int }
	if err := mgr.SetBlob(ctx, "p", payload{V: 1}, time.Second); err != nil {
		t.Fatalf("SetBlob: %v", err)
	}
	clk.advance(time.Second)
	var p payload
	if err := mgr.GetBlob(ctx, "p", &p); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetBlob after expiry = %v, want ErrNotFound", err)
	}
}

func TestLocal_Del(t *testing.T) {
	mgr, _, cleanup := newTestLocal(t)
	defer cleanup()
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
	// Del a missing key is a no-op.
	if err := mgr.Del(ctx, "nope"); err != nil {
		t.Fatalf("Del missing = %v, want nil", err)
	}
}

func TestLocal_EvictBackground(t *testing.T) {
	// Use a short sweep interval to confirm the background loop removes
	// expired keys on its own — independent of any Get triggering lazy
	// deletion. We verify via the internal map after a real tick fires.
	clk := newFakeClock()
	mgr := NewLocal(WithNow(clk.Now), WithEvictInterval(20*time.Millisecond))
	defer func() {
		if c, ok := mgr.(*localCache); ok {
			_ = c.Close()
		}
	}()
	ctx := context.Background()

	if err := mgr.Set(ctx, "k", "v", time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(2 * time.Second)

	// Poll the internal map for up to ~0.5s for the sweep to clear the key.
	// This relies on real wall-clock ticks; the fake clock only drives the
	// expiration decision, while the ticker cadence is real.
	c := mgr.(*localCache)
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		c.lock.RLock()
		_, present := c.m["k"]
		c.lock.RUnlock()
		if !present {
			return // swept
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("background sweep did not remove the expired key in time")
}

func TestLocal_Close_Idempotent(t *testing.T) {
	mgr := NewLocal(WithEvictInterval(time.Hour))
	c := mgr.(*localCache)
	if err := c.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := c.Close(); err != nil { // must not panic on double-close
		t.Fatalf("second Close: %v", err)
	}
}

func TestLocal_Inactive(t *testing.T) {
	// A nil receiver cache reports ErrInActive for every op.
	var nilMgr Manager = (*localCache)(nil)
	ctx := context.Background()
	if _, err := nilMgr.Get(ctx, "k"); !errors.Is(err, ErrInActive) {
		t.Fatalf("nil Get = %v, want ErrInActive", err)
	}
	if err := nilMgr.Set(ctx, "k", "v", 0); !errors.Is(err, ErrInActive) {
		t.Fatalf("nil Set = %v, want ErrInActive", err)
	}
}

func TestLocal_Concurrent(t *testing.T) {
	// Stress the read/write lock under -race.
	mgr, _, cleanup := newTestLocal(t)
	defer cleanup()
	ctx := context.Background()

	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func(off int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				key := strconv.Itoa(off*1000 + i)
				_ = mgr.Set(ctx, key, "v", 0)
				_, _ = mgr.Get(ctx, key)
				_ = mgr.Del(ctx, key)
			}
		}(g)
	}
	wg.Wait()
}
