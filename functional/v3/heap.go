package fn

import (
	"container/heap"
)

// minHeap 是 TopK/BottomK 内部使用的通用 min-heap。它把目前为止见到的 k 个
// 最小元素保存在 root,这样比 root 更大的候选元素就能驱逐它。顺序由 less
// 函数提供,每次调用捕获一次。
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
	h.items[n-1] = zero // 避免引用残留
	h.items = h.items[:n-1]
	return x
}

// topKHeap 运行 O((n-k) log k) 选择:在前 k 个元素上维护一个大小为 k 的
// min-heap,然后对每个剩余元素,当候选元素优于 root 时驱逐 root 并插入该候选。
//
// lessRootWins 决定驱逐语义:
//
//	TopK    (保留最大 k 个): 当 candidate > root 时驱逐 root
//	BottomK (保留最小 k 个): 当 candidate < root 时驱逐 root
//
// 选择完成后,heap 以 min-heap 顺序持有 k 个幸存者;调用方将其排空为
// 所需的最终顺序。
func topKHeap[T any](s []T, k int, less func(a, b T) bool, keep func(root, cand T) bool) []T {
	if k <= 0 || len(s) == 0 {
		return nil
	}
	n := len(s)
	if k >= n {
		k = n
	}

	// 用前 k 个元素做种子(复制以避免 alias 输入)。
	h := &minHeap[T]{
		items: append(make([]T, 0, k), s[:k]...),
		less:  less,
	}
	heap.Init(h)

	for i := k; i < n; i++ {
		// h.items[0] 是当前 root(最可能被驱逐的元素)。
		if keep(h.items[0], s[i]) {
			h.items[0] = s[i]
			heap.Fix(h, 0)
		}
	}
	return h.items
}
