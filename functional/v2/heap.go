package function

// topKHeap is a min-heap used internally by TopK.
type topKHeap[T any] struct {
	items []T
	less  func(a, b T) bool
}

func newTopKHeap[T any](items []T, less func(a, b T) bool) *topKHeap[T] {
	return &topKHeap[T]{
		items: items,
		less:  less,
	}
}

func (h *topKHeap[T]) Len() int           { return len(h.items) }
func (h *topKHeap[T]) Less(i, j int) bool { return h.less(h.items[i], h.items[j]) }
func (h *topKHeap[T]) Swap(i, j int)      { h.items[i], h.items[j] = h.items[j], h.items[i] }

func (h *topKHeap[T]) Push(x any) {
	h.items = append(h.items, x.(T))
}

func (h *topKHeap[T]) Pop() any {
	n := len(h.items)
	x := h.items[n-1]
	h.items = h.items[:n-1]
	return x
}
