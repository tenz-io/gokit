package collection

// Queue is a FIFO (first-in-first-out) queue backed by a ring buffer.
//
// v2 backed the queue with a slice and Dequeued via `data = data[1:]`, which
// advances the window over the same backing array forever — dequeued slots
// are never reclaimed and the array only grows. The v3 ring buffer recycles
// slots as they are dequeued, so Enqueue/Dequeue are amortized O(1) and
// memory stays bounded over the queue's lifetime. When the buffer fills, it
// grows to 2× and linearizes the wrapped contents back to the front.
type Queue[T any] struct {
	buf   []T
	head  int // index of the front element
	tail  int // index where the next Enqueue writes
	count int // number of elements
}

// NewQueue creates an empty queue with the default initial capacity.
func NewQueue[T any]() *Queue[T] {
	return NewQueueWithCap[T](defaultCap)
}

// NewQueueWithCap creates an empty queue pre-sized for cap elements. A
// non-positive cap falls back to the default capacity.
func NewQueueWithCap[T any](cap int) *Queue[T] {
	if cap <= 0 {
		cap = defaultCap
	}
	return &Queue[T]{buf: make([]T, cap)}
}

// Enqueue adds v to the back of the queue. Amortized O(1): the buffer doubles
// and re-linearizes when it fills.
func (q *Queue[T]) Enqueue(v T) {
	if q.count == len(q.buf) {
		q.grow()
	}
	q.buf[q.tail] = v
	q.tail = (q.tail + 1) % len(q.buf)
	q.count++
}

// Dequeue removes and returns the front element. It returns (zero, false)
// when the queue is empty. The dequeued slot is zeroed so the GC can reclaim
// any reference it held — this is the bug v2's slice[1:] approach never
// addressed.
func (q *Queue[T]) Dequeue() (T, bool) {
	var zero T
	if q.count == 0 {
		return zero, false
	}
	v := q.buf[q.head]
	q.buf[q.head] = zero
	q.head = (q.head + 1) % len(q.buf)
	q.count--
	return v, true
}

// Peek returns the front element without removing it. It returns (zero,
// false) when the queue is empty.
func (q *Queue[T]) Peek() (T, bool) {
	var zero T
	if q.count == 0 {
		return zero, false
	}
	return q.buf[q.head], true
}

// Len returns the number of elements.
func (q *Queue[T]) Len() int { return q.count }

// IsEmpty reports whether the queue has no elements.
func (q *Queue[T]) IsEmpty() bool { return q.count == 0 }

// Clear removes all elements, zeroes the buffer, and resets the head/tail
// pointers. The buffer capacity is retained for reuse.
func (q *Queue[T]) Clear() {
	var zero T
	for i := range q.buf {
		q.buf[i] = zero
	}
	q.head, q.tail, q.count = 0, 0, 0
}

// Values returns a copy of the elements ordered front-to-back. The returned
// slice is independent; mutating it does not affect the queue.
func (q *Queue[T]) Values() []T {
	out := make([]T, q.count)
	for i := 0; i < q.count; i++ {
		out[i] = q.buf[(q.head+i)%len(q.buf)]
	}
	return out
}

// Clone returns an independent copy of the queue. The clone is laid out
// linearly (head=0) even if the source wraps around.
func (q *Queue[T]) Clone() *Queue[T] {
	cap := max(q.count, defaultCap)
	out := make([]T, cap, cap)
	copy(out, q.Values())
	return &Queue[T]{buf: out, head: 0, tail: q.count, count: q.count}
}

// grow doubles the buffer and linearizes the wrapped contents so head=0 and
// tail=count. Called only when the buffer is full.
func (q *Queue[T]) grow() {
	newCap := max(len(q.buf)*2, defaultCap)
	next := make([]T, newCap, newCap)
	// Copy the logical contents (head..head+count) into the front of next.
	// Two copies cover the wrap-around without branching in the hot path.
	n := copy(next, q.buf[q.head:])
	copy(next[n:], q.buf[:q.count-n])
	q.buf = next
	q.head = 0
	q.tail = q.count
}
