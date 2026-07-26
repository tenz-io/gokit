// Package collection provides generic, allocation-aware data structures:
// [Stack], [Queue], [Heap], and [Set].
//
// V3 is a from-scratch rewrite of the collection module and is NOT compatible
// with v2. The design follows three goals — reasonable, efficient, usable —
// realized through:
//
//   - Better tables: [Queue] is a ring buffer (fixing v2's slice[1:] memory
//     leak); [Heap] exposes [NewMinHeap]/[NewMaxHeap] for [cmp.Ordered]
//     types; [Set] hides its backing map behind a struct so the layout can
//     evolve without breaking callers.
//   - Better idioms: every container exposes [Stack.All]/[Queue.All]/
//     [Heap.All]/[Set.All] returning an [iter.Seq], so `for v := range s.All()`
//     works and the containers compose with the standard library's
//     [slices]/[maps] packages. [Stack.Pop]/[Queue.Dequeue]/[Heap.Peek]
//     return (T, bool); zero values are cleared to let the GC reclaim
//     references.
//   - Easier use: [Set] supports method-chained algebra
//     (`a.Union(b).Intersect(c)`) alongside free-function aliases, and [Heap]
//     has cmp.Ordered constructors so callers rarely write a `less` func.
//
// Module boundary:
//
//	Order-preserving dedup / membership → functional/v3.OrderedSet.
//	Unordered set algebra (union/intersect/diff) → this package's Set.
//	Slice-level FP (Map/Filter/Reduce on []T) → functional/v3.
//
// All operations are pointer-receiver methods; copy a container with Clone
// before mutating if you need to share state.
package collection
