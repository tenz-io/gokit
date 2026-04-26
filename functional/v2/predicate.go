package function

// All returns true if all elements satisfy the predicate.
// Returns true for an empty slice (vacuous truth).
func All[T any](list []T, predicate func(T) bool) bool {
	for _, elem := range list {
		if !predicate(elem) {
			return false
		}
	}
	return true
}

// Any returns true if at least one element satisfies the predicate.
func Any[T any](list []T, predicate func(T) bool) bool {
	for _, elem := range list {
		if predicate(elem) {
			return true
		}
	}
	return false
}

// None returns true if no elements satisfy the predicate.
// Returns true for an empty slice.
func None[T any](list []T, predicate func(T) bool) bool {
	for _, elem := range list {
		if predicate(elem) {
			return false
		}
	}
	return true
}

// Contains returns true if elem is found in the slice.
func Contains[T comparable](list []T, elem T) bool {
	for _, item := range list {
		if item == elem {
			return true
		}
	}
	return false
}

// ContainsBy returns true if the slice contains an element whose key matches the given key.
func ContainsBy[T any, K comparable](list []T, key K, keyFn func(T) K) bool {
	for _, item := range list {
		if keyFn(item) == key {
			return true
		}
	}
	return false
}
