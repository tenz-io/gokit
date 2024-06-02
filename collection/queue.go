package collection

import "fmt"

// Queue is a queue data structure
type Queue[T any] struct {
	data []T
}

// NewQueue creates a new queue
func NewQueue[T any](initSize int) *Queue[T] {
	if initSize <= 0 {
		initSize = 16
	}
	return &Queue[T]{
		data: make([]T, 0, initSize),
	}
}

// Enqueue adds an element to the back of the queue
func (q *Queue[T]) Enqueue(a T) {
	q.data = append(q.data, a)
}

// Dequeue removes an element from the front of the queue
// if the queue is empty, return false, otherwise return true
// and the element
func (q *Queue[T]) Dequeue() (T, bool) {
	if len(q.data) == 0 {
		var zero T
		return zero, false
	}
	a := q.data[0]
	q.data = q.data[1:]
	return a, true
}

// Peek returns the element at the front of the queue
// without removing it
func (q *Queue[T]) Peek() (T, bool) {
	if len(q.data) == 0 {
		var zero T
		return zero, false
	}
	return q.data[0], true
}

// Size returns the size of the queue
func (q *Queue[T]) Size() int {
	return len(q.data)
}

// IsEmpty checks if the queue is empty
func (q *Queue[T]) IsEmpty() bool {
	return len(q.data) == 0
}

// Clear clears the queue
func (q *Queue[T]) Clear() {
	q.data = q.data[:0]
}

// Values returns the values of the queue
func (q *Queue[T]) Values() []T {
	values := make([]T, len(q.data))
	copy(values, q.data)
	return values
}

// Clone returns a new queue with the same elements
func (q *Queue[T]) Clone() *Queue[T] {
	return &Queue[T]{
		data: q.Values(),
	}
}

// String returns the string representation of the queue
func (q *Queue[T]) String() string {
	return fmt.Sprintf("%v", q.data)
}
