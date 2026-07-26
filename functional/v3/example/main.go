// Example: functional/v3 — generic functional-programming over slices.
//
// Demonstrates the dual-track API:
//   - Standalone functions (Map / Filter / Reduce / TopK / ...)
//   - Fluent Chain (ChainOf(s).Map(...).Filter(...).TopK(k, key).Collect())
//   - Lazy Seq with short-circuit reads (Any / All / Find)
//   - In-place / zero-allocation variants (MapInPlace / FilterInPlace / ...)
//   - key-extractor-based aggregation (TopK / MinByKey / MaxByKey)
//   - OrderedSet for O(1) membership after a one-time build
package main

import (
	"fmt"

	fn "github.com/tenz-io/gokit/functional/v3"
)

type User struct {
	ID     int
	Name   string
	Active bool
	Score  int
	Tags   []string
}

func main() {
	users := []User{
		{ID: 1, Name: "alice", Active: true, Score: 90, Tags: []string{"admin", "vip"}},
		{ID: 2, Name: "bob", Active: false, Score: 70, Tags: []string{"user"}},
		{ID: 3, Name: "carol", Active: true, Score: 85, Tags: []string{"vip"}},
		{ID: 4, Name: "dave", Active: true, Score: 100, Tags: []string{"admin", "vip", "beta"}},
		{ID: 5, Name: "eve", Active: false, Score: 60, Tags: []string{"user", "beta"}},
	}

	// --- 1. Standalone functions ---
	ids := fn.Map(users, func(u User) int { return u.ID })
	fmt.Println("ids:                    ", ids)

	active := fn.Filter(users, func(u User) bool { return u.Active })
	totalScore := fn.Reduce(active, func(acc int, u User) int { return acc + u.Score }, 0)
	fmt.Printf("active users: %d, total score: %d\n", len(active), totalScore)

	// Index-aware map (carry position into the result).
	names := fn.MapIdx(users, func(i int, u User) string {
		return fmt.Sprintf("%d:%s", i, u.Name)
	})
	fmt.Println("indexed names:          ", names)

	// --- 2. key-extractor TopK (no hand-written less func) ---
	top2 := fn.TopK(users, 2, fn.Key[User, int](func(u User) int { return u.Score }))
	fmt.Println("top-2 by score:         ", namesOf(top2))

	bottom2 := fn.BottomK(users, 2, fn.Key[User, int](func(u User) int { return u.Score }))
	fmt.Println("bottom-2 by score:      ", namesOf(bottom2))

	// Min/Max by key.
	minUser, _ := fn.MinByKey(users, fn.Key[User, int](func(u User) int { return u.Score }))
	maxUser, _ := fn.MaxByKey(users, fn.Key[User, int](func(u User) int { return u.Score }))
	fmt.Printf("min/max by score:       %s (%d) / %s (%d)\n",
		minUser.Name, minUser.Score, maxUser.Name, maxUser.Score)

	// --- 3. Fluent Chain ---
	results := fn.ChainOf(users).
		Filter(func(u User) bool { return u.Active }).
		TopK(3, func(u User) int { return u.Score }). // chain TopK uses an int key
		Collect()
	fmt.Println("chain: active top-3:     ", namesOf(results))

	// Type-changing map is a free function (Go methods can't change type params).
	activeNames := fn.MapTo(
		fn.ChainOf(users).Filter(func(u User) bool { return u.Active }),
		func(u User) string { return u.Name },
	).Collect()
	fmt.Println("chain: active names:     ", activeNames)

	// SortBy (descending by score) via a cmp-style comparator.
	descByScore := fn.ChainOf(users).
		SortBy(func(a, b User) int { return b.Score - a.Score }).
		Collect()
	fmt.Println("chain: sort desc score: ", namesOf(descByScore))

	// --- 4. Lazy Seq (short-circuit, zero-allocation reads) ---
	has100 := fn.SeqOf(users).Any(func(u User) bool { return u.Score >= 100 })
	allHaveID := fn.SeqOf(users).All(func(u User) bool { return u.ID > 0 })
	firstVIP, found := fn.SeqOf(users).Find(func(u User) bool {
		return fn.Contains(u.Tags, "vip")
	})
	fmt.Printf("seq: has100=%v allHaveID=%v firstVIP=%s(found=%v)\n",
		has100, allHaveID, firstVIP.Name, found)

	// Fused filter+map on a Seq, materialized only at Collect.
	scoreDoubles := fn.MapSeq(
		fn.SeqOf(users).Filter(func(u User) bool { return u.Score >= 80 }),
		func(u User) int { return u.Score * 2 },
	).Collect()
	fmt.Println("seq: filtered+doubled:  ", scoreDoubles)

	// --- 5. In-place / zero-allocation variants ---
	scores := []int{10, 20, 30, 40}
	fn.MapInPlace(scores, func(i int) int { return i * 10 })
	fmt.Println("inplace MapInPlace:     ", scores)

	mixed := []int{1, 2, 3, 4, 5, 6, 7, 8}
	evens := fn.FilterInPlace(mixed, func(i int) bool { return i%2 == 0 })
	fmt.Println("inplace FilterInPlace:  ", evens)

	dup := []int{1, 2, 2, 3, 1, 4}
	uniq := fn.DeduplicateInPlace(dup)
	fmt.Println("inplace Deduplicate:     ", uniq)

	// --- 6. Chunk / Window / Zip / FlatMap ---
	chunked := fn.Chunk([]int{1, 2, 3, 4, 5}, 2)
	fmt.Println("chunk(size=2):          ", chunked)

	windows := fn.Window([]int{1, 2, 3, 4}, 3)
	fmt.Println("window(size=3):         ", windows)

	allTags := fn.FlatMap(users, func(u User) []string { return u.Tags })
	fmt.Println("flatMap tags:           ", allTags)

	zipped := fn.Zip([]int{1, 2, 3}, []string{"a", "b", "c"})
	fmt.Printf("zip:                    %+v\n", zipped)

	// --- 7. GroupBy / Partition ---
	byActive := fn.GroupBy(users, func(u User) bool { return u.Active })
	fmt.Printf("groupBy active:         active=%d inactive=%d\n",
		len(byActive[true]), len(byActive[false]))

	tagCounts := fn.GroupByCount(allTags, func(s string) string { return s })
	fmt.Println("groupByCount tags:      ", tagCounts)

	matched, unmatched := fn.Partition(users, func(u User) bool { return u.Score >= 80 })
	fmt.Printf("partition(>=80):        matched=%d unmatched=%d\n", len(matched), len(unmatched))

	// --- 8. OrderedSet (one-time build, O(1) membership) ---
	seen := fn.NewOrderedSet(ids...)
	fmt.Println("ordered set:            ", seen.ToSlice())
	fmt.Println("set contains ID 3?      ", seen.Contains(3))
	fmt.Println("set contains ID 99?     ", seen.Contains(99))
	seen.Remove(3)
	fmt.Println("after remove 3:         ", seen.ToSlice())

	// --- 9. Conditionals ---
	fallback := fn.Coalesce("", "", "default-value")
	fmt.Println("coalesce first non-empty:", fallback)
	def := fn.Default("", "fallback")
	fmt.Println("default empty -> fallback:", def)
	gated := fn.When(true, 5, func(i int) int { return i * 2 })
	fmt.Println("when(true, 5, *2):      ", gated)
}

// namesOf is a tiny local helper to render a user slice compactly.
func namesOf(us []User) []string {
	return fn.Map(us, func(u User) string { return u.Name })
}
