package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewStack(t *testing.T) {
	stack := NewStack[int](10)
	assert.NotNil(t, stack)
	assert.Equal(t, 0, stack.Size())
}

func Test_StackPush(t *testing.T) {
	stack := NewStack[int](2)
	stack.Push(1)
	assert.Equal(t, 1, stack.Size())

	stack.Push(2)
	assert.Equal(t, 2, stack.Size())

	stack.Push(3)
	assert.Equal(t, 3, stack.Size())
}

func Test_StackPop(t *testing.T) {
	stack := NewStack[int](10)
	_, ok := stack.Pop()
	assert.False(t, ok)

	// push: 1, 2, 3
	// pop: 3, 2, 1

	stack.Push(1)
	stack.Push(2)
	stack.Push(3)

	val, ok := stack.Pop()
	assert.True(t, ok)
	assert.Equal(t, 3, val)
	assert.Equal(t, 2, stack.Size())

	val, ok = stack.Pop()
	assert.True(t, ok)
	assert.Equal(t, 2, val)
	assert.Equal(t, 1, stack.Size())

	val, ok = stack.Pop()
	assert.True(t, ok)
	assert.Equal(t, 1, val)
	assert.Equal(t, 0, stack.Size())

	_, ok = stack.Pop()
	assert.False(t, ok)
}

func Test_StackPeek(t *testing.T) {
	stack := NewStack[int](10)
	_, ok := stack.Peek()
	assert.False(t, ok)

	stack.Push(1)
	stack.Push(2)

	val, ok := stack.Peek()
	assert.True(t, ok)
	assert.Equal(t, 2, val)

	stack.Pop()

	val, ok = stack.Peek()
	assert.True(t, ok)
	assert.Equal(t, 1, val)
}

func Test_StackSize(t *testing.T) {
	stack := NewStack[int](10)
	assert.Equal(t, 0, stack.Size())

	stack.Push(1)
	stack.Push(2)
	assert.Equal(t, 2, stack.Size())

	stack.Pop()
	assert.Equal(t, 1, stack.Size())

	stack.Clear()
	assert.Equal(t, 0, stack.Size())
}

func Test_StackIsEmpty(t *testing.T) {
	stack := NewStack[int](10)
	assert.True(t, stack.IsEmpty())

	stack.Push(1)
	assert.False(t, stack.IsEmpty())

	stack.Pop()
	assert.True(t, stack.IsEmpty())
}

func Test_StackClear(t *testing.T) {
	stack := NewStack[int](10)
	stack.Push(1)
	stack.Push(2)
	assert.Equal(t, 2, stack.Size())

	stack.Clear()
	assert.Equal(t, 0, stack.Size())
	assert.True(t, stack.IsEmpty())
}

func Test_StackValues(t *testing.T) {
	stack := NewStack[int](10)
	stack.Push(1)
	stack.Push(2)
	values := stack.Values()
	assert.ElementsMatch(t, []int{1, 2}, values)
	assert.Equal(t, 2, len(values))
}

func Test_StackString(t *testing.T) {
	stack := NewStack[int](10)
	stack.Push(1)
	stack.Push(2)
	assert.Equal(t, "[1 2]", stack.String())
}

func Test_StackClone(t *testing.T) {
	stack := NewStack[int](10)
	stack.Push(1)
	stack.Push(2)

	copyStack := stack.Clone()
	assert.NotSame(t, stack, copyStack)
	assert.ElementsMatch(t, stack.Values(), copyStack.Values())

	stack.Push(3)
	assert.NotEqual(t, stack.Size(), copyStack.Size())
}
