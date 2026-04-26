package collection

// Stack is a LIFO (last-in-first-out) data structure.
type Stack[T any] struct {
	data []T
}

// NewStack creates a stack with the default initial capacity (16).
func NewStack[T any]() *Stack[T] {
	return NewStackWithCap[T](16)
}

// NewStackWithCap creates a stack with the given initial capacity.
func NewStackWithCap[T any](cap int) *Stack[T] {
	if cap <= 0 {
		cap = 16
	}
	return &Stack[T]{data: make([]T, 0, cap)}
}

// Push adds an element to the top of the stack.
func (s *Stack[T]) Push(v T) {
	s.data = append(s.data, v)
}

// Pop removes and returns the top element.
// Returns (zero, false) if the stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.data) == 0 {
		var zero T
		return zero, false
	}
	n := len(s.data) - 1
	v := s.data[n]
	s.data[n] = *new(T) // allow GC
	s.data = s.data[:n]
	return v, true
}

// Peek returns the top element without removing it.
// Returns (zero, false) if the stack is empty.
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.data) == 0 {
		var zero T
		return zero, false
	}
	return s.data[len(s.data)-1], true
}

// Len returns the number of elements in the stack.
func (s *Stack[T]) Len() int { return len(s.data) }

// Size is an alias for Len.
func (s *Stack[T]) Size() int { return s.Len() }

// IsEmpty returns true if the stack has no elements.
func (s *Stack[T]) IsEmpty() bool { return len(s.data) == 0 }

// Clear removes all elements from the stack.
func (s *Stack[T]) Clear() {
	clear(s.data)
	s.data = s.data[:0]
}

// Values returns a copy of all elements (top is last in the slice).
func (s *Stack[T]) Values() []T {
	out := make([]T, len(s.data))
	copy(out, s.data)
	return out
}

// Clone returns a deep copy of the stack.
func (s *Stack[T]) Clone() *Stack[T] {
	return &Stack[T]{data: s.Values()}
}
