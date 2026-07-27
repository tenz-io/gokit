package collection

import "cmp"

// Heap 是一个基于二叉堆的 priority queue。
//
// 顺序由一个 less 函数定义:less(a, b) == true 表示 a 的优先级高于 b,因此 a 先出堆。
// 常见情形 —— 有序类型上的 min-heap 与 max-heap —— 有专用构造器
// [NewMinHeap] 和 [NewMaxHeap],因此调用方几乎不必手写 less 函数。对于自定义排序
// (多字段、计算优先级)请使用 [NewHeap] 并显式传入 less。
//
// Push 和 Pop 是 O(log n);Peek 和 Len 是 O(1)。
type Heap[T any] struct {
	data []T
	less func(a, b T) bool
}

// NewHeap 创建一个按 less 排序、使用默认 capacity 的空 heap。
// less(a, b) == true 表示 a 的优先级高于 b。
func NewHeap[T any](less func(a, b T) bool) *Heap[T] {
	return NewHeapWithCap[T](defaultCap, less)
}

// NewHeapWithCap 创建一个按 less 排序、为 cap 个元素预分配大小的空 heap。非正的 cap 会回退到默认 capacity。
func NewHeapWithCap[T any](cap int, less func(a, b T) bool) *Heap[T] {
	if cap <= 0 {
		cap = defaultCap
	}
	return &Heap[T]{data: make([]T, 0, cap), less: less}
}

// NewMinHeap 创建一个基于有序类型的空 min-heap:最小元素具有最高优先级,最先出堆。
func NewMinHeap[T cmp.Ordered]() *Heap[T] {
	return NewHeap[T](func(a, b T) bool { return cmp.Compare(a, b) < 0 })
}

// NewMaxHeap 创建一个基于有序类型的空 max-heap:最大元素具有最高优先级,最先出堆。
func NewMaxHeap[T cmp.Ordered]() *Heap[T] {
	return NewHeap[T](func(a, b T) bool { return cmp.Compare(a, b) > 0 })
}

// Push 将 v 添加到 heap。O(log n)。
func (h *Heap[T]) Push(v T) {
	h.data = append(h.data, v)
	h.siftUp(len(h.data) - 1)
}

// Pop 移除并返回最高优先级的元素。当 heap 为空时返回 (zero, false)。O(log n)。
func (h *Heap[T]) Pop() (T, bool) {
	var zero T
	if len(h.data) == 0 {
		return zero, false
	}
	top := h.data[0]
	n := len(h.data) - 1
	h.data[0] = h.data[n]
	h.data[n] = zero // 让 GC 回收被移动的引用
	h.data = h.data[:n]
	if n > 0 {
		h.siftDown(0)
	}
	return top, true
}

// Peek 返回最高优先级的元素但不移除它。O(1)。当 heap 为空时返回 (zero, false)。
func (h *Heap[T]) Peek() (T, bool) {
	var zero T
	if len(h.data) == 0 {
		return zero, false
	}
	return h.data[0], true
}

// Len 返回元素数量。
func (h *Heap[T]) Len() int { return len(h.data) }

// IsEmpty 报告 heap 是否没有元素。
func (h *Heap[T]) IsEmpty() bool { return len(h.data) == 0 }

// Clear 移除所有元素并清零 backing array,以便 GC 回收持有的引用。capacity 保留以便复用。
func (h *Heap[T]) Clear() {
	var zero T
	for i := range h.data {
		h.data[i] = zero
	}
	h.data = h.data[:0]
}

// Values 返回底层 heap slice 的副本。顺序是 heap 的内部布局,而非优先级顺序 —— 若要按优先级顺序排空,请调用 Pop 直到为空。返回的 slice 是独立的。
func (h *Heap[T]) Values() []T {
	out := make([]T, len(h.data))
	copy(out, h.data)
	return out
}

// Clone 返回 heap 的独立副本,共享相同的排序方式。
func (h *Heap[T]) Clone() *Heap[T] {
	out := make([]T, len(h.data))
	copy(out, h.data)
	return &Heap[T]{data: out, less: h.less}
}

// siftUp 将索引 i 处的元素向上移动,直到恢复 heap 性质。O(log n)。
func (h *Heap[T]) siftUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !h.less(h.data[i], h.data[parent]) {
			break
		}
		h.data[i], h.data[parent] = h.data[parent], h.data[i]
		i = parent
	}
}

// siftDown 将索引 i 处的元素向下移动,直到恢复 heap 性质。O(log n)。
func (h *Heap[T]) siftDown(i int) {
	n := len(h.data)
	for {
		left := 2*i + 1
		right := 2*i + 2
		best := i
		if left < n && h.less(h.data[left], h.data[best]) {
			best = left
		}
		if right < n && h.less(h.data[right], h.data[best]) {
			best = right
		}
		if best == i {
			break
		}
		h.data[i], h.data[best] = h.data[best], h.data[i]
		i = best
	}
}
