package fn

// Seq is a lazy, callback-style iterator over a sequence of T. It is the
// pull-free "push" model: a consumer passes a yield callback, and the
// producer pushes each element to it. Returning false from yield stops
// iteration early (short-circuit).
//
// Seq is the zero-allocation complement to Chain: where Chain materializes a
// slice per step, Seq fuses operations into a single callback chain and can
// stop as soon as a result is known (Any/All/Find). For building a result
// slice, Chain/Collect is usually clearer.
//
// Construct with SeqOf (from a slice). Filter and Map compose lazily.
type Seq[T any] func(yield func(T) bool)

// SeqOf creates a Seq that yields the elements of s in order. The slice is
// referenced (not copied); do not mutate s while iterating.
func SeqOf[T any](s []T) Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// Filter returns a Seq that yields only elements of q for which pred holds.
// It is lazy: pred is evaluated only for elements actually consumed.
func (q Seq[T]) Filter(pred func(T) bool) Seq[T] {
	return func(yield func(T) bool) {
		q(func(v T) bool {
			if !pred(v) {
				return true // skip, keep iterating
			}
			return yield(v)
		})
	}
}

// MapSeq returns a Seq[U] that yields f(v) for each v produced by q. It is
// the type-changing lazy map, expressed as a free function since a method
// cannot change the receiver's type parameter.
func MapSeq[T, U any](q Seq[T], f func(T) U) Seq[U] {
	return func(yield func(U) bool) {
		q(func(v T) bool {
			return yield(f(v))
		})
	}
}

// ForEach calls fn on every element of q. It does not support early
// termination; for short-circuit consumption use Any/All/Find.
func (q Seq[T]) ForEach(fn func(T)) {
	q(func(v T) bool {
		fn(v)
		return true
	})
}

// Count returns the number of elements produced by q. It consumes q fully.
func (q Seq[T]) Count() int {
	n := 0
	q(func(v T) bool {
		n++
		return true
	})
	return n
}

// Any returns true if any element of q satisfies pred. It short-circuits,
// stopping at the first match.
func (q Seq[T]) Any(pred func(T) bool) bool {
	found := false
	q(func(v T) bool {
		if pred(v) {
			found = true
			return false // stop
		}
		return true
	})
	return found
}

// All returns true if every element of q satisfies pred. It short-circuits,
// stopping at the first non-match.
func (q Seq[T]) All(pred func(T) bool) bool {
	ok := true
	q(func(v T) bool {
		if !pred(v) {
			ok = false
			return false // stop
		}
		return true
	})
	return ok
}

// Find returns the first element of q satisfying pred, or (zero, false). It
// short-circuits at the first match.
func (q Seq[T]) Find(pred func(T) bool) (T, bool) {
	var result T
	found := false
	q(func(v T) bool {
		if pred(v) {
			result = v
			found = true
			return false // stop
		}
		return true
	})
	return result, found
}

// First returns the first element of q, or (zero, false) if q is empty.
func (q Seq[T]) First() (T, bool) {
	var result T
	found := false
	q(func(v T) bool {
		result = v
		found = true
		return false // stop after first
	})
	return result, found
}

// Collect materializes q into a slice. Because Seq is push-based, the result
// length is not known up front; elements are appended.
func (q Seq[T]) Collect() []T {
	out := make([]T, 0)
	q(func(v T) bool {
		out = append(out, v)
		return true
	})
	return out
}
