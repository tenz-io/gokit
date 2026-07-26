package fn

import (
	"reflect"
	"testing"
)

// ---- helpers ----

func eq[T any](t *testing.T, name string, got, want []T) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// ---- Map / MapIdx / MapInPlace ----

func TestMap(t *testing.T) {
	double := func(i int) int { return i * 2 }
	eq(t, "nil", Map([]int(nil), double), []int{})
	eq(t, "empty", Map([]int{}, double), []int{})
	eq(t, "single", Map([]int{5}, double), []int{10})
	eq(t, "multi", Map([]int{1, 2, 3}, double), []int{2, 4, 6})
	// type-changing
	strs := Map([]int{1, 2, 3}, func(i int) string { return string(rune('0' + i)) })
	eq(t, "type-change", strs, []string{"1", "2", "3"})
}

func TestMapIdx(t *testing.T) {
	got := MapIdx([]string{"a", "b", "c"}, func(i int, s string) string {
		return s + string(rune('0'+i))
	})
	eq(t, "MapIdx", got, []string{"a0", "b1", "c2"})
}

func TestMapInPlace(t *testing.T) {
	s := []int{1, 2, 3}
	out := MapInPlace(s, func(i int) int { return i * 10 })
	eq(t, "MapInPlace value", out, []int{10, 20, 30})
	if &out[0] != &s[0] {
		t.Error("MapInPlace should reuse backing array")
	}
}

// ---- Filter / FilterIdx / FilterInPlace ----

func TestFilter(t *testing.T) {
	even := func(i int) bool { return i%2 == 0 }
	eq(t, "nil", Filter([]int(nil), even), []int{})
	eq(t, "empty", Filter([]int{}, even), []int{})
	eq(t, "no match", Filter([]int{1, 3, 5}, even), []int{})
	eq(t, "all match", Filter([]int{2, 4, 6}, even), []int{2, 4, 6})
	eq(t, "some match", Filter([]int{1, 2, 3, 4, 5}, even), []int{2, 4})
}

func TestFilterIdx(t *testing.T) {
	eq(t, "FilterIdx", FilterIdx([]int{10, 20, 30}, func(i, v int) bool { return i == 1 }), []int{20})
}

func TestFilterInPlace(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6}
	out := FilterInPlace(s, func(i int) bool { return i%2 == 0 })
	eq(t, "FilterInPlace", out, []int{2, 4, 6})
	if cap(out) != cap(s) {
		t.Error("FilterInPlace should reuse backing array")
	}
	// single, no match
	s2 := []int{1, 3, 5}
	eq(t, "FilterInPlace none", FilterInPlace(s2, func(i int) bool { return i%2 == 0 }), []int{})
}

// ---- Reduce / ReduceIdx / ForEach / ForEachIdx ----

func TestReduce(t *testing.T) {
	if got := Reduce([]int{1, 2, 3, 4}, func(acc, e int) int { return acc + e }, 0); got != 10 {
		t.Errorf("sum = %v, want 10", got)
	}
	if got := Reduce([]int(nil), func(acc, e int) int { return acc + e }, 10); got != 10 {
		t.Errorf("nil = %v, want 10", got)
	}
	if got := Reduce([]int{}, func(acc, e int) int { return acc + e }, 5); got != 5 {
		t.Errorf("empty = %v, want 5", got)
	}
	if got := Reduce([]string{"a", "b", "c"}, func(acc, e string) string { return acc + e }, ""); got != "abc" {
		t.Errorf("concat = %v, want abc", got)
	}
}

func TestReduceIdx(t *testing.T) {
	got := ReduceIdx([]int{10, 20, 30}, func(acc, i, e int) int { return acc + i*e }, 0)
	// 0*10 + 1*20 + 2*30 = 80
	if got != 80 {
		t.Errorf("ReduceIdx = %v, want 80", got)
	}
}

func TestForEach(t *testing.T) {
	var sum int
	ForEach([]int{1, 2, 3}, func(i int) { sum += i })
	if sum != 6 {
		t.Errorf("ForEach sum = %v, want 6", sum)
	}
	ForEach([]int(nil), func(i int) { sum += i }) // must not panic
}

func TestForEachIdx(t *testing.T) {
	var idxSum int
	ForEachIdx([]int{5, 5, 5}, func(i, v int) { idxSum += i + v })
	if idxSum != (0+5)+(1+5)+(2+5) {
		t.Errorf("ForEachIdx = %v, want 18", idxSum)
	}
}

// ---- Flatten / FlatMap ----

func TestFlatten(t *testing.T) {
	eq(t, "nil", Flatten([][]int(nil)), []int{})
	eq(t, "empty outer", Flatten([][]int{}), []int{})
	eq(t, "empty inner", Flatten([][]int{{}, {}}), []int{})
	eq(t, "single", Flatten([][]int{{1, 2, 3}}), []int{1, 2, 3})
	eq(t, "multi", Flatten([][]int{{1, 2}, {3, 4}, {5}}), []int{1, 2, 3, 4, 5})
	eq(t, "mixed empty", Flatten([][]int{{1}, {}, {2, 3}}), []int{1, 2, 3})
}

func TestFlatMap(t *testing.T) {
	got := FlatMap([]int{1, 2, 3}, func(i int) []int { return []int{i, i * 10} })
	eq(t, "FlatMap", got, []int{1, 10, 2, 20, 3, 30})
	eq(t, "FlatMap nil", FlatMap([]int(nil), func(i int) []int { return []int{i} }), []int{})
}

// ---- Reverse / ReverseInPlace ----

func TestReverse(t *testing.T) {
	eq(t, "nil", Reverse([]int(nil)), []int{})
	eq(t, "empty", Reverse([]int{}), []int{})
	eq(t, "single", Reverse([]int{1}), []int{1})
	eq(t, "two", Reverse([]int{1, 2}), []int{2, 1})
	eq(t, "odd", Reverse([]int{1, 2, 3}), []int{3, 2, 1})
	eq(t, "even", Reverse([]int{1, 2, 3, 4}), []int{4, 3, 2, 1})
}

func TestReverseInPlace(t *testing.T) {
	cases := []struct {
		name string
		in   []int
		want []int
	}{
		{"empty", []int{}, []int{}},
		{"single", []int{1}, []int{1}},
		{"two", []int{1, 2}, []int{2, 1}},
		{"odd", []int{1, 2, 3}, []int{3, 2, 1}},
		{"even", []int{1, 2, 3, 4}, []int{4, 3, 2, 1}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := ReverseInPlace(tt.in)
			eq(t, "ReverseInPlace "+tt.name, got, tt.want)
		})
	}
	// nil must not panic and stays empty.
	_ = ReverseInPlace([]int(nil))
}

// ---- Chunk / Window / Zip / Concat / Repeat ----

func TestChunk(t *testing.T) {
	eq(t, "empty", Chunk([]int{}, 2), [][]int(nil))
	eq(t, "exact", Chunk([]int{1, 2, 3, 4}, 2), [][]int{{1, 2}, {3, 4}})
	eq(t, "short last", Chunk([]int{1, 2, 3, 4, 5}, 2), [][]int{{1, 2}, {3, 4}, {5}})
	eq(t, "single", Chunk([]int{1, 2, 3}, 5), [][]int{{1, 2, 3}})
}

func TestChunkPanicsOnBadSize(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Chunk(.,0) should panic")
		}
	}()
	_ = Chunk([]int{1}, 0)
}

func TestWindow(t *testing.T) {
	eq(t, "empty", Window([]int{}, 2), [][]int(nil))
	eq(t, "too small", Window([]int{1, 2}, 3), [][]int(nil))
	eq(t, "basic", Window([]int{1, 2, 3, 4}, 3), [][]int{{1, 2, 3}, {2, 3, 4}})
	eq(t, "size1", Window([]int{1, 2}, 1), [][]int{{1}, {2}})
}

func TestZip(t *testing.T) {
	got := Zip([]int{1, 2, 3}, []string{"a", "b"})
	if len(got) != 2 {
		t.Fatalf("Zip len = %d, want 2", len(got))
	}
	if got[0].A != 1 || got[0].B != "a" {
		t.Errorf("Zip[0] = %v, want {1 a}", got[0])
	}
	if got[1].A != 2 || got[1].B != "b" {
		t.Errorf("Zip[1] = %v, want {2 b}", got[1])
	}
}

func TestConcat(t *testing.T) {
	eq(t, "concat", Concat([]int{1, 2}, []int{3}, []int{4, 5}), []int{1, 2, 3, 4, 5})
	eq(t, "empty", Concat([]int{}, []int{}), []int{})
}

func TestRepeat(t *testing.T) {
	eq(t, "repeat", Repeat(7, 3), []int{7, 7, 7})
	eq(t, "zero count", Repeat(7, 0), []int{})
}

// ---- Predicates: All / Any / None / Contains / ContainsBy / Count ----

func TestAll(t *testing.T) {
	even := func(i int) bool { return i%2 == 0 }
	if !All([]int(nil), even) {
		t.Error("All(nil) should be true (vacuous)")
	}
	if !All([]int{}, even) {
		t.Error("All(empty) should be true (vacuous)")
	}
	if !All([]int{2, 4, 6}, even) {
		t.Error("All(even) should be true")
	}
	if All([]int{2, 3, 4}, even) {
		t.Error("All(mixed) should be false")
	}
}

func TestAny(t *testing.T) {
	even := func(i int) bool { return i%2 == 0 }
	if Any([]int(nil), even) {
		t.Error("Any(nil) should be false")
	}
	if Any([]int{1, 3, 5}, even) {
		t.Error("Any(odd) should be false")
	}
	if !Any([]int{1, 2, 3}, even) {
		t.Error("Any(mixed) should be true")
	}
}

func TestNone(t *testing.T) {
	even := func(i int) bool { return i%2 == 0 }
	if !None([]int{1, 3, 5}, even) {
		t.Error("None(odd) should be true")
	}
	if None([]int{1, 2, 3}, even) {
		t.Error("None(mixed) should be false")
	}
}

func TestContains(t *testing.T) {
	if Contains([]int{1, 2, 3}, 2) != true {
		t.Error("Contains 2")
	}
	if Contains([]int{1, 2, 3}, 4) != false {
		t.Error("not Contains 4")
	}
}

func TestContainsBy(t *testing.T) {
	type user struct{ ID int }
	list := []user{{1}, {2}, {3}}
	if !ContainsBy(list, 2, func(u user) int { return u.ID }) {
		t.Error("ContainsBy 2")
	}
	if ContainsBy(list, 4, func(u user) int { return u.ID }) {
		t.Error("not ContainsBy 4")
	}
}

func TestCount(t *testing.T) {
	even := func(i int) bool { return i%2 == 0 }
	if c := Count([]int{1, 2, 3, 4, 5}, even); c != 2 {
		t.Errorf("Count = %d, want 2", c)
	}
}

func TestCountBy(t *testing.T) {
	type u struct{ G string }
	list := []u{{"a"}, {"b"}, {"a"}, {"c"}, {"a"}}
	if c := CountBy(list, "a", func(v u) string { return v.G }); c != 3 {
		t.Errorf("CountBy = %d, want 3", c)
	}
}

// ---- Find / FindIndex / FindLast / FindLastIndex / IndexOf / LastIndexOf ----

func TestFind(t *testing.T) {
	gt2 := func(i int) bool { return i > 2 }
	if v, ok := Find([]int{1, 3, 5}, gt2); !ok || v != 3 {
		t.Errorf("Find = (%v,%v), want (3,true)", v, ok)
	}
	if _, ok := Find([]int{1, 2}, gt2); ok {
		t.Error("Find none should be false")
	}
}

func TestFindIndex(t *testing.T) {
	if i, ok := FindIndex([]int{1, 2, 3}, func(i int) bool { return i == 3 }); !ok || i != 2 {
		t.Errorf("FindIndex = (%d,%v), want (2,true)", i, ok)
	}
	if _, ok := FindIndex([]int{1}, func(i int) bool { return i == 9 }); ok {
		t.Error("FindIndex none should be false")
	}
}

func TestFindLast(t *testing.T) {
	v, ok := FindLast([]int{1, 2, 3, 2, 1}, func(i int) bool { return i == 2 })
	if !ok || v != 2 {
		t.Errorf("FindLast = (%v,%v), want (2,true)", v, ok)
	}
}

func TestFindLastIndex(t *testing.T) {
	i, ok := FindLastIndex([]int{1, 2, 3, 2, 1}, func(i int) bool { return i == 2 })
	if !ok || i != 3 {
		t.Errorf("FindLastIndex = (%d,%v), want (3,true)", i, ok)
	}
}

func TestIndexOf(t *testing.T) {
	if i, ok := IndexOf([]int{1, 2, 3}, 2); !ok || i != 1 {
		t.Errorf("IndexOf = (%d,%v), want (1,true)", i, ok)
	}
	if _, ok := IndexOf([]int{1, 2, 3}, 9); ok {
		t.Error("IndexOf none should be false")
	}
}

func TestLastIndexOf(t *testing.T) {
	if i, ok := LastIndexOf([]int{1, 2, 3, 2, 1}, 2); !ok || i != 3 {
		t.Errorf("LastIndexOf = (%d,%v), want (3,true)", i, ok)
	}
}

// ---- Min / Max / Sum / Avg ----

func TestMin(t *testing.T) {
	if v, ok := Min([]int{3, 1, 2}); !ok || v != 1 {
		t.Errorf("Min = (%v,%v), want (1,true)", v, ok)
	}
	if _, ok := Min([]int{}); ok {
		t.Error("Min empty should be false")
	}
}

func TestMax(t *testing.T) {
	if v, ok := Max([]int{1, 3, 2}); !ok || v != 3 {
		t.Errorf("Max = (%v,%v), want (3,true)", v, ok)
	}
}

func TestSum(t *testing.T) {
	if Sum([]int{1, 2, 3, 4}) != 10 {
		t.Error("Sum = 10")
	}
	if Sum([]int{}) != 0 {
		t.Error("Sum empty = 0")
	}
}

func TestAvg(t *testing.T) {
	if v, ok := Avg([]int{1, 2, 3, 4}); !ok || v != 2.5 {
		t.Errorf("Avg = (%v,%v), want (2.5,true)", v, ok)
	}
	if _, ok := Avg([]int{}); ok {
		t.Error("Avg empty should be false")
	}
}

// ---- MinByKey / MaxByKey / MinBy / MaxBy ----

func TestMinByKey(t *testing.T) {
	type u struct{ ID, Score int }
	key := Key[u, int](func(x u) int { return x.Score })
	v, ok := MinByKey([]u{{3, 90}, {1, 70}, {2, 85}}, key)
	if !ok || v.ID != 1 {
		t.Errorf("MinByKey = (%v,%v), want ID 1", v, ok)
	}
}

func TestMaxByKey(t *testing.T) {
	type u struct{ ID, Score int }
	key := Key[u, int](func(x u) int { return x.Score })
	v, ok := MaxByKey([]u{{3, 90}, {1, 70}, {2, 85}}, key)
	if !ok || v.ID != 3 {
		t.Errorf("MaxByKey = (%v,%v), want ID 3", v, ok)
	}
}

func TestMinBy(t *testing.T) {
	type u struct{ ID int }
	by := By[u](func(a, b u) int { return a.ID - b.ID })
	v, ok := MinBy([]u{{3}, {1}, {2}}, by)
	if !ok || v.ID != 1 {
		t.Errorf("MinBy = (%v,%v), want 1", v, ok)
	}
}

func TestMaxBy(t *testing.T) {
	type u struct{ ID int }
	by := By[u](func(a, b u) int { return a.ID - b.ID })
	v, ok := MaxBy([]u{{1}, {3}, {2}}, by)
	if !ok || v.ID != 3 {
		t.Errorf("MaxBy = (%v,%v), want 3", v, ok)
	}
}

// ---- TopK / BottomK (key) ----

func TestTopK(t *testing.T) {
	type item struct {
		ID   int
		Name string
	}
	key := Key[item, int](func(x item) int { return x.ID })
	cases := []struct {
		name string
		list []item
		k    int
		want []item
	}{
		{"nil", nil, 3, nil},
		{"empty", []item{}, 3, nil},
		{"k zero", []item{{1, "a"}}, 0, nil},
		{"k negative", []item{{1, "a"}}, -1, nil},
		{"k larger", []item{{1, "a"}, {3, "c"}, {2, "b"}}, 5, []item{{3, "c"}, {2, "b"}, {1, "a"}}},
		{"k equals", []item{{1, "a"}, {3, "c"}, {2, "b"}}, 3, []item{{3, "c"}, {2, "b"}, {1, "a"}}},
		{"top2 of 5", []item{{3, "c"}, {1, "a"}, {5, "e"}, {2, "b"}, {4, "d"}}, 2, []item{{5, "e"}, {4, "d"}}},
		{"top3 of 5", []item{{3, "c"}, {1, "a"}, {5, "e"}, {2, "b"}, {4, "d"}}, 3, []item{{5, "e"}, {4, "d"}, {3, "c"}}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			eq(t, "TopK "+tt.name, TopK(tt.list, tt.k, key), tt.want)
		})
	}
}

func TestTopK_Strings(t *testing.T) {
	key := Key[string, string](func(s string) string { return s })
	got := TopK([]string{"banana", "apple", "cherry", "date", "elderberry"}, 3, key)
	eq(t, "TopK strings", got, []string{"elderberry", "date", "cherry"})
}

func TestBottomK(t *testing.T) {
	key := Key[int, int](func(i int) int { return i })
	eq(t, "BottomK", BottomK([]int{5, 3, 1, 4, 2}, 3, key), []int{1, 2, 3})
	eq(t, "BottomK k>=n", BottomK([]int{3, 1, 2}, 5, key), []int{1, 2, 3})
	eq(t, "BottomK empty", BottomK([]int{}, 3, key), nil)
}

func TestTopKBy(t *testing.T) {
	type item struct{ ID int }
	by := By[item](func(a, b item) int { return a.ID - b.ID })
	eq(t, "TopKBy", TopKBy([]item{{3}, {1}, {5}, {2}, {4}}, 2, by), []item{{5}, {4}})
}

func TestBottomKBy(t *testing.T) {
	type item struct{ ID int }
	by := By[item](func(a, b item) int { return a.ID - b.ID })
	eq(t, "BottomKBy", BottomKBy([]item{{5}, {3}, {1}, {4}, {2}}, 3, by), []item{{1}, {2}, {3}})
}

// ---- Deduplicate / DeduplicateBy / InPlace ----

func TestDeduplicate(t *testing.T) {
	eq(t, "nil", Deduplicate([]int(nil)), nil)
	eq(t, "empty", Deduplicate([]int{}), nil)
	eq(t, "no dup", Deduplicate([]int{1, 2, 3}), []int{1, 2, 3})
	eq(t, "dup", Deduplicate([]int{1, 2, 2, 3, 1}), []int{1, 2, 3})
	eq(t, "all same", Deduplicate([]int{1, 1, 1}), []int{1})
}

func TestDeduplicateBy(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}
	key := func(u user) int { return u.ID }
	eq(t, "DeduplicateBy", DeduplicateBy([]user{{1, "a"}, {2, "b"}, {1, "c"}}, key), []user{{1, "a"}, {2, "b"}})
}

func TestDeduplicateInPlace(t *testing.T) {
	s := []int{1, 2, 2, 3, 1}
	out := DeduplicateInPlace(s)
	eq(t, "DeduplicateInPlace", out, []int{1, 2, 3})
	if cap(out) != cap(s) {
		t.Error("DeduplicateInPlace should reuse backing array")
	}
}

// ---- GroupBy / GroupByCount / Partition / PartitionInPlace ----

func TestGroupBy(t *testing.T) {
	type order struct {
		Product string
		Amount  int
	}
	list := []order{{"apple", 1}, {"banana", 2}, {"apple", 3}, {"cherry", 4}, {"banana", 5}}
	got := GroupBy(list, func(o order) string { return o.Product })
	if len(got) != 3 || len(got["apple"]) != 2 || len(got["banana"]) != 2 || len(got["cherry"]) != 1 {
		t.Errorf("GroupBy = %v", got)
	}
}

func TestGroupByCount(t *testing.T) {
	list := []string{"a", "b", "a", "c", "a"}
	got := GroupByCount(list, func(s string) string { return s })
	if got["a"] != 3 || got["b"] != 1 || got["c"] != 1 {
		t.Errorf("GroupByCount = %v", got)
	}
}

func TestPartition(t *testing.T) {
	even := func(i int) bool { return i%2 == 0 }
	m, u := Partition([]int{1, 2, 3, 4, 5}, even)
	eq(t, "matched", m, []int{2, 4})
	eq(t, "unmatched", u, []int{1, 3, 5})
}

func TestPartitionInPlace(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6}
	k := PartitionInPlace(s, func(i int) bool { return i%2 == 0 })
	if k != 3 {
		t.Errorf("PartitionInPlace k = %d, want 3", k)
	}
	eq(t, "matched prefix", s[:k], []int{2, 4, 6})
	eq(t, "unmatched tail", s[k:], []int{1, 3, 5})
}

// ---- OrderedSet ----

func TestOrderedSet(t *testing.T) {
	s := NewOrderedSet(1, 2, 3, 2, 1)
	if s.Len() != 3 {
		t.Errorf("Len = %d, want 3", s.Len())
	}
	eq(t, "ToSlice order", s.ToSlice(), []int{1, 2, 3})
	if !s.Contains(2) {
		t.Error("Contains 2")
	}
	if s.Contains(9) {
		t.Error("not Contains 9")
	}
	if !s.Add(4) {
		t.Error("Add 4 should be newly added")
	}
	if s.Add(2) {
		t.Error("Add 2 (dup) should return false")
	}
}

func TestOrderedSetRemove(t *testing.T) {
	s := NewOrderedSet(1, 2, 3)
	if !s.Remove(2) {
		t.Error("Remove 2 should return true")
	}
	if s.Contains(2) {
		t.Error("Contains 2 after remove")
	}
	eq(t, "ToSlice after remove", s.ToSlice(), []int{1, 3})
	// tombstone must not leak the zero-value element
	s2 := NewOrderedSet(0, 1, 2)
	s2.Remove(0)
	eq(t, "zero-value removed", s2.ToSlice(), []int{1, 2})
}

// ---- Conditional ----

func TestIf(t *testing.T) {
	if If(true, 1, 2) != 1 || If(false, 1, 2) != 2 {
		t.Error("If")
	}
}

func TestWhen(t *testing.T) {
	if When(true, 5, func(i int) int { return i * 2 }) != 10 {
		t.Error("When true")
	}
	if When(false, 5, func(i int) int { return i * 2 }) != 5 {
		t.Error("When false")
	}
}

func TestIfElse(t *testing.T) {
	if IfElse(true, f(1), f(2)) != 1 || IfElse(false, f(1), f(2)) != 2 {
		t.Error("IfElse")
	}
	called := false
	IfElse(true, f(0), func() int { called = true; return 1 })
	if called {
		t.Error("IfElse(true) should not call elseFn")
	}
}

func f(v int) func() int { return func() int { return v } }

func TestCoalesce(t *testing.T) {
	if Coalesce("", "", "x", "y") != "x" {
		t.Error("Coalesce first non-empty")
	}
	if Coalesce("", "") != "" {
		t.Error("Coalesce all empty")
	}
}

func TestDefault(t *testing.T) {
	if Default("", "fallback") != "fallback" {
		t.Error("Default fallback")
	}
	if Default("v", "fallback") != "v" {
		t.Error("Default keeps v")
	}
}

// ---- Chain ----

func TestChain(t *testing.T) {
	type user struct{ ID, Score int }
	users := []user{{1, 90}, {2, 70}, {3, 85}, {4, 100}}

	// Filter + TopK + Collect
	got := ChainOf(users).
		Filter(func(u user) bool { return u.Score >= 85 }).
		TopK(2, func(u user) int { return u.Score }).
		Collect()
	eq(t, "chain filter+topk", got, []user{{4, 100}, {1, 90}})

	// MapTo type change
	ids := MapTo(ChainOf(users), func(u user) int { return u.ID }).Collect()
	eq(t, "chain MapTo", ids, []int{1, 2, 3, 4})

	// Take / Drop
	eq(t, "chain Take", ChainOf([]int{1, 2, 3, 4, 5}).Take(2).Collect(), []int{1, 2})
	eq(t, "chain Drop", ChainOf([]int{1, 2, 3, 4, 5}).Drop(2).Collect(), []int{3, 4, 5})

	// SortChain / SortBy / Reverse
	eq(t, "chain Sort", SortChain(ChainOf([]int{3, 1, 2})).Collect(), []int{1, 2, 3})
	eq(t, "chain SortBy desc",
		ChainOf([]int{1, 2, 3}).SortBy(func(a, b int) int { return b - a }).Collect(),
		[]int{3, 2, 1})
	eq(t, "chain Reverse", ChainOf([]int{1, 2, 3}).Reverse().Collect(), []int{3, 2, 1})

	// DeduplicateByChain
	eq(t, "chain DeduplicateBy",
		DeduplicateByChain(ChainOf([]int{1, 2, 2, 3, 1}), func(i int) int { return i }).Collect(),
		[]int{1, 2, 3})

	// Concat
	eq(t, "chain Concat", ChainOf([]int{1, 2}).Concat([]int{3, 4}).Collect(), []int{1, 2, 3, 4})

	// Any/All/Find
	if !ChainOf([]int{1, 2, 3}).Any(func(i int) bool { return i == 2 }) {
		t.Error("chain Any")
	}
	if !ChainOf([]int{2, 4, 6}).All(func(i int) bool { return i%2 == 0 }) {
		t.Error("chain All")
	}
	if v, ok := ChainOf([]int{1, 2, 3}).Find(func(i int) bool { return i > 1 }); !ok || v != 2 {
		t.Error("chain Find")
	}
}

// ---- Seq ----

func TestSeq(t *testing.T) {
	q := SeqOf([]int{1, 2, 3, 4, 5})
	if q.Count() != 5 {
		t.Errorf("Count = %d, want 5", q.Count())
	}
	if !q.Any(func(i int) bool { return i == 3 }) {
		t.Error("Any 3")
	}
	if q.All(func(i int) bool { return i < 5 }) {
		t.Error("All should be false (5 not <5)")
	}
	if !q.All(func(i int) bool { return i <= 5 }) {
		t.Error("All <=5 should be true")
	}
	if v, ok := q.Find(func(i int) bool { return i > 2 }); !ok || v != 3 {
		t.Errorf("Find = (%v,%v), want (3,true)", v, ok)
	}
	if v, ok := q.First(); !ok || v != 1 {
		t.Errorf("First = (%v,%v), want (1,true)", v, ok)
	}
}

func TestSeqFilterMap(t *testing.T) {
	q := SeqOf([]int{1, 2, 3, 4, 5, 6})
	evens := q.Filter(func(i int) bool { return i%2 == 0 })
	doubled := MapSeq(evens, func(i int) int { return i * 2 })
	eq(t, "seq filter+map", doubled.Collect(), []int{4, 8, 12})

	// short-circuit: Any stops at the first match.
	hits := 0
	big := SeqOf(make([]int, 1000))
	big.Any(func(i int) bool { hits++; return true })
	if hits != 1 {
		t.Errorf("Any short-circuit hits = %d, want 1", hits)
	}
}

// ---- Benchmarks (vs v2 baseline feel) ----

func BenchmarkMap(b *testing.B) {
	list := make([]int, 1000)
	for i := range list {
		list[i] = i
	}
	f := func(i int) int { return i * 2 }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Map(list, f)
	}
}

func BenchmarkFilter(b *testing.B) {
	list := make([]int, 1000)
	for i := range list {
		list[i] = i
	}
	p := func(i int) bool { return i%2 == 0 }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Filter(list, p)
	}
}

func BenchmarkFilterInPlace(b *testing.B) {
	list := make([]int, 1000)
	for i := range list {
		list[i] = i
	}
	p := func(i int) bool { return i%2 == 0 }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FilterInPlace(append(list[:0:0], list...), p)
	}
}

func BenchmarkTopK(b *testing.B) {
	list := make([]int, 10000)
	for i := range list {
		list[i] = (i * 7) % 10000
	}
	key := Key[int, int](func(i int) int { return i })
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TopK(list, 100, key)
	}
}

func BenchmarkChain(b *testing.B) {
	list := make([]int, 1000)
	for i := range list {
		list[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ChainOf(list).
			Filter(func(i int) bool { return i%2 == 0 }).
			Map(func(i int) int { return i * 2 }).
			TopK(100, func(i int) int { return i }).
			Collect()
	}
}
