package fn

// --- Quantifiers ---

// All returns true if every element of s satisfies pred.
// An empty slice returns true (vacuous truth), matching standard quantifier
// semantics and matching [slices.IndexFunc] loop invariants.
func All[T any](s []T, pred func(T) bool) bool {
	for _, v := range s {
		if !pred(v) {
			return false
		}
	}
	return true
}

// Any returns true if at least one element of s satisfies pred.
// An empty slice returns false.
func Any[T any](s []T, pred func(T) bool) bool {
	for _, v := range s {
		if pred(v) {
			return true
		}
	}
	return false
}

// None returns true if no element of s satisfies pred.
// An empty slice returns true. None is the negation of Any.
func None[T any](s []T, pred func(T) bool) bool {
	for _, v := range s {
		if pred(v) {
			return false
		}
	}
	return true
}

// --- Membership ---

// Contains reports whether v is present in s. It is O(n); for repeated
// membership checks over the same set, build an OrderedSet for O(1) lookups.
func Contains[T comparable](s []T, v T) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// ContainsBy reports whether any element of s maps to key under keyFn.
func ContainsBy[T any, K comparable](s []T, key K, keyFn func(T) K) bool {
	for _, x := range s {
		if keyFn(x) == key {
			return true
		}
	}
	return false
}

// Count returns the number of elements of s that satisfy pred.
func Count[T any](s []T, pred func(T) bool) int {
	n := 0
	for _, v := range s {
		if pred(v) {
			n++
		}
	}
	return n
}

// CountBy returns the number of elements whose key (under keyFn) equals key.
func CountBy[T any, K comparable](s []T, key K, keyFn func(T) K) int {
	n := 0
	for _, x := range s {
		if keyFn(x) == key {
			n++
		}
	}
	return n
}

// --- Find / Index ---

// Find returns the first element of s satisfying pred, or (zero, false) if
// none match.
func Find[T any](s []T, pred func(T) bool) (T, bool) {
	for _, v := range s {
		if pred(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FindIndex returns the index of the first element of s satisfying pred, or
// (-1, false) if none match.
func FindIndex[T any](s []T, pred func(T) bool) (int, bool) {
	for i, v := range s {
		if pred(v) {
			return i, true
		}
	}
	return -1, false
}

// FindLast returns the last element of s satisfying pred, or (zero, false).
func FindLast[T any](s []T, pred func(T) bool) (T, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if pred(s[i]) {
			return s[i], true
		}
	}
	var zero T
	return zero, false
}

// FindLastIndex returns the index of the last element of s satisfying pred,
// or (-1, false) if none match.
func FindLastIndex[T any](s []T, pred func(T) bool) (int, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if pred(s[i]) {
			return i, true
		}
	}
	return -1, false
}

// IndexOf returns the index of the first occurrence of v in s, or (-1, false).
func IndexOf[T comparable](s []T, v T) (int, bool) {
	for i, x := range s {
		if x == v {
			return i, true
		}
	}
	return -1, false
}

// LastIndexOf returns the index of the last occurrence of v in s, or (-1, false).
func LastIndexOf[T comparable](s []T, v T) (int, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == v {
			return i, true
		}
	}
	return -1, false
}
