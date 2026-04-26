package collection

// PriorityQueue is a min-heap based priority queue.
// The less function defines the priority: less(a, b) == true means a has higher priority than b.
type PriorityQueue[T any] struct {
	data []T
	less func(a, b T) bool
}

// NewPriorityQueue creates a priority queue with the default capacity (16).
func NewPriorityQueue[T any](less func(a, b T) bool) *PriorityQueue[T] {
	return NewPriorityQueueWithCap(16, less)
}

// NewPriorityQueueWithCap creates a priority queue with the given initial capacity.
func NewPriorityQueueWithCap[T any](cap int, less func(a, b T) bool) *PriorityQueue[T] {
	if cap <= 0 {
		cap = 16
	}
	return &PriorityQueue[T]{
		data: make([]T, 0, cap),
		less: less,
	}
}

// Push adds an element. O(log n).
func (pq *PriorityQueue[T]) Push(v T) {
	pq.data = append(pq.data, v)
	pq.bubbleUp(len(pq.data) - 1)
}

// Pop removes and returns the highest-priority element. O(log n).
// Returns (zero, false) if the queue is empty.
func (pq *PriorityQueue[T]) Pop() (T, bool) {
	if len(pq.data) == 0 {
		var zero T
		return zero, false
	}
	top := pq.data[0]
	n := len(pq.data) - 1
	pq.data[0] = pq.data[n]
	pq.data[n] = *new(T) // allow GC
	pq.data = pq.data[:n]
	if n > 0 {
		pq.bubbleDown(0)
	}
	return top, true
}

// Peek returns the highest-priority element without removing it. O(1).
// Returns (zero, false) if the queue is empty.
func (pq *PriorityQueue[T]) Peek() (T, bool) {
	if len(pq.data) == 0 {
		var zero T
		return zero, false
	}
	return pq.data[0], true
}

// Len returns the number of elements.
func (pq *PriorityQueue[T]) Len() int { return len(pq.data) }

// Size is an alias for Len.
func (pq *PriorityQueue[T]) Size() int { return pq.Len() }

// IsEmpty returns true if the queue has no elements.
func (pq *PriorityQueue[T]) IsEmpty() bool { return len(pq.data) == 0 }

// Clear removes all elements.
func (pq *PriorityQueue[T]) Clear() {
	clear(pq.data)
	pq.data = pq.data[:0]
}

// Values returns a copy of the underlying heap slice (no ordering guarantee).
func (pq *PriorityQueue[T]) Values() []T {
	out := make([]T, len(pq.data))
	copy(out, pq.data)
	return out
}

// Clone returns a deep copy of the priority queue.
func (pq *PriorityQueue[T]) Clone() *PriorityQueue[T] {
	return &PriorityQueue[T]{
		data: pq.Values(),
		less: pq.less,
	}
}

func (pq *PriorityQueue[T]) bubbleUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !pq.less(pq.data[i], pq.data[parent]) {
			break
		}
		pq.data[i], pq.data[parent] = pq.data[parent], pq.data[i]
		i = parent
	}
}

func (pq *PriorityQueue[T]) bubbleDown(i int) {
	for {
		left := 2*i + 1
		right := 2*i + 2
		minIdx := i

		if left < len(pq.data) && pq.less(pq.data[left], pq.data[minIdx]) {
			minIdx = left
		}
		if right < len(pq.data) && pq.less(pq.data[right], pq.data[minIdx]) {
			minIdx = right
		}
		if minIdx == i {
			break
		}
		pq.data[i], pq.data[minIdx] = pq.data[minIdx], pq.data[i]
		i = minIdx
	}
}
