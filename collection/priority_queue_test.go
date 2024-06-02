package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func intLess(a, b int) bool {
	return a < b
}

func Test_NewPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue(2, intLess)
	assert.NotNil(t, pq)
	assert.Equal(t, 0, pq.Size())
}

func Test_PriorityQueuePush(t *testing.T) {
	pq := NewPriorityQueue(2, intLess)
	pq.Push(1)
	assert.Equal(t, 1, pq.Size())

	pq.Push(2)
	assert.Equal(t, 2, pq.Size())

	pq.Push(3)
	assert.Equal(t, 3, pq.Size())
}

func Test_PriorityQueuePop(t *testing.T) {
	pq := NewPriorityQueue(2, intLess)
	_, ok := pq.Pop()
	assert.False(t, ok)

	// push 2, 1, 3
	// pop 1, 2, 3

	pq.Push(2)
	pq.Push(1)
	pq.Push(3)
	val, ok := pq.Pop()
	assert.True(t, ok)
	assert.Equal(t, 1, val)
	assert.Equal(t, 2, pq.Size())

	val, ok = pq.Pop()
	assert.True(t, ok)
	assert.Equal(t, 2, val)
	assert.Equal(t, 1, pq.Size())

	val, ok = pq.Pop()
	assert.True(t, ok)
	assert.Equal(t, 3, val)

	val, ok = pq.Pop()
	assert.False(t, ok)
	assert.Equal(t, 0, pq.Size())

}

func Test_PriorityQueuePeek(t *testing.T) {
	pq := NewPriorityQueue(2, intLess)
	_, ok := pq.Peek()
	assert.False(t, ok)

	pq.Push(2)
	pq.Push(1)
	val, ok := pq.Peek()
	assert.True(t, ok)
	assert.Equal(t, 1, val)
	assert.Equal(t, 2, pq.Size())

	pq.Pop()
	val, ok = pq.Peek()
	assert.True(t, ok)
	assert.Equal(t, 2, val)
	assert.Equal(t, 1, pq.Size())
}

func Test_PriorityQueueSize(t *testing.T) {
	pq := NewPriorityQueue(2, intLess)
	assert.Equal(t, 0, pq.Size())

	pq.Push(1)
	pq.Push(2)
	assert.Equal(t, 2, pq.Size())

	pq.Pop()
	assert.Equal(t, 1, pq.Size())

	pq.Clear()
	assert.Equal(t, 0, pq.Size())
}

func Test_PriorityQueueIsEmpty(t *testing.T) {
	pq := NewPriorityQueue(2, intLess)
	assert.True(t, pq.IsEmpty())

	pq.Push(1)
	assert.False(t, pq.IsEmpty())

	pq.Pop()
	assert.True(t, pq.IsEmpty())
}

func Test_PriorityQueueClear(t *testing.T) {
	pq := NewPriorityQueue(2, intLess)
	pq.Push(1)
	pq.Push(2)
	assert.Equal(t, 2, pq.Size())

	pq.Clear()
	assert.Equal(t, 0, pq.Size())
	assert.True(t, pq.IsEmpty())
}

func Test_PriorityQueueValues(t *testing.T) {
	pq := NewPriorityQueue(2, intLess)
	pq.Push(1)
	pq.Push(2)
	values := pq.Values()
	assert.ElementsMatch(t, []int{1, 2}, values)
	assert.Equal(t, 2, len(values))
}

func Test_PriorityQueueString(t *testing.T) {
	pq := NewPriorityQueue(2, intLess)
	pq.Push(1)
	pq.Push(2)
	assert.Equal(t, "[1 2]", pq.String())
}

func Test_PriorityQueueClone(t *testing.T) {
	pq := NewPriorityQueue(10, intLess)
	pq.Push(1)
	pq.Push(2)

	copyPQ := pq.Clone()
	assert.NotSame(t, pq, copyPQ)
	assert.ElementsMatch(t, pq.Values(), copyPQ.Values())

	pq.Push(3)
	t.Logf("pq: %v, copyPQ: %v", pq.Values(), copyPQ.Values())
	assert.NotEqual(t, pq.Size(), copyPQ.Size())
}
