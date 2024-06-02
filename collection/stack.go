package collection

import "fmt"

// Stack is a stack data structure
type Stack[T any] struct {
	data []T
}

// NewStack creates a new stack
func NewStack[T any](initSize int) *Stack[T] {
	if initSize <= 0 {
		initSize = 16
	}
	return &Stack[T]{
		data: make([]T, 0, initSize),
	}
}

// Push adds an element to the top of the stack
func (s *Stack[T]) Push(a T) {
	s.data = append(s.data, a)
}

// Pop removes an element from the top of the stack
// if the stack is empty, return an error
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.data) == 0 {
		var zero T
		return zero, false
	}
	a := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return a, true
}

// Peek returns the element at the top of the stack
// without removing it
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.data) == 0 {
		var zero T
		return zero, false
	}
	return s.data[len(s.data)-1], true
}

// Size returns the size of the stack
func (s *Stack[T]) Size() int {
	return len(s.data)
}

// IsEmpty checks if the stack is empty
func (s *Stack[T]) IsEmpty() bool {
	return len(s.data) == 0
}

// Clear clears the stack
func (s *Stack[T]) Clear() {
	s.data = s.data[:0]
}

// Values returns the values of the stack
func (s *Stack[T]) Values() []T {
	values := make([]T, len(s.data))
	copy(values, s.data)
	return values
}

// String returns the string representation of the stack
func (s *Stack[T]) String() string {
	return fmt.Sprintf("%v", s.data)
}

// Clone returns a copy of the stack
func (s *Stack[T]) Clone() *Stack[T] {
	return &Stack[T]{
		data: s.Values(),
	}
}
