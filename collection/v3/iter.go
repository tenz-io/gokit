package collection

import "iter"

// All returns an [iter.Seq] over the stack's elements from bottom (first
// pushed) to top (last pushed). It is safe to break out of the range early.
//
//	for v := range s.All() { ... }
func (s *Stack[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s.data {
			if !yield(v) {
				return
			}
		}
	}
}

// All returns an [iter.Seq] over the queue's elements from front (oldest) to
// back (newest). It is safe to break out of the range early.
func (q *Queue[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		n := len(q.buf)
		for i := 0; i < q.count; i++ {
			if !yield(q.buf[(q.head+i)%n]) {
				return
			}
		}
	}
}

// All returns an [iter.Seq] over the heap's elements. The order is the
// internal heap layout, NOT priority order — to drain in priority order call
// [Heap.Pop] until empty. It is safe to break out of the range early.
func (h *Heap[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range h.data {
			if !yield(v) {
				return
			}
		}
	}
}

// All returns an [iter.Seq] over the set's elements. The order is
// non-deterministic (map iteration order). It is safe to break out of the
// range early.
func (s Set[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for k := range s.m {
			if !yield(k) {
				return
			}
		}
	}
}
