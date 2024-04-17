package function

import (
	"container/heap"

	"cmp"
	"slices"
)

// Map is a function that maps a list of items to another list of items based on a function.
func Map[T, U any](list []T, mapper func(T) U) []U {
	if list == nil {
		return nil
	}

	results := make([]U, 0, len(list))
	for _, item := range list {
		results = append(results, mapper(item))
	}
	return results
}

// Filter is a function that filters a list based on a predicate function.
func Filter[T any](list []T, predicate func(T) bool) []T {
	if list == nil {
		return nil
	}

	results := make([]T, 0, len(list))
	for _, elem := range list {
		if predicate(elem) {
			results = append(results, elem)
		}
	}
	return results
}

// GroupBy is a function that groups elements in a slice by a common value, generated by a provided function.
func GroupBy[T any, K comparable](list []T, keyFn func(T) K) map[K][]T {
	if list == nil {
		return nil
	}

	groups := make(map[K][]T, len(list))
	for _, elem := range list {
		key := keyFn(elem)
		groups[key] = append(groups[key], elem)
	}
	return groups
}

// Flatten is a function that flattens a slice of slices.
func Flatten[T any](list [][]T) []T {
	if list == nil {
		return nil
	}

	results := make([]T, 0, len(list))
	for _, subList := range list {
		results = append(results, subList...)
	}
	return results
}

// If is a function that returns a value if the condition is true, otherwise returns another value.
func If[T any](cond bool, ifVal, elseVal T) T {
	if cond {
		return ifVal
	}
	return elseVal
}

// IfThen is a function that applies a function to a value if the condition is true, otherwise returns the value.
func IfThen[T any](cond bool, val T, apply func(T) T) T {
	if cond {
		return apply(val)
	}
	return val
}

// IfElse is a function that returns a value if the condition is true, otherwise returns the result of another function.
func IfElse[T any](cond bool, ifFn, elseFn func() T) T {
	if cond {
		return ifFn()
	}
	return elseFn()
}

// Deduplicate is a function that removes duplicates from a slice.
func Deduplicate[T comparable](list []T) []T {
	if list == nil {
		return nil
	}

	seen := make(map[T]struct{}, len(list))
	results := make([]T, 0, len(list))
	for _, elem := range list {
		if _, ok := seen[elem]; ok {
			continue
		}
		seen[elem] = struct{}{}
		results = append(results, elem)
	}
	return results
}

// DeduplicateWith is a function that removes duplicates from a slice based on a key function.
func DeduplicateWith[T any, K comparable](list []T, keyFn func(T) K) []T {
	if list == nil {
		return nil
	}

	seen := make(map[K]struct{}, len(list))
	results := make([]T, 0, len(list))
	for _, elem := range list {
		key := keyFn(elem)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		results = append(results, elem)
	}
	return results
}

// Reverse is a function that reverses a slice.
func Reverse[T any](list []T) []T {
	if list == nil {
		return nil
	}

	results := make([]T, len(list))
	for i, j := 0, len(list)-1; i < len(list); i, j = i+1, j-1 {
		results[i] = list[j]
	}
	return results
}

// Contains is a function that checks if a slice contains an element.
func Contains[T comparable](list []T, elem T) bool {
	if list == nil {
		return false
	}

	for _, item := range list {
		if item == elem {
			return true
		}
	}
	return false
}

// ContainsWith is a function that checks if a slice contains an element based on a key function.
func ContainsWith[T any, K comparable](list []T, elem T, keyFn func(T) K) bool {
	if list == nil {
		return false
	}

	key := keyFn(elem)
	for _, item := range list {
		if keyFn(item) == key {
			return true
		}
	}
	return false
}

// All is a function that checks if all elements in a slice satisfy a predicate function.
func All[T any](list []T, predicate func(T) bool) bool {
	if len(list) == 0 {
		return false
	}

	for _, elem := range list {
		if !predicate(elem) {
			return false
		}
	}
	return true
}

// Any is a function that checks if any element in a slice satisfies a predicate function.
func Any[T any](list []T, predicate func(T) bool) bool {
	if len(list) == 0 {
		return false
	}

	for _, elem := range list {
		if predicate(elem) {
			return true
		}
	}
	return false
}

// None is a function that checks if no elements in a slice satisfy a predicate function.
func None[T any](list []T, predicate func(T) bool) bool {
	if len(list) == 0 {
		return true
	}

	for _, elem := range list {
		if predicate(elem) {
			return false
		}
	}
	return true
}

type number cmp.Ordered

// Min is a function that returns the minimum element in a slice.
func Min[T number](list []T) (T, bool) {
	if len(list) == 0 {
		var zero T
		return zero, false
	}

	curMin := list[0]
	for _, elem := range list {
		if elem < curMin {
			curMin = elem
		}
	}

	return curMin, true

}

// MinWith is a function that returns the minimum element in a slice based on a comparison function.
func MinWith[T any](list []T, less func(T, T) bool) (T, bool) {
	if len(list) == 0 {
		var zero T
		return zero, false
	}

	curMin := list[0]
	for _, elem := range list {
		if less(elem, curMin) {
			curMin = elem
		}
	}
	return curMin, true
}

// Max is a function that returns the maximum element in a slice.
func Max[T number](list []T) (T, bool) {
	if len(list) == 0 {
		var zero T
		return zero, false
	}

	curMax := list[0]
	for _, elem := range list {
		if elem > curMax {
			curMax = elem
		}
	}
	return curMax, true
}

// MaxWith is a function that returns the maximum element in a slice based on a comparison function.
func MaxWith[T any](list []T, less func(t1, t2 T) bool) (T, bool) {
	if len(list) == 0 {
		var zero T
		return zero, false
	}

	curMax := list[0]
	for _, elem := range list {
		if less(curMax, elem) {
			curMax = elem
		}
	}
	return curMax, true
}

// Sum is a function that returns the sum of elements in a slice.
func Sum[T number](list []T) T {
	if len(list) == 0 {
		var zero T
		return zero
	}

	var sum T
	for _, elem := range list {
		sum += elem
	}
	return sum
}

// Median is a function that returns the median of elements in a slice.
func Median[T number](list []T) (val T, ok bool) {
	if len(list) == 0 {
		var zero T
		return zero, false
	}

	mid := len(list) / 2
	subList := TopK(list, mid+1, func(t1, t2 T) int {
		if t1 < t2 {
			return -1
		}
		if t1 > t2 {
			return 1
		}
		return 0
	})
	if len(subList) == 0 {
		return val, false
	}

	return subList[len(subList)-1], true

}

// Partition is a function that partitions a slice into two slices based on a predicate function.
// The first slice contains elements that satisfy the predicate, and the second slice contains elements that do not.
func Partition[T any](list []T, predicate func(T) bool) (satisfied, unsatisfied []T) {
	if list == nil {
		return nil, nil
	}

	satisfied = make([]T, 0, len(list))
	unsatisfied = make([]T, 0, len(list))
	for _, elem := range list {
		if predicate(elem) {
			satisfied = append(satisfied, elem)
		} else {
			unsatisfied = append(unsatisfied, elem)
		}
	}
	return satisfied, unsatisfied
}

// TopK is a function that returns the top k elements in a slice based on a comparison function.
// cmp should return a positive number if t1 is greater than t2, a negative number if t1 is cmp than t2, and 0 if they are equal.
// The function uses a min heap to keep track of the top k elements.
func TopK[T any](list []T, k int, cmp func(t1, t2 T) int) []T {
	if len(list) == 0 || k == 0 {
		return []T{}
	}
	if k >= len(list) {
		res := make([]T, len(list))
		copy(res, list)
		slices.SortFunc(res, func(t1, t2 T) int {
			return cmp(t2, t1)
		})
		return res
	}

	it := newItemHeap(list[:k], cmp)
	heap.Init(it)

	for i := k; i < len(list); i++ {
		it.Update(list[i])
	}

	res := make([]T, k)
	for i := len(res) - 1; i >= 0; i-- {
		res[i] = heap.Pop(it).(T)
	}

	return res
}
