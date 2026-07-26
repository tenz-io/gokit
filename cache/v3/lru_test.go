package cache

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newTestLRU builds an lruManager (Manager) with a fake clock so expiration
// tests need no real sleeps. The sweep option is irrelevant for the LRU
// backend (it expires lazily) and is omitted.
func newTestLRU(t *testing.T, capability int) (Manager, *fakeClock) {
	t.Helper()
	clk := newFakeClock()
	mgr := NewLRU(capability, nil, WithNow(clk.Now))
	return mgr, clk
}

func TestLRU_Get_Set(t *testing.T) {
	mgr, _ := newTestLRU(t, 10)

	if _, err := mgr.Get("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get missing = %v, want ErrNotFound", err)
	}
	if err := mgr.Set("k", "v", 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := mgr.Get("k")
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
	mgr := NewLRU(2, func(key string, _ []byte) { evicted = append(evicted, key) }, WithNow(clk.Now))

	if err := mgr.Set("a", "1", 0); err != nil {
		t.Fatalf("Set a: %v", err)
	}
	if err := mgr.Set("b", "2", 0); err != nil {
		t.Fatalf("Set b: %v", err)
	}
	if err := mgr.Set("c", "3", 0); err != nil { // evicts "a"
		t.Fatalf("Set c: %v", err)
	}
	if len(evicted) != 1 || evicted[0] != "a" {
		t.Fatalf("evicted = %v, want [a]", evicted)
	}
	if _, err := mgr.Get("a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get a after eviction = %v, want ErrNotFound", err)
	}
}

func TestLRU_Expire(t *testing.T) {
	mgr, clk := newTestLRU(t, 0)

	if err := mgr.Set("k", "v", time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Second)
	if _, err := mgr.Get("k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after expiry = %v, want ErrNotFound", err)
	}
}

func TestLRU_NeverExpire(t *testing.T) {
	mgr, clk := newTestLRU(t, 10)

	if err := mgr.Set("k", "v", 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Hour)
	if got, err := mgr.Get("k"); err != nil || got != "v" {
		t.Fatalf("Get after long time = (%q,%v), want (v,nil)", got, err)
	}
}

func TestLRU_NegativeTTLNeverExpires(t *testing.T) {
	mgr, clk := newTestLRU(t, 10)

	if err := mgr.Set("k", "v", -1); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Hour)
	if got, err := mgr.Get("k"); err != nil || got != "v" {
		t.Fatalf("Get with negative TTL = (%q,%v), want (v,nil)", got, err)
	}
}

func TestLRU_SetNx(t *testing.T) {
	mgr, _ := newTestLRU(t, 10)

	if existing, err := mgr.SetNx("k", "v1", 0); err != nil || existing {
		t.Fatalf("first SetNx = (%v,%v), want (false,nil)", existing, err)
	}
	if existing, err := mgr.SetNx("k", "v2", 0); err != nil || !existing {
		t.Fatalf("second SetNx = (%v,%v), want (true,nil)", existing, err)
	}
	if got, _ := mgr.Get("k"); got != "v1" {
		t.Fatalf("value = %q, SetNx must not overwrite an existing key", got)
	}
}

func TestLRU_SetNx_Atomic(t *testing.T) {
	// Concurrent SetNx on one absent key: exactly one winner.
	mgr, _ := newTestLRU(t, 0)
	const n = 100
	var winners int64
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if existing, _ := mgr.SetNx("race", "v", 0); !existing {
				atomic.AddInt64(&winners, 1)
			}
		}()
	}
	wg.Wait()
	if winners != 1 {
		t.Fatalf("SetNx had %d winners, want exactly 1", winners)
	}
}

func TestLRU_SetNx_WritesThroughExpired(t *testing.T) {
	mgr, clk := newTestLRU(t, 0)

	if err := mgr.Set("k", "old", time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Second)
	if existing, err := mgr.SetNx("k", "new", 0); err != nil || existing {
		t.Fatalf("SetNx on expired key = (%v,%v), want (false,nil)", existing, err)
	}
	if got, _ := mgr.Get("k"); got != "new" {
		t.Fatalf("value = %q, want 'new' after replacing expired key", got)
	}
}

func TestLRU_Blob(t *testing.T) {
	mgr, _ := newTestLRU(t, 10)

	type item struct{ Name string }
	if err := mgr.SetBlob("i", item{Name: "tom"}, 0); err != nil {
		t.Fatalf("SetBlob: %v", err)
	}
	var got item
	if err := mgr.GetBlob("i", &got); err != nil {
		t.Fatalf("GetBlob: %v", err)
	}
	if got.Name != "tom" {
		t.Fatalf("decoded = %+v, want {tom}", got)
	}
}

func TestLRU_Del(t *testing.T) {
	mgr, _ := newTestLRU(t, 10)

	if err := mgr.Set("k", "v", 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := mgr.Del("k"); err != nil {
		t.Fatalf("Del: %v", err)
	}
	if _, err := mgr.Get("k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after Del = %v, want ErrNotFound", err)
	}
}

func TestLRU_Expire_Method(t *testing.T) {
	mgr, clk := newTestLRU(t, 10)

	if err := mgr.Set("k", "v", 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := mgr.Expire("k", time.Second); err != nil {
		t.Fatalf("Expire: %v", err)
	}
	clk.advance(time.Second)
	if _, err := mgr.Get("k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after reset expiry = %v, want ErrNotFound", err)
	}

	if err := mgr.Expire("nope", time.Second); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expire missing = %v, want ErrNotFound", err)
	}
}

func TestLRU_Expire_DoesNotResurrect(t *testing.T) {
	mgr, clk := newTestLRU(t, 10)

	if err := mgr.Set("k", "v", time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Second) // expired
	if err := mgr.Expire("k", time.Hour); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expire on expired key = %v, want ErrNotFound (no resurrect)", err)
	}
	if _, err := mgr.Get("k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("key should still be gone after no-op Expire, got %v", err)
	}
}

func TestLRU_Inactive(t *testing.T) {
	var nilMgr Manager = (*lruManager)(nil)
	if _, err := nilMgr.Get("k"); !errors.Is(err, ErrInactive) {
		t.Fatalf("nil Get = %v, want ErrInactive", err)
	}
}

// ----- underlying generic lruCache[K,V] (now in-package, unexported) -----

func newTestLRURaw(t *testing.T, capability int) (*lruCache[string, int], *fakeClock) {
	t.Helper()
	clk := newFakeClock()
	c := newLRU[string, int](capability, nil).withNow(clk.Now)
	return c, clk
}

func TestLRU_SetAndGet(t *testing.T) {
	c, _ := newTestLRURaw(t, 0)
	c.set("k", 1234, 0)
	got, ok := c.get("k")
	if !ok {
		t.Fatal("expected a hit")
	}
	if got != 1234 {
		t.Fatalf("got %v, want 1234", got)
	}
}

func TestLRU_Get_Miss(t *testing.T) {
	c, _ := newTestLRURaw(t, 10)
	if _, ok := c.get("missing"); ok {
		t.Fatal("expected a miss for a never-set key")
	}
}

func TestLRU_Set_UpdateInPlace(t *testing.T) {
	c, _ := newTestLRURaw(t, 0)
	c.set("k", 1, 0)
	c.set("k", 2, 0)
	if got, _ := c.get("k"); got != 2 {
		t.Fatalf("got %v, want updated value 2", got)
	}
	if c.len() != 1 {
		t.Fatalf("len = %d, want 1 (update must not grow the cache)", c.len())
	}
}

func TestLRU_RawExpire(t *testing.T) {
	c, clk := newTestLRURaw(t, 0)
	c.set("k", 1234, time.Second)
	if _, ok := c.get("k"); !ok {
		t.Fatal("expected a hit before expiry")
	}
	clk.advance(time.Second)
	if _, ok := c.get("k"); ok {
		t.Fatal("expected a miss after expiry")
	}
	if c.len() != 0 {
		t.Fatalf("len = %d, want 0 (expired entry should be removed)", c.len())
	}
}

func TestLRU_RawExpire_Method(t *testing.T) {
	c, clk := newTestLRURaw(t, 0)
	c.set("k", 1, 0)
	if !c.expire("k", time.Second) {
		t.Fatal("expire on a live key should return ok=true")
	}
	if _, ok := c.get("k"); !ok {
		t.Fatal("expected a hit before the new expiry")
	}
	clk.advance(time.Second + time.Nanosecond)
	if _, ok := c.get("k"); ok {
		t.Fatal("expected a miss after the reset expiry")
	}
}

func TestLRU_RawExpire_NeverExpire(t *testing.T) {
	c, clk := newTestLRURaw(t, 0)
	c.set("k", 1, time.Second)
	if !c.expire("k", 0) { // 0 → never expire
		t.Fatal("expire on a live key should return ok=true")
	}
	clk.advance(time.Hour)
	if got, ok := c.get("k"); !ok || got != 1 {
		t.Fatalf("got (%v,%v), want (1,true) — never-expire entry was evicted", got, ok)
	}
}

func TestLRU_RawExpire_DoesNotResurrect(t *testing.T) {
	c, clk := newTestLRURaw(t, 0)
	c.set("k", 1, time.Second)
	clk.advance(time.Second)
	if c.expire("k", time.Hour) {
		t.Fatal("expire on an expired key should return ok=false (no resurrect)")
	}
	if _, ok := c.get("k"); ok {
		t.Fatal("expected the key to remain gone after a no-op expire")
	}
}

func TestLRU_EvictOnCapacity_Raw(t *testing.T) {
	var evicted []string
	clk := newFakeClock()
	c := newLRU[string, int](2, func(key string, _ int) { evicted = append(evicted, key) }).withNow(clk.Now)
	c.set("a", 1, 0)
	c.set("b", 2, 0)
	c.set("c", 3, 0) // evicts "a" (LRU)

	if !reflect.DeepEqual(evicted, []string{"a"}) {
		t.Fatalf("evicted = %v, want [a]", evicted)
	}
	if _, ok := c.get("a"); ok {
		t.Fatal("expected 'a' to be evicted")
	}
	if c.len() != 2 {
		t.Fatalf("len = %d, want 2", c.len())
	}
}

func TestLRU_Get_PromotesRecency(t *testing.T) {
	var evicted []string
	clk := newFakeClock()
	c := newLRU[string, int](2, func(key string, _ int) { evicted = append(evicted, key) }).withNow(clk.Now)
	c.set("a", 1, 0)
	c.set("b", 2, 0)
	c.get("a")       // touch "a" → most-recently used
	c.set("c", 3, 0) // evicts "b", not "a"

	if !reflect.DeepEqual(evicted, []string{"b"}) {
		t.Fatalf("evicted = %v, want [b]", evicted)
	}
}

func TestLRU_Remove(t *testing.T) {
	var evicted []string
	clk := newFakeClock()
	c := newLRU[string, int](0, func(key string, _ int) { evicted = append(evicted, key) }).withNow(clk.Now)
	c.set("k", 1, 0)
	c.remove("k")
	if _, ok := c.get("k"); ok {
		t.Fatal("expected a miss after Remove")
	}
	if !reflect.DeepEqual(evicted, []string{"k"}) {
		t.Fatalf("expected onEvict for removed key, got %v", evicted)
	}
}

func TestLRU_RemoveOldest(t *testing.T) {
	var evicted []string
	clk := newFakeClock()
	c := newLRU[string, int](0, func(key string, _ int) { evicted = append(evicted, key) }).withNow(clk.Now)
	c.set("a", 1, 0)
	c.set("b", 2, 0)
	c.removeOldest()
	if !reflect.DeepEqual(evicted, []string{"a"}) {
		t.Fatalf("evicted = %v, want [a]", evicted)
	}
}

func TestLRU_RemoveExpired(t *testing.T) {
	c, clk := newTestLRURaw(t, 0)
	c.set("a", 1, time.Second)
	c.set("b", 2, 0) // never expires
	clk.advance(2 * time.Second)

	c.removeExpired()
	if c.len() != 1 {
		t.Fatalf("len = %d, want 1 after sweep", c.len())
	}
	if _, ok := c.get("b"); !ok {
		t.Fatal("expected 'b' to survive (no TTL)")
	}
}

func TestLRU_Clear(t *testing.T) {
	var evicted []string
	clk := newFakeClock()
	c := newLRU[string, int](0, func(key string, _ int) { evicted = append(evicted, key) }).withNow(clk.Now)
	for i := 0; i < 3; i++ {
		c.set(fmt.Sprintf("k%d", i), i, 0)
	}
	c.clear()
	if c.len() != 0 {
		t.Fatalf("len = %d, want 0 after Clear", c.len())
	}
	if len(evicted) != 3 {
		t.Fatalf("onEvict called %d times, want 3", len(evicted))
	}
}

func TestLRU_SetNx_Atomic_Raw(t *testing.T) {
	// Two concurrent setNx on the same absent key: exactly one must see
	// existing=false (the winner).
	const n = 100
	var winners int64
	var wg sync.WaitGroup
	c, _ := newTestLRURaw(t, 0)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !c.setNx("race", 1, 0) {
				atomic.AddInt64(&winners, 1)
			}
		}()
	}
	wg.Wait()
	if winners != 1 {
		t.Fatalf("setNx had %d winners, want exactly 1 (atomic check-and-set)", winners)
	}
	if got, _ := c.get("race"); got != 1 {
		t.Fatalf("winning value = %v, want 1", got)
	}
}

func TestLRU_SetNx_WritesThroughExpired_Raw(t *testing.T) {
	c, clk := newTestLRURaw(t, 0)
	c.set("k", 1, time.Second)
	clk.advance(time.Second)
	if c.setNx("k", 2, 0) {
		t.Fatal("setNx on an expired key should write through (existing=false)")
	}
	if got, _ := c.get("k"); got != 2 {
		t.Fatalf("value = %v, want 2 after replacing expired key", got)
	}
}

func TestLRU_SetNx_CapacityEvictionFiresCallback(t *testing.T) {
	// A setNx that inserts and triggers capacity eviction must still fire
	// onEvict (outside the lock) — this guards the insert path of setNx.
	var evicted []string
	clk := newFakeClock()
	c := newLRU[string, int](1, func(key string, _ int) { evicted = append(evicted, key) }).withNow(clk.Now)
	c.set("a", 1, 0)
	if c.setNx("b", 2, 0) {
		t.Fatal("setNx on absent key should report existing=false")
	}
	if !reflect.DeepEqual(evicted, []string{"a"}) {
		t.Fatalf("evicted = %v, want [a] from setNx capacity eviction", evicted)
	}
}

func TestLRU_OnEvicted_CanReenterCache(t *testing.T) {
	// A reentrant callback (one that calls back into the same cache) must
	// not deadlock — callbacks run outside the lock. The callback performs a
	// reentrant read (get) under the callback; if the callback ran under the
	// lock, the nested get would deadlock. We deliberately do NOT write into
	// the same bounded cache from the callback (an insert under capacity
	// pressure would itself evict and recurse the callback).
	var evicted atomic.Int32
	var c *lruCache[string, int]
	c = newLRU[string, int](2, func(key string, val int) {
		// Reentrant read from inside the eviction callback:
		_, _ = c.get(key)
		evicted.Add(1)
	}).withNow(newFakeClock().Now)
	c.set("a", 1, 0)
	c.set("b", 2, 0)
	c.set("c", 3, 0) // evicts "a"; callback reenters (get)
	if evicted.Load() < 1 {
		t.Fatal("eviction callback never ran")
	}
}

func TestLRU_OnEvicted_CanReenterWrite(t *testing.T) {
	// Separately prove a reentrant *write* also survives: use an unbounded
	// cache so the callback's set can never trigger a cascading eviction.
	var wrote atomic.Bool
	var c *lruCache[string, int]
	clk := newFakeClock()
	c = newLRU[string, int](0, func(key string, val int) {
		// No capacity → this set never evicts → no recursion.
		c.set("seen:"+key, val, 0)
		wrote.Store(true)
	}).withNow(clk.Now)
	// Expire-driven eviction (the only eviction path when unbounded) fires
	// on get of an expired entry.
	c.set("a", 1, time.Millisecond)
	clk.advance(2 * time.Millisecond)
	_, _ = c.get("a") // triggers eviction of expired "a" → callback
	if !wrote.Load() {
		t.Fatal("eviction callback never ran")
	}
	if _, ok := c.get("seen:a"); !ok {
		t.Fatal("reentrant write inside the callback did not land (callback ran under the lock?)")
	}
}

func TestLRU_Concurrent(t *testing.T) {
	// Stress the mutex under -race: concurrent set/setNx/get/remove across
	// many goroutines must not race or panic. Use an int-keyed LRU so keys
	// can be computed arithmetically.
	c := newLRU[int, int](100, nil)
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func(off int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				k := off*1000 + i
				c.set(k, i, 0)
				_ = c.setNx(k, i, 0)
				_, _ = c.get(k)
				c.remove(k)
			}
		}(g)
	}
	wg.Wait()
}
