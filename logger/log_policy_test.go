package logger

import (
	"github.com/smarty/assertions"
	"math"
	"testing"
	"time"
)

func TestRateLimitPolicy_Allow(t *testing.T) {
	t.Run("test rate limit policy", func(t *testing.T) {
		rp := NewRateLimitPolicy(2, 1)
		assertions.ShouldBeTrue(rp.Allow())
		assertions.ShouldBeTrue(rp.Allow())
		assertions.ShouldBeFalse(rp.Allow())
		assertions.ShouldBeFalse(rp.Allow())

		time.Sleep(1 * time.Second)
		assertions.ShouldBeTrue(rp.Allow())
	})
}

func TestSamplingPolicy_Allow(t *testing.T) {
	t.Run("test sampling policy", func(t *testing.T) {
		ratio := 0.1
		sp := NewSamplingPolicy(ratio)
		total := 1000000
		allowCount := 0
		for i := 0; i < total; i++ {
			if sp.Allow() {
				allowCount++
			}
		}

		gotRatio := float64(allowCount) / float64(total)
		t.Logf("total: %d, allow: %d, ratio: %f", total, allowCount, gotRatio)
		diff := math.Abs(gotRatio - ratio)
		t.Logf("diff: %f", diff)
		assertions.ShouldBeTrue(diff < 0.01)
	})
}
