package lru

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

// fakeClock is a controllable time source for expiration tests. It advances
// only when the test calls advance, so no test depends on real sleeps.
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

func TestCache_SetAndGet(t *testing.T) {
	c := New[string, int](0, nil, 0)
	c.Set("k", 1234, 0)

	got, ok := c.Get("k")
	if !ok {
		t.Fatal("expected a hit")
	}
	if got != 1234 {
		t.Fatalf("got %v, want 1234", got)
	}
}

func TestCache_Get_Miss(t *testing.T) {
	c := New[string, int](10, nil, 0)
	if _, ok := c.Get("missing"); ok {
		t.Fatal("expected a miss for a never-set key")
	}
}

func TestCache_Set_UpdateInPlace(t *testing.T) {
	c := New[string, int](0, nil, 0)
	c.Set("k", 1, 0)
	c.Set("k", 2, 0)
	if got, _ := c.Get("k"); got != 2 {
		t.Fatalf("got %v, want updated value 2", got)
	}
	if c.Len() != 1 {
		t.Fatalf("len = %d, want 1 (update must not grow the cache)", c.Len())
	}
}

func TestCache_Expire(t *testing.T) {
	clk := newFakeClock()
	c := New[string, int](0, nil, 0).WithNow(clk.Now)

	c.Set("k", 1234, time.Second)
	if _, ok := c.Get("k"); !ok {
		t.Fatal("expected a hit before expiry")
	}

	clk.advance(time.Second)
	if _, ok := c.Get("k"); ok {
		t.Fatal("expected a miss after expiry")
	}
	if c.Len() != 0 {
		t.Fatalf("len = %d, want 0 (expired entry should be removed)", c.Len())
	}
}

func TestCache_Expire_Method(t *testing.T) {
	clk := newFakeClock()
	c := New[string, int](0, nil, 0).WithNow(clk.Now)

	// Set with no TTL, then shorten via Expire.
	c.Set("k", 1, 0)
	c.Expire("k", time.Second)
	if _, ok := c.Get("k"); !ok {
		t.Fatal("expected a hit before the new expiry")
	}
	clk.advance(time.Second + time.Nanosecond)
	if _, ok := c.Get("k"); ok {
		t.Fatal("expected a miss after the reset expiry")
	}
}

func TestCache_Expire_NeverExpire(t *testing.T) {
	clk := newFakeClock()
	c := New[string, int](0, nil, 0).WithNow(clk.Now)

	c.Set("k", 1, time.Second)
	c.Expire("k", -1) // negative → never expire
	clk.advance(time.Hour)
	if got, ok := c.Get("k"); !ok || got != 1 {
		t.Fatalf("got (%v,%v), want (1,true) — never-expire entry was evicted", got, ok)
	}
}

func TestCache_EvictOnCapacity(t *testing.T) {
	var evicted []string
	onEvict := func(key string, _ int) { evicted = append(evicted, key) }

	c := New[string, int](2, onEvict, 0)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Set("c", 3, 0) // evicts "a" (LRU)

	if !reflect.DeepEqual(evicted, []string{"a"}) {
		t.Fatalf("evicted = %v, want [a]", evicted)
	}
	if _, ok := c.Get("a"); ok {
		t.Fatal("expected 'a' to be evicted")
	}
	if c.Len() != 2 {
		t.Fatalf("len = %d, want 2", c.Len())
	}
}

func TestCache_Get_PromotesRecency(t *testing.T) {
	var evicted []string
	onEvict := func(key string, _ int) { evicted = append(evicted, key) }

	c := New[string, int](2, onEvict, 0)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Get("a")       // touch "a" → most-recently used
	c.Set("c", 3, 0) // evicts "b", not "a"

	if !reflect.DeepEqual(evicted, []string{"b"}) {
		t.Fatalf("evicted = %v, want [b]", evicted)
	}
}

func TestCache_Remove(t *testing.T) {
	var evicted []string
	onEvict := func(key string, _ int) { evicted = append(evicted, key) }

	c := New[string, int](0, onEvict, 0)
	c.Set("k", 1, 0)
	c.Remove("k")
	if _, ok := c.Get("k"); ok {
		t.Fatal("expected a miss after Remove")
	}
	if !reflect.DeepEqual(evicted, []string{"k"}) {
		t.Fatalf("expected onEvict for removed key, got %v", evicted)
	}
}

func TestCache_RemoveOldest(t *testing.T) {
	var evicted []string
	onEvict := func(key string, _ int) { evicted = append(evicted, key) }

	c := New[string, int](0, onEvict, 0)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.RemoveOldest()
	if !reflect.DeepEqual(evicted, []string{"a"}) {
		t.Fatalf("evicted = %v, want [a]", evicted)
	}
}

func TestCache_RemoveExpired(t *testing.T) {
	clk := newFakeClock()
	c := New[string, int](0, nil, 0).WithNow(clk.Now)

	c.Set("a", 1, time.Second)
	c.Set("b", 2, 0) // never expires
	clk.advance(2 * time.Second)

	c.RemoveExpired()
	if c.Len() != 1 {
		t.Fatalf("len = %d, want 1 after sweep", c.Len())
	}
	if _, ok := c.Get("b"); !ok {
		t.Fatal("expected 'b' to survive (no TTL)")
	}
}

func TestCache_Clear(t *testing.T) {
	var evicted []string
	onEvict := func(key string, _ int) { evicted = append(evicted, key) }

	c := New[string, int](0, onEvict, 0)
	for i := 0; i < 3; i++ {
		c.Set(fmt.Sprintf("k%d", i), i, 0)
	}
	c.Clear()
	if c.Len() != 0 {
		t.Fatalf("len = %d, want 0 after Clear", c.Len())
	}
	if len(evicted) != 3 {
		t.Fatalf("onEvict called %d times, want 3", len(evicted))
	}
}

func TestCache_DefaultExpiration(t *testing.T) {
	clk := newFakeClock()
	// A default TTL of 5s applies when Set is called with duration 0.
	c := New[string, int](0, nil, 5*time.Second).WithNow(clk.Now)

	c.Set("k", 1, 0) // 0 → use default 5s
	clk.advance(4 * time.Second)
	if _, ok := c.Get("k"); !ok {
		t.Fatal("expected a hit at 4s (within 5s default TTL)")
	}
	clk.advance(2 * time.Second)
	if _, ok := c.Get("k"); ok {
		t.Fatal("expected a miss at 6s (past 5s default TTL)")
	}
}

func TestCache_Concurrent(t *testing.T) {
	// Stress the mutex under -race: concurrent Set/Get/Remove across many
	// goroutines must not race or panic.
	c := New[int, int](100, nil, 0)
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func(off int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				k := off*1000 + i
				c.Set(k, i, 0)
				_, _ = c.Get(k)
				c.Remove(k)
			}
		}(g)
	}
	wg.Wait()
}
