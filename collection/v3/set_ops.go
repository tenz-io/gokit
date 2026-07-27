package collection

// 本文件为 set 代数和函数式操作提供自由函数入口。在 Go 中,方法与包级函数不能同名,
// 因此代数自由函数加上 *Of 后缀([UnionOf]、[IntersectOf]、...),而链式方法保持简短
// ([Set.Union]、[Set.Intersect]、...)。在调用点选用最易读的形式:
//
//	collection.UnionOf(a, b).Intersect(c)        // 函数式起点,再链式调用
//	a.Union(b).Intersect(c).Subtract(d)          // 全链式

// UnionOf 返回包含 a 和 b 所有元素的新 set(a ∪ b)。
func UnionOf[T comparable](a, b Set[T]) Set[T] { return a.Union(b) }

// IntersectOf 返回同时存在于 a 和 b 中的元素的新 set(a ∩ b)。
func IntersectOf[T comparable](a, b Set[T]) Set[T] { return a.Intersect(b) }

// Difference 返回在 a 中但不在 b 中的元素的新 set(a \ b)。它是 [Set.Subtract] 的自由函数形式;保留 v2 的名称以方便调用方迁移调用点。
func Difference[T comparable](a, b Set[T]) Set[T] { return a.Subtract(b) }

// SymmetricDifference 返回恰好在 a 或 b 其中之一中的元素的新 set(a △ b)。
func SymmetricDifference[T comparable](a, b Set[T]) Set[T] {
	return a.SymmetricDifference(b)
}

// IsSubset 报告 a 的每个元素是否都在 b 中(a ⊆ b)。
func IsSubset[T comparable](a, b Set[T]) bool { return a.IsSubset(b) }

// IsSuperset 报告 b 的每个元素是否都在 a 中(a ⊇ b)。
func IsSuperset[T comparable](a, b Set[T]) bool { return a.IsSuperset(b) }

// IsDisjoint 报告 a 和 b 是否没有共同元素(a ∩ b = ∅)。
func IsDisjoint[T comparable](a, b Set[T]) bool { return a.IsDisjoint(b) }

// Equal 报告 a 和 b 是否包含相同元素。
func Equal[T comparable](a, b Set[T]) bool { return a.Equal(b) }

// Clone 返回 s 的独立副本。它是 [Set.Clone] 的自由函数形式。
func Clone[T comparable](s Set[T]) Set[T] { return s.Clone() }

// --- 函数式操作 ---

// Find 返回满足 predicate 的第一个元素(按 map 迭代顺序,该顺序不确定),否则返回 (zero, false)。
func Find[T comparable](s Set[T], predicate func(T) bool) (T, bool) {
	for k := range s.m {
		if predicate(k) {
			return k, true
		}
	}
	var zero T
	return zero, false
}

// FindAll 返回包含满足 predicate 的每个元素的新 set。
func FindAll[T comparable](s Set[T], predicate func(T) bool) Set[T] {
	out := make(map[T]struct{}, len(s.m))
	for k := range s.m {
		if predicate(k) {
			out[k] = struct{}{}
		}
	}
	return Set[T]{m: out}
}

// Partition 将 s 拆分为两个 set:matched 持有 predicate 为真的元素,unmatched 持有其余元素。
func Partition[T comparable](s Set[T], predicate func(T) bool) (matched, unmatched Set[T]) {
	matched = Set[T]{m: make(map[T]struct{}, len(s.m))}
	unmatched = Set[T]{m: make(map[T]struct{}, len(s.m))}
	for k := range s.m {
		if predicate(k) {
			matched.m[k] = struct{}{}
		} else {
			unmatched.m[k] = struct{}{}
		}
	}
	return
}

// Map 通过 fn 变换 s 的每个元素并返回一个新 set。映射到同一值的多个源元素会合并为一个。
func Map[T comparable, U comparable](s Set[T], fn func(T) U) Set[U] {
	out := make(map[U]struct{}, len(s.m))
	for k := range s.m {
		out[fn(k)] = struct{}{}
	}
	return Set[U]{m: out}
}

// Reduce 从 initial 开始,将 s 折叠为单个值。
func Reduce[T comparable, U any](s Set[T], reducer func(acc U, elem T) U, initial U) U {
	acc := initial
	for k := range s.m {
		acc = reducer(acc, k)
	}
	return acc
}

// ForEach 对每个元素调用 fn。访问顺序不确定。
func ForEach[T comparable](s Set[T], fn func(T)) {
	for k := range s.m {
		fn(k)
	}
}

// Any 报告是否至少有一个元素满足 predicate。
func Any[T comparable](s Set[T], predicate func(T) bool) bool {
	for k := range s.m {
		if predicate(k) {
			return true
		}
	}
	return false
}

// All 报告是否每个元素都满足 predicate。空 set 返回 true(空真命题)。
func All[T comparable](s Set[T], predicate func(T) bool) bool {
	for k := range s.m {
		if !predicate(k) {
			return false
		}
	}
	return true
}

// None 报告是否没有元素满足 predicate。
func None[T comparable](s Set[T], predicate func(T) bool) bool {
	for k := range s.m {
		if predicate(k) {
			return false
		}
	}
	return true
}
