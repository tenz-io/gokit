package function

import (
	"container/heap"
	"fmt"
	"sync"
)

type entity[T any] struct {
	val T
}

type itemHeap[T any] struct {
	arr  []*entity[T]
	cmp  func(t1, t2 T) int
	lock sync.RWMutex
}

func newItemHeap[T any](items []T, cmp func(t1, t2 T) int) *itemHeap[T] {
	dst := make([]T, len(items))
	copy(dst, items)
	arr := make([]*entity[T], len(items))
	for i, e := range dst {
		v := e
		arr[i] = &entity[T]{val: v}
	}

	return &itemHeap[T]{
		arr: arr,
		cmp: cmp,
	}
}

func (h *itemHeap[T]) Len() int {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return len(h.arr)
}

func (h *itemHeap[T]) Less(i, j int) bool {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.cmp(h.arr[i].val, h.arr[j].val) < 0
}

func (h *itemHeap[T]) Swap(i, j int) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.arr[i], h.arr[j] = h.arr[j], h.arr[i]
}

func (h *itemHeap[T]) Push(x any) {
	h.lock.Lock()
	defer h.lock.Unlock()
	v := x.(T)
	h.arr = append(h.arr, &entity[T]{val: v})
}

func (h *itemHeap[T]) Pop() any {
	h.lock.Lock()
	defer h.lock.Unlock()

	tailIdx := len(h.arr) - 1
	tail := h.arr[tailIdx]
	h.arr = h.arr[:tailIdx]
	return tail.val
}

func (h *itemHeap[T]) Peek() T {
	return h.GetVal(0)
}

func (h *itemHeap[T]) Update(x T) T {
	if h.Len() == 0 {
		heap.Push(h, x)
		return x
	}
	if h.cmp(x, h.Peek()) < 0 {
		//ignore
		return x
	}

	heap.Push(h, x)
	return heap.Pop(h).(T)
}

func (h *itemHeap[T]) GetVal(i int) T {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.arr[i].val
}

func (h *itemHeap[T]) String() string {
	h.lock.RLock()
	defer h.lock.RUnlock()
	arr := make([]T, 0)
	for _, e := range h.arr {
		arr = append(arr, e.val)
	}
	return fmt.Sprintf("%v", arr)
}
