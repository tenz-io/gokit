package fn

// --- 去重 ---

// Deduplicate 返回一个去除重复元素后的新 slice,保留首次出现顺序。
// 输入 slice 不会被修改。
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

// DeduplicateBy 返回一个新 slice,其 key(由 keyFn 给出)重复的元素被去除,
// 保留首次出现顺序。输入 slice 不会被修改。
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

// DeduplicateInPlace 原地从 s 中去除重复元素,保留首次出现顺序,
// 并返回(缩短后的)前缀 s[:k]。它复用 s 的 backing array。被丢弃的尾部
// 会被清零以避免保留活跃引用。
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

// DeduplicateByInPlace 是 DeduplicateBy 的 in-place 变体。
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

// Unique 和 UniqueBy 是 Deduplicate 与 DeduplicateBy 的别名,供偏好更常见
// 命名的调用方使用。
//
// Deprecated alias:为清晰起见优先使用 Deduplicate;保留 Unique 是为了
// 方便来自其他语言标准库的读者。
func Unique[T comparable](s []T) []T { return Deduplicate(s) }

// UniqueBy 是 DeduplicateBy 的别名。
func UniqueBy[T any, K comparable](s []T, keyFn func(T) K) []T {
	return DeduplicateBy(s, keyFn)
}

// --- 分组 / 划分 ---

// GroupBy 将 s 按 keyFn(element) 作为 key 分区到一个 map 中。在每个 group 内,
// 元素保留原有的相对顺序。输入为空时返回空(非 nil)map。
func GroupBy[T any, K comparable](s []T, keyFn func(T) K) map[K][]T {
	groups := make(map[K][]T, len(s))
	for _, v := range s {
		k := keyFn(v)
		groups[k] = append(groups[k], v)
	}
	return groups
}

// GroupByCount 返回一个从 key 到该 key 元素数量的 map。它是 GroupBy 针对
// 计数负载的特化版本,避免每个 group 分配 slice。
func GroupByCount[T any, K comparable](s []T, keyFn func(T) K) map[K]int {
	groups := make(map[K]int, len(s))
	for _, v := range s {
		groups[keyFn(v)]++
	}
	return groups
}

// Partition 将 s 拆分为两个 slice:pred 返回 true 的元素(matched)和
// 其余元素(unmatched),两者都保留输入顺序。
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

// PartitionInPlace 原地对 s 分区,使所有 pred 返回 true 的元素排在前面,
// 其余元素紧随其后,并返回匹配元素的个数(unmatched 尾部开始的 index)。
// 这是一个 stable partition:两侧各自保留相对顺序。
//
// matched 前缀 s[:k] 原地压缩;unmatched 尾部 s[k:] 按原始顺序持有其余元素。
// 仅对 unmatched 半段使用一个小的辅助 buffer,使 matched 前缀压缩能在一次
// 正向扫描中完成;两半最终都落在 s 的 backing array 中。
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
