package fn

import (
	"cmp"
	"slices"
)

// number is the set of built-in numeric types whose values can be widened to
// float64. It mirrors the subset of cmp.Ordered that is arithmetic (i.e. it
// excludes strings). The standard library does not expose a numeric
// constraint, so we declare one here.
type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// --- Min / Max (ordered) ---

// Min returns the minimum element of s, or (zero, false) if s is empty.
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

// Max returns the maximum element of s, or (zero, false) if s is empty.
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

// --- Sum / Avg ---

// Sum returns the sum of the elements of s. Returns the zero value for an
// empty slice.
func Sum[T cmp.Ordered](s []T) T {
	var sum T
	for _, v := range s {
		sum += v
	}
	return sum
}

// Avg returns the arithmetic mean of s as a float64, and ok=false if s is
// empty. T must be a built-in numeric type (integer or float); values are
// widened to float64 for the computation.
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

// --- Min / Max by key extractor ---

// MinByKey returns the element of s with the smallest key (under key), or
// (zero, false) if s is empty. Ties are broken by first occurrence.
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

// MaxByKey returns the element of s with the largest key (under key), or
// (zero, false) if s is empty. Ties are broken by first occurrence.
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

// --- Min / Max by comparator ---

// MinBy returns the element of s for which cmp returns the smallest value, or
// (zero, false) if s is empty. c is a cmp-style comparator
// (negative when a < b).
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

// MaxBy returns the element of s for which c reports it as largest, or
// (zero, false) if s is empty. c is a cmp-style comparator.
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

// --- TopK / BottomK by key extractor ---

// TopK returns the k elements of s with the largest keys, in descending key
// order. If k >= len(s), all elements are returned, descending.
//
// The key extractor lets you order by a single cmp.Ordered field — e.g.
// TopK(users, 10, fn.Key(func(u User) int { return u.Score })).
func TopK[T any, K cmp.Ordered](s []T, k int, key Key[T, K]) []T {
	if k <= 0 || len(s) == 0 {
		return nil
	}
	n := len(s)
	if k >= n {
		out := append(make([]T, 0, n), s...)
		slices.SortFunc(out, func(a, b T) int {
			// Descending by key.
			return cmp.Compare(key(b), key(a))
		})
		return out
	}
	// Selection: keep a min-heap of the k largest seen so far. Evict the root
	// (smallest of the survivors) when a candidate has a larger key.
	root := topKHeap(s, k,
		func(a, b T) bool { return key(a) < key(b) }, // min-heap by key
		func(root, cand T) bool { return key(cand) > key(root) },
	)
	// root holds the k survivors in min-heap order; sort to descending.
	slices.SortFunc(root, func(a, b T) int { return cmp.Compare(key(b), key(a)) })
	return root
}

// BottomK returns the k elements of s with the smallest keys, in ascending key
// order. If k >= len(s), all elements are returned, ascending.
func BottomK[T any, K cmp.Ordered](s []T, k int, key Key[T, K]) []T {
	if k <= 0 || len(s) == 0 {
		return nil
	}
	n := len(s)
	if k >= n {
		out := append(make([]T, 0, n), s...)
		slices.SortFunc(out, func(a, b T) int {
			return cmp.Compare(key(a), key(b)) // Ascending by key.
		})
		return out
	}
	// Selection: keep a max-heap of the k smallest seen so far. Model a
	// max-heap by inverting the min-heap's comparator.
	root := topKHeap(s, k,
		func(a, b T) bool { return key(a) > key(b) }, // max-heap via reversed key
		func(root, cand T) bool { return key(cand) < key(root) },
	)
	slices.SortFunc(root, func(a, b T) int { return cmp.Compare(key(a), key(b)) }) // ascending
	return root
}

// --- TopK / BottomK by comparator ---

// TopKBy returns the k largest elements per the cmp-style comparator c, in
// descending order per c. Use this when ordering cannot be expressed by a
// single ordered key.
func TopKBy[T any](s []T, k int, c By[T]) []T {
	if k <= 0 || len(s) == 0 {
		return nil
	}
	n := len(s)
	if k >= n {
		out := append(make([]T, 0, n), s...)
		slices.SortFunc(out, func(a, b T) int { return c(b, a) }) // descending
		return out
	}
	root := topKHeap(s, k,
		func(a, b T) bool { return c(a, b) < 0 }, // min-heap: a "less" than b
		func(root, cand T) bool { return c(cand, root) > 0 },
	)
	slices.SortFunc(root, func(a, b T) int { return c(b, a) }) // descending
	return root
}

// BottomKBy returns the k smallest elements per the cmp-style comparator c,
// in ascending order per c.
func BottomKBy[T any](s []T, k int, c By[T]) []T {
	if k <= 0 || len(s) == 0 {
		return nil
	}
	n := len(s)
	if k >= n {
		out := append(make([]T, 0, n), s...)
		slices.SortFunc(out, c) // ascending
		return out
	}
	root := topKHeap(s, k,
		func(a, b T) bool { return c(a, b) > 0 }, // max-heap via reversed comparator
		func(root, cand T) bool { return c(cand, root) < 0 },
	)
	slices.SortFunc(root, c) // ascending
	return root
}
