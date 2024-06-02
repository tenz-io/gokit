package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewSet(t *testing.T) {
	set := NewSet[int]()
	assert.NotNil(t, set)
	assert.Equal(t, 0, set.Size())
}

func Test_SetAdd(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	assert.True(t, set.Contains(1))
	assert.Equal(t, 1, set.Size())
}

func Test_SetRemove(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Remove(1)
	assert.False(t, set.Contains(1))
	assert.Equal(t, 0, set.Size())
}

func Test_SetContains(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	assert.True(t, set.Contains(1))
	assert.False(t, set.Contains(2))
}

func Test_SetSize(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	assert.Equal(t, 2, set.Size())
}

func Test_SetClear(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Clear()
	assert.Equal(t, 0, set.Size())
}

func Test_SetValues(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	values := set.Values()
	assert.ElementsMatch(t, []int{1, 2}, values)
}

func Test_SetIntersection(t *testing.T) {
	a := NewSet[int]()
	a.Add(1)
	a.Add(2)
	b := NewSet[int]()
	b.Add(2)
	b.Add(3)
	result := Intersection(a, b)
	assert.True(t, result.Contains(2))
	assert.False(t, result.Contains(1))
	assert.False(t, result.Contains(3))
}

func Test_SetUnion(t *testing.T) {
	a := NewSet[int]()
	a.Add(1)
	a.Add(2)
	b := NewSet[int]()
	b.Add(2)
	b.Add(3)
	result := Union(a, b)
	assert.True(t, result.Contains(1))
	assert.True(t, result.Contains(2))
	assert.True(t, result.Contains(3))
}

func Test_SetDifference(t *testing.T) {
	a := NewSet[int]()
	a.Add(1)
	a.Add(2)
	b := NewSet[int]()
	b.Add(2)
	b.Add(3)
	result := Difference(a, b)
	assert.True(t, result.Contains(1))
	assert.False(t, result.Contains(2))
	assert.False(t, result.Contains(3))
}

func Test_SetSymmetricDifference(t *testing.T) {
	a := NewSet[int]()
	a.Add(1)
	a.Add(2)
	b := NewSet[int]()
	b.Add(2)
	b.Add(3)
	result := SymmetricDifference(a, b)
	assert.True(t, result.Contains(1))
	assert.False(t, result.Contains(2))
	assert.True(t, result.Contains(3))
}

func Test_SetIsSubset(t *testing.T) {
	a := NewSet[int]()
	a.Add(1)
	a.Add(2)
	b := NewSet[int]()
	b.Add(1)
	b.Add(2)
	b.Add(3)
	assert.True(t, IsSubset(a, b))
	assert.False(t, IsSubset(b, a))
}

func Test_SetIsSuperset(t *testing.T) {
	a := NewSet[int]()
	a.Add(1)
	a.Add(2)
	b := NewSet[int]()
	b.Add(1)
	b.Add(2)
	b.Add(3)
	assert.False(t, IsSuperset(a, b))
	assert.True(t, IsSuperset(b, a))
}

func Test_SetIsDisjoint(t *testing.T) {
	a := NewSet[int]()
	a.Add(1)
	b := NewSet[int]()
	b.Add(2)
	assert.True(t, IsDisjoint(a, b))
	a.Add(2)
	assert.False(t, IsDisjoint(a, b))
}

func Test_SetEqual(t *testing.T) {
	a := NewSet[int]()
	a.Add(1)
	a.Add(2)
	b := NewSet[int]()
	b.Add(1)
	b.Add(2)
	assert.True(t, Equal(a, b))
	b.Add(3)
	assert.False(t, Equal(a, b))
}

func Test_SetClone(t *testing.T) {
	a := NewSet[int]()
	a.Add(1)
	a.Add(2)
	b := Clone(a)
	assert.True(t, Equal(a, b))
	a.Add(3)
	assert.False(t, Equal(a, b))
}

func Test_SetFilter(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(3)
	result := Filter(set, func(x int) bool { return x%2 == 0 })
	assert.True(t, result.Contains(2))
	assert.False(t, result.Contains(1))
	assert.False(t, result.Contains(3))
}

func Test_SetMap(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(3)
	result := Map(set, func(x int) int { return x * x })
	assert.True(t, result.Contains(1))
	assert.True(t, result.Contains(4))
	assert.True(t, result.Contains(9))
}

func Test_SetReduce(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(3)
	result := Reduce(set, func(acc, x int) int { return acc + x }, 0)
	assert.Equal(t, 6, result)
}

func Test_SetForEach(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(3)
	sum := 0
	ForEach(set, func(x int) { sum += x })
	assert.Equal(t, 6, sum)
}

func Test_SetAny(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(3)
	assert.True(t, Any(set, func(x int) bool { return x%2 == 0 }))
	assert.False(t, Any(set, func(x int) bool { return x > 3 }))
}

func Test_SetAll(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(3)
	assert.True(t, All(set, func(x int) bool { return x > 0 }))
	assert.False(t, All(set, func(x int) bool { return x%2 == 0 }))
}

func Test_SetNone(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(3)
	assert.True(t, None(set, func(x int) bool { return x > 3 }))
	assert.False(t, None(set, func(x int) bool { return x > 0 }))
}

func Test_SetFind(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(3)
	result, found := Find(set, func(x int) bool { return x%2 == 0 })
	assert.True(t, found)
	assert.Equal(t, 2, result)
	result, found = Find(set, func(x int) bool { return x > 3 })
	assert.False(t, found)
}

func Test_SetFindAll(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(3)
	result := FindAll(set, func(x int) bool { return x%2 == 0 })
	assert.True(t, result.Contains(2))
	assert.False(t, result.Contains(1))
	assert.False(t, result.Contains(3))
}

func Test_SetPartition(t *testing.T) {
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(3)
	even, odd := Partition(set, func(x int) bool { return x%2 == 0 })
	assert.True(t, even.Contains(2))
	assert.False(t, even.Contains(1))
	assert.False(t, even.Contains(3))
	assert.True(t, odd.Contains(1))
	assert.True(t, odd.Contains(3))
	assert.False(t, odd.Contains(2))
}
