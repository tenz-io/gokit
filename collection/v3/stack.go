package collection

const defaultCap = 16

// Stack is a LIFO (last-in-first-out) stack backed by a slice.
//
// All mutating methods take a pointer receiver so the underlying slice header
// (length, capacity, pointer) is shared; a value copy would alias the backing
// array. Clone a stack with [Stack.Clone] before mutating if you need to
// diverge.
type Stack[T any] struct {
	data []T
}

// NewStack creates an empty stack with the default initial capacity.
func NewStack[T any]() *Stack[T] {
	return NewStackWithCap[T](defaultCap)
}

// NewStackWithCap creates an empty stack pre-sized for cap elements. A non-positive
// cap falls back to the default capacity so callers can pass through a zero
// value without conditionals.
func NewStackWithCap[T any](cap int) *Stack[T] {
	if cap <= 0 {
		cap = defaultCap
	}
	return &Stack[T]{data: make([]T, 0, cap)}
}

// Push adds v to the top of the stack. Amortized O(1).
func (s *Stack[T]) Push(v T) {
	s.data = append(s.data, v)
}

// Pop removes and returns the top element. It returns (zero, false) when the
// stack is empty. The popped slot is zeroed so the GC can reclaim any
// reference it held.
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.data) == 0 {
		return zero, false
	}
	n := len(s.data) - 1
	v := s.data[n]
	s.data[n] = zero
	s.data = s.data[:n]
	return v, true
}

// Peek returns the top element without removing it. It returns (zero, false)
// when the stack is empty.
func (s *Stack[T]) Peek() (T, bool) {
	var zero T
	if len(s.data) == 0 {
		return zero, false
	}
	return s.data[len(s.data)-1], true
}

// Len returns the number of elements.
func (s *Stack[T]) Len() int { return len(s.data) }

// IsEmpty reports whether the stack has no elements.
func (s *Stack[T]) IsEmpty() bool { return len(s.data) == 0 }

// Clear removes all elements and zeroes the backing array up to its length so
// the GC can reclaim held references. The capacity is retained for reuse.
func (s *Stack[T]) Clear() {
	var zero T
	for i := range s.data {
		s.data[i] = zero
	}
	s.data = s.data[:0]
}

// Values returns a copy of the elements ordered bottom-to-top (the first
// pushed element is first in the slice). The returned slice is independent;
// mutating it does not affect the stack.
func (s *Stack[T]) Values() []T {
	out := make([]T, len(s.data))
	copy(out, s.data)
	return out
}

// Clone returns an independent copy of the stack.
func (s *Stack[T]) Clone() *Stack[T] {
	return &Stack[T]{data: s.Values()}
}
