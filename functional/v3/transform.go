package fn

// --- 映射(Map) ---

// Map 对 s 中每个元素应用 f,返回结果的新 slice。
//
// 它分配一个长度为 len(s) 的 slice 并按 index 写入(比按 append 增长更快)。
// 输入 slice 不会被修改。
func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, len(s))
	for i := range s {
		out[i] = f(s[i])
	}
	return out
}

// MapIdx 类似 Map,但函数还接收元素的 index。
func MapIdx[T, U any](s []T, f func(int, T) U) []U {
	out := make([]U, len(s))
	for i := range s {
		out[i] = f(i, s[i])
	}
	return out
}

// MapInPlace 对 s 中每个元素应用 f,并将结果写回同一个 slice。slice
// 长度不变。当输出类型与输入类型相同时,它是 zero-allocation 热路径。
//
// 返回 s 以便链式调用。
func MapInPlace[T any](s []T, f func(T) T) []T {
	for i := range s {
		s[i] = f(s[i])
	}
	return s
}

// --- 过滤(Filter) ---

// Filter 返回一个持有 s 中 pred 返回 true 元素的新 slice,保留顺序。
// 输入 slice 不会被修改。
//
// 结果按 len(s)/2(向上取整)预分配,作为典型选择率的一个启发式估计;
// 若幸存元素更多则通过 append 增长。
func Filter[T any](s []T, pred func(T) bool) []T {
	out := make([]T, 0, (len(s)+1)/2)
	for _, v := range s {
		if pred(v) {
			out = append(out, v)
		}
	}
	return out
}

// FilterIdx 类似 Filter,但 predicate 还接收 index。
func FilterIdx[T any](s []T, pred func(int, T) bool) []T {
	out := make([]T, 0, (len(s)+1)/2)
	for i, v := range s {
		if pred(i, v) {
			out = append(out, v)
		}
	}
	return out
}

// FilterInPlace 原地重写 s,仅保留 pred 返回 true 的元素,并返回
// (缩短后的)前缀 s[:k]。它复用 s 的 backing array——无分配——
// 是调用方拥有 s 且不再需要被丢弃尾部时的首选热路径。
//
// 返回的 slice alias s 的 backing 内存;不要保留对被丢弃元素的引用。
func FilterInPlace[T any](s []T, pred func(T) bool) []T {
	k := 0
	for i := range s {
		if pred(s[i]) {
			s[k] = s[i]
			k++
		}
	}
	// 将被丢弃的尾部清零,使幸存元素不会通过 backing array 保留活跃引用。
	// 对值类型而言这是 no-op 成本;对 pointer/iface 类型可避免无意中的引用残留。
	clear(s[k:])
	return s[:k]
}

// --- 归约 / 遍历 ---

// Reduce 使用 reducer 和初始累加值将 s 折叠为单个值。s 为空时返回 initial 不变。
func Reduce[T, U any](s []T, reducer func(acc U, elem T) U, initial U) U {
	acc := initial
	for _, v := range s {
		acc = reducer(acc, v)
	}
	return acc
}

// ReduceIdx 类似 Reduce,但 reducer 还接收元素的 index。
func ReduceIdx[T, U any](s []T, reducer func(acc U, idx int, elem T) U, initial U) U {
	acc := initial
	for i, v := range s {
		acc = reducer(acc, i, v)
	}
	return acc
}

// ForEach 按顺序对 s 中每个元素调用 fn。用于产生副作用。
func ForEach[T any](s []T, fn func(T)) {
	for _, v := range s {
		fn(v)
	}
}

// ForEachIdx 类似 ForEach,但 callback 还接收 index。
func ForEachIdx[T any](s []T, fn func(int, T)) {
	for i, v := range s {
		fn(i, v)
	}
}

// --- 展平 / FlatMap ---

// Flatten 将 s 的各 sub-slice 拼接为一个扁平 slice。总长度预先计算,
// 因此结果只分配一次。
func Flatten[T any](s [][]T) []T {
	total := 0
	for _, sub := range s {
		total += len(sub)
	}
	out := make([]T, 0, total)
	for _, sub := range s {
		out = append(out, sub...)
	}
	return out
}

// FlatMap 对 s 中每个元素应用 f(返回一个 slice)并拼接结果。
// 它是 Map 再 Flatten 在单次扫描中完成的组合。
func FlatMap[T, U any](s []T, f func(T) []U) []U {
	// 两轮扫描:先确定输出大小,再填充。避免 append 再扩容。
	total := 0
	for _, v := range s {
		total += len(f(v))
	}
	out := make([]U, 0, total)
	for _, v := range s {
		out = append(out, f(v)...)
	}
	return out
}

// --- 反转 ---

// Reverse 返回元素顺序与 s 相反的新 slice。输入 slice 不会被修改。
func Reverse[T any](s []T) []T {
	n := len(s)
	out := make([]T, n)
	for i, j := 0, n-1; i < n; i, j = i+1, j-1 {
		out[i] = s[j]
	}
	return out
}

// ReverseInPlace 原地反转 s 并返回 s(以便链式调用)。
func ReverseInPlace[T any](s []T) []T {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// --- 分块 / 窗口 ---

// Chunk 将 s 切分为至多 size 个元素的连续 chunk。最后一个 chunk 可能更短。
// 空输入返回 nil。
//
//	Chunk([1,2,3,4,5], 2) -> [[1,2],[3,4],[5]]
func Chunk[T any](s []T, size int) [][]T {
	if size <= 0 {
		panic("fn.Chunk: size must be positive")
	}
	if len(s) == 0 {
		return nil
	}
	n := (len(s) + size - 1) / size
	out := make([][]T, 0, n)
	for i := 0; i < len(s); i += size {
		end := i + size
		if end > len(s) {
			end = len(s)
		}
		// 复制以避免 chunk alias 输入的 backing array。
		chunk := make([]T, end-i)
		copy(chunk, s[i:end])
		out = append(out, chunk)
	}
	return out
}

// Window 返回 s 上大小为 n 的全部滑动窗口。
//
//	Window([1,2,3,4], 3) -> [[1,2,3],[2,3,4]]
//
// 当 n <= 0 或 len(s) < n 时返回 nil。
func Window[T any](s []T, n int) [][]T {
	if n <= 0 || len(s) < n {
		return nil
	}
	count := len(s) - n + 1
	out := make([][]T, 0, count)
	for i := 0; i < count; i++ {
		w := make([]T, n)
		copy(w, s[i:i+n])
		out = append(out, w)
	}
	return out
}

// --- 拼合 / 拼接 / 重复 ---

// Zip 按 index 将 a 与 b 的元素配对,直到较短者结束。第 i 个 Pair 持有
// a[i] 与 b[i]。
func Zip[A, B any](a []A, b []B) []Pair[A, B] {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	out := make([]Pair[A, B], n)
	for i := 0; i < n; i++ {
		out[i] = Pair[A, B]{A: a[i], B: b[i]}
	}
	return out
}

// Concat 将给定的多个 slice 拼接为单个 slice。总长度预先求和,
// 因此结果只分配一次。
func Concat[T any](slices ...[]T) []T {
	total := 0
	for _, sl := range slices {
		total += len(sl)
	}
	out := make([]T, 0, total)
	for _, sl := range slices {
		out = append(out, sl...)
	}
	return out
}

// Repeat 返回长度为 count 的 slice,每个元素都是 v。
// count 必须非负;count 为零时返回空 slice。
func Repeat[T any](v T, count int) []T {
	if count < 0 {
		panic("fn.Repeat: count must be non-negative")
	}
	out := make([]T, count)
	for i := range out {
		out[i] = v
	}
	return out
}
