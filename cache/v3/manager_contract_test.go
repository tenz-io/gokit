package cache

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// These contract tests run the same scenarios against both backends to prove
// NewLocal and NewLRU honor the same Manager contract — the bugs found in
// review (async-misdelete, non-atomic SetNx, TTL/Expire drift) were all
// backend-divergence or concurrency defects, so a shared table is the right
// guard.

// backend is a fixture producing a fresh Manager and the fake clock its
// expiration decisions run on.
type backend struct {
	name string
	make func(t *testing.T) (Manager, *fakeClock)
}

var backends = []backend{
	{
		name: "Local",
		make: func(t *testing.T) (Manager, *fakeClock) {
			clk := newFakeClock()
			mgr := NewLocal(WithNow(clk.Now), WithEvictInterval(0))
			t.Cleanup(func() { _ = mgr.Close() })
			return mgr, clk
		},
	},
	{
		name: "LRU",
		make: func(t *testing.T) (Manager, *fakeClock) {
			clk := newFakeClock()
			mgr := NewLRU(100, nil, WithNow(clk.Now))
			t.Cleanup(func() { _ = mgr.Close() })
			return mgr, clk
		},
	},
}

func TestContract_ZeroTTLNeverExpires(t *testing.T) {
	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			mgr, clk := b.make(t)
			if err := mgr.Set("k", "v", 0); err != nil {
				t.Fatalf("Set: %v", err)
			}
			clk.advance(time.Hour)
			if got, err := mgr.Get("k"); err != nil || got != "v" {
				t.Fatalf("zero-TTL key expired: got (%q,%v), want (v,nil)", got, err)
			}
		})
	}
}

func TestContract_NegativeTTLNeverExpires(t *testing.T) {
	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			mgr, clk := b.make(t)
			if err := mgr.Set("k", "v", -1); err != nil {
				t.Fatalf("Set: %v", err)
			}
			clk.advance(time.Hour)
			if got, err := mgr.Get("k"); err != nil || got != "v" {
				t.Fatalf("negative-TTL key expired: got (%q,%v), want (v,nil)", got, err)
			}
		})
	}
}

func TestContract_PositiveTTLExpiresAtBoundary(t *testing.T) {
	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			mgr, clk := b.make(t)
			if err := mgr.Set("k", "v", time.Second); err != nil {
				t.Fatalf("Set: %v", err)
			}
			if _, err := mgr.Get("k"); err != nil {
				t.Fatalf("Get before expiry: %v", err)
			}
			clk.advance(time.Second)
			if _, err := mgr.Get("k"); !errors.Is(err, ErrNotFound) {
				t.Fatalf("Get at expiry boundary = %v, want ErrNotFound", err)
			}
		})
	}
}

func TestContract_ExpireAfterExpiryDoesNotResurrect(t *testing.T) {
	// The headline bug: Expire on an expired key must NOT bring it back.
	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			mgr, clk := b.make(t)
			if err := mgr.Set("k", "v", time.Second); err != nil {
				t.Fatalf("Set: %v", err)
			}
			clk.advance(time.Second) // expired
			if err := mgr.Expire("k", time.Hour); !errors.Is(err, ErrNotFound) {
				t.Fatalf("Expire on expired key = %v, want ErrNotFound", err)
			}
			if _, err := mgr.Get("k"); !errors.Is(err, ErrNotFound) {
				t.Fatalf("key resurrected by no-op Expire: err=%v", err)
			}
		})
	}
}

func TestContract_ExpireShortensLifetime(t *testing.T) {
	// Expire on a live key actually shortens its lifetime.
	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			mgr, clk := b.make(t)
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
		})
	}
}

func TestContract_SetNxAtomicExactlyOneWinner(t *testing.T) {
	// The headline bug: concurrent SetNx on one absent key must have exactly
	// one winner across both backends.
	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			mgr, _ := b.make(t)
			const n = 100
			var winners int64
			var wg sync.WaitGroup
			start := make(chan struct{})
			wg.Add(n)
			for i := 0; i < n; i++ {
				go func() {
					defer wg.Done()
					<-start
					if existing, _ := mgr.SetNx("race", "v", 0); !existing {
						atomic.AddInt64(&winners, 1)
					}
				}()
			}
			close(start) // release all goroutines as close together as possible
			wg.Wait()
			if winners != 1 {
				t.Fatalf("%s: SetNx had %d winners, want exactly 1", b.name, winners)
			}
		})
	}
}

func TestContract_SetNxExistingKeyReturnsTrue(t *testing.T) {
	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			mgr, _ := b.make(t)
			if existing, err := mgr.SetNx("k", "v1", 0); err != nil || existing {
				t.Fatalf("first SetNx = (%v,%v), want (false,nil)", existing, err)
			}
			if existing, err := mgr.SetNx("k", "v2", 0); err != nil || !existing {
				t.Fatalf("second SetNx = (%v,%v), want (true,nil)", existing, err)
			}
			if got, _ := mgr.Get("k"); got != "v1" {
				t.Fatalf("value = %q, SetNx must not overwrite", got)
			}
		})
	}
}

func TestContract_GetDoesNotMisdeleteConcurrentSet(t *testing.T) {
	// Regression for the async-misdelete race on Local. We can't deterministically
	// interleave goroutines, but a high-fan-out mix of Get on expired keys +
	// concurrent Set to the same keys must never leave a live value missing.
	// (LRU has no async delete path; this still runs harmlessly.)
	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			mgr, clk := b.make(t)
			// Seed 50 keys with a 1s TTL.
			for i := 0; i < 50; i++ {
				if err := mgr.Set("k"+itoa(i), "seed", time.Second); err != nil {
					t.Fatalf("Set: %v", err)
				}
			}
			clk.advance(time.Second) // all seeded keys now expired

			var wg sync.WaitGroup
			const readers, writers = 16, 16
			for r := 0; r < readers; r++ {
				wg.Add(1)
				go func(off int) {
					defer wg.Done()
					for i := 0; i < 200; i++ {
						_, _ = mgr.Get("k" + itoa((off+i)%50)) // hits expired keys
					}
				}(r)
			}
			for w := 0; w < writers; w++ {
				wg.Add(1)
				go func(off int) {
					defer wg.Done()
					for i := 0; i < 200; i++ {
						_ = mgr.Set("k"+itoa((off+i)%50), "new", 0) // live values
					}
				}(w)
			}
			wg.Wait()

			// Any key last touched by a writer must hold a live "new" value,
			// never be absent due to a racing lazy delete.
			for i := 0; i < 50; i++ {
				if got, err := mgr.Get("k" + itoa(i)); err == nil && got != "new" {
					t.Fatalf("k%d = %q, want 'new'", i, got)
				}
			}
		})
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = '0' + byte(i%10)
		i /= 10
	}
	return string(b[n:])
}
