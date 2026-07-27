package collection

// Queue 是一个基于 ring buffer 的 FIFO(先入先出)queue。
//
// v2 用 slice 作为 queue 的底层,通过 `data = data[1:]` 执行 Dequeue,这会使窗口
// 永远在同一个 backing array 上推进 —— 出队的槽位永远不会被回收,数组只增不减。
// v3 的 ring buffer 在出队时回收槽位,因此 Enqueue/Dequeue 是均摊 O(1),
// 且内存在 queue 的整个生命周期内保持有界。当 buffer 填满时,它会扩容到 2× 并将环绕的内容
// 重新线性化到前端。
type Queue[T any] struct {
	buf   []T
	head  int // 前端元素的索引
	tail  int // 下一次 Enqueue 写入的索引
	count int // 元素数量
}

// NewQueue 创建一个使用默认初始 capacity 的空 queue。
func NewQueue[T any]() *Queue[T] {
	return NewQueueWithCap[T](defaultCap)
}

// NewQueueWithCap 创建一个为 cap 个元素预分配大小的空 queue。非正的 cap 会回退到默认 capacity。
func NewQueueWithCap[T any](cap int) *Queue[T] {
	if cap <= 0 {
		cap = defaultCap
	}
	return &Queue[T]{buf: make([]T, cap)}
}

// Enqueue 将 v 添加到 queue 的后端。均摊 O(1):buffer 填满时会翻倍并重新线性化。
func (q *Queue[T]) Enqueue(v T) {
	if q.count == len(q.buf) {
		q.grow()
	}
	q.buf[q.tail] = v
	q.tail = (q.tail + 1) % len(q.buf)
	q.count++
}

// Dequeue 移除并返回前端元素。当 queue 为空时返回 (zero, false)。出队的槽位会被清零,
// 以便 GC 回收它持有的任何引用 —— 这正是 v2 的 slice[1:] 方式从未解决的问题。
func (q *Queue[T]) Dequeue() (T, bool) {
	var zero T
	if q.count == 0 {
		return zero, false
	}
	v := q.buf[q.head]
	q.buf[q.head] = zero
	q.head = (q.head + 1) % len(q.buf)
	q.count--
	return v, true
}

// Peek 返回前端元素但不移除它。当 queue 为空时返回 (zero, false)。
func (q *Queue[T]) Peek() (T, bool) {
	var zero T
	if q.count == 0 {
		return zero, false
	}
	return q.buf[q.head], true
}

// Len 返回元素数量。
func (q *Queue[T]) Len() int { return q.count }

// IsEmpty 报告 queue 是否没有元素。
func (q *Queue[T]) IsEmpty() bool { return q.count == 0 }

// Clear 移除所有元素,清零 buffer,并重置 head/tail 指针。buffer 的 capacity 保留以便复用。
func (q *Queue[T]) Clear() {
	var zero T
	for i := range q.buf {
		q.buf[i] = zero
	}
	q.head, q.tail, q.count = 0, 0, 0
}

// Values 返回按从前到后顺序排列的元素副本。返回的 slice 是独立的;修改它不会影响 queue。
func (q *Queue[T]) Values() []T {
	out := make([]T, q.count)
	for i := 0; i < q.count; i++ {
		out[i] = q.buf[(q.head+i)%len(q.buf)]
	}
	return out
}

// Clone 返回 queue 的独立副本。即便源 queue 发生环绕,副本仍线性排列(head=0)。
func (q *Queue[T]) Clone() *Queue[T] {
	cap := max(q.count, defaultCap)
	out := make([]T, cap, cap)
	copy(out, q.Values())
	return &Queue[T]{buf: out, head: 0, tail: q.count, count: q.count}
}

// grow 将 buffer 翻倍并线性化环绕的内容,使 head=0 且 tail=count。仅在 buffer 填满时调用。
func (q *Queue[T]) grow() {
	newCap := max(len(q.buf)*2, defaultCap)
	next := make([]T, newCap, newCap)
	// 将逻辑内容(head..head+count)拷贝到 next 的前端。
	// 两次拷贝覆盖了环绕情况,热路径中无需分支。
	n := copy(next, q.buf[q.head:])
	copy(next[n:], q.buf[:q.count-n])
	q.buf = next
	q.head = 0
	q.tail = q.count
}
