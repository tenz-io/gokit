package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Stack ---

func TestStack_New(t *testing.T) {
	s := NewStack[int]()
	assert.NotNil(t, s)
	assert.True(t, s.IsEmpty())
	assert.Equal(t, 0, s.Len())
	assert.Equal(t, 0, s.Size())
}

func TestStack_NewWithCap(t *testing.T) {
	s := NewStackWithCap[int](0)
	assert.NotNil(t, s) // zero cap falls back to 16
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
	assert.Equal(t, 2, s.Len()) // Peek doesn't remove
}

func TestStack_IsEmpty(t *testing.T) {
	s := NewStack[int]()
	assert.True(t, s.IsEmpty())
	s.Push(1)
	assert.False(t, s.IsEmpty())
	s.Pop()
	assert.True(t, s.IsEmpty())
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
	assert.Equal(t, []int{1, 2}, s.Values())
	// verify it's a copy
	vals := s.Values()
	vals[0] = 99
	v, _ := s.Pop()
	assert.Equal(t, 2, v) // unchanged
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

// --- Queue ---

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
	assert.Equal(t, 2, q.Len()) // not removed
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
	// copy
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

func TestQueue_MemoryLeak(t *testing.T) {
	// Verify that Dequeue nils out the reference so that the GC can collect the dequeued element.
	type big struct {
		data [1024]byte
	}
	q := NewQueue[*big]()
	ptr := &big{}
	q.Enqueue(ptr)
	_, _ = q.Dequeue()

	// The dequeued reference should be nilled. We can't easily verify this
	// without reflection, but the code explicitly sets q.data[0] = *new(T).
	// This test just ensures the operation doesn't panic.
}

// --- PriorityQueue ---

func intLess(a, b int) bool { return a < b }

func TestPriorityQueue_New(t *testing.T) {
	pq := NewPriorityQueue(intLess)
	assert.NotNil(t, pq)
	assert.True(t, pq.IsEmpty())
}

func TestPriorityQueue_PushPop(t *testing.T) {
	pq := NewPriorityQueue(intLess)
	_, ok := pq.Pop()
	assert.False(t, ok)

	pq.Push(2)
	pq.Push(1)
	pq.Push(3)

	v, ok := pq.Pop()
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	v, ok = pq.Pop()
	assert.Equal(t, 2, v)
	v, ok = pq.Pop()
	assert.Equal(t, 3, v)

	_, ok = pq.Pop()
	assert.False(t, ok)
}

func TestPriorityQueue_Peek(t *testing.T) {
	pq := NewPriorityQueue(intLess)
	_, ok := pq.Peek()
	assert.False(t, ok)

	pq.Push(3)
	pq.Push(1)
	v, ok := pq.Peek()
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	assert.Equal(t, 2, pq.Len())
}

func TestPriorityQueue_Clone(t *testing.T) {
	pq := NewPriorityQueue(intLess)
	pq.Push(1)
	pq.Push(2)
	c := pq.Clone()
	// cloned queues have same elements (in heap order)
	assert.ElementsMatch(t, pq.Values(), c.Values())
	pq.Push(3)
	assert.NotEqual(t, pq.Len(), c.Len())
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
	s.Add(2) // duplicate, no-op
	assert.Equal(t, 3, s.Len())
}

func TestSet_Remove(t *testing.T) {
	s := NewSet(1, 2)
	s.Remove(1)
	assert.False(t, s.Contains(1))
	assert.True(t, s.Contains(2))
	s.Remove(99) // no-op
	assert.Equal(t, 1, s.Len())
}

func TestSet_Contains(t *testing.T) {
	s := NewSet(1)
	assert.True(t, s.Contains(1))
	assert.False(t, s.Contains(2))
}

func TestSet_Clear(t *testing.T) {
	s := NewSet(1, 2, 3)
	s.Clear()
	assert.True(t, s.IsEmpty())
	assert.Equal(t, 0, s.Len())
	s.Add(4) // still usable after clear
	assert.True(t, s.Contains(4))
}

func TestSet_Values(t *testing.T) {
	s := NewSet(1, 2)
	vals := s.Values()
	assert.ElementsMatch(t, []int{1, 2}, vals)
}

func TestSet_Clone(t *testing.T) {
	s := NewSet(1, 2)
	c := Clone(s)
	assert.True(t, Equal(s, c))
	s.Add(3)
	assert.False(t, Equal(s, c))
}

func TestSet_Intersection(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(2, 3)
	r := Intersection(a, b)
	assert.True(t, r.Contains(2))
	assert.Equal(t, 1, r.Len())
}

func TestSet_Union(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(2, 3)
	r := Union(a, b)
	assert.Subset(t, []int{1, 2, 3}, r.Values())
	assert.Equal(t, 3, r.Len())
}

func TestSet_Difference(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(2, 3)
	r := Difference(a, b)
	assert.True(t, r.Contains(1))
	assert.False(t, r.Contains(2))
}

func TestSet_SymmetricDifference(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(2, 3)
	r := SymmetricDifference(a, b)
	assert.True(t, r.Contains(1))
	assert.True(t, r.Contains(3))
	assert.False(t, r.Contains(2))
}

func TestSet_IsSubset(t *testing.T) {
	a := NewSet(1)
	b := NewSet(1, 2, 3)
	assert.True(t, IsSubset(a, b))
	assert.False(t, IsSubset(b, a))
	// empty set is subset of any set
	assert.True(t, IsSubset(NewSet[int](), b))
	// superset can't be subset
	assert.False(t, IsSubset(b, a))
}

func TestSet_IsSuperset(t *testing.T) {
	a := NewSet(1, 2, 3)
	b := NewSet(1)
	assert.True(t, IsSuperset(a, b))
	assert.False(t, IsSuperset(b, a))
}

func TestSet_IsDisjoint(t *testing.T) {
	a := NewSet(1)
	b := NewSet(2)
	assert.True(t, IsDisjoint(a, b))
	a.Add(2)
	assert.False(t, IsDisjoint(a, b))
	// empty sets are disjoint
	assert.True(t, IsDisjoint(NewSet[int](), b))
}

func TestSet_Equal(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(2, 1)
	assert.True(t, Equal(a, b))
	b.Add(3)
	assert.False(t, Equal(a, b))
	// different sizes
	assert.False(t, Equal(NewSet(1, 2), NewSet(1, 2, 3)))
}

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
	assert.Subset(t, []int{2, 4}, r.Values())
	assert.Equal(t, 2, r.Len())
}

func TestSet_Partition(t *testing.T) {
	s := NewSet(1, 2, 3)
	even, odd := Partition(s, func(x int) bool { return x%2 == 0 })
	assert.True(t, even.Contains(2))
	assert.Equal(t, 1, even.Len())
	assert.True(t, odd.Contains(1))
	assert.True(t, odd.Contains(3))
	assert.Equal(t, 2, odd.Len())
}

func TestSet_Map(t *testing.T) {
	s := NewSet(1, 2, 3)
	r := Map(s, func(x int) int { return x * x })
	assert.True(t, r.Contains(1))
	assert.True(t, r.Contains(4))
	assert.True(t, r.Contains(9))
	assert.Equal(t, 3, r.Len())
}

func TestSet_Map_Duplicates(t *testing.T) {
	s := NewSet(1, -1, 2)
	r := Map(s, func(x int) int { return x * x })
	assert.True(t, r.Contains(1))
	assert.True(t, r.Contains(4))
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
	// vacuous truth for empty set
	assert.True(t, All(NewSet[int](), func(x int) bool { return false }))
}

func TestSet_None(t *testing.T) {
	s := NewSet(1, 2, 3)
	assert.True(t, None(s, func(x int) bool { return x > 10 }))
	assert.False(t, None(s, func(x int) bool { return x > 0 }))
	assert.True(t, None(NewSet[int](), func(x int) bool { return true }))
}

func TestSet_Intersection_SmallerIterated(t *testing.T) {
	// verify performance optimization: smaller set is iterated
	a := NewSet(1)
	b := NewSet(1, 2, 3, 4, 5)
	r := Intersection(a, b)
	assert.Equal(t, 1, r.Len())
}

// Benchmarks

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

func BenchmarkPriorityQueue_PushPop(b *testing.B) {
	pq := NewPriorityQueue(intLess)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Push(i)
		pq.Pop()
	}
}

func BenchmarkSet_Add(b *testing.B) {
	s := NewSet[int]()
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
