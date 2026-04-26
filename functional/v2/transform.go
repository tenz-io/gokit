package function

// Deduplicate removes duplicate elements while preserving order of first occurrence.
func Deduplicate[T comparable](list []T) []T {
	seen := make(map[T]struct{}, len(list))
	results := make([]T, 0, len(list))
	for _, elem := range list {
		if _, ok := seen[elem]; ok {
			continue
		}
		seen[elem] = struct{}{}
		results = append(results, elem)
	}
	return results
}

// DeduplicateBy removes elements with duplicate keys while preserving order.
func DeduplicateBy[T any, K comparable](list []T, keyFn func(T) K) []T {
	seen := make(map[K]struct{}, len(list))
	results := make([]T, 0, len(list))
	for _, elem := range list {
		key := keyFn(elem)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		results = append(results, elem)
	}
	return results
}

// GroupBy groups elements by a key derived from keyFn.
func GroupBy[T any, K comparable](list []T, keyFn func(T) K) map[K][]T {
	groups := make(map[K][]T, len(list))
	for _, elem := range list {
		key := keyFn(elem)
		groups[key] = append(groups[key], elem)
	}
	return groups
}

// Partition splits a slice into two: elements that satisfy the predicate and those that don't.
func Partition[T any](list []T, predicate func(T) bool) (matched, unmatched []T) {
	matched = make([]T, 0, len(list))
	unmatched = make([]T, 0, len(list))
	for _, elem := range list {
		if predicate(elem) {
			matched = append(matched, elem)
		} else {
			unmatched = append(unmatched, elem)
		}
	}
	return
}
