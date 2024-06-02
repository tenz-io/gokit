package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewQueue(t *testing.T) {
	queue := NewQueue[int](10)
	assert.NotNil(t, queue)
	assert.Equal(t, 0, queue.Size())
}

func Test_QueueEnqueue(t *testing.T) {
	queue := NewQueue[int](2)
	queue.Enqueue(1)
	assert.Equal(t, 1, queue.Size())

	queue.Enqueue(2)
	assert.Equal(t, 2, queue.Size())

	queue.Enqueue(3)
	assert.Equal(t, 3, queue.Size())
}

func Test_QueueDequeue(t *testing.T) {
	queue := NewQueue[int](10)
	_, ok := queue.Dequeue()
	assert.False(t, ok)

	queue.Enqueue(1)
	queue.Enqueue(2)
	val, ok := queue.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, 1, val)
	assert.Equal(t, 1, queue.Size())

	val, ok = queue.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, 2, val)
	assert.Equal(t, 0, queue.Size())

	_, ok = queue.Dequeue()
	assert.False(t, ok)

}

func Test_QueuePeek(t *testing.T) {
	queue := NewQueue[int](10)
	_, ok := queue.Peek()
	assert.False(t, ok)

	queue.Enqueue(1)
	queue.Enqueue(2)

	val, ok := queue.Peek()
	assert.True(t, ok)
	assert.Equal(t, 1, val)
	assert.Equal(t, 2, queue.Size())

	queue.Dequeue()

	val, ok = queue.Peek()
	assert.True(t, ok)
	assert.Equal(t, 2, val)
	assert.Equal(t, 1, queue.Size())
}

func Test_QueueSize(t *testing.T) {
	queue := NewQueue[int](10)
	assert.Equal(t, 0, queue.Size())

	queue.Enqueue(1)
	queue.Enqueue(2)
	assert.Equal(t, 2, queue.Size())

	queue.Dequeue()
	assert.Equal(t, 1, queue.Size())

	queue.Clear()
	assert.Equal(t, 0, queue.Size())
}

func Test_QueueIsEmpty(t *testing.T) {
	queue := NewQueue[int](10)
	assert.True(t, queue.IsEmpty())

	queue.Enqueue(1)
	assert.False(t, queue.IsEmpty())

	queue.Dequeue()
	assert.True(t, queue.IsEmpty())
}

func Test_QueueClear(t *testing.T) {
	queue := NewQueue[int](10)
	queue.Enqueue(1)
	queue.Enqueue(2)
	assert.Equal(t, 2, queue.Size())

	queue.Clear()
	assert.Equal(t, 0, queue.Size())
	assert.True(t, queue.IsEmpty())
}

func Test_QueueValues(t *testing.T) {
	queue := NewQueue[int](10)
	queue.Enqueue(1)
	queue.Enqueue(2)
	values := queue.Values()
	assert.ElementsMatch(t, []int{1, 2}, values)
	assert.Equal(t, 2, len(values))
}

func Test_QueueString(t *testing.T) {
	queue := NewQueue[int](10)
	queue.Enqueue(1)
	queue.Enqueue(2)
	assert.Equal(t, "[1 2]", queue.String())
}

func Test_QueueClone(t *testing.T) {
	queue := NewQueue[int](10)
	queue.Enqueue(1)
	queue.Enqueue(2)

	copyQueue := queue.Clone()
	assert.NotSame(t, queue, copyQueue)
	assert.ElementsMatch(t, queue.Values(), copyQueue.Values())

	queue.Enqueue(3)
	assert.NotEqual(t, queue.Size(), copyQueue.Size())
}
