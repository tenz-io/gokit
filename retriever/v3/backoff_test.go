package retriever

import (
	"math"
	"testing"
	"time"
)

func TestConstant(t *testing.T) {
	c := Constant(50 * time.Millisecond)
	for _, attempt := range []int{0, 1, 5} {
		if got := c.Next(attempt); got != 50*time.Millisecond {
			t.Errorf("Constant.Next(%d) = %v, want 50ms", attempt, got)
		}
	}
}

func TestConstant_Zero(t *testing.T) {
	if got := Constant(0).Next(0); got != 0 {
		t.Errorf("Constant(0).Next(0) = %v, want 0", got)
	}
}

func TestLinear(t *testing.T) {
	l := Linear{Base: 100 * time.Millisecond, Step: 50 * time.Millisecond}
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond}, // 100 + 50*0
		{1, 150 * time.Millisecond}, // 100 + 50*1
		{3, 250 * time.Millisecond}, // 100 + 50*3
	}
	for _, c := range cases {
		if got := l.Next(c.attempt); got != c.want {
			t.Errorf("Linear.Next(%d) = %v, want %v", c.attempt, got, c.want)
		}
	}
}

func TestLinear_ZeroStep(t *testing.T) {
	l := Linear{Base: 100 * time.Millisecond}
	if got := l.Next(5); got != 100*time.Millisecond {
		t.Errorf("Linear{Base:100ms}.Next(5) = %v, want 100ms", got)
	}
}

func TestLinear_NegativeAttempt(t *testing.T) {
	l := Linear{Base: 100 * time.Millisecond, Step: 50 * time.Millisecond}
	if got := l.Next(-1); got != 100*time.Millisecond {
		t.Errorf("Linear.Next(-1) = %v, want 100ms (clamped)", got)
	}
}

func TestExponential(t *testing.T) {
	e := Exponential{Base: 100 * time.Millisecond, Factor: 2}
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond}, // 100 * 2^0
		{1, 200 * time.Millisecond}, // 100 * 2^1
		{2, 400 * time.Millisecond}, // 100 * 2^2
		{3, 800 * time.Millisecond}, // 100 * 2^3
	}
	for _, c := range cases {
		if got := e.Next(c.attempt); got != c.want {
			t.Errorf("Exponential.Next(%d) = %v, want %v", c.attempt, got, c.want)
		}
	}
}

func TestExponential_Zero(t *testing.T) {
	e := Exponential{Base: 0, Factor: 2}
	if got := e.Next(3); got != 0 {
		t.Errorf("Exponential{Base:0}.Next(3) = %v, want 0", got)
	}
	e = Exponential{Base: 100 * time.Millisecond, Factor: 0}
	if got := e.Next(3); got != 0 {
		t.Errorf("Exponential{Factor:0}.Next(3) = %v, want 0", got)
	}
}

func TestJitter_NoFactor(t *testing.T) {
	j := Jitter{Backoff: Constant(100 * time.Millisecond), Factor: 0}
	if got := j.Next(0); got != 100*time.Millisecond {
		t.Errorf("Jitter{Factor:0}.Next(0) = %v, want 100ms", got)
	}
}

func TestJitter_NilBackoff(t *testing.T) {
	j := Jitter{Backoff: nil, Factor: 0.3}
	if got := j.Next(0); got != 0 {
		t.Errorf("Jitter{Backoff:nil}.Next(0) = %v, want 0", got)
	}
}

func TestJitter_Range(t *testing.T) {
	// wait in [100ms, 130ms) since Factor=0.3 adds [0, 30ms)
	j := Jitter{Backoff: Constant(100 * time.Millisecond), Factor: 0.3}
	for i := 0; i < 50; i++ {
		got := j.Next(0)
		if got < 100*time.Millisecond || got >= 130*time.Millisecond {
			t.Errorf("Jitter.Next(0) = %v, want in [100ms, 130ms)", got)
		}
	}
}

func TestJitter_WrapsExponential(t *testing.T) {
	j := Jitter{
		Backoff: Exponential{Base: 100 * time.Millisecond, Factor: 2},
		Factor:  0.5,
	}
	// attempt 1: base wait 200ms, jitter [0, 100ms) => [200ms, 300ms)
	for i := 0; i < 50; i++ {
		got := j.Next(1)
		if got < 200*time.Millisecond || got >= 300*time.Millisecond {
			t.Errorf("Jitter(Exp).Next(1) = %v, want in [200ms, 300ms)", got)
		}
	}
}

func TestExponential_FractionalFactor(t *testing.T) {
	// 非整数 Factor(如 1.5):100 * 1.5^2 = 225ms。
	e := Exponential{Base: 100 * time.Millisecond, Factor: 1.5}
	if got := e.Next(2); got != 225*time.Millisecond {
		t.Errorf("Exponential{Factor:1.5}.Next(2) = %v, want 225ms", got)
	}
}

func TestExponential_NegativeAttempt(t *testing.T) {
	// 负 attempt 应被钳制为 0,等价于首次重试前的等待。
	e := Exponential{Base: 100 * time.Millisecond, Factor: 2}
	if got := e.Next(-3); got != 100*time.Millisecond {
		t.Errorf("Exponential.Next(-3) = %v, want 100ms (clamped to 0)", got)
	}
}

func TestLinear_NegativeBase(t *testing.T) {
	// Base<=0:文档未承诺,但应返回 Step*attempt(不 panic)。
	l := Linear{Base: 0, Step: 50 * time.Millisecond}
	if got := l.Next(2); got != 100*time.Millisecond {
		t.Errorf("Linear{Base:0}.Next(2) = %v, want 100ms", got)
	}
}

func TestJitter_WrapsLinear(t *testing.T) {
	// Jitter 装饰 Linear:attempt 2 => base 100+50*2=200ms,jitter [0,60ms) => [200,260)ms。
	j := Jitter{
		Backoff: Linear{Base: 100 * time.Millisecond, Step: 50 * time.Millisecond},
		Factor:  0.3,
	}
	for i := 0; i < 50; i++ {
		got := j.Next(2)
		if got < 200*time.Millisecond || got >= 260*time.Millisecond {
			t.Errorf("Jitter(Linear).Next(2) = %v, want in [200ms, 260ms)", got)
		}
	}
}

func TestJitter_WrapsConstantZeroWait(t *testing.T) {
	// 被包装退避返回 0 时,jitter 也应返回 0(不放大)。
	j := Jitter{Backoff: Constant(0), Factor: 0.5}
	if got := j.Next(0); got != 0 {
		t.Errorf("Jitter{wait:0}.Next(0) = %v, want 0", got)
	}
}

func TestJitter_HalfOpenRange(t *testing.T) {
	// 验证抖动为半开区间 [wait, wait+factor*wait):理论上界不可达。
	// Factor=1.0 => [100ms, 200ms)。统计上界:rand.Float64()<1 恒成立,
	// 故 jitter < 100ms,即 got < 200ms。这里做大量抽样,确认无一越界。
	j := Jitter{Backoff: Constant(100 * time.Millisecond), Factor: 1.0}
	for i := 0; i < 1000; i++ {
		got := j.Next(0)
		if got < 100*time.Millisecond || got >= 200*time.Millisecond {
			t.Fatalf("Jitter{Factor:1}.Next(0) = %v, want in [100ms, 200ms)", got)
		}
	}
}

func TestBackoffStrategies_Table(t *testing.T) {
	// 一张表覆盖三种无抖动策略在 attempt=0 的精确预期(无抖动 => 确定性)。
	cases := []struct {
		name    string
		backoff Backoff
		want0   time.Duration
	}{
		{"Constant", Constant(50 * time.Millisecond), 50 * time.Millisecond},
		{"Linear", Linear{Base: 10 * time.Millisecond, Step: 10 * time.Millisecond}, 10 * time.Millisecond},
		{"Exponential", Exponential{Base: 10 * time.Millisecond, Factor: 2}, 10 * time.Millisecond},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.backoff.Next(0)
			if got != c.want0 {
				t.Errorf("Next(0) = %v, want %v", got, c.want0)
			}
		})
	}
}

// --- 溢出 / 饱和运算回归 --------------------------------------------------

func TestExponential_OverflowClamped(t *testing.T) {
	// 极大 attempt 不应导致无界增长或溢出回绕为负:结果被钳制到 [maxDuration]。
	// 这回归了两个早期版本 bug:(a) 把 ms 钳到 MaxInt64 再乘 time.Millisecond 回绕为负;
	// (b) 钳到 maxDuration 后 Jitter 再做 wait+jitter 仍回绕为负。
	e := Exponential{Base: 100 * time.Millisecond, Factor: 2}
	got := e.Next(1000)
	if got != maxDuration {
		t.Fatalf("Exponential.Next(1000) = %v, want maxDuration (%v)", got, maxDuration)
	}
	if got < 0 {
		t.Fatalf("Exponential.Next(1000) = %v, must not be negative", got)
	}
	// 钳制后稳定(确定性)。
	if got2 := e.Next(1000); got2 != got {
		t.Errorf("Exponential.Next(1000) not stable: %v then %v", got, got2)
	}
}

func TestExponential_NaNsAndInfFactor(t *testing.T) {
	// NaN / +Inf / -Inf 的 Factor 视为非法:返回 0(立即重试),不产生 NaN Duration。
	for _, f := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		e := Exponential{Base: 100 * time.Millisecond, Factor: f}
		if got := e.Next(3); got != 0 {
			t.Errorf("Factor=%v: Next(3) = %v, want 0", f, got)
		}
	}
}

func TestLinear_OverflowClamped(t *testing.T) {
	// 巨大 attempt 与 Step 的乘积溢出时应饱和为 [maxDuration],不回绕为负。
	l := Linear{Base: time.Millisecond, Step: time.Hour}
	got := l.Next(1 << 30) // 1<<30 * 1h 远超 maxDuration
	if got != maxDuration {
		t.Fatalf("Linear.Next(big) = %v, want maxDuration (%v)", got, maxDuration)
	}
	if got < 0 {
		t.Fatalf("Linear.Next(big) = %v, must not be negative", got)
	}
}

func TestLinear_SaturatingAdd(t *testing.T) {
	// Base 接近上限时 + 增长项应饱和,不回绕。
	l := Linear{Base: maxDuration - time.Hour, Step: time.Hour}
	got := l.Next(2) // (maxDuration-1h) + 2h => 应饱和到 maxDuration
	if got != maxDuration {
		t.Errorf("Linear.Next(2) = %v, want maxDuration (saturated)", got)
	}
}

func TestJitter_SaturatesOnMaxDurationWait(t *testing.T) {
	// 被包装退避返回 maxDuration 时,叠加抖动应饱和到 maxDuration,
	// 绝不因 wait+jitter 回绕为负导致 time.NewTimer 立即触发(密集重试)。
	// 回归 v3 早期 "wait+jitter 回绕为 -1793433h" 的 bug。
	j := Jitter{
		Backoff: Exponential{Base: 100 * time.Millisecond, Factor: 2}, // Next(1000)=maxDuration
		Factor:  0.3,
	}
	for i := 0; i < 200; i++ {
		got := j.Next(1000)
		if got < 0 {
			t.Fatalf("Jitter.Next(1000) = %v, must not be negative (saturated)", got)
		}
		if got != maxDuration {
			// 严格来说,wait=maxDuration 时任一 jitter>0 都应饱和到 maxDuration。
			t.Fatalf("Jitter.Next(1000) = %v, want maxDuration (saturated add)", got)
		}
	}
}

func TestJitter_NonMaxWaitStillJitters(t *testing.T) {
	// 非极端情况下抖动仍正常叠加(确保饱和逻辑没把正常路径也压平)。
	j := Jitter{
		Backoff: Exponential{Base: 100 * time.Millisecond, Factor: 2}, // Next(2)=400ms
		Factor:  0.5,
	}
	sawNonBase := false
	for i := 0; i < 100; i++ {
		got := j.Next(2)
		if got < 400*time.Millisecond || got >= 600*time.Millisecond {
			t.Fatalf("Jitter.Next(2) = %v, want in [400ms, 600ms)", got)
		}
		if got > 400*time.Millisecond {
			sawNonBase = true
		}
	}
	if !sawNonBase {
		t.Error("expected jitter to produce values > 400ms at least once")
	}
}

func TestExponential_FiniteOverflowClamped(t *testing.T) {
	// 产出一个"有限但 >= maxDuration"的 ns,命中 Exponential.Next 的
	// `case ns >= float64(maxDuration)` 分支(而非 +Inf 走的 IsInf 分支)。
	// Base=1ns, Factor=2, attempt=63 => ns = 2^63 = 9.22e18 >= float64(maxDuration), 有限。
	e := Exponential{Base: 1, Factor: 2}
	got := e.Next(63)
	if got != maxDuration {
		t.Errorf("Exponential{Base:1ns}.Next(63) = %v, want maxDuration (finite overflow clamp)", got)
	}
}

func TestAddSat_NegativeInputs(t *testing.T) {
	// addSat 的防御性分支:负输入被视为非法,取较大者返回(避免负 Duration 污染)。
	// Linear/Jitter 正常路径不会产生负输入,这里直接单测该防御逻辑。
	if got := addSat(-time.Second, time.Millisecond); got != time.Millisecond {
		t.Errorf("addSat(-1s, 1ms) = %v, want 1ms (keep larger)", got)
	}
	if got := addSat(time.Millisecond, -time.Second); got != time.Millisecond {
		t.Errorf("addSat(1ms, -1s) = %v, want 1ms (keep larger)", got)
	}
}
