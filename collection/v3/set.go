package collection

// Set 是一个基于 map 的、存储 comparable 值的无序 set。
//
// 与 v2(暴露 `type Set map[T]struct{}`)不同,v3 将 map 隐藏在
// struct 之后,使内部布局可在不破坏调用方的情况下演进,并使 set 代数可以表示为方法链:
//
//	a.Union(b).Intersect(c).Subtract(d)
//
// 所有代数方法都返回一个新的 [Set];它们从不修改接收者。
// 为偏好函数式风格或对 slice 中存储的值进行操作的调用方提供了
// 自由函数别名([Union]、[Intersect]、[Difference]、...)。
//
// 如需保序 set(去重 + 按插入顺序迭代),请改用 functional/v3 的 OrderedSet。
type Set[T comparable] struct {
	m map[T]struct{}
}

// NewSet 创建一个 set,可选地预先填入 values。
func NewSet[T comparable](values ...T) Set[T] {
	s := Set[T]{m: make(map[T]struct{}, len(values))}
	s.Add(values...)
	return s
}

// NewSetWithCap 创建一个为 cap 个元素预分配大小的空 set。
func NewSetWithCap[T comparable](cap int) Set[T] {
	if cap <= 0 {
		cap = defaultCap
	}
	return Set[T]{m: make(map[T]struct{}, cap)}
}

// Add 将 values 插入 set。重复值是 no-op。它返回接收者,以便 Add 可与代数方法链式调用:`s.Add(1, 2).Union(b)`。
func (s Set[T]) Add(values ...T) Set[T] {
	for _, v := range values {
		s.m[v] = struct{}{}
	}
	return s
}

// Remove 从 set 中删除 v。若 v 不存在则是 no-op。
func (s Set[T]) Remove(v T) {
	delete(s.m, v)
}

// Contains 报告 v 是否存在于 set 中。
func (s Set[T]) Contains(v T) bool {
	_, ok := s.m[v]
	return ok
}

// Len 返回元素数量。
func (s Set[T]) Len() int { return len(s.m) }

// IsEmpty 报告 set 是否没有元素。
func (s Set[T]) IsEmpty() bool { return len(s.m) == 0 }

// Clear 移除所有元素。内部 map 保留以便复用。
func (s Set[T]) Clear() {
	clear(s.m)
}

// Values 以 slice 形式返回所有元素。顺序不确定(map 迭代顺序);返回的 slice 是独立的。
func (s Set[T]) Values() []T {
	out := make([]T, 0, len(s.m))
	for k := range s.m {
		out = append(out, k)
	}
	return out
}

// Clone 返回 set 的独立副本。
func (s Set[T]) Clone() Set[T] {
	out := make(map[T]struct{}, len(s.m))
	for k := range s.m {
		out[k] = struct{}{}
	}
	return Set[T]{m: out}
}

// --- 链式 set 代数(返回新 Set,接收者不变)---

// Union 返回包含 s 和 other 所有元素的新 set(s ∪ other)。
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

// Intersect 返回同时存在于 s 和 other 中的元素的新 set(s ∩ other)。为提高效率,它会遍历较小的 set。
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

// Subtract 返回在 s 中但不在 other 中的元素的新 set(s \ other)。
func (s Set[T]) Subtract(other Set[T]) Set[T] {
	out := make(map[T]struct{}, len(s.m))
	for k := range s.m {
		if _, ok := other.m[k]; !ok {
			out[k] = struct{}{}
		}
	}
	return Set[T]{m: out}
}

// SymmetricDifference 返回恰好在 s 或 other 其中之一中的元素的新 set(s △ other)。
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

// --- set 关系(谓词)---

// IsSubset 报告 s 的每个元素是否都在 other 中(s ⊆ other)。
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

// IsSuperset 报告 other 的每个元素是否都在 s 中(s ⊇ other)。
func (s Set[T]) IsSuperset(other Set[T]) bool {
	return other.IsSubset(s)
}

// IsDisjoint 报告 s 和 other 是否没有共同元素(s ∩ other = ∅)。
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

// Equal 报告 s 和 other 是否包含相同元素。
func (s Set[T]) Equal(other Set[T]) bool {
	return len(s.m) == len(other.m) && s.IsSubset(other)
}
