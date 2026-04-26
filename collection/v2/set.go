package collection

// Set is a set data structure backed by a map.
type Set[T comparable] map[T]struct{}

// NewSet creates a set, optionally pre-populated with values.
func NewSet[T comparable](values ...T) Set[T] {
	s := make(Set[T], len(values))
	for _, v := range values {
		s[v] = struct{}{}
	}
	return s
}

// NewSetWithCap creates an empty set with a capacity hint.
func NewSetWithCap[T comparable](cap int) Set[T] {
	return make(Set[T], cap)
}

// Add inserts elements into the set.
func (s Set[T]) Add(values ...T) {
	for _, v := range values {
		s[v] = struct{}{}
	}
}

// Remove deletes an element from the set. No-op if not present.
func (s Set[T]) Remove(v T) {
	delete(s, v)
}

// Contains returns true if the element is in the set.
func (s Set[T]) Contains(v T) bool {
	_, ok := s[v]
	return ok
}

// Len returns the number of elements in the set.
func (s Set[T]) Len() int { return len(s) }

// Size is an alias for Len.
func (s Set[T]) Size() int { return s.Len() }

// IsEmpty returns true if the set has no elements.
func (s Set[T]) IsEmpty() bool { return len(s) == 0 }

// Clear removes all elements from the set.
func (s Set[T]) Clear() {
	clear(s)
}

// Values returns all elements as a slice (order is non-deterministic).
func (s Set[T]) Values() []T {
	out := make([]T, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return out
}

// Clone returns a shallow copy of the set.
func Clone[T comparable](s Set[T]) Set[T] {
	out := make(Set[T], len(s))
	for k := range s {
		out[k] = struct{}{}
	}
	return out
}

// --- Set algebra ---

// Intersection returns elements present in both sets (a ∩ b).
func Intersection[T comparable](a, b Set[T]) Set[T] {
	// iterate the smaller set for efficiency
	if len(a) > len(b) {
		a, b = b, a
	}
	out := make(Set[T], len(a))
	for k := range a {
		if b.Contains(k) {
			out[k] = struct{}{}
		}
	}
	return out
}

// Union returns all elements from both sets (a ∪ b).
func Union[T comparable](a, b Set[T]) Set[T] {
	out := make(Set[T], len(a)+len(b))
	for k := range a {
		out[k] = struct{}{}
	}
	for k := range b {
		out[k] = struct{}{}
	}
	return out
}

// Difference returns elements in a but not in b (a - b).
func Difference[T comparable](a, b Set[T]) Set[T] {
	out := make(Set[T], len(a))
	for k := range a {
		if !b.Contains(k) {
			out[k] = struct{}{}
		}
	}
	return out
}

// SymmetricDifference returns elements in exactly one of the sets (a ∆ b).
func SymmetricDifference[T comparable](a, b Set[T]) Set[T] {
	out := make(Set[T], len(a)+len(b))
	for k := range a {
		if !b.Contains(k) {
			out[k] = struct{}{}
		}
	}
	for k := range b {
		if !a.Contains(k) {
			out[k] = struct{}{}
		}
	}
	return out
}

// IsSubset returns true if every element of a is in b (a ⊆ b).
func IsSubset[T comparable](a, b Set[T]) bool {
	if len(a) > len(b) {
		return false
	}
	for k := range a {
		if !b.Contains(k) {
			return false
		}
	}
	return true
}

// IsSuperset returns true if every element of b is in a (a ⊇ b).
func IsSuperset[T comparable](a, b Set[T]) bool {
	return IsSubset(b, a)
}

// IsDisjoint returns true if the sets share no elements (a ∩ b = ∅).
func IsDisjoint[T comparable](a, b Set[T]) bool {
	// iterate the smaller set
	if len(a) > len(b) {
		a, b = b, a
	}
	for k := range a {
		if b.Contains(k) {
			return false
		}
	}
	return true
}

// Equal returns true if both sets contain the same elements.
func Equal[T comparable](a, b Set[T]) bool {
	return len(a) == len(b) && IsSubset(a, b)
}

// --- Functional operations on sets ---

// Find returns the first element (arbitrary order) satisfying the predicate.
func Find[T comparable](s Set[T], predicate func(T) bool) (T, bool) {
	for k := range s {
		if predicate(k) {
			return k, true
		}
	}
	var zero T
	return zero, false
}

// FindAll returns all elements satisfying the predicate.
func FindAll[T comparable](s Set[T], predicate func(T) bool) Set[T] {
	out := make(Set[T], len(s))
	for k := range s {
		if predicate(k) {
			out[k] = struct{}{}
		}
	}
	return out
}

// Partition splits the set into two based on the predicate.
func Partition[T comparable](s Set[T], predicate func(T) bool) (matched, unmatched Set[T]) {
	matched = make(Set[T], len(s))
	unmatched = make(Set[T], len(s))
	for k := range s {
		if predicate(k) {
			matched[k] = struct{}{}
		} else {
			unmatched[k] = struct{}{}
		}
	}
	return
}

// Map transforms each element and returns a new set.
// Duplicates (multiple source elements mapping to the same value) are collapsed.
func Map[T comparable, U comparable](s Set[T], fn func(T) U) Set[U] {
	out := make(Set[U], len(s))
	for k := range s {
		out[fn(k)] = struct{}{}
	}
	return out
}

// Reduce folds the set into a single value.
func Reduce[T comparable, U any](s Set[T], reducer func(acc U, elem T) U, initial U) U {
	acc := initial
	for k := range s {
		acc = reducer(acc, k)
	}
	return acc
}

// ForEach applies fn to each element.
func ForEach[T comparable](s Set[T], fn func(T)) {
	for k := range s {
		fn(k)
	}
}

// Any returns true if at least one element satisfies the predicate.
func Any[T comparable](s Set[T], predicate func(T) bool) bool {
	for k := range s {
		if predicate(k) {
			return true
		}
	}
	return false
}

// All returns true if every element satisfies the predicate.
// Returns true for an empty set (vacuous truth).
func All[T comparable](s Set[T], predicate func(T) bool) bool {
	for k := range s {
		if !predicate(k) {
			return false
		}
	}
	return true
}

// None returns true if no element satisfies the predicate.
func None[T comparable](s Set[T], predicate func(T) bool) bool {
	for k := range s {
		if predicate(k) {
			return false
		}
	}
	return true
}
