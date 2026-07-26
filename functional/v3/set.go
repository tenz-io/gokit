package fn

// OrderedSet is a set of comparable values that remembers the order in which
// elements were first inserted. Membership checks are O(1); iteration and
// ToSlice return elements in insertion order.
//
// It is the natural counterpart to the order-preserving Deduplicate helpers:
// build a set once and test membership repeatedly without re-scanning a slice.
// For full set algebra (union, intersection, difference) over unordered sets,
// use collection/v2's Set instead.
type OrderedSet[T comparable] struct {
	index map[T]int // value -> position in order (for stable removal)
	order []T
}

// NewOrderedSet creates an OrderedSet, optionally seeded with values. Later
// duplicates among the seed values keep only the first occurrence.
func NewOrderedSet[T comparable](values ...T) *OrderedSet[T] {
	s := &OrderedSet[T]{index: make(map[T]int, len(values))}
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// NewOrderedSetWithCapacity creates an empty OrderedSet pre-sized for cap
// distinct elements.
func NewOrderedSetWithCapacity[T comparable](cap int) *OrderedSet[T] {
	return &OrderedSet[T]{index: make(map[T]int, cap)}
}

// Add inserts v if absent and returns whether it was newly added.
func (s *OrderedSet[T]) Add(v T) bool {
	if _, ok := s.index[v]; ok {
		return false
	}
	s.index[v] = len(s.order)
	s.order = append(s.order, v)
	return true
}

// Contains reports whether v is present.
func (s *OrderedSet[T]) Contains(v T) bool {
	_, ok := s.index[v]
	return ok
}

// Has is an alias for Contains.
func (s *OrderedSet[T]) Has(v T) bool { return s.Contains(v) }

// Remove deletes v if present and returns whether it was removed.
//
// Removal marks the slot as logically empty rather than shifting the order
// slice, so subsequent ToSlice/ForEach skip gaps in O(n). This keeps removal
// O(1) at the cost of slots only being reclaimed by Clone.
func (s *OrderedSet[T]) Remove(v T) bool {
	pos, ok := s.index[v]
	if !ok {
		return false
	}
	delete(s.index, v)
	var zero T
	s.order[pos] = zero // tombstone; ToSlice/ForEach skip zero-tombstones
	return true
}

// Len returns the number of elements currently in the set.
func (s *OrderedSet[T]) Len() int { return len(s.index) }

// IsEmpty reports whether the set has no elements.
func (s *OrderedSet[T]) IsEmpty() bool { return len(s.index) == 0 }

// ToSlice returns the elements in insertion order, skipping any tombstones
// left by Remove. The returned slice is a copy; mutating it does not affect
// the set.
func (s *OrderedSet[T]) ToSlice() []T {
	out := make([]T, 0, len(s.index))
	for _, v := range s.order {
		if _, ok := s.index[v]; !ok {
			continue
		}
		out = append(out, v)
	}
	return out
}

// ForEach calls fn on each element in insertion order, skipping tombstones.
func (s *OrderedSet[T]) ForEach(fn func(T)) {
	for _, v := range s.order {
		if _, ok := s.index[v]; !ok {
			continue
		}
		fn(v)
	}
}

// Clone returns a copy of the set with tombstones removed (order compacted).
func (s *OrderedSet[T]) Clone() *OrderedSet[T] {
	out := NewOrderedSetWithCapacity[T](len(s.index))
	for _, v := range s.order {
		if _, ok := s.index[v]; !ok {
			continue
		}
		out.Add(v)
	}
	return out
}
