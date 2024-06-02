package collection

import "fmt"

// PriorityQueue represents a priority queue.
type PriorityQueue[T any] struct {
	data []T
	less func(a, b T) bool
}

// NewPriorityQueue creates a new priority queue with capacity and a less function.
// The less function should return true if a is less than b.
// The priority queue is a min heap.
func NewPriorityQueue[T any](initSize int, less func(a, b T) bool) *PriorityQueue[T] {
	if initSize <= 0 {
		initSize = 16
	}
	return &PriorityQueue[T]{
		data: make([]T, 0, initSize),
		less: less,
	}
}

// Push adds an element to the priority queue.
// The time complexity is O(log n).
func (pq *PriorityQueue[T]) Push(a T) {
	pq.data = append(pq.data, a)
	pq.bubbleUp()
}

// Pop removes the element with the highest priority from the priority queue.
// If the priority queue is empty, return false. Otherwise, return true.
// The time complexity is O(log n).
func (pq *PriorityQueue[T]) Pop() (T, bool) {
	if len(pq.data) == 0 {
		var zero T
		return zero, false
	}
	top := pq.data[0]
	pq.data[0] = pq.data[len(pq.data)-1]
	pq.data = pq.data[:len(pq.data)-1]
	pq.bubbleDown()
	return top, true
}

func (pq *PriorityQueue[T]) bubbleUp() {
	idx := len(pq.data) - 1
	for idx > 0 {
		parent := (idx - 1) / 2
		if pq.less(pq.data[idx], pq.data[parent]) {
			pq.data[idx], pq.data[parent] = pq.data[parent], pq.data[idx]
			idx = parent
		} else {
			break
		}
	}
}

func (pq *PriorityQueue[T]) bubbleDown() {
	idx := 0
	for {
		left := 2*idx + 1
		right := 2*idx + 2
		minIdx := idx

		if left < len(pq.data) && pq.less(pq.data[left], pq.data[minIdx]) {
			minIdx = left
		}
		if right < len(pq.data) && pq.less(pq.data[right], pq.data[minIdx]) {
			minIdx = right
		}

		if minIdx != idx {
			pq.data[idx], pq.data[minIdx] = pq.data[minIdx], pq.data[idx]
			idx = minIdx
		} else {
			break
		}
	}
}

// Peek returns the element with the highest priority from the priority queue.
// If the priority queue is empty, return false. Otherwise, return true.
// The time complexity is O(1).
func (pq *PriorityQueue[T]) Peek() (T, bool) {
	if len(pq.data) == 0 {
		var zero T
		return zero, false
	}
	return pq.data[0], true
}

// Size returns the size of the priority queue.
func (pq *PriorityQueue[T]) Size() int {
	return len(pq.data)
}

// IsEmpty checks if the priority queue is empty.
func (pq *PriorityQueue[T]) IsEmpty() bool {
	return len(pq.data) == 0
}

// Clear clears the priority queue.
func (pq *PriorityQueue[T]) Clear() {
	pq.data = pq.data[:0]
}

// Values returns the values of the priority queue.
func (pq *PriorityQueue[T]) Values() []T {
	values := make([]T, len(pq.data))
	copy(values, pq.data)
	return values
}

// Clone returns a new priority queue with the same elements.
func (pq *PriorityQueue[T]) Clone() *PriorityQueue[T] {
	return &PriorityQueue[T]{
		data: pq.Values(),
		less: pq.less,
	}
}

// String returns the string representation of the priority queue.
func (pq *PriorityQueue[T]) String() string {
	return fmt.Sprintf("%v", pq.data)
}
