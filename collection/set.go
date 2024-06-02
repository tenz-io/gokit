package collection

// Set is a set data structure
type Set[T comparable] map[T]struct{}

// NewSet creates a new set
func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

// Add adds an element to the set
func (s Set[T]) Add(a T) {
	s[a] = struct{}{}
}

// Remove removes an element from the set
func (s Set[T]) Remove(a T) {
	delete(s, a)
}

// Contains checks if the set contains an element
func (s Set[T]) Contains(a T) bool {
	_, ok := s[a]
	return ok
}

// Size returns the size of the set
func (s Set[T]) Size() int {
	return len(s)
}

// Clear clears the set
func (s Set[T]) Clear() {
	for k := range s {
		delete(s, k)
	}
}

// Values returns the values of the set
func (s Set[T]) Values() []T {
	values := make([]T, 0, len(s))
	for k := range s {
		values = append(values, k)
	}
	return values
}

// Intersection returns the intersection of two sets;
// elements in both a and b
//
// a ∩ b
func Intersection[T comparable](a, b Set[T]) Set[T] {
	result := NewSet[T]()
	for k := range a {
		if b.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

// Union returns the union of two sets;
// all elements in a and b
//
// a ∪ b
func Union[T comparable](a, b Set[T]) Set[T] {
	result := NewSet[T]()
	for k := range a {
		result.Add(k)
	}
	for k := range b {
		result.Add(k)
	}
	return result
}

// Difference returns the difference of two sets;
// elements in a but not in b
//
// a - b
func Difference[T comparable](a, b Set[T]) Set[T] {
	result := NewSet[T]()
	for k := range a {
		if !b.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

// SymmetricDifference returns the symmetric difference of two sets;
// elements in a or b but not in both
//
// a ∆ b
func SymmetricDifference[T comparable](a, b Set[T]) Set[T] {
	result := NewSet[T]()
	for k := range a {
		if !b.Contains(k) {
			result.Add(k)
		}
	}
	for k := range b {
		if !a.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

// IsSubset checks if a is a subset of b
//
// a ⊆ b
func IsSubset[T comparable](a, b Set[T]) bool {
	for k := range a {
		if !b.Contains(k) {
			return false
		}
	}
	return true
}

// IsSuperset checks if a is a superset of b
//
// a ⊇ b
func IsSuperset[T comparable](a, b Set[T]) bool {
	return IsSubset(b, a)
}

// IsDisjoint checks if a and b are disjoint
//
// a ∩ b = ∅
func IsDisjoint[T comparable](a, b Set[T]) bool {
	for k := range a {
		if b.Contains(k) {
			return false
		}
	}
	return true
}

// Equal checks if a and b are equal
//
// a = b
func Equal[T comparable](a, b Set[T]) bool {
	return IsSubset(a, b) && IsSuperset(a, b)
}

// Clone returns a shallow copy of the set
func Clone[T comparable](s Set[T]) Set[T] {
	result := NewSet[T]()
	for k := range s {
		result.Add(k)
	}
	return result
}

// Filter returns a new set with elements that satisfy the predicate
// predicate is a function that takes an element and returns a boolean
// e.g. filter all even elements in the set
func Filter[T comparable](s Set[T], predicate func(T) bool) Set[T] {
	result := NewSet[T]()
	for k := range s {
		if predicate(k) {
			result.Add(k)
		}
	}
	return result
}

// Map returns a new set with elements that are transformed by the function
// function is a function that takes an element and returns a new element
// e.g. square all elements in the set
func Map[T comparable, U comparable](s Set[T], function func(T) U) Set[U] {
	result := NewSet[U]()
	for k := range s {
		result.Add(function(k))
	}
	return result
}

// Reduce reduces the set to a single value
// function is a function that takes an accumulator and an element and returns a new accumulator
// e.g. sum all elements in the set
func Reduce[T comparable, U comparable](s Set[T], function func(U, T) U, initial U) U {
	accumulator := initial
	for k := range s {
		accumulator = function(accumulator, k)
	}
	return accumulator
}

// ForEach applies the function to each element in the set
// function is a function that takes an element
// e.g. print all elements in the set
func ForEach[T comparable](s Set[T], function func(T)) {
	for k := range s {
		function(k)
	}
}

// Any checks if any element in the set satisfies the predicate
// predicate is a function that takes an element and returns a boolean
// e.g. check if any element is even in the set
func Any[T comparable](s Set[T], predicate func(T) bool) bool {
	for k := range s {
		if predicate(k) {
			return true
		}
	}
	return false
}

// All checks if all elements in the set satisfy the predicate
// predicate is a function that takes an element and returns a boolean
// e.g. check if all elements are even in the set
func All[T comparable](s Set[T], predicate func(T) bool) bool {
	for k := range s {
		if !predicate(k) {
			return false
		}
	}
	return true
}

// None checks if no elements in the set satisfy the predicate
// predicate is a function that takes an element and returns a boolean
// e.g. check if no elements are even in the set
func None[T comparable](s Set[T], predicate func(T) bool) bool {
	for k := range s {
		if predicate(k) {
			return false
		}
	}
	return true
}

// Find returns the first element that satisfies the predicate
// predicate is a function that takes an element and returns a boolean
// e.g. find the first even element in the set
func Find[T comparable](s Set[T], predicate func(T) bool) (T, bool) {
	for k := range s {
		if predicate(k) {
			return k, true
		}
	}
	var zero T
	return zero, false
}

// FindAll returns all elements that satisfy the predicate
// predicate is a function that takes an element and returns a boolean
// e.g. find all even elements in the set
func FindAll[T comparable](s Set[T], predicate func(T) bool) Set[T] {
	result := NewSet[T]()
	for k := range s {
		if predicate(k) {
			result.Add(k)
		}
	}
	return result
}

// Partition partitions the set into two sets based on the predicate
// predicate is a function that takes an element and returns a boolean
// e.g. partition the set into even and odd elements
func Partition[T comparable](s Set[T], predicate func(T) bool) (Set[T], Set[T]) {
	a := NewSet[T]()
	b := NewSet[T]()
	for k := range s {
		if predicate(k) {
			a.Add(k)
		} else {
			b.Add(k)
		}
	}
	return a, b
}
