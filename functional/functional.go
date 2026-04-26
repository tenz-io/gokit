// Package function provides generic functional programming utilities for Go slices.
//
// Operations are organized into:
//   - Core transformations: Map, Filter, Reduce, ForEach, Flatten, Reverse
//   - Predicates: All, Any, None, Contains, ContainsBy
//   - Aggregations: Min, Max, Sum, MinBy, MaxBy, TopK
//   - Conditionals: If, When, IfElse
//   - Transformations: Deduplicate, DeduplicateBy, GroupBy, Partition
package function

// Map transforms each element in a slice using the mapper function.
func Map[T, U any](list []T, mapper func(T) U) []U {
	results := make([]U, 0, len(list))
	for _, item := range list {
		results = append(results, mapper(item))
	}
	return results
}

// Filter returns elements that satisfy the predicate.
func Filter[T any](list []T, predicate func(T) bool) []T {
	results := make([]T, 0, len(list))
	for _, elem := range list {
		if predicate(elem) {
			results = append(results, elem)
		}
	}
	return results
}

// Reduce folds a slice into a single value using a reducer function and an initial accumulator.
func Reduce[T, U any](list []T, reducer func(accumulator U, elem T) U, initial U) U {
	acc := initial
	for _, item := range list {
		acc = reducer(acc, item)
	}
	return acc
}

// ForEach applies fn to each element in the slice.
func ForEach[T any](list []T, fn func(T)) {
	for _, elem := range list {
		fn(elem)
	}
}

// Flatten concatenates all sub-slices into a single flat slice.
func Flatten[T any](list [][]T) []T {
	var total int
	for _, sub := range list {
		total += len(sub)
	}
	results := make([]T, 0, total)
	for _, sub := range list {
		results = append(results, sub...)
	}
	return results
}

// Reverse returns a new slice with elements in reverse order.
func Reverse[T any](list []T) []T {
	if len(list) == 0 {
		return []T{}
	}
	results := make([]T, len(list))
	for i, j := 0, len(list)-1; i < len(list); i, j = i+1, j-1 {
		results[i] = list[j]
	}
	return results
}

// ReverseInPlace reverses the slice in place.
func ReverseInPlace[T any](list []T) {
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
}

// Find returns the first element matching the predicate, or (zero, false) if none found.
func Find[T any](list []T, predicate func(T) bool) (T, bool) {
	for _, elem := range list {
		if predicate(elem) {
			return elem, true
		}
	}
	var zero T
	return zero, false
}

// Count returns the number of elements satisfying the predicate.
func Count[T any](list []T, predicate func(T) bool) int {
	var n int
	for _, elem := range list {
		if predicate(elem) {
			n++
		}
	}
	return n
}
