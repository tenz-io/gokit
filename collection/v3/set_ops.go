package collection

// This file provides free-function entry points for the set algebra and
// functional operations. In Go a method and a package-level function cannot
// share a name, so the algebra free-functions take an *Of suffix
// ([UnionOf], [IntersectOf], ...) while the chaining methods stay short
// ([Set.Union], [Set.Intersect], ...). Use whichever reads best at the call
// site:
//
//	collection.UnionOf(a, b).Intersect(c)        // functional start, then chain
//	a.Union(b).Intersect(c).Subtract(d)          // fully chained

// UnionOf returns a new set with all elements from a and b (a ∪ b).
func UnionOf[T comparable](a, b Set[T]) Set[T] { return a.Union(b) }

// IntersectOf returns a new set with elements present in both a and b
// (a ∩ b).
func IntersectOf[T comparable](a, b Set[T]) Set[T] { return a.Intersect(b) }

// Difference returns a new set with elements in a but not in b (a \ b). It is
// the free-function form of [Set.Subtract]; the v2 name is preserved for
// callers migrating their call sites.
func Difference[T comparable](a, b Set[T]) Set[T] { return a.Subtract(b) }

// SymmetricDifference returns a new set with elements in exactly one of a or
// b (a △ b).
func SymmetricDifference[T comparable](a, b Set[T]) Set[T] {
	return a.SymmetricDifference(b)
}

// IsSubset reports whether every element of a is in b (a ⊆ b).
func IsSubset[T comparable](a, b Set[T]) bool { return a.IsSubset(b) }

// IsSuperset reports whether every element of b is in a (a ⊇ b).
func IsSuperset[T comparable](a, b Set[T]) bool { return a.IsSuperset(b) }

// IsDisjoint reports whether a and b share no elements (a ∩ b = ∅).
func IsDisjoint[T comparable](a, b Set[T]) bool { return a.IsDisjoint(b) }

// Equal reports whether a and b contain the same elements.
func Equal[T comparable](a, b Set[T]) bool { return a.Equal(b) }

// Clone returns an independent copy of s. It is the free-function form of
// [Set.Clone].
func Clone[T comparable](s Set[T]) Set[T] { return s.Clone() }

// --- Functional operations ---

// Find returns the first element (in map iteration order, which is
// non-deterministic) satisfying predicate, or (zero, false).
func Find[T comparable](s Set[T], predicate func(T) bool) (T, bool) {
	for k := range s.m {
		if predicate(k) {
			return k, true
		}
	}
	var zero T
	return zero, false
}

// FindAll returns a new set with every element satisfying predicate.
func FindAll[T comparable](s Set[T], predicate func(T) bool) Set[T] {
	out := make(map[T]struct{}, len(s.m))
	for k := range s.m {
		if predicate(k) {
			out[k] = struct{}{}
		}
	}
	return Set[T]{m: out}
}

// Partition splits s into two sets: matched holds elements for which predicate
// is true, unmatched holds the rest.
func Partition[T comparable](s Set[T], predicate func(T) bool) (matched, unmatched Set[T]) {
	matched = Set[T]{m: make(map[T]struct{}, len(s.m))}
	unmatched = Set[T]{m: make(map[T]struct{}, len(s.m))}
	for k := range s.m {
		if predicate(k) {
			matched.m[k] = struct{}{}
		} else {
			unmatched.m[k] = struct{}{}
		}
	}
	return
}

// Map transforms each element of s via fn and returns a new set. Multiple
// source elements mapping to the same value collapse into one.
func Map[T comparable, U comparable](s Set[T], fn func(T) U) Set[U] {
	out := make(map[U]struct{}, len(s.m))
	for k := range s.m {
		out[fn(k)] = struct{}{}
	}
	return Set[U]{m: out}
}

// Reduce folds s into a single value, starting from initial.
func Reduce[T comparable, U any](s Set[T], reducer func(acc U, elem T) U, initial U) U {
	acc := initial
	for k := range s.m {
		acc = reducer(acc, k)
	}
	return acc
}

// ForEach calls fn on each element. The visit order is non-deterministic.
func ForEach[T comparable](s Set[T], fn func(T)) {
	for k := range s.m {
		fn(k)
	}
}

// Any reports whether at least one element satisfies predicate.
func Any[T comparable](s Set[T], predicate func(T) bool) bool {
	for k := range s.m {
		if predicate(k) {
			return true
		}
	}
	return false
}

// All reports whether every element satisfies predicate. An empty set returns
// true (vacuous truth).
func All[T comparable](s Set[T], predicate func(T) bool) bool {
	for k := range s.m {
		if !predicate(k) {
			return false
		}
	}
	return true
}

// None reports whether no element satisfies predicate.
func None[T comparable](s Set[T], predicate func(T) bool) bool {
	for k := range s.m {
		if predicate(k) {
			return false
		}
	}
	return true
}
