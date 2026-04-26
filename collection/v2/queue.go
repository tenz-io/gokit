package collection

// Queue is a FIFO (first-in-first-out) data structure.
type Queue[T any] struct {
	data []T
}

// NewQueue creates a queue with the default initial capacity (16).
func NewQueue[T any]() *Queue[T] {
	return NewQueueWithCap[T](16)
}

// NewQueueWithCap creates a queue with the given initial capacity.
func NewQueueWithCap[T any](cap int) *Queue[T] {
	if cap <= 0 {
		cap = 16
	}
	return &Queue[T]{data: make([]T, 0, cap)}
}

// Enqueue adds an element to the back of the queue.
func (q *Queue[T]) Enqueue(v T) {
	q.data = append(q.data, v)
}

// Dequeue removes and returns the front element.
// Returns (zero, false) if the queue is empty.
func (q *Queue[T]) Dequeue() (T, bool) {
	if len(q.data) == 0 {
		var zero T
		return zero, false
	}
	v := q.data[0]
	q.data[0] = *new(T) // allow GC of the dequeued element
	q.data = q.data[1:]
	return v, true
}

// Peek returns the front element without removing it.
// Returns (zero, false) if the queue is empty.
func (q *Queue[T]) Peek() (T, bool) {
	if len(q.data) == 0 {
		var zero T
		return zero, false
	}
	return q.data[0], true
}

// Len returns the number of elements in the queue.
func (q *Queue[T]) Len() int { return len(q.data) }

// Size is an alias for Len.
func (q *Queue[T]) Size() int { return q.Len() }

// IsEmpty returns true if the queue has no elements.
func (q *Queue[T]) IsEmpty() bool { return len(q.data) == 0 }

// Clear removes all elements from the queue.
func (q *Queue[T]) Clear() {
	clear(q.data)
	q.data = q.data[:0]
}

// Values returns a copy of all elements (front is first in the slice).
func (q *Queue[T]) Values() []T {
	out := make([]T, len(q.data))
	copy(out, q.data)
	return out
}

// Clone returns a deep copy of the queue.
func (q *Queue[T]) Clone() *Queue[T] {
	return &Queue[T]{data: q.Values()}
}
