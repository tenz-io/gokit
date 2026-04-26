package function

import (
	"cmp"
	"container/heap"
	"slices"
)

// Min returns the minimum element in a slice of ordered values.
func Min[T cmp.Ordered](list []T) (T, bool) {
	if len(list) == 0 {
		var zero T
		return zero, false
	}
	cur := list[0]
	for _, elem := range list[1:] {
		if elem < cur {
			cur = elem
		}
	}
	return cur, true
}

// Max returns the maximum element in a slice of ordered values.
func Max[T cmp.Ordered](list []T) (T, bool) {
	if len(list) == 0 {
		var zero T
		return zero, false
	}
	cur := list[0]
	for _, elem := range list[1:] {
		if elem > cur {
			cur = elem
		}
	}
	return cur, true
}

// Sum returns the sum of elements.
// Returns the zero value for an empty slice.
func Sum[T cmp.Ordered](list []T) T {
	var sum T
	for _, elem := range list {
		sum += elem
	}
	return sum
}

// MinBy returns the minimum element based on a less function.
func MinBy[T any](list []T, less func(a, b T) bool) (T, bool) {
	if len(list) == 0 {
		var zero T
		return zero, false
	}
	cur := list[0]
	for _, elem := range list[1:] {
		if less(elem, cur) {
			cur = elem
		}
	}
	return cur, true
}

// MaxBy returns the maximum element based on a less function.
func MaxBy[T any](list []T, less func(a, b T) bool) (T, bool) {
	if len(list) == 0 {
		var zero T
		return zero, false
	}
	cur := list[0]
	for _, elem := range list[1:] {
		if less(cur, elem) {
			cur = elem
		}
	}
	return cur, true
}

// TopK returns the top-k largest elements according to the less function.
// The result is sorted in descending order (largest first).
// If k >= len(list), all elements are returned in descending order.
func TopK[T any](list []T, k int, less func(a, b T) bool) []T {
	if len(list) == 0 || k <= 0 {
		return []T{}
	}
	if k >= len(list) {
		res := make([]T, len(list))
		copy(res, list)
		slices.SortFunc(res, func(a, b T) int {
			if less(a, b) {
				return 1
			}
			if less(b, a) {
				return -1
			}
			return 0
		})
		return res
	}

	h := newTopKHeap(list[:k], less)
	heap.Init(h)

	for i := k; i < len(list); i++ {
		if less(h.items[0], list[i]) {
			h.items[0] = list[i]
			heap.Fix(h, 0)
		}
	}

	res := make([]T, k)
	for i := k - 1; i >= 0; i-- {
		res[i] = heap.Pop(h).(T)
	}
	return res
}
