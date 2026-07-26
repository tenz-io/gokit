package collection

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Stack ---

func TestStack_New(t *testing.T) {
	s := NewStack[int]()
	assert.NotNil(t, s)
	assert.True(t, s.IsEmpty())
	assert.Equal(t, 0, s.Len())
}

func TestStack_NewWithCap(t *testing.T) {
	s := NewStackWithCap[int](0) // non-positive cap falls back to default
	assert.NotNil(t, s)
	assert.True(t, s.IsEmpty())
}

func TestStack_PushPop(t *testing.T) {
	s := NewStack[int]()
	_, ok := s.Pop()
	assert.False(t, ok)

	s.Push(1)
	s.Push(2)
	s.Push(3)
	assert.Equal(t, 3, s.Len())

	v, ok := s.Pop()
	assert.True(t, ok)
	assert.Equal(t, 3, v)
	v, ok = s.Pop()
	assert.Equal(t, 2, v)
	v, ok = s.Pop()
	assert.Equal(t, 1, v)

	_, ok = s.Pop()
	assert.False(t, ok)
}

func TestStack_Peek(t *testing.T) {
	s := NewStack[int]()
	_, ok := s.Peek()
	assert.False(t, ok)

	s.Push(1)
	s.Push(2)
	v, ok := s.Peek()
	assert.True(t, ok)
	assert.Equal(t, 2, v)
	assert.Equal(t, 2, s.Len()) // Peek does not remove
}

func TestStack_Clear(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Clear()
	assert.True(t, s.IsEmpty())
	assert.Equal(t, 0, s.Len())
}

func TestStack_Values(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	assert.Equal(t, []int{1, 2}, s.Values()) // bottom-to-top
	vals := s.Values()
	vals[0] = 99
	v, _ := s.Pop()
	assert.Equal(t, 2, v) // stack unaffected by mutating the copy
}

func TestStack_Clone(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	c := s.Clone()
	assert.Equal(t, s.Values(), c.Values())
	s.Push(3)
	assert.NotEqual(t, s.Len(), c.Len())
}

// --- Queue (ring buffer) ---

func TestQueue_New(t *testing.T) {
	q := NewQueue[int]()
	assert.NotNil(t, q)
	assert.True(t, q.IsEmpty())
}

func TestQueue_EnqueueDequeue(t *testing.T) {
	q := NewQueue[int]()
	_, ok := q.Dequeue()
	assert.False(t, ok)

	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	v, ok := q.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	v, ok = q.Dequeue()
	assert.Equal(t, 2, v)
	v, ok = q.Dequeue()
	assert.Equal(t, 3, v)

	_, ok = q.Dequeue()
	assert.False(t, ok)
}

func TestQueue_Peek(t *testing.T) {
	q := NewQueue[int]()
	_, ok := q.Peek()
	assert.False(t, ok)

	q.Enqueue(1)
	q.Enqueue(2)
	v, ok := q.Peek()
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	assert.Equal(t, 2, q.Len())
}

func TestQueue_Clear(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Clear()
	assert.True(t, q.IsEmpty())
}

func TestQueue_Values(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	assert.Equal(t, []int{1, 2}, q.Values())
	vals := q.Values()
	vals[0] = 99
	v, _ := q.Peek()
	assert.Equal(t, 1, v)
}

func TestQueue_Clone(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	c := q.Clone()
	assert.Equal(t, q.Values(), c.Values())
	q.Enqueue(3)
	assert.NotEqual(t, q.Len(), c.Len())
}

// TestQueue_RingBuffer_Wrap exercises the wrap-around path: a small-capacity
// queue that wraps many times must still return elements in FIFO order.
func TestQueue_RingBuffer_Wrap(t *testing.T) {
	q := NewQueueWithCap[int](4)
	for round := 0; round < 5; round++ {
		for i := 0; i < 100; i++ {
			q.Enqueue(i)
		}
		if q.Len() != 100 {
			t.Fatalf("round %d: len=%d want 100", round, q.Len())
		}
		for i := 0; i < 100; i++ {
			v, ok := q.Dequeue()
			if !ok || v != i {
				t.Fatalf("round %d dequeue %d: got %v ok=%v", round, i, v, ok)
			}
		}
		if !q.IsEmpty() {
			t.Fatalf("round %d not empty: len=%d", round, q.Len())
		}
	}
}

// TestQueue_RingBuffer_NoMemoryLeak guards the bug v2 never fixed: a
// long-lived queue that enqueues and dequeues the same number of elements
// must keep its length at zero and a stable (bounded) backing capacity — the
// slice[1:] approach in v2 left dequeued slots unreclaimed and the array grew
// without bound.
func TestQueue_RingBuffer_NoMemoryLeak(t *testing.T) {
	q := NewQueueWithCap[int](8)
	const rounds = 1000
	for r := 0; r < rounds; r++ {
		for i := 0; i < 10; i++ {
			q.Enqueue(i)
		}
		for i := 0; i < 10; i++ {
			q.Dequeue()
		}
	}
	assert.Equal(t, 0, q.Len(), "queue must drain to empty")
	// Capacity is allowed to grow from the initial 8, but it must NOT grow
	// without bound across rounds of steady-state enqueue/dequeue. After the
	// first round (10 > 8) it doubles once; thereafter it stays put.
	assert.LessOrEqual(t, cap(q.buf), 64,
		"capacity must stay bounded after steady-state use, got %d", cap(q.buf))
}

func TestQueue_RingBuffer_PartialWrap(t *testing.T) {
	q := NewQueueWithCap[int](2)
	q.Enqueue(10)
	q.Enqueue(11)
	q.Dequeue()   // remove 10, head now at 11
	q.Enqueue(12) // wraps to slot 0

	p, ok := q.Peek()
	assert.True(t, ok)
	assert.Equal(t, 11, p)

	assert.Equal(t, []int{11, 12}, q.Values())

	v, _ := q.Dequeue()
	assert.Equal(t, 11, v)
	v, _ = q.Dequeue()
	assert.Equal(t, 12, v)
	assert.True(t, q.IsEmpty())
}

// --- Heap ---

func intLess(a, b int) bool { return a < b }

func TestHeap_New(t *testing.T) {
	h := NewHeap(intLess)
	assert.NotNil(t, h)
	assert.True(t, h.IsEmpty())
}

func TestHeap_PushPop(t *testing.T) {
	h := NewHeap(intLess)
	_, ok := h.Pop()
	assert.False(t, ok)

	h.Push(2)
	h.Push(1)
	h.Push(3)

	v, ok := h.Pop()
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	v, ok = h.Pop()
	assert.Equal(t, 2, v)
	v, ok = h.Pop()
	assert.Equal(t, 3, v)

	_, ok = h.Pop()
	assert.False(t, ok)
}

func TestHeap_Peek(t *testing.T) {
	h := NewHeap(intLess)
	_, ok := h.Peek()
	assert.False(t, ok)

	h.Push(3)
	h.Push(1)
	v, ok := h.Peek()
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	assert.Equal(t, 2, h.Len())
}

func TestHeap_Clone(t *testing.T) {
	h := NewHeap(intLess)
	h.Push(1)
	h.Push(2)
	c := h.Clone()
	assert.ElementsMatch(t, h.Values(), c.Values())
	h.Push(3)
	assert.NotEqual(t, h.Len(), c.Len())
}

func TestMinHeap(t *testing.T) {
	h := NewMinHeap[int]()
	h.Push(5)
	h.Push(1)
	h.Push(3)
	h.Push(2)
	h.Push(4)

	var got []int
	for {
		v, ok := h.Pop()
		if !ok {
			break
		}
		got = append(got, v)
	}
	assert.Equal(t, []int{1, 2, 3, 4, 5}, got) // ascending — min first
}

func TestMaxHeap(t *testing.T) {
	h := NewMaxHeap[int]()
	h.Push(5)
	h.Push(1)
	h.Push(3)
	h.Push(2)
	h.Push(4)

	var got []int
	for {
		v, ok := h.Pop()
		if !ok {
			break
		}
		got = append(got, v)
	}
	assert.Equal(t, []int{5, 4, 3, 2, 1}, got) // descending — max first
}

func TestMaxHeap_Strings(t *testing.T) {
	h := NewMaxHeap[string]() // cmp.Ordered covers strings
	h.Push("a")
	h.Push("c")
	h.Push("b")
	v, _ := h.Pop()
	assert.Equal(t, "c", v)
	v, _ = h.Pop()
	assert.Equal(t, "b", v)
	v, _ = h.Pop()
	assert.Equal(t, "a", v)
}

// --- Set ---

func TestSet_New(t *testing.T) {
	s := NewSet(1, 2, 3)
	assert.Equal(t, 3, s.Len())
	assert.True(t, s.Contains(1))
	assert.True(t, s.Contains(2))
	assert.True(t, s.Contains(3))
}

func TestSet_NewEmpty(t *testing.T) {
	s := NewSet[int]()
	assert.Equal(t, 0, s.Len())
	assert.True(t, s.IsEmpty())
}

func TestSet_NewWithCap(t *testing.T) {
	s := NewSetWithCap[int](100)
	assert.Equal(t, 0, s.Len())
}

func TestSet_Add(t *testing.T) {
	s := NewSet[int]()
	s.Add(1, 2, 3)
	assert.Equal(t, 3, s.Len())
	s.Add(2) // duplicate
	assert.Equal(t, 3, s.Len())
}

func TestSet_Add_Chained(t *testing.T) {
	s := NewSet[int]().Add(1, 2).Add(3)
	assert.Equal(t, 3, s.Len())
	assert.True(t, s.Contains(3))
}

func TestSet_Remove(t *testing.T) {
	s := NewSet(1, 2)
	s.Remove(1)
	assert.False(t, s.Contains(1))
	assert.True(t, s.Contains(2))
	s.Remove(99)
	assert.Equal(t, 1, s.Len())
}

func TestSet_Clear(t *testing.T) {
	s := NewSet(1, 2, 3)
	s.Clear()
	assert.True(t, s.IsEmpty())
	s.Add(4) // still usable
	assert.True(t, s.Contains(4))
}

func TestSet_Values(t *testing.T) {
	s := NewSet(1, 2)
	assert.ElementsMatch(t, []int{1, 2}, s.Values())
}

func TestSet_Clone(t *testing.T) {
	s := NewSet(1, 2)
	c := s.Clone()
	assert.True(t, s.Equal(c))
	s.Add(3)
	assert.False(t, s.Equal(c))
}

// --- set algebra (chained methods) ---

func TestSet_Union(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(2, 3)
	r := a.Union(b)
	assert.Equal(t, 3, r.Len())
	assert.True(t, r.Contains(1))
	assert.True(t, r.Contains(2))
	assert.True(t, r.Contains(3))
	assert.Equal(t, 2, a.Len()) // receiver untouched
}

func TestSet_Intersect(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(2, 3)
	r := a.Intersect(b)
	assert.Equal(t, 1, r.Len())
	assert.True(t, r.Contains(2))
}

func TestSet_Subtract(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(2, 3)
	r := a.Subtract(b)
	assert.True(t, r.Contains(1))
	assert.False(t, r.Contains(2))
	assert.Equal(t, 1, r.Len())
}

func TestSet_SymmetricDifference(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(2, 3)
	r := a.SymmetricDifference(b)
	assert.True(t, r.Contains(1))
	assert.True(t, r.Contains(3))
	assert.False(t, r.Contains(2))
	assert.Equal(t, 2, r.Len())
}

func TestSet_Chained(t *testing.T) {
	a := NewSet(1, 2, 3)
	b := NewSet(2, 3, 4)
	c := NewSet(3, 4, 5)
	// a ∪ b = {1,2,3,4}; then \ c = {1,2}
	r := a.Union(b).Subtract(c)
	assert.Equal(t, 2, r.Len())
	assert.True(t, r.Contains(1))
	assert.True(t, r.Contains(2))
}

func TestSet_Intersect_SmallerIterated(t *testing.T) {
	a := NewSet(1)
	b := NewSet(1, 2, 3, 4, 5)
	r := a.Intersect(b)
	assert.Equal(t, 1, r.Len())
}

// --- free-function aliases ---

func TestSet_FreeFunctions(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(2, 3)
	assert.True(t, Equal(UnionOf(a, b), NewSet(1, 2, 3)))
	assert.True(t, Equal(IntersectOf(a, b), NewSet(2)))
	assert.True(t, Equal(Difference(a, b), NewSet(1)))
	assert.True(t, Equal(SymmetricDifference(a, b), NewSet(1, 3)))
	assert.True(t, Equal(Clone(a), a))
}

// --- set relations ---

func TestSet_IsSubset(t *testing.T) {
	a := NewSet(1)
	b := NewSet(1, 2, 3)
	assert.True(t, a.IsSubset(b))
	assert.False(t, b.IsSubset(a))
	assert.True(t, NewSet[int]().IsSubset(b)) // empty is subset of any
}

func TestSet_IsSuperset(t *testing.T) {
	a := NewSet(1, 2, 3)
	b := NewSet(1)
	assert.True(t, a.IsSuperset(b))
	assert.False(t, b.IsSuperset(a))
}

func TestSet_IsDisjoint(t *testing.T) {
	a := NewSet(1)
	b := NewSet(2)
	assert.True(t, a.IsDisjoint(b))
	a.Add(2)
	assert.False(t, a.IsDisjoint(b))
	assert.True(t, NewSet[int]().IsDisjoint(b))
}

func TestSet_Equal(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(2, 1)
	assert.True(t, a.Equal(b))
	b.Add(3)
	assert.False(t, a.Equal(b))
	assert.False(t, NewSet(1, 2).Equal(NewSet(1, 2, 3)))
}

// --- functional ops ---

func TestSet_Find(t *testing.T) {
	s := NewSet(1, 2, 3)
	v, ok := Find(s, func(x int) bool { return x%2 == 0 })
	assert.True(t, ok)
	assert.Equal(t, 2, v)
	_, ok = Find(s, func(x int) bool { return x > 10 })
	assert.False(t, ok)
}

func TestSet_FindAll(t *testing.T) {
	s := NewSet(1, 2, 3, 4)
	r := FindAll(s, func(x int) bool { return x%2 == 0 })
	assert.Equal(t, 2, r.Len())
	assert.True(t, r.Contains(2))
	assert.True(t, r.Contains(4))
}

func TestSet_Partition(t *testing.T) {
	s := NewSet(1, 2, 3)
	even, odd := Partition(s, func(x int) bool { return x%2 == 0 })
	assert.Equal(t, 1, even.Len())
	assert.True(t, even.Contains(2))
	assert.Equal(t, 2, odd.Len())
	assert.True(t, odd.Contains(1))
	assert.True(t, odd.Contains(3))
}

func TestSet_Map(t *testing.T) {
	s := NewSet(1, 2, 3)
	r := Map(s, func(x int) int { return x * x })
	assert.Equal(t, 3, r.Len())
	assert.True(t, r.Contains(1))
	assert.True(t, r.Contains(4))
	assert.True(t, r.Contains(9))
}

func TestSet_Map_Duplicates(t *testing.T) {
	s := NewSet(1, -1, 2)
	r := Map(s, func(x int) int { return x * x })
	assert.Equal(t, 2, r.Len()) // 1 and -1 both square to 1
}

func TestSet_Reduce(t *testing.T) {
	s := NewSet(1, 2, 3)
	sum := Reduce(s, func(acc int, x int) int { return acc + x }, 0)
	assert.Equal(t, 6, sum)

	// U is any, not necessarily comparable
	concat := Reduce(s, func(acc string, x int) string { return acc + string(rune('0'+x)) }, "")
	assert.Equal(t, 3, len(concat))
}

func TestSet_ForEach(t *testing.T) {
	s := NewSet(1, 2, 3)
	sum := 0
	ForEach(s, func(x int) { sum += x })
	assert.Equal(t, 6, sum)
}

func TestSet_Any(t *testing.T) {
	s := NewSet(1, 2, 3)
	assert.True(t, Any(s, func(x int) bool { return x%2 == 0 }))
	assert.False(t, Any(s, func(x int) bool { return x > 10 }))
	assert.False(t, Any(NewSet[int](), func(x int) bool { return true }))
}

func TestSet_All(t *testing.T) {
	s := NewSet(1, 2, 3)
	assert.True(t, All(s, func(x int) bool { return x > 0 }))
	assert.False(t, All(s, func(x int) bool { return x%2 == 0 }))
	assert.True(t, All(NewSet[int](), func(x int) bool { return false })) // vacuous truth
}

func TestSet_None(t *testing.T) {
	s := NewSet(1, 2, 3)
	assert.True(t, None(s, func(x int) bool { return x > 10 }))
	assert.False(t, None(s, func(x int) bool { return x > 0 }))
	assert.True(t, None(NewSet[int](), func(x int) bool { return true }))
}

// --- iter.Seq integration ---

func TestIter_Stack(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)
	var got []int
	for v := range s.All() {
		got = append(got, v)
	}
	assert.Equal(t, []int{1, 2, 3}, got) // bottom-to-top
}

func TestIter_Stack_EarlyBreak(t *testing.T) {
	s := NewStack[int]()
	for i := 0; i < 5; i++ {
		s.Push(i)
	}
	n := 0
	for range s.All() {
		n++
		if n == 2 {
			break
		}
	}
	assert.Equal(t, 2, n)
}

func TestIter_Queue(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)
	var got []int
	for v := range q.All() {
		got = append(got, v)
	}
	assert.Equal(t, []int{1, 2, 3}, got) // front-to-back
}

func TestIter_Heap(t *testing.T) {
	h := NewMinHeap[int]()
	h.Push(3)
	h.Push(1)
	h.Push(2)
	got := slices.Collect(h.All()) // order is heap layout, not sorted
	assert.ElementsMatch(t, []int{1, 2, 3}, got)
}

func TestIter_Set(t *testing.T) {
	s := NewSet(1, 2, 3)
	got := slices.Collect(s.All())
	assert.ElementsMatch(t, []int{1, 2, 3}, got)
}

func TestIter_Set_EarlyBreak(t *testing.T) {
	s := NewSet(1, 2, 3, 4, 5)
	n := 0
	for range s.All() {
		n++
		if n == 3 {
			break
		}
	}
	assert.Equal(t, 3, n)
}

// --- benchmarks ---

func BenchmarkStack_PushPop(b *testing.B) {
	s := NewStack[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(i)
		s.Pop()
	}
}

func BenchmarkQueue_EnqueueDequeue(b *testing.B) {
	q := NewQueue[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i)
		q.Dequeue()
	}
}

// BenchmarkQueue_LongLived models a queue that stays alive across many
// enqueue/dequeue pairs — the exact workload v2 leaked under.
func BenchmarkQueue_LongLived(b *testing.B) {
	q := NewQueueWithCap[int](8)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i)
		q.Dequeue()
	}
}

func BenchmarkHeap_PushPop(b *testing.B) {
	h := NewMinHeap[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Push(i)
		h.Pop()
	}
}

func BenchmarkSet_Add(b *testing.B) {
	s := NewSetWithCap[int](b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}
}

func BenchmarkSet_Contains(b *testing.B) {
	s := NewSet[int]()
	for i := 0; i < 10000; i++ {
		s.Add(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Contains(i % 10000)
	}
}

func BenchmarkSet_Intersect(b *testing.B) {
	a := NewSet[int]()
	bb := NewSet[int]()
	for i := 0; i < 1000; i++ {
		a.Add(i)
		bb.Add(i + 500)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Intersect(bb)
	}
}
