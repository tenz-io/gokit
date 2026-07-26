package collection

import "cmp"

// Heap is a binary-heap priority queue.
//
// The ordering is defined by a less function: less(a, b) == true means a has
// higher priority than b and therefore leaves the heap first. The common
// cases — min-heap and max-heap over an ordered type — have dedicated
// constructors [NewMinHeap] and [NewMaxHeap] so callers rarely hand-write a
// less func. For custom orderings (multi-field, computed priority) use
// [NewHeap] with an explicit less.
//
// Push and Pop are O(log n); Peek and Len are O(1).
type Heap[T any] struct {
	data []T
	less func(a, b T) bool
}

// NewHeap creates an empty heap ordered by less with the default capacity.
// less(a, b) == true means a has higher priority than b.
func NewHeap[T any](less func(a, b T) bool) *Heap[T] {
	return NewHeapWithCap[T](defaultCap, less)
}

// NewHeapWithCap creates an empty heap ordered by less, pre-sized for cap
// elements. A non-positive cap falls back to the default capacity.
func NewHeapWithCap[T any](cap int, less func(a, b T) bool) *Heap[T] {
	if cap <= 0 {
		cap = defaultCap
	}
	return &Heap[T]{data: make([]T, 0, cap), less: less}
}

// NewMinHeap creates an empty min-heap over an ordered type: the smallest
// element has the highest priority and leaves first.
func NewMinHeap[T cmp.Ordered]() *Heap[T] {
	return NewHeap[T](func(a, b T) bool { return cmp.Compare(a, b) < 0 })
}

// NewMaxHeap creates an empty max-heap over an ordered type: the largest
// element has the highest priority and leaves first.
func NewMaxHeap[T cmp.Ordered]() *Heap[T] {
	return NewHeap[T](func(a, b T) bool { return cmp.Compare(a, b) > 0 })
}

// Push adds v to the heap. O(log n).
func (h *Heap[T]) Push(v T) {
	h.data = append(h.data, v)
	h.siftUp(len(h.data) - 1)
}

// Pop removes and returns the highest-priority element. It returns (zero,
// false) when the heap is empty. O(log n).
func (h *Heap[T]) Pop() (T, bool) {
	var zero T
	if len(h.data) == 0 {
		return zero, false
	}
	top := h.data[0]
	n := len(h.data) - 1
	h.data[0] = h.data[n]
	h.data[n] = zero // let the GC reclaim the moved reference
	h.data = h.data[:n]
	if n > 0 {
		h.siftDown(0)
	}
	return top, true
}

// Peek returns the highest-priority element without removing it. O(1). It
// returns (zero, false) when the heap is empty.
func (h *Heap[T]) Peek() (T, bool) {
	var zero T
	if len(h.data) == 0 {
		return zero, false
	}
	return h.data[0], true
}

// Len returns the number of elements.
func (h *Heap[T]) Len() int { return len(h.data) }

// IsEmpty reports whether the heap has no elements.
func (h *Heap[T]) IsEmpty() bool { return len(h.data) == 0 }

// Clear removes all elements and zeroes the backing array so the GC can
// reclaim held references. The capacity is retained for reuse.
func (h *Heap[T]) Clear() {
	var zero T
	for i := range h.data {
		h.data[i] = zero
	}
	h.data = h.data[:0]
}

// Values returns a copy of the underlying heap slice. The order is the heap's
// internal layout, NOT priority order — to drain in priority order, call Pop
// until empty. The returned slice is independent.
func (h *Heap[T]) Values() []T {
	out := make([]T, len(h.data))
	copy(out, h.data)
	return out
}

// Clone returns an independent copy of the heap sharing the same ordering.
func (h *Heap[T]) Clone() *Heap[T] {
	out := make([]T, len(h.data))
	copy(out, h.data)
	return &Heap[T]{data: out, less: h.less}
}

// siftUp moves the element at index i up the heap until the heap property is
// restored. O(log n).
func (h *Heap[T]) siftUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !h.less(h.data[i], h.data[parent]) {
			break
		}
		h.data[i], h.data[parent] = h.data[parent], h.data[i]
		i = parent
	}
}

// siftDown moves the element at index i down the heap until the heap property
// is restored. O(log n).
func (h *Heap[T]) siftDown(i int) {
	n := len(h.data)
	for {
		left := 2*i + 1
		right := 2*i + 2
		best := i
		if left < n && h.less(h.data[left], h.data[best]) {
			best = left
		}
		if right < n && h.less(h.data[right], h.data[best]) {
			best = right
		}
		if best == i {
			break
		}
		h.data[i], h.data[best] = h.data[best], h.data[i]
		i = best
	}
}
