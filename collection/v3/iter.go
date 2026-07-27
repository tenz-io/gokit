package collection

import "iter"

// All 返回一个遍历 stack 元素的 [iter.Seq],顺序从底部(最先压入)到顶部(最后压入)。提前 break 出 range 是安全的。
//
//	for v := range s.All() { ... }
func (s *Stack[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s.data {
			if !yield(v) {
				return
			}
		}
	}
}

// All 返回一个遍历 queue 元素的 [iter.Seq],顺序从前端(最旧)到后端(最新)。提前 break 出 range 是安全的。
func (q *Queue[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		n := len(q.buf)
		for i := 0; i < q.count; i++ {
			if !yield(q.buf[(q.head+i)%n]) {
				return
			}
		}
	}
}

// All 返回一个遍历 heap 元素的 [iter.Seq]。顺序是内部 heap 布局,而非优先级顺序 —— 若要按优先级顺序排空,请调用 [Heap.Pop] 直到为空。提前 break 出 range 是安全的。
func (h *Heap[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range h.data {
			if !yield(v) {
				return
			}
		}
	}
}

// All 返回一个遍历 set 元素的 [iter.Seq]。顺序不确定(map 迭代顺序)。提前 break 出 range 是安全的。
func (s Set[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for k := range s.m {
			if !yield(k) {
				return
			}
		}
	}
}
