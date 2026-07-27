package fn

import (
	"cmp"
	"slices"
)

// Chain 是构建在 slice 之上的 fluent builder。每个操作都物化出一个新 slice,
// 因此 Chain 易于阅读和调试,通常在 Go 中与一堆 closure 一样快甚至更快。
//
// 用 ChainOf 构造;用 Collect(或其别名 ToSlice)终止。
//
//	ChainOf(users).
//	    Filter(isActive).
//	    Map(func(u User) User { return u /* ... */ }).
//	    TopK(10, func(u User) int { return u.Score }).
//	    Collect()
//
// 改变类型的 map 使用自由函数 MapTo,因为 Go 的方法 receiver 不能改变
// receiver 的类型参数:
//
//	scores := MapTo(ChainOf(users), func(u User) int { return u.Score }).Collect()
type Chain[T any] struct {
	s []T
}

// ChainOf 将 s 包装为 Chain。被包装的 slice 共享输入的 backing array;
// 读侧操作为 zero-copy,写操作采用 copy on write。
func ChainOf[T any](s []T) Chain[T] {
	return Chain[T]{s: s}
}

// MapTo 将 c 中每个元素转换为类型 U,返回 Chain[U]。它是 fluent API 中
// 改变类型的 map,以自由函数形式提供,因为 Go 方法 receiver 无法返回
// 不同类型参数的 Chain。
func MapTo[T, U any](c Chain[T], f func(T) U) Chain[U] {
	return Chain[U]{s: Map(c.s, f)}
}

// SortChain 在 T 为 cmp.Ordered 时将 Chain 升序排序。它是自由函数,
// 因为 Go 方法无法为 receiver 添加类型参数 constraint;对任意 T 基于
// comparator 排序请用 SortBy。
func SortChain[T cmp.Ordered](c Chain[T]) Chain[T] {
	out := append(make([]T, 0, len(c.s)), c.s...)
	slices.Sort(out)
	return Chain[T]{s: out}
}

// Len 返回 chain 中的元素个数。
func (c Chain[T]) Len() int { return len(c.s) }

// Slice 返回当前 backing slice(非副本)。修改它会修改 chain;
// 需要独立副本请用 Collect。
func (c Chain[T]) Slice() []T { return c.s }

// Map 原地转换每个元素(T -> T)。T -> U 请用 MapTo。
func (c Chain[T]) Map(f func(T) T) Chain[T] { return Chain[T]{s: Map(c.s, f)} }

// Filter 保留 pred 返回 true 的元素。
func (c Chain[T]) Filter(pred func(T) bool) Chain[T] {
	return Chain[T]{s: Filter(c.s, pred)}
}

// FilterIdx 保留 pred(index, value) 返回 true 的元素。
func (c Chain[T]) FilterIdx(pred func(int, T) bool) Chain[T] {
	return Chain[T]{s: FilterIdx(c.s, pred)}
}

// FlatMap 将每个元素映射为 slice 并展开结果(T -> []T)。
func (c Chain[T]) FlatMap(f func(T) []T) Chain[T] {
	return Chain[T]{s: FlatMap(c.s, f)}
}

// Take 返回一个持有至多前 n 个元素的 Chain。
func (c Chain[T]) Take(n int) Chain[T] {
	if n < 0 {
		n = 0
	}
	if n > len(c.s) {
		n = len(c.s)
	}
	out := make([]T, n)
	copy(out, c.s[:n])
	return Chain[T]{s: out}
}

// Drop 返回一个持有除前 n 个之外全部元素的 Chain。
func (c Chain[T]) Drop(n int) Chain[T] {
	if n < 0 {
		n = 0
	}
	if n > len(c.s) {
		n = len(c.s)
	}
	out := make([]T, len(c.s)-n)
	copy(out, c.s[n:])
	return Chain[T]{s: out}
}

// DeduplicateByChain 返回一个去除 key(由 keyFn 给出)重复元素后的 Chain,
// 保留首次出现。它是自由函数,因为 Go 方法不能声明类型参数;
// comparable constraint 在此处落在 K 上。
func DeduplicateByChain[T any, K comparable](c Chain[T], keyFn func(T) K) Chain[T] {
	return Chain[T]{s: DeduplicateBy(c.s, keyFn)}
}

// Reverse 返回一个元素逆序的 Chain。
func (c Chain[T]) Reverse() Chain[T] { return Chain[T]{s: Reverse(c.s)} }

// SortBy 使用 cmp 风格 comparator(a < b 时为负)对 chain 排序。
func (c Chain[T]) SortBy(less By[T]) Chain[T] {
	out := append(make([]T, 0, len(c.s)), c.s...)
	slices.SortFunc(out, less)
	return Chain[T]{s: out}
}

// TopK 保留 k 个最大的元素(按整数 key),按降序排列。
// 非整数 key 请使用 standalone 的 TopK 函数。
func (c Chain[T]) TopK(k int, key func(T) int) Chain[T] {
	return Chain[T]{s: TopK(c.s, k, Key[T, int](key))}
}

// Concat 将 other 的元素追加到本 chain 的副本之后。
func (c Chain[T]) Concat(other []T) Chain[T] {
	return Chain[T]{s: Concat(c.s, other)}
}

// Collect 返回 chain slice 的副本。
func (c Chain[T]) Collect() []T {
	out := make([]T, len(c.s))
	copy(out, c.s)
	return out
}

// ToSlice 是 Collect 的别名。
func (c Chain[T]) ToSlice() []T { return c.Collect() }

// ForEach 对 chain 中每个元素调用 fn。
func (c Chain[T]) ForEach(fn func(T)) { ForEach(c.s, fn) }

// Reduce 将 chain 折叠为单个值。
func (c Chain[T]) Reduce(reducer func(acc T, elem T) T, initial T) T {
	return Reduce(c.s, reducer, initial)
}

// Any 当任一元素满足 pred 时返回 true(short-circuit)。
func (c Chain[T]) Any(pred func(T) bool) bool { return Any(c.s, pred) }

// All 当所有元素满足 pred 时返回 true(short-circuit)。
func (c Chain[T]) All(pred func(T) bool) bool { return All(c.s, pred) }

// Find 返回首个满足 pred 的元素,否则返回 (zero, false)。
func (c Chain[T]) Find(pred func(T) bool) (T, bool) { return Find(c.s, pred) }
