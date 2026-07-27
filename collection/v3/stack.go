package collection

const defaultCap = 16

// Stack 是一个基于 slice 的 LIFO(后入先出)stack。
//
// 所有修改方法都使用指针接收者,以便共享底层 slice header(长度、capacity、指针);
// 值拷贝会导致 backing array 发生别名。若需要分叉,请在修改前用 [Stack.Clone] 复制 stack。
type Stack[T any] struct {
	data []T
}

// NewStack 创建一个使用默认初始 capacity 的空 stack。
func NewStack[T any]() *Stack[T] {
	return NewStackWithCap[T](defaultCap)
}

// NewStackWithCap 创建一个为 cap 个元素预分配大小的空 stack。非正的 cap 会回退到默认 capacity,因此调用方可以直接透传零值而无需条件判断。
func NewStackWithCap[T any](cap int) *Stack[T] {
	if cap <= 0 {
		cap = defaultCap
	}
	return &Stack[T]{data: make([]T, 0, cap)}
}

// Push 将 v 添加到 stack 的顶部。均摊 O(1)。
func (s *Stack[T]) Push(v T) {
	s.data = append(s.data, v)
}

// Pop 移除并返回顶部元素。当 stack 为空时返回 (zero, false)。出栈的槽位会被清零,以便 GC 回收它持有的任何引用。
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

// Peek 返回顶部元素但不移除它。当 stack 为空时返回 (zero, false)。
func (s *Stack[T]) Peek() (T, bool) {
	var zero T
	if len(s.data) == 0 {
		return zero, false
	}
	return s.data[len(s.data)-1], true
}

// Len 返回元素数量。
func (s *Stack[T]) Len() int { return len(s.data) }

// IsEmpty 报告 stack 是否没有元素。
func (s *Stack[T]) IsEmpty() bool { return len(s.data) == 0 }

// Clear 移除所有元素,并将 backing array 中不超过其长度的部分清零,以便 GC 回收持有的引用。capacity 保留以便复用。
func (s *Stack[T]) Clear() {
	var zero T
	for i := range s.data {
		s.data[i] = zero
	}
	s.data = s.data[:0]
}

// Values 返回按从底部到顶部顺序排列的元素副本(最先压入的元素在 slice 的最前面)。返回的 slice 是独立的;修改它不会影响 stack。
func (s *Stack[T]) Values() []T {
	out := make([]T, len(s.data))
	copy(out, s.data)
	return out
}

// Clone 返回 stack 的独立副本。
func (s *Stack[T]) Clone() *Stack[T] {
	return &Stack[T]{data: s.Values()}
}
