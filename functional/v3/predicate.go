package fn

// --- 量词 ---

// All 当 s 中每个元素都满足 pred 时返回 true。
// 空 slice 返回 true(空真),符合标准量词语义,并与 [slices.IndexFunc] 的
// 循环不变式一致。
func All[T any](s []T, pred func(T) bool) bool {
	for _, v := range s {
		if !pred(v) {
			return false
		}
	}
	return true
}

// Any 当 s 中至少一个元素满足 pred 时返回 true。
// 空 slice 返回 false。
func Any[T any](s []T, pred func(T) bool) bool {
	for _, v := range s {
		if pred(v) {
			return true
		}
	}
	return false
}

// None 当 s 中没有元素满足 pred 时返回 true。
// 空 slice 返回 true。None 是 Any 的否定。
func None[T any](s []T, pred func(T) bool) bool {
	for _, v := range s {
		if pred(v) {
			return false
		}
	}
	return true
}

// --- 成员判定 ---

// Contains 报告 v 是否存在于 s 中。它是 O(n);对同一集合反复做 membership
// 检查时,构建一个 OrderedSet 以获得 O(1) 查找。
func Contains[T comparable](s []T, v T) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// ContainsBy 报告 s 中是否有元素在 keyFn 下映射到 key。
func ContainsBy[T any, K comparable](s []T, key K, keyFn func(T) K) bool {
	for _, x := range s {
		if keyFn(x) == key {
			return true
		}
	}
	return false
}

// Count 返回 s 中满足 pred 的元素个数。
func Count[T any](s []T, pred func(T) bool) int {
	n := 0
	for _, v := range s {
		if pred(v) {
			n++
		}
	}
	return n
}

// CountBy 返回 key(由 keyFn 给出)等于 key 的元素个数。
func CountBy[T any, K comparable](s []T, key K, keyFn func(T) K) int {
	n := 0
	for _, x := range s {
		if keyFn(x) == key {
			n++
		}
	}
	return n
}

// --- 查找 / 索引 ---

// Find 返回 s 中首个满足 pred 的元素,若无匹配则返回 (zero, false)。
func Find[T any](s []T, pred func(T) bool) (T, bool) {
	for _, v := range s {
		if pred(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FindIndex 返回 s 中首个满足 pred 元素的 index,若无匹配则返回 (-1, false)。
func FindIndex[T any](s []T, pred func(T) bool) (int, bool) {
	for i, v := range s {
		if pred(v) {
			return i, true
		}
	}
	return -1, false
}

// FindLast 返回 s 中最后一个满足 pred 的元素,否则返回 (zero, false)。
func FindLast[T any](s []T, pred func(T) bool) (T, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if pred(s[i]) {
			return s[i], true
		}
	}
	var zero T
	return zero, false
}

// FindLastIndex 返回 s 中最后一个满足 pred 元素的 index,
// 若无匹配则返回 (-1, false)。
func FindLastIndex[T any](s []T, pred func(T) bool) (int, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if pred(s[i]) {
			return i, true
		}
	}
	return -1, false
}

// IndexOf 返回 v 在 s 中首次出现位置的 index,否则返回 (-1, false)。
func IndexOf[T comparable](s []T, v T) (int, bool) {
	for i, x := range s {
		if x == v {
			return i, true
		}
	}
	return -1, false
}

// LastIndexOf 返回 v 在 s 中最后出现位置的 index,否则返回 (-1, false)。
func LastIndexOf[T comparable](s []T, v T) (int, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == v {
			return i, true
		}
	}
	return -1, false
}
