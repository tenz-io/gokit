// Example: collection/v3 — Stack, Queue (ring buffer), Heap, and Set.
package main

import (
	"fmt"

	"github.com/tenz-io/gokit/collection/v3"
)

func main() {
	stack()
	queue()
	heap()
	setAlgebra()
}

func stack() {
	// Bracket-matching via a LIFO stack.
	const expr = "{[()()]}"
	pairs := map[rune]rune{')': '(', ']': '[', '}': '{'}
	s := collection.NewStack[rune]()
	balanced := true
	for _, r := range expr {
		switch r {
		case '(', '[', '{':
			s.Push(r)
		case ')', ']', '}':
			top, ok := s.Pop()
			if !ok || top != pairs[r] {
				balanced = false
			}
		}
	}
	fmt.Printf("Stack:   %q balanced=%v\n", expr, balanced && s.IsEmpty())

	// range over All(): bottom-to-top order.
	si := collection.NewStack[int]()
	si.Push(1)
	si.Push(2)
	si.Push(3)
	var vals []int
	for v := range si.All() {
		vals = append(vals, v)
	}
	fmt.Printf("         stack.All() = %v\n", vals)
}

func queue() {
	// A long-lived task buffer: enqueue/dequeue many pairs. The ring buffer
	// recycles slots, so memory stays bounded.
	q := collection.NewQueue[int]()
	for i := 0; i < 1_000_000; i++ {
		q.Enqueue(i)
		q.Dequeue()
	}
	fmt.Printf("Queue:   drained %d enqueue/dequeue pairs, len=%d\n", 1_000_000, q.Len())
}

func heap() {
	// Min-heap over ints: drains smallest-first.
	h := collection.NewMinHeap[int]()
	for _, v := range []int{5, 1, 3, 2, 4} {
		h.Push(v)
	}
	var sorted []int
	for {
		v, ok := h.Pop()
		if !ok {
			break
		}
		sorted = append(sorted, v)
	}
	fmt.Printf("Heap:    min-heap drains -> %v\n", sorted)

	// Max-heap over strings (cmp.Ordered).
	mh := collection.NewMaxHeap[string]()
	mh.Push("go")
	mh.Push("rust")
	mh.Push("zig")
	fmt.Printf("         max-heap top = %q\n", mustTop(mh))
}

func mustTop(h *collection.Heap[string]) string {
	v, _ := h.Peek()
	return v
}

func setAlgebra() {
	// Chained set algebra: roles of users.
	devs := collection.NewSet("alice", "bob", "carol")
	qa := collection.NewSet("bob", "carol", "dave")
	ops := collection.NewSet("carol", "dave", "erin")

	// Everyone who touches anything, minus pure ops (no dev/qa overlap):
	cross := devs.Union(qa).Subtract(ops)
	fmt.Printf("Set:     dev∪qa \\ ops = %v\n", cross.Values())

	// Relations: is QA a subset of the union of dev+ops?
	all := devs.Union(ops)
	fmt.Printf("         qa ⊆ dev∪ops? %v\n", qa.IsSubset(all))
	fmt.Printf("         devs ∩ qa == {bob,carol}? %v\n",
		devs.Intersect(qa).Equal(collection.NewSet("bob", "carol")))

	// range over All() with early break.
	n := 0
	for range cross.All() {
		n++
		if n == 10 {
			break
		}
	}
	fmt.Printf("         stopped after %d elements\n", n)
}
