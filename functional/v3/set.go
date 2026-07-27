package fn

// OrderedSet 是一个 comparable 值集合,会记住元素首次插入的顺序。
// membership 检查为 O(1);迭代与 ToSlice 按插入顺序返回元素。
//
// 它是与保留顺序的 Deduplicate 辅助函数的天然对应物:构建一次集合,
// 随后可反复做 membership 测试,无需重新扫描 slice。对于无序集合的
// 完整集合代数(并、交、差),请改用 collection/v2 的 Set。
type OrderedSet[T comparable] struct {
	index map[T]int // value -> 在 order 中的位置(用于稳定删除)
	order []T
}

// NewOrderedSet 创建一个 OrderedSet,可选地用 values 做种子。种子值中
// 后续重复者只保留首次出现。
func NewOrderedSet[T comparable](values ...T) *OrderedSet[T] {
	s := &OrderedSet[T]{index: make(map[T]int, len(values))}
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// NewOrderedSetWithCapacity 创建一个空 OrderedSet,按 cap 个不同元素预分配。
func NewOrderedSetWithCapacity[T comparable](cap int) *OrderedSet[T] {
	return &OrderedSet[T]{index: make(map[T]int, cap)}
}

// Add 在 v 不存在时插入,并返回它是否为新增。
func (s *OrderedSet[T]) Add(v T) bool {
	if _, ok := s.index[v]; ok {
		return false
	}
	s.index[v] = len(s.order)
	s.order = append(s.order, v)
	return true
}

// Contains 报告 v 是否存在。
func (s *OrderedSet[T]) Contains(v T) bool {
	_, ok := s.index[v]
	return ok
}

// Has 是 Contains 的别名。
func (s *OrderedSet[T]) Has(v T) bool { return s.Contains(v) }

// Remove 在 v 存在时删除,并返回它是否被移除。
//
// 删除把该槽标记为逻辑空,而非平移 order slice,因此随后的
// ToSlice/ForEach 会在 O(n) 中跳过空隙。这让删除保持 O(1),
// 代价是槽只能由 Clone 回收。
func (s *OrderedSet[T]) Remove(v T) bool {
	pos, ok := s.index[v]
	if !ok {
		return false
	}
	delete(s.index, v)
	var zero T
	s.order[pos] = zero // tombstone;ToSlice/ForEach 跳过 zero-tombstone
	return true
}

// Len 返回集合中当前的元素个数。
func (s *OrderedSet[T]) Len() int { return len(s.index) }

// IsEmpty 报告集合是否无元素。
func (s *OrderedSet[T]) IsEmpty() bool { return len(s.index) == 0 }

// ToSlice 按插入顺序返回元素,跳过 Remove 留下的 tombstone。
// 返回的 slice 是副本;修改它不会影响集合。
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

// ForEach 按插入顺序对每个元素调用 fn,跳过 tombstone。
func (s *OrderedSet[T]) ForEach(fn func(T)) {
	for _, v := range s.order {
		if _, ok := s.index[v]; !ok {
			continue
		}
		fn(v)
	}
}

// Clone 返回集合的副本,并移除 tombstone(顺序被压缩)。
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
