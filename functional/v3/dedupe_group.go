package fn

// --- Deduplicate ---

// Deduplicate returns a new slice with duplicates removed, preserving the
// order of first occurrence. The input slice is not modified.
func Deduplicate[T comparable](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	seen := make(map[T]struct{}, len(s))
	out := make([]T, 0, len(s))
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// DeduplicateBy returns a new slice with elements whose key (under keyFn) is
// duplicated removed, preserving the order of first occurrence. The input
// slice is not modified.
func DeduplicateBy[T any, K comparable](s []T, keyFn func(T) K) []T {
	if len(s) == 0 {
		return nil
	}
	seen := make(map[K]struct{}, len(s))
	out := make([]T, 0, len(s))
	for _, v := range s {
		k := keyFn(v)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, v)
	}
	return out
}

// DeduplicateInPlace removes duplicates from s in place, preserving first
// occurrence, and returns the (shortened) prefix s[:k]. It reuses s's backing
// array. The dropped tail is zeroed to avoid retaining live references.
func DeduplicateInPlace[T comparable](s []T) []T {
	if len(s) == 0 {
		return s[:0]
	}
	seen := make(map[T]struct{}, len(s))
	k := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[k] = v
		k++
	}
	clear(s[k:])
	return s[:k]
}

// DeduplicateByInPlace is the in-place variant of DeduplicateBy.
func DeduplicateByInPlace[T any, K comparable](s []T, keyFn func(T) K) []T {
	if len(s) == 0 {
		return s[:0]
	}
	seen := make(map[K]struct{}, len(s))
	k := 0
	for _, v := range s {
		key := keyFn(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		s[k] = v
		k++
	}
	clear(s[k:])
	return s[:k]
}

// Unique and UniqueBy are aliases for Deduplicate and DeduplicateBy for
// callers that prefer the more common name.
//
// Deprecated alias: prefer Deduplicate for clarity; Unique is kept as a
// convenience for readers coming from other languages' stdlibs.
func Unique[T comparable](s []T) []T { return Deduplicate(s) }

// UniqueBy is the alias for DeduplicateBy.
func UniqueBy[T any, K comparable](s []T, keyFn func(T) K) []T {
	return DeduplicateBy(s, keyFn)
}

// --- GroupBy / Partition ---

// GroupBy partitions s into a map keyed by keyFn(element). Within each group,
// elements keep their original relative order. Returns an empty (non-nil) map
// for an empty input.
func GroupBy[T any, K comparable](s []T, keyFn func(T) K) map[K][]T {
	groups := make(map[K][]T, len(s))
	for _, v := range s {
		k := keyFn(v)
		groups[k] = append(groups[k], v)
	}
	return groups
}

// GroupByCount returns a map from key to the number of elements with that key.
// It is a specialization of GroupBy for counting workloads, avoiding slice
// allocation per group.
func GroupByCount[T any, K comparable](s []T, keyFn func(T) K) map[K]int {
	groups := make(map[K]int, len(s))
	for _, v := range s {
		groups[keyFn(v)]++
	}
	return groups
}

// Partition splits s into two slices: elements for which pred returns true
// (matched) and the rest (unmatched), both preserving input order.
func Partition[T any](s []T, pred func(T) bool) (matched, unmatched []T) {
	matched = make([]T, 0, (len(s)+1)/2)
	unmatched = make([]T, 0, (len(s)+1)/2)
	for _, v := range s {
		if pred(v) {
			matched = append(matched, v)
		} else {
			unmatched = append(unmatched, v)
		}
	}
	return matched, unmatched
}

// PartitionInPlace partitions s in place so that all elements for which pred
// returns true come first, followed by the rest, and returns the count of
// matched elements (the index where the unmatched tail begins). It is a
// stable partition: relative order is preserved within each side.
//
// The matched prefix s[:k] is compacted in place; the unmatched tail s[k:]
// holds the remaining elements in original order. A small auxiliary buffer is
// used only for the unmatched half so the matched-prefix compaction can run in
// a single forward pass; both halves end up in s's backing array.
func PartitionInPlace[T any](s []T, pred func(T) bool) int {
	n := len(s)
	unmatched := make([]T, 0, n)
	k := 0
	for i := 0; i < n; i++ {
		if pred(s[i]) {
			s[k] = s[i]
			k++
		} else {
			unmatched = append(unmatched, s[i])
		}
	}
	copy(s[k:], unmatched)
	return k
}
