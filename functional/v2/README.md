# function

泛型函数式编程工具库 — 为 Go 切片提供零依赖、高性能的函数式操作（`Map`、`Filter`、`Reduce` 等），完全基于 Go 1.21+ 泛型实现。

## 快速开始

```go
import "github.com/tenz-io/gokit/functional"

// Map：对切片每个元素执行转换
ids := function.Map(users, func(u User) int { return u.ID })

// Filter：按条件过滤
activeUsers := function.Filter(users, func(u User) bool { return u.Active })

// Reduce：折叠为单个值
totalAge := function.Reduce(users, func(acc int, u User) int { return acc + u.Age }, 0)

// 链式调用：找到所有活跃用户中年龄最大的 3 位
active := function.Filter(users, func(u User) bool { return u.Active })
top3 := function.TopK(active, 3, func(a, b User) bool { return a.Age < b.Age })
```

## API 参考

### 核心变换

| 函数 | 签名 | 说明 |
|------|------|------|
| `Map` | `Map[T, U any](list []T, mapper func(T) U) []U` | 对每个元素执行映射，返回等长切片。空/nil 切片返回空切片 |
| `Filter` | `Filter[T any](list []T, predicate func(T) bool) []T` | 保留满足条件的所有元素，保持顺序 |
| `Reduce` | `Reduce[T, U any](list []T, reducer func(acc U, elem T) U, initial U) U` | 从左到右折叠，空切片返回 initial |
| `ForEach` | `ForEach[T any](list []T, fn func(T))` | 对每个元素执行副作用操作 |
| `Flatten` | `Flatten[T any](list [][]T) []T` | 展平二维切片，内部会预计算容量避免多次扩容 |
| `Reverse` | `Reverse[T any](list []T) []T` | 返回反转后的新切片，原切片不变 |
| `ReverseInPlace` | `ReverseInPlace[T any](list []T)` | 原地反转切片，nil/空切片无操作 |

### 断言

| 函数 | 签名 | 说明 |
|------|------|------|
| `All` | `All[T any](list []T, predicate func(T) bool) bool` | 所有元素均满足时返回 `true`。空切片返回 `true`（空洞真值） |
| `Any` | `Any[T any](list []T, predicate func(T) bool) bool` | 任意元素满足时返回 `true`。空切片返回 `false` |
| `None` | `None[T any](list []T, predicate func(T) bool) bool` | 没有元素满足时返回 `true`。空切片返回 `true` |
| `Contains` | `Contains[T comparable](list []T, elem T) bool` | 判断切片是否包含指定值 |
| `ContainsBy` | `ContainsBy[T any, K comparable](list []T, key K, keyFn func(T) K) bool` | 判断切片是否存在指定键的元素（直接传入键值，无需构造占位元素） |

### 聚合

| 函数 | 签名 | 说明 |
|------|------|------|
| `Min` | `Min[T cmp.Ordered](list []T) (T, bool)` | 返回最小元素。空切片返回 `(zero, false)` |
| `Max` | `Max[T cmp.Ordered](list []T) (T, bool)` | 返回最大元素。空切片返回 `(zero, false)` |
| `Sum` | `Sum[T cmp.Ordered](list []T) T` | 返回元素之和。空切片返回零值（数学上，空集的和为 0） |
| `MinBy` | `MinBy[T any](list []T, less func(a, b T) bool) (T, bool)` | 依据自定义 `less` 函数返回最小元素 |
| `MaxBy` | `MaxBy[T any](list []T, less func(a, b T) bool) (T, bool)` | 依据自定义 `less` 函数返回最大元素 |
| `TopK` | `TopK[T any](list []T, k int, less func(a, b T) bool) []T` | 返回前 k 个最大元素（降序）。k≤0 或空切片返回空切片；k≥len 时返回全量降序排列。算法复杂度 O(n log k) |

### 条件运算

| 函数 | 签名 | 说明 |
|------|------|------|
| `If` | `If[T any](cond bool, ifVal, elseVal T) T` | 泛型三目运算符 |
| `When` | `When[T any](cond bool, val T, fn func(T) T) T` | 条件为真时对值应用函数，否则返回原值不变 |
| `IfElse` | `IfElse[T any](cond bool, ifFn, elseFn func() T) T` | 惰性求值版三目 — 只执行命中的分支，适合分支开销较大的场景 |

### 变换

| 函数 | 签名 | 说明 |
|------|------|------|
| `Deduplicate` | `Deduplicate[T comparable](list []T) []T` | 去重，保持首次出现的顺序。时间复杂度 O(n) |
| `DeduplicateBy` | `DeduplicateBy[T any, K comparable](list []T, keyFn func(T) K) []T` | 按指定键去重，保持首次出现的顺序 |
| `GroupBy` | `GroupBy[T any, K comparable](list []T, keyFn func(T) K) map[K][]T` | 按键分组，组内保持原始顺序 |
| `Partition` | `Partition[T any](list []T, predicate func(T) bool) (matched, unmatched []T)` | 按条件拆分为两组 |

### 查找

| 函数 | 签名 | 说明 |
|------|------|------|
| `Find` | `Find[T any](list []T, predicate func(T) bool) (T, bool)` | 返回第一个满足条件的元素 |
| `Count` | `Count[T any](list []T, predicate func(T) bool) int` | 返回满足条件的元素个数 |

## 最佳实践

### 善用空切片语义

`nil` 切片和空切片在语义上是等价的——所有函数对 nil 切片都安全（不会 panic）。返回的空切片（`[]T{}`）可以直接用于 JSON 序列化（序列化为 `[]` 而非 `null`）。

```go
// 这些操作都是安全的，不会 panic
function.Map([]int(nil), fn)      // 返回 []U{}
function.Filter([]int(nil), pred) // 返回 []int{}
```

### 选择合适的断言函数

当处理复杂类型时，优先使用 `ContainsBy` 而非构造占位元素：

```go
// 不推荐：构造占位元素
function.Contains(users, User{ID: 42})

// 推荐：直接按键查找
function.ContainsBy(users, 42, func(u User) int { return u.ID })
```

### 链式组合

多个函数可以组合使用表达复杂逻辑：

```go
// 分组统计
groups := function.GroupBy(orders, func(o Order) string { return o.Product })
totals := function.Map(groups, func(k string, orders []Order) Summary {
    sum := function.Reduce(orders, func(acc int, o Order) int { return acc + o.Amount }, 0)
    return Summary{Product: k, Total: sum}
})

// 验证
isValid := function.All(items, func(i Item) bool {
    return i.Price > 0 && i.Stock > 0
})
```

### 避免在热路径上大量分配

`Map`、`Filter` 等函数每次调用都会分配新切片。在热路径上可以复用已有的切片：

```go
// 需要新切片时直接用 Map
newIDs := function.Map(users, func(u User) int { return u.ID })

// 热路径上可以用 ForEach 避免分配
var ids []int
function.ForEach(users, func(u User) { ids = append(ids, u.ID) })
```

### TopK 性能提示

当 k << len(list) 时，`TopK` 使用最小堆，复杂度 O(n log k)，远优于全局排序的 O(n log n)。但当 k 接近 len 时，退化为全量排序，与 `slices.SortFunc` 性能相当。

## 与 v1 的区别（破坏性变更）

本版本相对于原 `functional` 模块做了大范围重构，不保证向后兼容：

| 变更项 | v1 | v2 |
|--------|----|----|
| `All` 空切片行为 | `false` | `true`（空洞真值） |
| `None` 空切片行为 | 与 `All` 不一致 | `true`，与 `All` 一致 |
| `IfThen` | 存在 | 重命名为 `When` |
| `MinWith` / `MaxWith` | 存在 | 重命名为 `MinBy` / `MaxBy` |
| `DeduplicateWith` | 存在 | 重命名为 `DeduplicateBy` |
| `ContainsWith` | `(list, elem, keyFn)` | 重命名为 `ContainsBy`，签名改为 `(list, key, keyFn)` |
| `TopK` cmp 参数 | `func(T, T) int` | `func(T, T) bool`（less 模式） |
| 新增函数 | — | `Find`、`Count`、`ReverseInPlace` |
| nil 切片返回值 | `nil` | 空切片 `[]T{}`（JSON 友好） |
| `TopK` k | 仅处理 `==0` | 同时处理 `<=0` |

## 性能特征

| 函数 | 时间复杂度 | 空间复杂度 |
|------|-----------|-----------|
| `Map` | O(n) | O(n) |
| `Filter` | O(n) | O(n) |
| `Reduce` | O(n) | O(1) |
| `Flatten` | O(n) | O(n)，预计算容量无额外扩容 |
| `Reverse` | O(n) | O(n) |
| `ReverseInPlace` | O(n/2) | O(1) |
| `TopK` | O(n log k) | O(k) |
| `Deduplicate` | O(n) | O(n) |
| 其余函数 | O(n) | O(1) |
