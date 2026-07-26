package fn

// --- Map ---

// Map applies f to each element of s and returns a new slice of results.
//
// It allocates a slice of length len(s) and writes by index (faster than a
// grow-by-append loop). The input slice is not modified.
func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, len(s))
	for i := range s {
		out[i] = f(s[i])
	}
	return out
}

// MapIdx is like Map but the function also receives the element's index.
func MapIdx[T, U any](s []T, f func(int, T) U) []U {
	out := make([]U, len(s))
	for i := range s {
		out[i] = f(i, s[i])
	}
	return out
}

// MapInPlace applies f to each element of s, writing the result back into the
// same slice. The slice length is unchanged. It is the zero-allocation hot
// path when the output type equals the input type.
//
// Returns s for chaining.
func MapInPlace[T any](s []T, f func(T) T) []T {
	for i := range s {
		s[i] = f(s[i])
	}
	return s
}

// --- Filter ---

// Filter returns a new slice holding the elements of s for which pred returns
// true, preserving order. The input slice is not modified.
//
// The result is preallocated to len(s)/2 (rounded up) as a heuristic for
// typical selectivity; it grows via append if more elements survive.
func Filter[T any](s []T, pred func(T) bool) []T {
	out := make([]T, 0, (len(s)+1)/2)
	for _, v := range s {
		if pred(v) {
			out = append(out, v)
		}
	}
	return out
}

// FilterIdx is like Filter but the predicate also receives the index.
func FilterIdx[T any](s []T, pred func(int, T) bool) []T {
	out := make([]T, 0, (len(s)+1)/2)
	for i, v := range s {
		if pred(i, v) {
			out = append(out, v)
		}
	}
	return out
}

// FilterInPlace rewrites s in place, keeping only elements for which pred
// returns true, and returns the (shortened) prefix s[:k]. It reuses the
// backing array of s — no allocation — and is the preferred hot path when the
// caller owns s and no longer needs the dropped tail.
//
// The returned slice aliases s's backing memory; do not retain references to
// the dropped elements.
func FilterInPlace[T any](s []T, pred func(T) bool) []T {
	k := 0
	for i := range s {
		if pred(s[i]) {
			s[k] = s[i]
			k++
		}
	}
	// Zero out the dropped tail so the retained elements don't keep live
	// references through the backing array. For value types this is a no-op
	// cost; for pointer/iface types it avoids unintentional retention.
	clear(s[k:])
	return s[:k]
}

// --- Reduce / ForEach ---

// Reduce folds s into a single value using reducer and an initial accumulator.
// Returns initial unchanged when s is empty.
func Reduce[T, U any](s []T, reducer func(acc U, elem T) U, initial U) U {
	acc := initial
	for _, v := range s {
		acc = reducer(acc, v)
	}
	return acc
}

// ReduceIdx is like Reduce but the reducer also receives the element's index.
func ReduceIdx[T, U any](s []T, reducer func(acc U, idx int, elem T) U, initial U) U {
	acc := initial
	for i, v := range s {
		acc = reducer(acc, i, v)
	}
	return acc
}

// ForEach calls fn on each element of s in order. It is for side effects.
func ForEach[T any](s []T, fn func(T)) {
	for _, v := range s {
		fn(v)
	}
}

// ForEachIdx is like ForEach but the callback also receives the index.
func ForEachIdx[T any](s []T, fn func(int, T)) {
	for i, v := range s {
		fn(i, v)
	}
}

// --- Flatten / FlatMap ---

// Flatten concatenates the sub-slices of s into a single flat slice. The
// total length is computed up front so the result is allocated once.
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

// FlatMap applies f to each element of s (which returns a slice) and
// concatenates the results. It is the composition of Map then Flatten in a
// single pass.
func FlatMap[T, U any](s []T, f func(T) []U) []U {
	// Two passes: size the output, then fill. Avoids append regrowth.
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

// --- Reverse ---

// Reverse returns a new slice with the elements of s in reverse order. The
// input slice is not modified.
func Reverse[T any](s []T) []T {
	n := len(s)
	out := make([]T, n)
	for i, j := 0, n-1; i < n; i, j = i+1, j-1 {
		out[i] = s[j]
	}
	return out
}

// ReverseInPlace reverses s in place and returns s (for chaining).
func ReverseInPlace[T any](s []T) []T {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// --- Chunk / Window ---

// Chunk splits s into consecutive chunks of at most size elements. The final
// chunk may be shorter. Returns nil for an empty input.
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
		// Copy so chunks don't alias the input backing array.
		chunk := make([]T, end-i)
		copy(chunk, s[i:end])
		out = append(out, chunk)
	}
	return out
}

// Window returns all sliding windows of size n over s.
//
//	Window([1,2,3,4], 3) -> [[1,2,3],[2,3,4]]
//
// Returns nil if n <= 0 or len(s) < n.
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

// --- Zip / Concat / Repeat ---

// Zip pairs elements from a and b by index up to the shorter length.
// The i-th Pair holds a[i] and b[i].
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

// Concat concatenates the given slices into a single slice. The total length
// is summed up front so the result is allocated once.
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

// Repeat returns a slice of length count where every element is v.
// count must be non-negative; a count of zero returns an empty slice.
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
