package cache

import (
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

// fakeClock is a controllable time source so expiration tests never depend
// on real time.Sleep.
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
	return mgr, clk, func() { _ = mgr.Close() }
}

func TestLocal_Get_Set(t *testing.T) {
	mgr, _, cleanup := newTestLocal(t)
	defer cleanup()

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

func TestLocal_Expire(t *testing.T) {
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()

	if err := mgr.Set("k", "v", time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Second)

	if _, err := mgr.Get("k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after expiry = %v, want ErrNotFound", err)
	}
}

func TestLocal_NeverExpire(t *testing.T) {
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()

	if err := mgr.Set("k", "v", 0); err != nil { // 0 → never expires
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Hour)
	if got, err := mgr.Get("k"); err != nil || got != "v" {
		t.Fatalf("Get after long time = (%q,%v), want (v,nil)", got, err)
	}
}

func TestLocal_NegativeTTLNeverExpires(t *testing.T) {
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()

	if err := mgr.Set("k", "v", -1); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Hour)
	if got, err := mgr.Get("k"); err != nil || got != "v" {
		t.Fatalf("Get with negative TTL = (%q,%v), want (v,nil)", got, err)
	}
}

func TestLocal_Expire_Method(t *testing.T) {
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()

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

	// Expire on a missing key returns ErrNotFound.
	if err := mgr.Expire("nope", time.Second); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expire missing = %v, want ErrNotFound", err)
	}
}

func TestLocal_Expire_DoesNotResurrect(t *testing.T) {
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()

	if err := mgr.Set("k", "v", time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Second) // key is now expired
	// Expire must not bring an expired key back to life.
	if err := mgr.Expire("k", time.Hour); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expire on expired key = %v, want ErrNotFound (no resurrect)", err)
	}
	if _, err := mgr.Get("k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("key should still be gone after no-op Expire, got %v", err)
	}
}

func TestLocal_SetNx(t *testing.T) {
	mgr, _, cleanup := newTestLocal(t)
	defer cleanup()

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

func TestLocal_SetNx_WritesThroughExpired(t *testing.T) {
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()

	if err := mgr.Set("k", "old", time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(time.Second)

	existing, err := mgr.SetNx("k", "new", 0)
	if err != nil || existing {
		t.Fatalf("SetNx on expired key = (%v,%v), want (false,nil)", existing, err)
	}
	if got, _ := mgr.Get("k"); got != "new" {
		t.Fatalf("value = %q, want 'new' (expired key should be replaceable)", got)
	}
}

func TestLocal_SetNx_Atomic(t *testing.T) {
	// Concurrent SetNx on one absent key: exactly one winner.
	mgr, _, cleanup := newTestLocal(t)
	defer cleanup()

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

func TestLocal_Get_DoesNotMisdeleteNewValue(t *testing.T) {
	// Regression for the async-misdelete race: a Get that observes an
	// expired key must not later delete a value just written by a
	// concurrent Set to the same key. We cannot reliably interleave two
	// goroutines here, so we simulate the exact sequence the fix guards:
	// Get observes expiry, then Set writes a new value, then the lazy
	// delete must leave the new value intact.
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()
	c := mgr.(*localCache)

	// Plant an expired item directly so Get observes the stale pointer.
	c.lock.Lock()
	stale := &item{raw: []byte("old"), expireAt: clk.Now().Add(-time.Second)}
	c.m["k"] = stale
	c.lock.Unlock()

	// A concurrent Set replaces the value with a live one.
	if err := mgr.Set("k", "new", 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	// Now trigger the conditional delete path with the stale pointer; it must
	// detect the replacement (cur != stale) and leave the new value.
	c.deleteIfExpired("k", stale)
	if got, err := mgr.Get("k"); err != nil || got != "new" {
		t.Fatalf("new value was misdeleted: got (%q,%v), want (new,nil)", got, err)
	}
}

func TestLocal_Blob(t *testing.T) {
	mgr, _, cleanup := newTestLocal(t)
	defer cleanup()

	type user struct {
		Name string
		Age  int
	}

	var u user
	if err := mgr.GetBlob("u:1", &u); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetBlob missing = %v, want ErrNotFound", err)
	}
	if err := mgr.SetBlob("u:1", user{Name: "tom", Age: 18}, 0); err != nil {
		t.Fatalf("SetBlob: %v", err)
	}
	var got user
	if err := mgr.GetBlob("u:1", &got); err != nil {
		t.Fatalf("GetBlob: %v", err)
	}
	if got.Name != "tom" || got.Age != 18 {
		t.Fatalf("decoded = %+v, want {tom 18}", got)
	}
}

func TestLocal_Blob_Expiry(t *testing.T) {
	mgr, clk, cleanup := newTestLocal(t)
	defer cleanup()

	type payload struct{ V int }
	if err := mgr.SetBlob("p", payload{V: 1}, time.Second); err != nil {
		t.Fatalf("SetBlob: %v", err)
	}
	clk.advance(time.Second)
	var p payload
	if err := mgr.GetBlob("p", &p); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetBlob after expiry = %v, want ErrNotFound", err)
	}
}

func TestLocal_Blob_DecodesOutsideLock(t *testing.T) {
	// Prove decoding happens after RUnlock: a value whose msgpack
	// CustomDecoder reenters the cache (a write) must not deadlock. If decode
	// ran under the read lock, the nested write lock would deadlock.
	mgr, _, cleanup := newTestLocal(t)
	defer cleanup()
	blobReenterTarget = mgr
	t.Cleanup(func() { blobReenterTarget = nil })

	// SetBlob stores msgpack bytes; the decoder below consumes exactly that.
	if err := mgr.SetBlob("r", "OK", 0); err != nil {
		t.Fatalf("SetBlob: %v", err)
	}

	var dst reentrantString
	if err := mgr.GetBlob("r", &dst); err != nil {
		t.Fatalf("GetBlob: %v", err)
	}
	if string(dst) != "OK" {
		t.Fatalf("decoded = %q, want OK", string(dst))
	}
	// The reentrant write during decode must have landed — proving the read
	// lock was released before DecodeMsgpack ran.
	if got, err := mgr.Get("sentinel"); err != nil || got != "from-unmarshal" {
		t.Fatalf("reentrant write missing: got (%q,%v)", got, err)
	}
}

// blobReenterTarget is the cache a reentrantString's decoder writes into.
// Set by the test before GetBlob; it avoids a closure (DecodeMsgpack can't be
// one) while keeping the test self-contained.
var blobReenterTarget Manager

// reentrantString decodes from msgpack and, mid-decode, writes back into the
// same cache (which needs the write lock — impossible under a held read lock).
type reentrantString string

func (r *reentrantString) DecodeMsgpack(dec *msgpack.Decoder) error {
	var s string
	if err := dec.Decode(&s); err != nil {
		return err
	}
	*r = reentrantString(s)
	if blobReenterTarget != nil {
		return blobReenterTarget.Set("sentinel", "from-unmarshal", 0)
	}
	return nil
}

func TestLocal_Del(t *testing.T) {
	mgr, _, cleanup := newTestLocal(t)
	defer cleanup()

	if err := mgr.Set("k", "v", 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := mgr.Del("k"); err != nil {
		t.Fatalf("Del: %v", err)
	}
	if _, err := mgr.Get("k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after Del = %v, want ErrNotFound", err)
	}
	if err := mgr.Del("nope"); err != nil {
		t.Fatalf("Del missing = %v, want nil", err)
	}
}

func TestLocal_EvictBackground(t *testing.T) {
	// Short sweep interval; confirm the background loop removes an expired
	// key on its own, independent of any Get, and that Close waits for the
	// goroutine to exit.
	clk := newFakeClock()
	mgr := NewLocal(WithNow(clk.Now), WithEvictInterval(20*time.Millisecond))
	c := mgr.(*localCache)

	if err := mgr.Set("k", "v", time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	clk.advance(2 * time.Second)

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		c.lock.RLock()
		_, present := c.m["k"]
		c.lock.RUnlock()
		if !present {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	c.lock.RLock()
	_, present := c.m["k"]
	c.lock.RUnlock()
	if present {
		t.Fatal("background sweep did not remove the expired key in time")
	}
	// Close must return only after the sweep goroutine has exited.
	done := make(chan struct{})
	go func() { _ = mgr.Close(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Close did not return within 1s (sweep goroutine never exited)")
	}
}

func TestLocal_Close_Idempotent(t *testing.T) {
	mgr := NewLocal(WithEvictInterval(time.Hour))
	if err := mgr.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := mgr.Close(); err != nil { // must not panic on double-close
		t.Fatalf("second Close: %v", err)
	}
}

func TestLocal_Inactive(t *testing.T) {
	// A nil receiver cache reports ErrInactive for every op.
	var nilMgr Manager = (*localCache)(nil)
	if _, err := nilMgr.Get("k"); !errors.Is(err, ErrInactive) {
		t.Fatalf("nil Get = %v, want ErrInactive", err)
	}
	if err := nilMgr.Set("k", "v", 0); !errors.Is(err, ErrInactive) {
		t.Fatalf("nil Set = %v, want ErrInactive", err)
	}
}

func TestLocal_Concurrent(t *testing.T) {
	// Stress the read/write lock under -race.
	mgr, _, cleanup := newTestLocal(t)
	defer cleanup()

	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func(off int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				key := strconv.Itoa(off*1000 + i)
				_ = mgr.Set(key, "v", 0)
				_, _ = mgr.Get(key)
				_ = mgr.Del(key)
			}
		}(g)
	}
	wg.Wait()
}
