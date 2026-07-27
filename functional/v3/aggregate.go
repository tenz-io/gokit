package fn

import (
	"cmp"
	"slices"
)

// number 是可被 widen 到 float64 的内置数值类型集合。它对应 cmp.Ordered
// 中算术子集(即排除 strings)。标准库未暴露数值 constraint,因此在此声明。
type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// --- 最小 / 最大(有序) ---

// Min 返回 s 的最小元素,s 为空时返回 (zero, false)。
func Min[T cmp.Ordered](s []T) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	cur := s[0]
	for _, v := range s[1:] {
		if v < cur {
			cur = v
		}
	}
	return cur, true
}

// Max 返回 s 的最大元素,s 为空时返回 (zero, false)。
func Max[T cmp.Ordered](s []T) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	cur := s[0]
	for _, v := range s[1:] {
		if v > cur {
			cur = v
		}
	}
	return cur, true
}

// --- 求和 / 均值 ---

// Sum 返回 s 各元素之和。空 slice 返回零值。
func Sum[T cmp.Ordered](s []T) T {
	var sum T
	for _, v := range s {
		sum += v
	}
	return sum
}

// Avg 以 float64 返回 s 的算术平均,s 为空时 ok=false。T 必须是内置数值
// 类型(整数或浮点);计算时值会被 widen 到 float64。
func Avg[T number](s []T) (float64, bool) {
	if len(s) == 0 {
		return 0, false
	}
	var sum float64
	for _, v := range s {
		sum += float64(v)
	}
	return sum / float64(len(s)), true
}

// --- 按 key 提取器取最小 / 最大 ---

// MinByKey 返回 s 中 key(由 key 给出)最小的元素,s 为空时返回
// (zero, false)。并列时按首次出现者取胜。
func MinByKey[T any, K cmp.Ordered](s []T, key Key[T, K]) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	cur := s[0]
	curKey := key(cur)
	for _, v := range s[1:] {
		k := key(v)
		if k < curKey {
			cur = v
			curKey = k
		}
	}
	return cur, true
}

// MaxByKey 返回 s 中 key(由 key 给出)最大的元素,s 为空时返回
// (zero, false)。并列时按首次出现者取胜。
func MaxByKey[T any, K cmp.Ordered](s []T, key Key[T, K]) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	cur := s[0]
	curKey := key(cur)
	for _, v := range s[1:] {
		k := key(v)
		if k > curKey {
			cur = v
			curKey = k
		}
	}
	return cur, true
}

// --- 按 comparator 取最小 / 最大 ---

// MinBy 返回 s 中由 cmp 判定为最小的元素,s 为空时返回 (zero, false)。
// c 是 cmp 风格的 comparator(a < b 时为负)。
func MinBy[T any](s []T, c By[T]) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	cur := s[0]
	for _, v := range s[1:] {
		if c(v, cur) < 0 {
			cur = v
		}
	}
	return cur, true
}

// MaxBy 返回 s 中由 c 判定为最大的元素,s 为空时返回 (zero, false)。
// c 是 cmp 风格的 comparator。
func MaxBy[T any](s []T, c By[T]) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	cur := s[0]
	for _, v := range s[1:] {
		if c(v, cur) > 0 {
			cur = v
		}
	}
	return cur, true
}

// --- 按 key 提取器的 TopK / BottomK ---

// TopK 返回 s 中 key 最大的 k 个元素,按 key 降序排列。
// 若 k >= len(s),返回全部元素,降序。
//
// key extractor 让你按单个 cmp.Ordered 字段排序——例如
// TopK(users, 10, fn.Key(func(u User) int { return u.Score }))。
func TopK[T any, K cmp.Ordered](s []T, k int, key Key[T, K]) []T {
	if k <= 0 || len(s) == 0 {
		return nil
	}
	n := len(s)
	if k >= n {
		out := append(make([]T, 0, n), s...)
		slices.SortFunc(out, func(a, b T) int {
			// 按 key 降序。
			return cmp.Compare(key(b), key(a))
		})
		return out
	}
	// 选择:维护一个含目前为止最大 k 个元素的 min-heap。当候选元素 key
	// 更大时驱逐 root(幸存者中最小的)。
	root := topKHeap(s, k,
		func(a, b T) bool { return key(a) < key(b) }, // 按 key 的 min-heap
		func(root, cand T) bool { return key(cand) > key(root) },
	)
	// root 以 min-heap 顺序持有 k 个幸存者;排序为降序。
	slices.SortFunc(root, func(a, b T) int { return cmp.Compare(key(b), key(a)) })
	return root
}

// BottomK 返回 s 中 key 最小的 k 个元素,按 key 升序排列。
// 若 k >= len(s),返回全部元素,升序。
func BottomK[T any, K cmp.Ordered](s []T, k int, key Key[T, K]) []T {
	if k <= 0 || len(s) == 0 {
		return nil
	}
	n := len(s)
	if k >= n {
		out := append(make([]T, 0, n), s...)
		slices.SortFunc(out, func(a, b T) int {
			return cmp.Compare(key(a), key(b)) // 按 key 升序。
		})
		return out
	}
	// 选择:维护一个含目前为止最小 k 个元素的 max-heap。通过反转
	// min-heap 的 comparator 来模拟 max-heap。
	root := topKHeap(s, k,
		func(a, b T) bool { return key(a) > key(b) }, // 反转 key 得到 max-heap
		func(root, cand T) bool { return key(cand) < key(root) },
	)
	slices.SortFunc(root, func(a, b T) int { return cmp.Compare(key(a), key(b)) }) // 升序
	return root
}

// --- 按 comparator 的 TopK / BottomK ---

// TopKBy 按 cmp 风格 comparator c 返回 k 个最大元素,按 c 降序排列。
// 当排序无法用单个有序 key 表达时使用。
func TopKBy[T any](s []T, k int, c By[T]) []T {
	if k <= 0 || len(s) == 0 {
		return nil
	}
	n := len(s)
	if k >= n {
		out := append(make([]T, 0, n), s...)
		slices.SortFunc(out, func(a, b T) int { return c(b, a) }) // 降序
		return out
	}
	root := topKHeap(s, k,
		func(a, b T) bool { return c(a, b) < 0 }, // min-heap:a 比 b "更小"
		func(root, cand T) bool { return c(cand, root) > 0 },
	)
	slices.SortFunc(root, func(a, b T) int { return c(b, a) }) // 降序
	return root
}

// BottomKBy 按 cmp 风格 comparator c 返回 k 个最小元素,按 c 升序排列。
func BottomKBy[T any](s []T, k int, c By[T]) []T {
	if k <= 0 || len(s) == 0 {
		return nil
	}
	n := len(s)
	if k >= n {
		out := append(make([]T, 0, n), s...)
		slices.SortFunc(out, c) // 升序
		return out
	}
	root := topKHeap(s, k,
		func(a, b T) bool { return c(a, b) > 0 }, // 反转 comparator 得到 max-heap
		func(root, cand T) bool { return c(cand, root) < 0 },
	)
	slices.SortFunc(root, c) // 升序
	return root
}
