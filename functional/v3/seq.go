package fn

// Seq 是遍历 T 序列的 lazy、callback 风格迭代器。它是无 pull 的 "push" 模型:
// 消费者传入一个 yield callback,生产者把每个元素 push 给它。从 yield
// 返回 false 可提前停止迭代(short-circuit)。
//
// Seq 是 Chain 的 zero-allocation 互补:Chain 在每步物化出一个 slice,
// 而 Seq 将操作融合为单个 callback chain,并可在得到结果时立即停止
// (Any/All/Find)。对于构建结果 slice,Chain/Collect 通常更清晰。
//
// 用 SeqOf(从 slice)构造。Filter 和 Map 以 lazy 方式组合。
type Seq[T any] func(yield func(T) bool)

// SeqOf 创建一个按顺序 yield s 元素的 Seq。slice 是被引用(非复制);
// 迭代期间不要修改 s。
func SeqOf[T any](s []T) Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// Filter 返回一个只 yield q 中 pred 成立元素的 Seq。它是 lazy 的:
// pred 仅对实际被消费的元素求值。
func (q Seq[T]) Filter(pred func(T) bool) Seq[T] {
	return func(yield func(T) bool) {
		q(func(v T) bool {
			if !pred(v) {
				return true // 跳过,继续迭代
			}
			return yield(v)
		})
	}
}

// MapSeq 返回一个 Seq[U],对 q 产出的每个 v yield f(v)。它是改变类型的
// lazy map,以自由函数形式提供,因为方法无法改变 receiver 的类型参数。
func MapSeq[T, U any](q Seq[T], f func(T) U) Seq[U] {
	return func(yield func(U) bool) {
		q(func(v T) bool {
			return yield(f(v))
		})
	}
}

// ForEach 对 q 的每个元素调用 fn。它不支持提前终止;需要 short-circuit
// 消费请用 Any/All/Find。
func (q Seq[T]) ForEach(fn func(T)) {
	q(func(v T) bool {
		fn(v)
		return true
	})
}

// Count 返回 q 产出的元素个数。它会完全消费 q。
func (q Seq[T]) Count() int {
	n := 0
	q(func(v T) bool {
		n++
		return true
	})
	return n
}

// Any 当 q 中任一元素满足 pred 时返回 true。它会 short-circuit,
// 在首个匹配处停止。
func (q Seq[T]) Any(pred func(T) bool) bool {
	found := false
	q(func(v T) bool {
		if pred(v) {
			found = true
			return false // 停止
		}
		return true
	})
	return found
}

// All 当 q 中每个元素都满足 pred 时返回 true。它会 short-circuit,
// 在首个不匹配处停止。
func (q Seq[T]) All(pred func(T) bool) bool {
	ok := true
	q(func(v T) bool {
		if !pred(v) {
			ok = false
			return false // 停止
		}
		return true
	})
	return ok
}

// Find 返回 q 中首个满足 pred 的元素,否则返回 (zero, false)。
// 它在首个匹配处 short-circuit。
func (q Seq[T]) Find(pred func(T) bool) (T, bool) {
	var result T
	found := false
	q(func(v T) bool {
		if pred(v) {
			result = v
			found = true
			return false // 停止
		}
		return true
	})
	return result, found
}

// First 返回 q 的首个元素,q 为空时返回 (zero, false)。
func (q Seq[T]) First() (T, bool) {
	var result T
	found := false
	q(func(v T) bool {
		result = v
		found = true
		return false // 取首个后停止
	})
	return result, found
}

// Collect 将 q 物化为 slice。由于 Seq 是 push 模型,结果长度事先未知,
// 元素通过 append 加入。
func (q Seq[T]) Collect() []T {
	out := make([]T, 0)
	q(func(v T) bool {
		out = append(out, v)
		return true
	})
	return out
}
