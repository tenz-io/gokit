package fn

import (
	"container/heap"
)

// minHeap is a generic min-heap used internally by TopK/BottomK. It keeps the
// k smallest elements seen so far at the root so a candidate larger than the
// root can evict it. The ordering is provided by the less function, which is
// captured once per call.
type minHeap[T any] struct {
	items []T
	less  func(a, b T) bool
}

func (h *minHeap[T]) Len() int           { return len(h.items) }
func (h *minHeap[T]) Less(i, j int) bool { return h.less(h.items[i], h.items[j]) }
func (h *minHeap[T]) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
}

func (h *minHeap[T]) Push(x any) {
	h.items = append(h.items, x.(T))
}

func (h *minHeap[T]) Pop() any {
	n := len(h.items)
	x := h.items[n-1]
	var zero T
	h.items[n-1] = zero // avoid retention
	h.items = h.items[:n-1]
	return x
}

// topKHeap runs the O((n-k) log k) selection: maintain a min-heap of size k
// over the first k elements, then for each remaining element, evict the root
// and insert the candidate whenever the candidate outranks the root.
//
// lessRootWins drives eviction semantics:
//
//	TopK    (keep largest k):    evict root when candidate > root
//	BottomK (keep smallest k):   evict root when candidate < root
//
// After selection the heap holds the k survivors in min-heap order; the caller
// drains them into the desired final order.
func topKHeap[T any](s []T, k int, less func(a, b T) bool, keep func(root, cand T) bool) []T {
	if k <= 0 || len(s) == 0 {
		return nil
	}
	n := len(s)
	if k >= n {
		k = n
	}

	// Seed with the first k elements (copied so we don't alias the input).
	h := &minHeap[T]{
		items: append(make([]T, 0, k), s[:k]...),
		less:  less,
	}
	heap.Init(h)

	for i := k; i < n; i++ {
		// h.items[0] is the current root (the element most likely to be evicted).
		if keep(h.items[0], s[i]) {
			h.items[0] = s[i]
			heap.Fix(h, 0)
		}
	}
	return h.items
}
