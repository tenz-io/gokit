package collection

// Set is an unordered set of comparable values backed by a map.
//
// Unlike v2 (which exposed `type Set map[T]struct{}`), v3 hides the map behind
// a struct so the internal layout can evolve without breaking callers, and so
// the set algebra can be expressed as method chains:
//
//	a.Union(b).Intersect(c).Subtract(d)
//
// All algebra methods return a fresh [Set]; they never mutate the receiver.
// Free-function aliases ([Union], [Intersect], [Difference], ...) are provided
// for callers that prefer a functional style or operate on values held in a
// slice.
//
// For an order-preserving set (dedup + iteration in insertion order) use
// functional/v3's OrderedSet instead.
type Set[T comparable] struct {
	m map[T]struct{}
}

// NewSet creates a set, optionally pre-populated with values.
func NewSet[T comparable](values ...T) Set[T] {
	s := Set[T]{m: make(map[T]struct{}, len(values))}
	s.Add(values...)
	return s
}

// NewSetWithCap creates an empty set pre-sized for cap elements.
func NewSetWithCap[T comparable](cap int) Set[T] {
	if cap <= 0 {
		cap = defaultCap
	}
	return Set[T]{m: make(map[T]struct{}, cap)}
}

// Add inserts the values into the set. Duplicates are no-ops. It returns the
// receiver so Add can be chained with algebra methods: `s.Add(1, 2).Union(b)`.
func (s Set[T]) Add(values ...T) Set[T] {
	for _, v := range values {
		s.m[v] = struct{}{}
	}
	return s
}

// Remove deletes v from the set. It is a no-op if v is absent.
func (s Set[T]) Remove(v T) {
	delete(s.m, v)
}

// Contains reports whether v is present in the set.
func (s Set[T]) Contains(v T) bool {
	_, ok := s.m[v]
	return ok
}

// Len returns the number of elements.
func (s Set[T]) Len() int { return len(s.m) }

// IsEmpty reports whether the set has no elements.
func (s Set[T]) IsEmpty() bool { return len(s.m) == 0 }

// Clear removes all elements. The internal map is retained for reuse.
func (s Set[T]) Clear() {
	clear(s.m)
}

// Values returns all elements as a slice. The order is non-deterministic
// (map iteration order); the returned slice is independent.
func (s Set[T]) Values() []T {
	out := make([]T, 0, len(s.m))
	for k := range s.m {
		out = append(out, k)
	}
	return out
}

// Clone returns an independent copy of the set.
func (s Set[T]) Clone() Set[T] {
	out := make(map[T]struct{}, len(s.m))
	for k := range s.m {
		out[k] = struct{}{}
	}
	return Set[T]{m: out}
}

// --- Chained set algebra (return new Set, receiver untouched) ---

// Union returns a new set with all elements from s and other (s ∪ other).
func (s Set[T]) Union(other Set[T]) Set[T] {
	out := make(map[T]struct{}, len(s.m)+len(other.m))
	for k := range s.m {
		out[k] = struct{}{}
	}
	for k := range other.m {
		out[k] = struct{}{}
	}
	return Set[T]{m: out}
}

// Intersect returns a new set with elements present in both s and other
// (s ∩ other). It iterates the smaller set for efficiency.
func (s Set[T]) Intersect(other Set[T]) Set[T] {
	a, b := s.m, other.m
	if len(a) > len(b) {
		a, b = b, a
	}
	out := make(map[T]struct{}, len(a))
	for k := range a {
		if _, ok := b[k]; ok {
			out[k] = struct{}{}
		}
	}
	return Set[T]{m: out}
}

// Subtract returns a new set with elements in s but not in other (s \ other).
func (s Set[T]) Subtract(other Set[T]) Set[T] {
	out := make(map[T]struct{}, len(s.m))
	for k := range s.m {
		if _, ok := other.m[k]; !ok {
			out[k] = struct{}{}
		}
	}
	return Set[T]{m: out}
}

// SymmetricDifference returns a new set with elements in exactly one of s or
// other (s △ other).
func (s Set[T]) SymmetricDifference(other Set[T]) Set[T] {
	out := make(map[T]struct{}, len(s.m)+len(other.m))
	for k := range s.m {
		if _, ok := other.m[k]; !ok {
			out[k] = struct{}{}
		}
	}
	for k := range other.m {
		if _, ok := s.m[k]; !ok {
			out[k] = struct{}{}
		}
	}
	return Set[T]{m: out}
}

// --- Set relations (predicates) ---

// IsSubset reports whether every element of s is in other (s ⊆ other).
func (s Set[T]) IsSubset(other Set[T]) bool {
	if len(s.m) > len(other.m) {
		return false
	}
	for k := range s.m {
		if _, ok := other.m[k]; !ok {
			return false
		}
	}
	return true
}

// IsSuperset reports whether every element of other is in s (s ⊇ other).
func (s Set[T]) IsSuperset(other Set[T]) bool {
	return other.IsSubset(s)
}

// IsDisjoint reports whether s and other share no elements (s ∩ other = ∅).
func (s Set[T]) IsDisjoint(other Set[T]) bool {
	a, b := s.m, other.m
	if len(a) > len(b) {
		a, b = b, a
	}
	for k := range a {
		if _, ok := b[k]; ok {
			return false
		}
	}
	return true
}

// Equal reports whether s and other contain the same elements.
func (s Set[T]) Equal(other Set[T]) bool {
	return len(s.m) == len(other.m) && s.IsSubset(other)
}
