package function

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestMap(t *testing.T) {
	type args[T any, U any] struct {
		list []T
		fn   func(T) U
	}
	type testCase[T any, U any] struct {
		name string
		args args[T, U]
		want []U
	}
	tests := []testCase[int, string]{
		{
			name: "when empty list then return empty list",
			args: args[int, string]{
				list: []int{},
				fn:   func(i int) string { return fmt.Sprintf("%d", i) },
			},
			want: []string{},
		},
		{
			name: "when list has elements then return list of results",
			args: args[int, string]{
				list: []int{1, 2, 3},
				fn:   func(i int) string { return fmt.Sprintf("%d", i) },
			},
			want: []string{"1", "2", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got := Map(tt.args.list, tt.args.fn)
			duration := time.Since(start)
			t.Logf("duration: %v, got: %+v", duration, got)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() = %v, want %v", got, tt.want)
				return
			}

		})
	}
}

func TestFilter(t *testing.T) {
	type args[T any] struct {
		list      []T
		predicate func(T) bool
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[int]{
		{
			name: "when empty list then return empty list",
			args: args[int]{
				list: []int{},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			want: []int{},
		},
		{
			name: "when list has elements then return list of results",
			args: args[int]{
				list: []int{1, 2, 3, 4, 5},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			want: []int{2, 4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filter(tt.args.list, tt.args.predicate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroupBy(t *testing.T) {
	type args[T any, K comparable] struct {
		list  []T
		keyFn func(T) K
	}
	type testCase[T any, K comparable] struct {
		name string
		args args[T, K]
		want map[K][]T
	}
	tests := []testCase[int, string]{
		{
			name: "when empty list then return empty map",
			args: args[int, string]{
				list: []int{},
				keyFn: func(i int) string {
					if i%2 == 0 {
						return "even"
					} else {
						return "odd"
					}
				},
			},
			want: map[string][]int{},
		},
		{
			name: "when list has elements then return map of results",
			args: args[int, string]{
				list: []int{1, 2, 3, 4, 5},
				keyFn: func(i int) string {
					if i%2 == 0 {
						return "even"
					} else {
						return "odd"
					}
				},
			},
			want: map[string][]int{
				"even": {2, 4},
				"odd":  {1, 3, 5},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GroupBy(tt.args.list, tt.args.keyFn)
			t.Logf("got: %+v", got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GroupBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlatten(t *testing.T) {
	type args[T any] struct {
		list [][]T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[int]{
		{
			name: "when empty list then return empty list",
			args: args[int]{
				list: [][]int{},
			},
			want: []int{},
		},
		{
			name: "when list has elements then return list of results",
			args: args[int]{
				list: [][]int{{1, 2}, {3, 4}, {5}},
			},
			want: []int{1, 2, 3, 4, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Flatten(tt.args.list); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Flatten() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIf(t *testing.T) {
	type args[T any] struct {
		cond    bool
		ifVal   T
		elseVal T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want T
	}
	tests := []testCase[int]{
		{
			name: "when condition is true then return ifVal",
			args: args[int]{
				cond:    true,
				ifVal:   1,
				elseVal: 2,
			},
			want: 1,
		},
		{
			name: "when condition is false then return elseVal",
			args: args[int]{
				cond:    false,
				ifVal:   1,
				elseVal: 2,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := If(tt.args.cond, tt.args.ifVal, tt.args.elseVal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("If() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIfThen(t *testing.T) {
	type args[T any] struct {
		cond  bool
		val   T
		apply func(T) T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want T
	}
	tests := []testCase[int]{
		{
			name: "when condition is true then apply function",
			args: args[int]{
				cond: true,
				val:  1,
				apply: func(i int) int {
					return i + 1
				},
			},
			want: 2,
		},
		{
			name: "when condition is false then return value",
			args: args[int]{
				cond: false,
				val:  1,
				apply: func(i int) int {
					return i + 1
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IfThen(tt.args.cond, tt.args.val, tt.args.apply); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IfThen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIfElse(t *testing.T) {
	type args[T any] struct {
		cond   bool
		ifFn   func() T
		elseFn func() T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want T
	}
	tests := []testCase[int]{
		{
			name: "when condition is true then return ifFn result",
			args: args[int]{
				cond: true,
				ifFn: func() int {
					return 1
				},
				elseFn: func() int {
					return 2
				},
			},
			want: 1,
		},
		{
			name: "when condition is false then return elseFn result",
			args: args[int]{
				cond: false,
				ifFn: func() int {
					return 1
				},
				elseFn: func() int {
					return 2
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IfElse(tt.args.cond, tt.args.ifFn, tt.args.elseFn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IfElse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeduplicate(t *testing.T) {
	type args[T comparable] struct {
		list []T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[int]{
		{
			name: "when empty list then return empty list",
			args: args[int]{
				list: []int{},
			},
			want: []int{},
		},
		{
			name: "when list has elements then return list of results",
			args: args[int]{
				list: []int{1, 2, 2, 3, 2, 1},
			},
			want: []int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Deduplicate(tt.args.list); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Deduplicate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeduplicateWith(t *testing.T) {
	type item struct {
		id   int
		desc string
	}
	type args[T any, K comparable] struct {
		list  []T
		keyFn func(T) K
	}
	type testCase[T any, K comparable] struct {
		name string
		args args[T, K]
		want []T
	}
	tests := []testCase[item, int]{
		{
			name: "when empty list then return empty list",
			args: args[item, int]{
				list: []item{},
				keyFn: func(i item) int {
					return i.id
				},
			},
			want: []item{},
		},
		{
			name: "when list has elements then return list of results",
			args: args[item, int]{
				list: []item{
					{1, "one"},
					{2, "two"},
					{2, "two"},
					{3, "three"},
					{2, "two"},
					{1, "one"},
				},
				keyFn: func(i item) int {
					return i.id
				},
			},
			want: []item{
				{1, "one"},
				{2, "two"},
				{3, "three"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeduplicateWith(tt.args.list, tt.args.keyFn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeduplicateWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	type args[T any] struct {
		list []T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[int]{
		{
			name: "when empty list then return empty list",
			args: args[int]{
				list: []int{},
			},
			want: []int{},
		},
		{
			name: "when list has one element then return same list",
			args: args[int]{
				list: []int{1},
			},
			want: []int{1},
		},
		{
			name: "when list has two elements then return reversed list",
			args: args[int]{
				list: []int{1, 2},
			},
			want: []int{2, 1},
		},
		{
			name: "when list has odd elements then return reversed list",
			args: args[int]{
				list: []int{1, 2, 3},
			},
			want: []int{3, 2, 1},
		},
		{
			name: "when list has even elements then return reversed list",
			args: args[int]{
				list: []int{1, 2, 3, 4},
			},
			want: []int{4, 3, 2, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reverse(tt.args.list); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reverse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	type args[T comparable] struct {
		list []T
		elem T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[int]{
		{
			name: "when empty list then return false",
			args: args[int]{
				list: []int{},
				elem: 1,
			},
			want: false,
		},
		{
			name: "when list has elements then return true if element is in list",
			args: args[int]{
				list: []int{1, 2, 3},
				elem: 2,
			},
			want: true,
		},
		{
			name: "when list has elements then return false if element is not in list",
			args: args[int]{
				list: []int{1, 2, 3},
				elem: 4,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.args.list, tt.args.elem); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsWith(t *testing.T) {
	type item struct {
		id   int
		desc string
	}
	type args[T any, K comparable] struct {
		list  []T
		elem  T
		keyFn func(T) K
	}
	type testCase[T any, K comparable] struct {
		name string
		args args[T, K]
		want bool
	}
	tests := []testCase[item, int]{
		{
			name: "when empty list then return false",
			args: args[item, int]{
				list: []item{},
				elem: item{1, "one"},
				keyFn: func(i item) int {
					return i.id
				},
			},
			want: false,
		},
		{
			name: "when list has elements then return true if element is in list",
			args: args[item, int]{
				list: []item{
					{1, "one"},
					{2, "two"},
					{3, "three"},
				},
				elem: item{2, "two"},
				keyFn: func(i item) int {
					return i.id
				},
			},
			want: true,
		},
		{
			name: "when list has elements then return false if element is not in list",
			args: args[item, int]{
				list: []item{
					{1, "one"},
					{2, "two"},
					{3, "three"},
				},
				elem: item{4, "four"},
				keyFn: func(i item) int {
					return i.id
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsWith(tt.args.list, tt.args.elem, tt.args.keyFn); got != tt.want {
				t.Errorf("ContainsWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAll(t *testing.T) {
	type args[T any] struct {
		list      []T
		predicate func(T) bool
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[int]{
		{
			name: "when empty list then return false",
			args: args[int]{
				list: []int{},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			want: false,
		},
		{
			name: "when all elements satisfy predicate then return true",
			args: args[int]{
				list: []int{2, 4, 6},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			want: true,
		},
		{
			name: "when some elements do not satisfy predicate then return false",
			args: args[int]{
				list: []int{2, 3, 4},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := All(tt.args.list, tt.args.predicate); got != tt.want {
				t.Errorf("All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAny(t *testing.T) {
	type args[T any] struct {
		list      []T
		predicate func(T) bool
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[int]{
		{
			name: "when empty list then return false",
			args: args[int]{
				list: []int{},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			want: false,
		},
		{
			name: "when some elements satisfy predicate then return true",
			args: args[int]{
				list: []int{1, 2, 3},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			want: true,
		},
		{
			name: "when no elements satisfy predicate then return false",
			args: args[int]{
				list: []int{1, 3, 5},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Any(tt.args.list, tt.args.predicate); got != tt.want {
				t.Errorf("Any() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNone(t *testing.T) {
	type args[T any] struct {
		list      []T
		predicate func(T) bool
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[int]{
		{
			name: "when empty list then return true",
			args: args[int]{
				list: []int{},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			want: true,
		},
		{
			name: "when no elements satisfy predicate then return true",
			args: args[int]{
				list: []int{1, 3, 5},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			want: true,
		},
		{
			name: "when some elements satisfy predicate then return false",
			args: args[int]{
				list: []int{1, 2, 3},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := None(tt.args.list, tt.args.predicate); got != tt.want {
				t.Errorf("None() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTopK(t *testing.T) {
	type item struct {
		id   int
		desc string
	}
	type args[T any] struct {
		list []T
		k    int
		less func(t1, t2 T) int
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[item]{
		{
			name: "when empty list then return empty list",
			args: args[item]{
				list: []item{},
				k:    1,
				less: func(t1, t2 item) int {
					return t1.id - t2.id
				},
			},
			want: []item{},
		},
		{
			name: "when k is greater than list length then return sorted list",
			args: args[item]{
				list: []item{
					{1, "one"},
					{3, "three"},
					{2, "two"},
					{4, "four"},
				},
				k: 5,
				less: func(t1, t2 item) int {
					return t1.id - t2.id
				},
			},
			want: []item{
				{4, "four"},
				{3, "three"},
				{2, "two"},
				{1, "one"},
			},
		},
		{
			name: "when k is cmp than list length then return top k elements",
			args: args[item]{
				list: []item{
					{1, "one"},
					{3, "three"},
					{2, "two"},
					{4, "four"},
				},
				k: 3,
				less: func(t1, t2 item) int {
					return t1.id - t2.id
				},
			},
			want: []item{
				{4, "four"},
				{3, "three"},
				{2, "two"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TopK(tt.args.list, tt.args.k, tt.args.less); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TopK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartition(t *testing.T) {
	type args[T any] struct {
		list      []T
		predicate func(T) bool
	}
	type testCase[T any] struct {
		name        string
		args        args[T]
		satisfied   []T
		unsatisfied []T
	}
	tests := []testCase[int]{
		{
			name: "when empty list then return empty lists",
			args: args[int]{
				list: []int{},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			satisfied:   []int{},
			unsatisfied: []int{},
		},
		{
			name: "when list has elements then return partitioned lists",
			args: args[int]{
				list: []int{1, 2, 3, 4, 5},
				predicate: func(i int) bool {
					return i%2 == 0
				},
			},
			satisfied:   []int{2, 4},
			unsatisfied: []int{1, 3, 5},
		},
		{
			name: "when list has elements split then return partitioned lists",
			args: args[int]{
				list: []int{1, 2, 3, 4, 5},
				predicate: func(i int) bool {
					return i > 3
				},
			},
			satisfied:   []int{4, 5},
			unsatisfied: []int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := Partition(tt.args.list, tt.args.predicate)
			if !reflect.DeepEqual(got, tt.satisfied) {
				t.Errorf("Partition() got = %v, want %v", got, tt.satisfied)
			}
			if !reflect.DeepEqual(got1, tt.unsatisfied) {
				t.Errorf("Partition() got1 = %v, want %v", got1, tt.unsatisfied)
			}
		})
	}
}

func TestSum(t *testing.T) {
	type args[T number] struct {
		list []T
	}
	type testCase[T number] struct {
		name string
		args args[T]
		want T
	}
	tests := []testCase[int]{
		{
			name: "when empty list then return 0",
			args: args[int]{
				list: []int{},
			},
			want: 0,
		},
		{
			name: "when list has elements then return sum",
			args: args[int]{
				list: []int{1, 2, 3},
			},
			want: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sum(tt.args.list); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	type args[T number] struct {
		list []T
	}
	type testCase[T number] struct {
		name    string
		args    args[T]
		wantVal T
		wantOk  bool
	}
	tests := []testCase[int]{
		{
			name: "when empty list then return false",
			args: args[int]{
				list: []int{},
			},
			wantVal: 0,
			wantOk:  false,
		},
		{
			name: "when list has elements then return min",
			args: args[int]{
				list: []int{1, 2, 3},
			},
			wantVal: 1,
			wantOk:  true,
		},
		{
			name: "when list has elements then return min",
			args: args[int]{
				list: []int{3, 2, 1},
			},
			wantVal: 1,
			wantOk:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := Min(tt.args.list)
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("Min() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Min() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestMinWith(t *testing.T) {
	type item struct {
		id   int
		desc string
	}
	type args[T any] struct {
		list []T
		less func(t1, t2 T) bool
	}
	type testCase[T any] struct {
		name    string
		args    args[T]
		wantVal T
		wantOk  bool
	}
	tests := []testCase[item]{
		{
			name: "when empty list then return false",
			args: args[item]{
				list: []item{},
				less: func(t1, t2 item) bool {
					return t1.id < t2.id
				},
			},
			wantVal: item{},
			wantOk:  false,
		},
		{
			name: "when list has elements then return min",
			args: args[item]{
				list: []item{
					{1, "one"},
					{2, "two"},
					{3, "three"},
				},
				less: func(t1, t2 item) bool {
					return t1.id < t2.id
				},
			},
			wantVal: item{1, "one"},
			wantOk:  true,
		},
		{
			name: "when list has raffled elements then return min",
			args: args[item]{
				list: []item{
					{3, "three"},
					{1, "one"},
					{2, "two"},
				},
				less: func(t1, t2 item) bool {
					return t1.id < t2.id
				},
			},
			wantVal: item{1, "one"},
			wantOk:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := MinWith(tt.args.list, tt.args.less)
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("MinWith() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			if gotOk != tt.wantOk {
				t.Errorf("MinWith() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestMax(t *testing.T) {
	type args[T number] struct {
		list []T
	}
	type testCase[T number] struct {
		name    string
		args    args[T]
		wantVal T
		wantOk  bool
	}
	tests := []testCase[int]{
		{
			name: "when empty list then return false",
			args: args[int]{
				list: []int{},
			},
			wantVal: 0,
			wantOk:  false,
		},
		{
			name: "when list has elements then return max",
			args: args[int]{
				list: []int{1, 3, 1, 2},
			},
			wantVal: 3,
			wantOk:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := Max(tt.args.list)
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("Max() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Max() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestMaxWith(t *testing.T) {
	type item struct {
		id   int
		desc string
	}
	type args[T any] struct {
		list []T
		less func(t1, t2 T) bool
	}
	type testCase[T any] struct {
		name    string
		args    args[T]
		wantVal T
		wantOk  bool
	}
	tests := []testCase[item]{
		{
			name: "when empty list then return false",
			args: args[item]{
				list: []item{},
				less: func(t1, t2 item) bool {
					return t1.id < t2.id
				},
			},
			wantVal: item{},
			wantOk:  false,
		},
		{
			name: "when list has elements then return max",
			args: args[item]{
				list: []item{
					{1, "one"},
					{3, "three"},
					{2, "two"},
				},
				less: func(t1, t2 item) bool {
					return t1.id < t2.id
				},
			},
			wantVal: item{3, "three"},
			wantOk:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := MaxWith(tt.args.list, tt.args.less)
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("MaxWith() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			if gotOk != tt.wantOk {
				t.Errorf("MaxWith() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
