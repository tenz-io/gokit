// Package fn provides generic functional-programming utilities for Go slices.
//
// V3 is a from-scratch rewrite of the functional toolkit. It is NOT compatible
// with v2. The design follows three goals — reasonable, efficient, usable —
// realized through a dual-track API:
//
//   - Standalone functions: a strict superset of v2's surface (Map, Filter,
//     Reduce, TopK, ...) with index-aware variants (MapIdx, FindIndex),
//     in-place/zero-allocation variants (MapInPlace, FilterInPlace), and
//     key-extractor-based aggregation (TopK/BottomK with a Key extractor).
//   - Fluent Chain: ChainOf(s).Map(...).Filter(...).TopK(k, key).Collect()
//     materializes a slice per step — typically faster and far easier to debug
//     than a chain of closures in Go.
//   - Lazy Seq: a callback-style iterator (compatible with the spirit of the
//     future iter.Seq) for short-circuit reads (Any/All/Find) over large or
//     generated inputs without allocating a materialized slice.
//
// Ordering model:
//
//	TopK returns the k largest elements in descending order.
//	BottomK returns the k smallest elements in ascending order.
//
// Key extractors vs comparators:
//
//	Key[T, K cmp.Ordered] extracts an ordered key — the common case (top-k by
//	score, min/max by id). Use it for TopK/BottomK/MinByKey/MaxByKey.
//	By[T] is a cmp-style comparator func(a, b T) int for the *By variants when
//	you need multi-field or non-key ordering. Prefer Key when a single ordered
//	field suffices.
//
// All operations are pure unless their name ends with InPlace. In-place
// variants reuse the input slice's backing array and return a (possibly
// shortened) view of it; they are the zero-allocation hot path.
package fn

import "cmp"

// Key is a function that extracts an ordered key from a value of type T.
//
// It is the primary input to the key-based aggregation functions
// (TopK/BottomK/MinByKey/MaxByKey). Prefer it over By when the ordering
// derives from a single cmp.Ordered field.
//
// Example:
//
//	fn.TopK(users, 10, fn.Key[User, int](func(u User) int { return u.Score }))
type Key[T any, K cmp.Ordered] func(T) K

// By is a cmp-style comparator: it returns a negative int when a < b, zero
// when equal, and a positive int when a > b — matching [cmp.Compare] and
// [slices.SortFunc] conventions.
//
// Use it for the *By variants (TopKBy/BottomKBy/MinBy/MaxBy) when ordering
// cannot be expressed by a single ordered key (multi-field, custom ranking).
type By[T any] func(a, b T) int

// Pair groups two values of (possibly) different types.
//
// It is returned by Zip and is the element type of a zipped slice.
type Pair[A, B any] struct {
	A A
	B B
}
