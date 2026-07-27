package retriever

import (
	"math"
	"math/rand"
	"time"
)

// Backoff 计算两次重试之间的等待时长。attempt 从首次重试前的第一次失败
// 开始按 0,1,2,... 编号。实现无需保持状态:同一 attempt 输入应得到可比的
// 等待时长(允许抖动产生随机性)。
//
// 本包的所有内置策略均为可比较字段、可直接用复合字面量构造的值类型,
// 并通过 [Jitter] 装饰器组合,例如:
//
//	retriever.Jitter{
//	    Backoff: retriever.Exponential{Base: 100 * time.Millisecond, Factor: 2},
//	    Factor:  0.3,
//	}
type Backoff interface {
	Next(attempt int) time.Duration
}

// Constant 在每次重试前等待相同时长。零值等待 0(等价于立即重试)。
//
//	retriever.Constant(50 * time.Millisecond)
type Constant time.Duration

// Next 实现 [Backoff]。
func (c Constant) Next(_ int) time.Duration { return time.Duration(c) }

// Linear 按线性增长等待:Base + Step*attempt。
// 零值 Step 退化为固定退避,零值 Base 退化为立即重试。
//
//	retriever.Linear{Base: 100 * time.Millisecond, Step: 50 * time.Millisecond}
type Linear struct {
	Base time.Duration
	Step time.Duration
}

// maxDuration 是 [time.Duration] 能表示的最大值(以纳秒计),
// 用于在指数/线性增长溢出时钳制结果,避免回绕为负数。
const maxDuration = time.Duration(math.MaxInt64)

// mulStep 返回 attempt*step 的饱和乘法:乘积溢出或非正时钳制到 [maxDuration]。
// 任何一项为 0 时返回 0;attempt 为负时按 0 处理。
func mulStep(attempt int, step time.Duration) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if attempt == 0 || step <= 0 {
		return 0
	}
	// int64 乘法在溢出时会回绕为任意符号的值,无法事后判别。
	// 改用 float64 预估,超过 maxDuration 直接饱和,避免回绕。
	p := float64(attempt) * float64(step)
	if p >= float64(maxDuration) {
		return maxDuration
	}
	return time.Duration(p)
}

// addSat 返回 a+b 的饱和加法:和溢出(回绕为非正)时钳制到 [maxDuration]。
// 两项均非负时,a+b 若为负即为溢出。
func addSat(a, b time.Duration) time.Duration {
	if a < 0 || b < 0 {
		// 负输入视为非法,直接取较大者(避免负 Duration 污染结果)。
		if a > b {
			return a
		}
		return b
	}
	s := a + b
	if s < a || s < b { // 回绕
		return maxDuration
	}
	return s
}

// Next 实现 [Backoff]:返回 Base + Step*attempt,乘法与加法均饱和,
// 溢出时钳制到 [maxDuration] 而非回绕为负。
func (l Linear) Next(attempt int) time.Duration {
	return addSat(l.Base, mulStep(attempt, l.Step))
}

// Exponential 按指数增长等待:Base * Factor^attempt。
// Base 是首次重试前的等待时长,Factor 是每次失败后的倍数。
// Base<=0、Factor<=0、Factor 为 NaN/Inf 时按立即重试处理。为避免无界增长
// 导致 [time.Duration] 溢出为负,结果会被钳制到 [maxDuration] 纳秒
// (约 292 年,即 [time.Duration] 的上界)以内。
//
//	retriever.Exponential{Base: 100 * time.Millisecond, Factor: 2}
type Exponential struct {
	Base   time.Duration
	Factor float64
}

// Next 实现 [Backoff]:返回 Base * Factor^attempt,溢出时钳制到 [maxDuration]。
// Factor 为 NaN/Inf 或 Base/Factor 非正时返回 0(立即重试)。
func (e Exponential) Next(attempt int) time.Duration {
	if e.Base <= 0 || math.IsNaN(e.Factor) || math.IsInf(e.Factor, 0) || e.Factor <= 0 {
		return 0
	}
	if attempt < 0 {
		attempt = 0
	}
	ns := float64(e.Base) * math.Pow(e.Factor, float64(attempt))
	switch {
	case math.IsNaN(ns) || math.IsInf(ns, 0):
		return maxDuration
	case ns >= float64(maxDuration):
		return maxDuration
	default:
		return time.Duration(ns)
	}
}

// Jitter 在被包装的 [Backoff] 基础上叠加随机抖动,等待时长变为
// [wait, wait+Factor*wait)。Factor<=0 时退化为不抖动。
// 抖动可避免大量客户端在退避后同步重试造成的"重试风暴"。
// 即便被包装退避返回 [maxDuration],叠加抖动也会饱和到 [maxDuration],
// 不会因 wait+jitter 回绕为负数而让 [time.NewTimer] 立即触发。
//
//	retriever.Jitter{Backoff: retriever.Exponential{Base: 100 * time.Millisecond, Factor: 2}, Factor: 0.3}
type Jitter struct {
	Backoff Backoff
	Factor  float64
}

// Next 实现 [Backoff]:返回 b.Backoff.Next(attempt) 加上
// [0, Factor*wait) 区间内的随机抖动,饱和加法保证不回绕为负。
func (j Jitter) Next(attempt int) time.Duration {
	if j.Backoff == nil {
		return 0
	}
	wait := j.Backoff.Next(attempt)
	if j.Factor <= 0 || wait <= 0 {
		return wait
	}
	jitter := time.Duration(rand.Float64() * j.Factor * float64(wait))
	return addSat(wait, jitter)
}
