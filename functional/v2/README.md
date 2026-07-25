# functional

泛型函数式编程：Map、Filter、Reduce、GroupBy、TopK、Flatten、Partition 等，覆盖 Go 切片操作中最常用的函数式套路，替代手写 for 循环样板代码。

## 功能特性

- 核心变换：`Map`、`Filter`、`Reduce`、`ForEach`、`Flatten`、`Reverse`/`ReverseInPlace`
- 查找与统计：`Find` 返回首个匹配元素，`Count` 统计满足条件的元素个数
- 谓词判断：`All`、`Any`、`None` 判断切片整体是否满足条件，`Contains`/`ContainsBy` 判断元素是否存在
- 聚合计算：`Min`/`Max`/`Sum` 及其 `MinBy`/`MaxBy` 变体，`TopK` 基于堆实现的 Top-K 排序
- 去重与分组：`Deduplicate`/`DeduplicateBy` 保序去重，`GroupBy` 按 key 分组，`Partition` 按谓词一分为二
- 条件表达式：`If`（三元运算）、`When`（条件应用函数）、`IfElse`（惰性分支求值）
- 全部基于 Go 泛型实现，无第三方依赖，可直接作用于任意元素类型的切片

## 快速开始

```go
import function "github.com/tenz-io/gokit/functional/v2"

func main() {
	type User struct {
		ID     int
		Name   string
		Active bool
		Score  int
	}

	users := []User{
		{ID: 1, Name: "alice", Active: true, Score: 90},
		{ID: 2, Name: "bob", Active: false, Score: 70},
		{ID: 3, Name: "carol", Active: true, Score: 85},
	}

	ids := function.Map(users, func(u User) int { return u.ID })
	active := function.Filter(users, func(u User) bool { return u.Active })
	totalScore := function.Reduce(active, func(acc int, u User) int { return acc + u.Score }, 0)

	byActive := function.GroupBy(users, func(u User) bool { return u.Active })
	top2 := function.TopK(users, 2, func(a, b User) bool { return a.Score < b.Score })

	fmt.Println(ids, totalScore, len(byActive[true]), top2)
}
```

## API 速查

| 名称 | 说明 |
|------|------|
| `Map[T, U](list []T, mapper func(T) U) []U` | 对每个元素做映射转换 |
| `Filter[T](list []T, predicate func(T) bool) []T` | 筛选出满足条件的元素 |
| `Reduce[T, U](list []T, reducer func(U, T) U, initial U) U` | 将切片折叠为单个值 |
| `ForEach[T](list []T, fn func(T))` | 对每个元素执行副作用函数 |
| `Flatten[T](list [][]T) []T` | 将二维切片拼接为一维切片 |
| `Reverse[T](list []T) []T` | 返回反转后的新切片 |
| `ReverseInPlace[T](list []T)` | 原地反转切片 |
| `Find[T](list []T, predicate func(T) bool) (T, bool)` | 返回首个满足条件的元素 |
| `Count[T](list []T, predicate func(T) bool) int` | 统计满足条件的元素个数 |
| `All[T](list []T, predicate func(T) bool) bool` | 判断是否全部元素满足条件 |
| `Any[T](list []T, predicate func(T) bool) bool` | 判断是否存在满足条件的元素 |
| `None[T](list []T, predicate func(T) bool) bool` | 判断是否没有元素满足条件 |
| `Contains[T comparable](list []T, elem T) bool` | 判断元素是否存在于切片中 |
| `ContainsBy[T, K comparable](list []T, key K, keyFn func(T) K) bool` | 按 key 判断是否存在匹配元素 |
| `Min[T cmp.Ordered](list []T) (T, bool)` | 返回切片中的最小值 |
| `Max[T cmp.Ordered](list []T) (T, bool)` | 返回切片中的最大值 |
| `Sum[T cmp.Ordered](list []T) T` | 返回切片元素之和 |
| `MinBy[T](list []T, less func(a, b T) bool) (T, bool)` | 按自定义比较函数返回最小元素 |
| `MaxBy[T](list []T, less func(a, b T) bool) (T, bool)` | 按自定义比较函数返回最大元素 |
| `TopK[T](list []T, k int, less func(a, b T) bool) []T` | 基于堆返回按降序排列的前 K 大元素 |
| `Deduplicate[T comparable](list []T) []T` | 保序去重 |
| `DeduplicateBy[T, K comparable](list []T, keyFn func(T) K) []T` | 按 key 保序去重 |
| `GroupBy[T, K comparable](list []T, keyFn func(T) K) map[K][]T` | 按 key 将元素分组 |
| `Partition[T](list []T, predicate func(T) bool) (matched, unmatched []T)` | 按谓词将切片一分为二 |
| `If[T](cond bool, ifVal, elseVal T) T` | 泛型三元运算符 |
| `When[T](cond bool, val T, fn func(T) T) T` | 条件为真时对值应用函数 |
| `IfElse[T](cond bool, ifFn, elseFn func() T) T` | 惰性求值的条件分支 |

引入路径：`github.com/tenz-io/gokit/functional/v2`（包名为 `function`）。
