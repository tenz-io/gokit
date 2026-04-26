package function

import (
	"reflect"
	"testing"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name   string
		list   []int
		mapper func(int) string
		want   []string
	}{
		{
			name:   "nil slice",
			list:   nil,
			mapper: func(i int) string { return string(rune('0' + i)) },
			want:   []string{},
		},
		{
			name:   "empty slice",
			list:   []int{},
			mapper: func(i int) string { return string(rune('0' + i)) },
			want:   []string{},
		},
		{
			name:   "single element",
			list:   []int{5},
			mapper: func(i int) string { return string(rune('0' + i)) },
			want:   []string{"5"},
		},
		{
			name:   "multiple elements",
			list:   []int{1, 2, 3},
			mapper: func(i int) string { return string(rune('0' + i)) },
			want:   []string{"1", "2", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Map(tt.list, tt.mapper)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name      string
		list      []int
		predicate func(int) bool
		want      []int
	}{
		{
			name:      "nil slice",
			list:      nil,
			predicate: func(i int) bool { return i%2 == 0 },
			want:      []int{},
		},
		{
			name:      "empty slice",
			list:      []int{},
			predicate: func(i int) bool { return i%2 == 0 },
			want:      []int{},
		},
		{
			name:      "no match",
			list:      []int{1, 3, 5},
			predicate: func(i int) bool { return i%2 == 0 },
			want:      []int{},
		},
		{
			name:      "all match",
			list:      []int{2, 4, 6},
			predicate: func(i int) bool { return i%2 == 0 },
			want:      []int{2, 4, 6},
		},
		{
			name:      "some match",
			list:      []int{1, 2, 3, 4, 5},
			predicate: func(i int) bool { return i%2 == 0 },
			want:      []int{2, 4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filter(tt.list, tt.predicate)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReduce(t *testing.T) {
	// int reduction
	if got := Reduce([]int{1, 2, 3, 4}, func(acc, e int) int { return acc + e }, 0); got != 10 {
		t.Errorf("Reduce sum = %v, want 10", got)
	}
	// nil slice returns initial
	if got := Reduce([]int(nil), func(acc, e int) int { return acc + e }, 10); got != 10 {
		t.Errorf("Reduce nil = %v, want 10", got)
	}
	// empty slice returns initial
	if got := Reduce([]int{}, func(acc, e int) int { return acc + e }, 5); got != 5 {
		t.Errorf("Reduce empty = %v, want 5", got)
	}
	// string concat
	if got := Reduce([]string{"a", "b", "c"}, func(acc string, e string) string { return acc + e }, ""); got != "abc" {
		t.Errorf("Reduce concat = %v, want abc", got)
	}
}

func TestForEach(t *testing.T) {
	var sum int
	ForEach([]int{1, 2, 3}, func(i int) { sum += i })
	if sum != 6 {
		t.Errorf("ForEach() sum = %v, want 6", sum)
	}

	// nil slice should not panic
	ForEach([]int(nil), func(i int) { sum += i })
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name string
		list [][]int
		want []int
	}{
		{
			name: "nil slice",
			list: nil,
			want: []int{},
		},
		{
			name: "empty outer",
			list: [][]int{},
			want: []int{},
		},
		{
			name: "empty inner",
			list: [][]int{{}, {}},
			want: []int{},
		},
		{
			name: "single group",
			list: [][]int{{1, 2, 3}},
			want: []int{1, 2, 3},
		},
		{
			name: "multiple groups",
			list: [][]int{{1, 2}, {3, 4}, {5}},
			want: []int{1, 2, 3, 4, 5},
		},
		{
			name: "mixed empty",
			list: [][]int{{1}, {}, {2, 3}},
			want: []int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Flatten(tt.list)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Flatten() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name string
		list []int
		want []int
	}{
		{"nil slice", nil, []int{}},
		{"empty", []int{}, []int{}},
		{"single", []int{1}, []int{1}},
		{"two elements", []int{1, 2}, []int{2, 1}},
		{"odd count", []int{1, 2, 3}, []int{3, 2, 1}},
		{"even count", []int{1, 2, 3, 4}, []int{4, 3, 2, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reverse(tt.list)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reverse() = %v, want %v", got, tt.want)
			}
			// original must be unchanged
			if len(tt.list) > 0 {
				reversed := Reverse(got)
				if !reflect.DeepEqual(reversed, tt.list) {
					t.Errorf("Reverse() not reversible: reversed=%v, original=%v", reversed, tt.list)
				}
			}
		})
	}
}

func TestReverseInPlace(t *testing.T) {
	tests := []struct {
		name string
		list []int
		want []int
	}{
		{"nil slice", nil, nil},
		{"empty", []int{}, []int{}},
		{"single", []int{1}, []int{1}},
		{"two", []int{1, 2}, []int{2, 1}},
		{"odd", []int{1, 2, 3}, []int{3, 2, 1}},
		{"even", []int{1, 2, 3, 4}, []int{4, 3, 2, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ReverseInPlace(tt.list)
			if !reflect.DeepEqual(tt.list, tt.want) {
				t.Errorf("ReverseInPlace() = %v, want %v", tt.list, tt.want)
			}
		})
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name      string
		list      []int
		predicate func(int) bool
		wantVal   int
		wantOk    bool
	}{
		{"nil slice", nil, func(i int) bool { return i > 2 }, 0, false},
		{"empty", []int{}, func(i int) bool { return i > 2 }, 0, false},
		{"found first", []int{1, 3, 5}, func(i int) bool { return i > 2 }, 3, true},
		{"found middle", []int{1, 2, 3, 4}, func(i int) bool { return i%2 == 0 }, 2, true},
		{"not found", []int{1, 3, 5}, func(i int) bool { return i > 10 }, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Find(tt.list, tt.predicate)
			if got != tt.wantVal || ok != tt.wantOk {
				t.Errorf("Find() = (%v, %v), want (%v, %v)", got, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name      string
		list      []int
		predicate func(int) bool
		want      int
	}{
		{"nil slice", nil, func(i int) bool { return i%2 == 0 }, 0},
		{"empty", []int{}, func(i int) bool { return i%2 == 0 }, 0},
		{"none match", []int{1, 3, 5}, func(i int) bool { return i%2 == 0 }, 0},
		{"all match", []int{2, 4, 6}, func(i int) bool { return i%2 == 0 }, 3},
		{"some match", []int{1, 2, 3, 4, 5}, func(i int) bool { return i%2 == 0 }, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Count(tt.list, tt.predicate); got != tt.want {
				t.Errorf("Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAll(t *testing.T) {
	tests := []struct {
		name      string
		list      []int
		predicate func(int) bool
		want      bool
	}{
		{"nil slice vacuous truth", nil, func(i int) bool { return i%2 == 0 }, true},
		{"empty vacuous truth", []int{}, func(i int) bool { return i%2 == 0 }, true},
		{"all satisfy", []int{2, 4, 6}, func(i int) bool { return i%2 == 0 }, true},
		{"one fails", []int{2, 3, 4}, func(i int) bool { return i%2 == 0 }, false},
		{"none satisfy", []int{1, 3, 5}, func(i int) bool { return i%2 == 0 }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := All(tt.list, tt.predicate); got != tt.want {
				t.Errorf("All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAny(t *testing.T) {
	tests := []struct {
		name      string
		list      []int
		predicate func(int) bool
		want      bool
	}{
		{"nil slice", nil, func(i int) bool { return i%2 == 0 }, false},
		{"empty", []int{}, func(i int) bool { return i%2 == 0 }, false},
		{"one matches", []int{1, 2, 3}, func(i int) bool { return i%2 == 0 }, true},
		{"all match", []int{2, 4, 6}, func(i int) bool { return i%2 == 0 }, true},
		{"none match", []int{1, 3, 5}, func(i int) bool { return i%2 == 0 }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Any(tt.list, tt.predicate); got != tt.want {
				t.Errorf("Any() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNone(t *testing.T) {
	tests := []struct {
		name      string
		list      []int
		predicate func(int) bool
		want      bool
	}{
		{"nil slice", nil, func(i int) bool { return i%2 == 0 }, true},
		{"empty", []int{}, func(i int) bool { return i%2 == 0 }, true},
		{"none match", []int{1, 3, 5}, func(i int) bool { return i%2 == 0 }, true},
		{"one matches", []int{1, 2, 3}, func(i int) bool { return i%2 == 0 }, false},
		{"all match", []int{2, 4, 6}, func(i int) bool { return i%2 == 0 }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := None(tt.list, tt.predicate); got != tt.want {
				t.Errorf("None() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name string
		list []int
		elem int
		want bool
	}{
		{"nil slice", nil, 1, false},
		{"empty", []int{}, 1, false},
		{"found first", []int{1, 2, 3}, 1, true},
		{"found middle", []int{1, 2, 3}, 2, true},
		{"found last", []int{1, 2, 3}, 3, true},
		{"not found", []int{1, 2, 3}, 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.list, tt.elem); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsBy(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}
	keyFn := func(u user) int { return u.ID }
	list := []user{{1, "alice"}, {2, "bob"}, {3, "carol"}}

	if got := ContainsBy(list, 2, keyFn); !got {
		t.Errorf("ContainsBy() should find key 2")
	}
	if got := ContainsBy(list, 4, keyFn); got {
		t.Errorf("ContainsBy() should not find key 4")
	}
	if got := ContainsBy([]user{}, 1, keyFn); got {
		t.Errorf("ContainsBy() on empty should return false")
	}
	if got := ContainsBy([]user(nil), 1, keyFn); got {
		t.Errorf("ContainsBy() on nil should return false")
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name    string
		list    []int
		wantVal int
		wantOk  bool
	}{
		{"nil slice", nil, 0, false},
		{"empty", []int{}, 0, false},
		{"single", []int{5}, 5, true},
		{"ascending", []int{1, 2, 3}, 1, true},
		{"descending", []int{3, 2, 1}, 1, true},
		{"mixed", []int{3, 1, 2}, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Min(tt.list)
			if got != tt.wantVal || ok != tt.wantOk {
				t.Errorf("Min() = (%v, %v), want (%v, %v)", got, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name    string
		list    []int
		wantVal int
		wantOk  bool
	}{
		{"nil slice", nil, 0, false},
		{"empty", []int{}, 0, false},
		{"single", []int{5}, 5, true},
		{"ascending", []int{1, 2, 3}, 3, true},
		{"descending", []int{3, 2, 1}, 3, true},
		{"mixed", []int{1, 3, 2}, 3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Max(tt.list)
			if got != tt.wantVal || ok != tt.wantOk {
				t.Errorf("Max() = (%v, %v), want (%v, %v)", got, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name string
		list []int
		want int
	}{
		{"nil slice", nil, 0},
		{"empty", []int{}, 0},
		{"single", []int{5}, 5},
		{"multiple", []int{1, 2, 3, 4}, 10},
		{"negative", []int{-1, 1}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sum(tt.list); got != tt.want {
				t.Errorf("Sum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMinBy(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}
	less := func(a, b user) bool { return a.ID < b.ID }

	tests := []struct {
		name    string
		list    []user
		wantVal user
		wantOk  bool
	}{
		{"nil slice", nil, user{}, false},
		{"empty", []user{}, user{}, false},
		{"single", []user{{1, "a"}}, user{1, "a"}, true},
		{"multiple", []user{{3, "c"}, {1, "a"}, {2, "b"}}, user{1, "a"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := MinBy(tt.list, less)
			if got != tt.wantVal || ok != tt.wantOk {
				t.Errorf("MinBy() = (%v, %v), want (%v, %v)", got, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}

func TestMaxBy(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}
	less := func(a, b user) bool { return a.ID < b.ID }

	tests := []struct {
		name    string
		list    []user
		wantVal user
		wantOk  bool
	}{
		{"nil slice", nil, user{}, false},
		{"empty", []user{}, user{}, false},
		{"single", []user{{1, "a"}}, user{1, "a"}, true},
		{"multiple", []user{{3, "c"}, {1, "a"}, {2, "b"}}, user{3, "c"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := MaxBy(tt.list, less)
			if got != tt.wantVal || ok != tt.wantOk {
				t.Errorf("MaxBy() = (%v, %v), want (%v, %v)", got, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}

func TestTopK(t *testing.T) {
	type item struct {
		ID   int
		Name string
	}
	less := func(a, b item) bool { return a.ID < b.ID }

	tests := []struct {
		name string
		list []item
		k    int
		want []item
	}{
		{
			name: "nil slice",
			list: nil,
			k:    3,
			want: []item{},
		},
		{
			name: "empty",
			list: []item{},
			k:    3,
			want: []item{},
		},
		{
			name: "k zero",
			list: []item{{1, "a"}, {2, "b"}},
			k:    0,
			want: []item{},
		},
		{
			name: "k negative",
			list: []item{{1, "a"}, {2, "b"}},
			k:    -1,
			want: []item{},
		},
		{
			name: "k larger than list",
			list: []item{{1, "a"}, {3, "c"}, {2, "b"}},
			k:    5,
			want: []item{{3, "c"}, {2, "b"}, {1, "a"}},
		},
		{
			name: "k equals list",
			list: []item{{1, "a"}, {3, "c"}, {2, "b"}},
			k:    3,
			want: []item{{3, "c"}, {2, "b"}, {1, "a"}},
		},
		{
			name: "top 2 of 5",
			list: []item{{3, "c"}, {1, "a"}, {5, "e"}, {2, "b"}, {4, "d"}},
			k:    2,
			want: []item{{5, "e"}, {4, "d"}},
		},
		{
			name: "top 3 of 5",
			list: []item{{3, "c"}, {1, "a"}, {5, "e"}, {2, "b"}, {4, "d"}},
			k:    3,
			want: []item{{5, "e"}, {4, "d"}, {3, "c"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TopK(tt.list, tt.k, less)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TopK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTopK_Strings(t *testing.T) {
	list := []string{"banana", "apple", "cherry", "date", "elderberry"}
	less := func(a, b string) bool { return a < b }
	got := TopK(list, 3, less)
	want := []string{"elderberry", "date", "cherry"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TopK() = %v, want %v", got, want)
	}
}

func TestIf(t *testing.T) {
	if got := If(true, 1, 2); got != 1 {
		t.Errorf("If(true) = %v, want 1", got)
	}
	if got := If(false, 1, 2); got != 2 {
		t.Errorf("If(false) = %v, want 2", got)
	}
}

func TestWhen(t *testing.T) {
	double := func(i int) int { return i * 2 }
	if got := When(true, 5, double); got != 10 {
		t.Errorf("When(true) = %v, want 10", got)
	}
	if got := When(false, 5, double); got != 5 {
		t.Errorf("When(false) = %v, want 5", got)
	}
}

func TestIfElse(t *testing.T) {
	if got := IfElse(true, func() int { return 1 }, func() int { return 2 }); got != 1 {
		t.Errorf("IfElse(true) = %v, want 1", got)
	}
	if got := IfElse(false, func() int { return 1 }, func() int { return 2 }); got != 2 {
		t.Errorf("IfElse(false) = %v, want 2", got)
	}

	// Verify lazy evaluation: elseFn should not be called when cond is true
	called := false
	IfElse(true, func() int { return 0 }, func() int { called = true; return 1 })
	if called {
		t.Error("IfElse(true) should not call elseFn")
	}
}

func TestDeduplicate(t *testing.T) {
	tests := []struct {
		name string
		list []int
		want []int
	}{
		{"nil slice", nil, []int{}},
		{"empty", []int{}, []int{}},
		{"no duplicates", []int{1, 2, 3}, []int{1, 2, 3}},
		{"duplicates", []int{1, 2, 2, 3, 1}, []int{1, 2, 3}},
		{"all same", []int{1, 1, 1}, []int{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Deduplicate(tt.list)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Deduplicate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeduplicateBy(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}
	keyFn := func(u user) int { return u.ID }

	tests := []struct {
		name string
		list []user
		want []user
	}{
		{"nil slice", nil, []user{}},
		{"empty", []user{}, []user{}},
		{"no dup keys", []user{{1, "a"}, {2, "b"}}, []user{{1, "a"}, {2, "b"}}},
		{
			"dup keys keeps first",
			[]user{{1, "a"}, {2, "b"}, {1, "c"}},
			[]user{{1, "a"}, {2, "b"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeduplicateBy(tt.list, keyFn)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeduplicateBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroupBy(t *testing.T) {
	type order struct {
		Product string
		Amount  int
	}
	list := []order{
		{"apple", 1},
		{"banana", 2},
		{"apple", 3},
		{"cherry", 4},
		{"banana", 5},
	}
	keyFn := func(o order) string { return o.Product }

	got := GroupBy(list, keyFn)
	if len(got) != 3 {
		t.Errorf("GroupBy() len = %v, want 3", len(got))
	}
	if len(got["apple"]) != 2 {
		t.Errorf("GroupBy() apple count = %v, want 2", len(got["apple"]))
	}
	if len(got["banana"]) != 2 {
		t.Errorf("GroupBy() banana count = %v, want 2", len(got["banana"]))
	}
	if len(got["cherry"]) != 1 {
		t.Errorf("GroupBy() cherry count = %v, want 1", len(got["cherry"]))
	}
}

func TestGroupBy_Empty(t *testing.T) {
	got := GroupBy([]int{}, func(i int) string { return "key" })
	if len(got) != 0 {
		t.Errorf("GroupBy() on empty should return empty map, got %v", got)
	}
}

func TestPartition(t *testing.T) {
	tests := []struct {
		name      string
		list      []int
		predicate func(int) bool
		matched   []int
		unmatched []int
	}{
		{
			name:      "nil slice",
			list:      nil,
			predicate: func(i int) bool { return i%2 == 0 },
			matched:   []int{},
			unmatched: []int{},
		},
		{
			name:      "empty",
			list:      []int{},
			predicate: func(i int) bool { return i%2 == 0 },
			matched:   []int{},
			unmatched: []int{},
		},
		{
			name:      "all match",
			list:      []int{2, 4, 6},
			predicate: func(i int) bool { return i%2 == 0 },
			matched:   []int{2, 4, 6},
			unmatched: []int{},
		},
		{
			name:      "none match",
			list:      []int{1, 3, 5},
			predicate: func(i int) bool { return i%2 == 0 },
			matched:   []int{},
			unmatched: []int{1, 3, 5},
		},
		{
			name:      "mixed",
			list:      []int{1, 2, 3, 4, 5},
			predicate: func(i int) bool { return i%2 == 0 },
			matched:   []int{2, 4},
			unmatched: []int{1, 3, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatched, gotUnmatched := Partition(tt.list, tt.predicate)
			if !reflect.DeepEqual(gotMatched, tt.matched) {
				t.Errorf("Partition() matched = %v, want %v", gotMatched, tt.matched)
			}
			if !reflect.DeepEqual(gotUnmatched, tt.unmatched) {
				t.Errorf("Partition() unmatched = %v, want %v", gotUnmatched, tt.unmatched)
			}
		})
	}
}

// Benchmarks
func BenchmarkMap(b *testing.B) {
	list := make([]int, 1000)
	for i := range list {
		list[i] = i
	}
	mapper := func(i int) string { return string(rune('0' + i%10)) }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Map(list, mapper)
	}
}

func BenchmarkFilter(b *testing.B) {
	list := make([]int, 1000)
	for i := range list {
		list[i] = i
	}
	pred := func(i int) bool { return i%2 == 0 }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Filter(list, pred)
	}
}

func BenchmarkTopK(b *testing.B) {
	list := make([]int, 10000)
	for i := range list {
		list[i] = (i * 7) % 10000
	}
	less := func(a, b int) bool { return a < b }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TopK(list, 100, less)
	}
}
