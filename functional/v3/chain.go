package fn

import (
	"cmp"
	"slices"
)

// Chain is a fluent builder over a slice. Each operation materializes a new
// slice, so a Chain is straightforward to read and debug and is typically as
// fast or faster than a stack of closures in Go.
//
// Construct with ChainOf; terminate with Collect (or ToSlice, an alias).
//
//	ChainOf(users).
//	    Filter(isActive).
//	    Map(func(u User) User { return u /* ... */ }).
//	    TopK(10, func(u User) int { return u.Score }).
//	    Collect()
//
// Type-changing maps use the free function MapTo, since Go method receivers
// cannot change the receiver's type parameter:
//
//	scores := MapTo(ChainOf(users), func(u User) int { return u.Score }).Collect()
type Chain[T any] struct {
	s []T
}

// ChainOf wraps s in a Chain. The wrapped slice shares the input's backing
// array; read-side operations are zero-copy, mutating operations copy on write.
func ChainOf[T any](s []T) Chain[T] {
	return Chain[T]{s: s}
}

// MapTo transforms each element of c into type U and returns a Chain[U]. It is
// the type-changing map for the fluent API, expressed as a free function
// because Go method receivers cannot return a Chain of a different type
// parameter.
func MapTo[T, U any](c Chain[T], f func(T) U) Chain[U] {
	return Chain[U]{s: Map(c.s, f)}
}

// SortChain sorts a Chain ascending when T is cmp.Ordered. It is a free
// function because Go methods cannot add type-parameter constraints to the
// receiver; use SortBy for a comparator-based sort on any T.
func SortChain[T cmp.Ordered](c Chain[T]) Chain[T] {
	out := append(make([]T, 0, len(c.s)), c.s...)
	slices.Sort(out)
	return Chain[T]{s: out}
}

// Len returns the number of elements in the chain.
func (c Chain[T]) Len() int { return len(c.s) }

// Slice returns the current backing slice (not a copy). Mutating it mutates
// the chain; use Collect for an independent copy.
func (c Chain[T]) Slice() []T { return c.s }

// Map transforms each element in place (T -> T). For T -> U use MapTo.
func (c Chain[T]) Map(f func(T) T) Chain[T] { return Chain[T]{s: Map(c.s, f)} }

// Filter keeps elements for which pred returns true.
func (c Chain[T]) Filter(pred func(T) bool) Chain[T] {
	return Chain[T]{s: Filter(c.s, pred)}
}

// FilterIdx keeps elements for which pred(index, value) returns true.
func (c Chain[T]) FilterIdx(pred func(int, T) bool) Chain[T] {
	return Chain[T]{s: FilterIdx(c.s, pred)}
}

// FlatMap maps each element to a slice and flattens the results (T -> []T).
func (c Chain[T]) FlatMap(f func(T) []T) Chain[T] {
	return Chain[T]{s: FlatMap(c.s, f)}
}

// Take returns a Chain holding at most the first n elements.
func (c Chain[T]) Take(n int) Chain[T] {
	if n < 0 {
		n = 0
	}
	if n > len(c.s) {
		n = len(c.s)
	}
	out := make([]T, n)
	copy(out, c.s[:n])
	return Chain[T]{s: out}
}

// Drop returns a Chain holding all but the first n elements.
func (c Chain[T]) Drop(n int) Chain[T] {
	if n < 0 {
		n = 0
	}
	if n > len(c.s) {
		n = len(c.s)
	}
	out := make([]T, len(c.s)-n)
	copy(out, c.s[n:])
	return Chain[T]{s: out}
}

// DeduplicateByChain returns a Chain with elements whose key (under keyFn) is
// duplicated removed, preserving first occurrence. This is a free function
// because Go methods cannot declare type parameters; the comparable constraint
// lives on K here.
func DeduplicateByChain[T any, K comparable](c Chain[T], keyFn func(T) K) Chain[T] {
	return Chain[T]{s: DeduplicateBy(c.s, keyFn)}
}

// Reverse returns a Chain with the elements in reverse order.
func (c Chain[T]) Reverse() Chain[T] { return Chain[T]{s: Reverse(c.s)} }

// SortBy sorts the chain using a cmp-style comparator (negative when a < b).
func (c Chain[T]) SortBy(less By[T]) Chain[T] {
	out := append(make([]T, 0, len(c.s)), c.s...)
	slices.SortFunc(out, less)
	return Chain[T]{s: out}
}

// TopK keeps the k largest elements (by integer key) in descending order. For
// non-integer keys use the standalone TopK function.
func (c Chain[T]) TopK(k int, key func(T) int) Chain[T] {
	return Chain[T]{s: TopK(c.s, k, Key[T, int](key))}
}

// Concat appends the elements of other to a copy of this chain.
func (c Chain[T]) Concat(other []T) Chain[T] {
	return Chain[T]{s: Concat(c.s, other)}
}

// Collect returns a copy of the chain's slice.
func (c Chain[T]) Collect() []T {
	out := make([]T, len(c.s))
	copy(out, c.s)
	return out
}

// ToSlice is an alias for Collect.
func (c Chain[T]) ToSlice() []T { return c.Collect() }

// ForEach calls fn on each element of the chain.
func (c Chain[T]) ForEach(fn func(T)) { ForEach(c.s, fn) }

// Reduce folds the chain into a single value.
func (c Chain[T]) Reduce(reducer func(acc T, elem T) T, initial T) T {
	return Reduce(c.s, reducer, initial)
}

// Any returns true if any element satisfies pred (short-circuit).
func (c Chain[T]) Any(pred func(T) bool) bool { return Any(c.s, pred) }

// All returns true if all elements satisfy pred (short-circuit).
func (c Chain[T]) All(pred func(T) bool) bool { return All(c.s, pred) }

// Find returns the first element satisfying pred, or (zero, false).
func (c Chain[T]) Find(pred func(T) bool) (T, bool) { return Find(c.s, pred) }
