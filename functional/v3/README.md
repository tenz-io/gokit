# functional

泛型函数式编程工具库：Map / Filter / Reduce / TopK / Chunk / Zip / OrderedSet 等，覆盖 Go 切片操作中最常用的函数式套路，并以**双轨 API**（独立函数 + 流式 Chain + 惰性 Seq）替代手写 for 循环样板代码。

```go
import fn "github.com/tenz-io/gokit/functional/v3"
```

## V3 相对 V2 的核心变化

- **双轨 API**：① 独立函数（v2 超集，补齐索引版 / 就地版 / key 提取版）；② 流式 `Chain` 链式 `ChainOf(s).Map(...).Filter(...).TopK(k, key).Collect()`；③ 惰性 `Seq` 回调迭代器，支持 `Any/All/Find` 短路读，零分配。
- **key 提取器优先**：`TopK` / `BottomK` / `MinByKey` / `MaxByKey` 用 `Key[T, K cmp.Ordered]` 提取器，无需手写 `less func(a,b)bool`；复杂排序再用 cmp 风格比较器 `By[T]` 的 `*By` 变体。
- **就地变体**：`MapInPlace` / `FilterInPlace` / `DeduplicateInPlace` / `ReverseInPlace` / `PartitionInPlace` 复用底层数组，零分配；非破坏版本（返回新切片）同时保留。
- **能力补齐**：新增 `MapIdx` / `FilterIdx` / `ReduceIdx`（带索引）、`FlatMap`、`Chunk` / `Window` / `Zip` / `Concat` / `Repeat`、`FindIndex` / `FindLast` / `IndexOf` / `LastIndexOf`、`GroupByCount`、`Avg`、`Coalesce` / `Default`、`OrderedSet`（保序集合，O(1) 成员判定）。
- **纯函数语义**：除名称以 `InPlace` 结尾的变体，所有操作都不修改入参。

> 集合代数（Union / Intersect / Difference）请使用 `collection/v2` 的 `Set`；本包的 `OrderedSet` 聚焦保序去重与成员判定。

## 快速开始

```go
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
}

func main() {
	users := []User{
		{1, "alice", true, 90},
		{2, "bob", false, 70},
		{3, "carol", true, 85},
	}

	// 独立函数
	ids := fn.Map(users, func(u User) int { return u.ID })
	active := fn.Filter(users, func(u User) bool { return u.Active })
	totalScore := fn.Reduce(active, func(acc int, u User) int { return acc + u.Score }, 0)

	// TopK：key 提取器，无需手写 less
	top2 := fn.TopK(users, 2, fn.Key[User, int](func(u User) int { return u.Score }))

	// 流式 Chain
	results := fn.ChainOf(users).
		Filter(func(u User) bool { return u.Active }).
		TopK(2, func(u User) int { return u.Score }).
		Collect()

	// 类型变换（自由函数，方法无法改变类型参数）
	names := fn.MapTo(fn.ChainOf(users), func(u User) string { return u.Name }).Collect()

	// 惰性 Seq + 短路
	has100 := fn.SeqOf(users).Any(func(u User) bool { return u.Score >= 100 })

	// 就地零分配
	scores := []int{1, 2, 3, 4}
	fn.MapInPlace(scores, func(i int) int { return i * 10 }) // scores == [10,20,30,40]

	// 保序集合：一次构建，O(1) 反复判定
	seen := fn.NewOrderedSet(ids...)
	_ = seen.Contains(3) // true

	fmt.Println(ids, totalScore, len(results), names, top2, has100)
}
```

## 能力清单

| 能力 | 含义 |
|------|------|
| 切片映射与折叠 | `Map`/`MapIdx`/`MapInPlace` 批量转换（含就地），`Reduce`/`ReduceIdx` 折叠为汇总值，`ForEach`/`ForEachIdx` 副作用遍历 |
| 二维扁平化 | `Flatten` 拼接嵌套切片，`FlatMap` 合并 Map+Flatten 为单次扫描 |
| 反转 | `Reverse` 返回新切片，`ReverseInPlace` 原地反转 |
| 分块与窗口 | `Chunk` 定长分块，`Window` 滑动窗口，`Zip` 按索引配对，`Concat` 拼接，`Repeat` 重复填充 |
| 条件筛选与统计 | `Filter`/`FilterIdx`/`FilterInPlace` 按谓词挑出元素，`Count`/`CountBy` 计数 |
| 查找与索引 | `Find`/`FindIndex`/`FindLast`/`FindLastIndex` 谓词查找，`IndexOf`/`LastIndexOf` 按值定位 |
| 整体谓词判断 | `All`/`Any`/`None` 短路判断整体是否满足/存在/全不满足 |
| 元素存在性 | `Contains` 按值，`ContainsBy` 按 key，`OrderedSet.Contains` O(1) 反复判定 |
| 极值与求和 | `Min`/`Max`/`Sum`/`Avg`，`MinByKey`/`MaxByKey` 按 key 提取器，`MinBy`/`MaxBy` 按比较器 |
| Top-K 排序 | `TopK`/`BottomK` 按 key 提取器，`TopKBy`/`BottomKBy` 按比较器；基于 min-heap O(n·log k) |
| 保序去重 | `Deduplicate`/`DeduplicateBy` 返回新切片，`DeduplicateInPlace`/`DeduplicateByInPlace` 原地紧凑 |
| 分组与拆分 | `GroupBy` 按 key 分组，`GroupByCount` 按 key 计数，`Partition` 一分为二，`PartitionInPlace` 原地稳定分区 |
| 保序集合 | `OrderedSet[T]`：插入序 + O(1) 成员判定 + O(1) 删除（墓碑标记） |
| 条件表达式 | `If`（三元）/ `When`（条件应用）/ `IfElse`（惰性分支）/ `Coalesce`（首个非零）/ `Default`（零值兜底） |
| 流式链 | `ChainOf(s).Map().Filter().FlatMap().Take().Drop().DeduplicateBy().Reverse().SortBy().TopK().Concat().Collect()` |
| 类型变换 | `MapTo(chain, f)` 流式类型变换，`SortChain` 有序类型排序，`MapSeq` 惰性类型变换 |
| 惰性迭代 | `SeqOf(s).Filter().Any()/All()/Find()/ForEach()/Count()/First()/Collect()`，短路零分配 |

## API 速查

### 变换

| 名称 | 说明 |
|------|------|
| `Map[T,U](s, f func(T) U) []U` | 对每个元素映射转换，返回新切片 |
| `MapIdx[T,U](s, f func(int, T) U) []U` | 带索引的 Map |
| `MapInPlace[T](s, f func(T) T) []T` | 原地改写，长度不变，零分配 |
| `Filter[T](s, pred func(T) bool) []T` | 返回满足条件的新切片 |
| `FilterIdx[T](s, pred func(int, T) bool) []T` | 带索引的 Filter |
| `FilterInPlace[T](s, pred func(T) bool) []T` | 紧凑重写 s[:k]，复用底层数组 |
| `Reduce[T,U](s, reducer func(U,T) U, initial U) U` | 折叠为汇总值 |
| `ReduceIdx[T,U](s, reducer func(U,int,T) U, initial U) U` | 带索引的 Reduce |
| `ForEach[T](s, fn func(T))` / `ForEachIdx[T](s, fn func(int, T))` | 副作用遍历 |
| `Flatten[T](s [][]T) []T` | 拼接二维切片 |
| `FlatMap[T,U](s, f func(T) []U) []U` | Map + Flatten 单次扫描 |
| `Reverse[T](s) []T` / `ReverseInPlace[T](s) []T` | 反转（新切片 / 原地） |
| `Chunk[T](s, n int) [][]T` | 定长分块，末块可短 |
| `Window[T](s, n int) [][]T` | 滑动窗口 |
| `Zip[A,B](a, b) []Pair[A,B]` | 按索引配对 |
| `Concat[T](slices ...[]T) []T` | 拼接多切片 |
| `Repeat[T](v T, count int) []T` | 重复填充 |

### 查找与谓词

| 名称 | 说明 |
|------|------|
| `Find[T](s, pred) (T, bool)` / `FindLast[T](s, pred) (T, bool)` | 首/末个满足谓词的元素 |
| `FindIndex[T](s, pred) (int, bool)` / `FindLastIndex[T](s, pred) (int, bool)` | 首/末个满足谓词的索引 |
| `IndexOf[T comparable](s, v) (int, bool)` / `LastIndexOf[T comparable](s, v) (int, bool)` | 按值定位索引 |
| `Contains[T comparable](s, v) bool` / `ContainsBy[T,K](s, key, keyFn) bool` | 存在性判定 |
| `Count[T](s, pred) int` / `CountBy[T,K](s, key, keyFn) int` | 计数 |
| `All[T](s, pred) bool` / `Any[T](s, pred) bool` / `None[T](s, pred) bool` | 短路整体判断 |

### 聚合

| 名称 | 说明 |
|------|------|
| `Min[T cmp.Ordered](s) (T, bool)` / `Max[T cmp.Ordered](s) (T, bool)` | 有序类型极值 |
| `Sum[T cmp.Ordered](s) T` | 求和（字符串可拼接） |
| `Avg[T number](s) (float64, bool)` | 数值均值 |
| `MinByKey[T,K](s, key Key[T,K]) (T, bool)` / `MaxByKey[T,K]` | 按 key 提取器极值 |
| `MinBy[T](s, c By[T]) (T, bool)` / `MaxBy[T](s, c By[T]) (T, bool)` | 按比较器极值 |
| `TopK[T,K](s, k int, key Key[T,K]) []T` | k 大元素，降序；O(n·log k) |
| `BottomK[T,K](s, k int, key Key[T,K]) []T` | k 小元素，升序 |
| `TopKBy[T](s, k int, c By[T]) []T` / `BottomKBy[T](s, k int, c By[T]) []T` | 按比较器的 Top/BottomK |

### 去重 / 分组 / 拆分

| 名称 | 说明 |
|------|------|
| `Deduplicate[T comparable](s) []T` / `DeduplicateBy[T,K](s, keyFn) []T` | 保序去重，返回新切片 |
| `DeduplicateInPlace[T comparable](s) []T` / `DeduplicateByInPlace[T,K](s, keyFn) []T` | 原地紧凑去重 |
| `Unique[T comparable]` / `UniqueBy[T,K]` | Deduplicate 别名 |
| `GroupBy[T,K](s, keyFn) map[K][]T` | 按 key 分组 |
| `GroupByCount[T,K](s, keyFn) map[K]int` | 按 key 计数 |
| `Partition[T](s, pred) (matched, unmatched []T)` | 一分为二 |
| `PartitionInPlace[T](s, pred) int` | 原地稳定分区，返回 matched 数量 |

### 保序集合

| 名称 | 说明 |
|------|------|
| `NewOrderedSet[T](vs ...T) *OrderedSet[T]` / `NewOrderedSetWithCapacity[T](cap) *OrderedSet[T]` | 构造 |
| `(*OrderedSet).Add(v) bool` / `Contains(v) bool` / `Has(v) bool` | 添加 / 判定 |
| `(*OrderedSet).Remove(v) bool` | O(1) 删除（墓碑标记） |
| `(*OrderedSet).Len() int` / `IsEmpty() bool` | 规模 |
| `(*OrderedSet).ToSlice() []T` / `ForEach(fn)` | 插入序输出（跳过墓碑） |
| `(*OrderedSet).Clone() *OrderedSet[T]` | 复制并压缩墓碑 |

### 条件表达式

| 名称 | 说明 |
|------|------|
| `If[T](cond, ifVal, elseVal) T` | 泛型三元（立即求值） |
| `When[T](cond, val, fn func(T) T) T` | 条件为真时应用函数 |
| `IfElse[T](cond, ifFn, elseFn func() T) T` | 惰性分支求值 |
| `Coalesce[T comparable](vs ...T) T` | 取首个非零值 |
| `Default[T comparable](v, def) T` | 零值兜底（Coalesce 二参特化） |

### 提取器类型

| 名称 | 说明 |
|------|------|
| `Key[T any, K cmp.Ordered] func(T) K` | 从 T 提取有序 key，用于 TopK/BottomK/MinByKey/MaxByKey |
| `By[T any] func(a, b T) int` | cmp 风格比较器（负/零/正），用于 *By 变体 |
| `Pair[A,B]` | Zip 返回的二元组 |

### 流式 Chain

| 名称 | 说明 |
|------|------|
| `ChainOf[T](s) Chain[T]` | 构造链 |
| `MapTo[T,U](c, f func(T) U) Chain[U]` | 类型变换 Map |
| `SortChain[T cmp.Ordered](c) Chain[T]` | 有序类型升序排序 |
| `(Chain).Map / Filter / FilterIdx / FlatMap / Take / Drop / Reverse / SortBy / TopK / Concat / Collect / ToSlice / ForEach / Reduce / Any / All / Find` | 链式方法 |
| `DeduplicateByChain[T,K](c, keyFn) Chain[T]` | 链式去重（自由函数，因方法不能加类型参数） |

### 惰性 Seq

| 名称 | 说明 |
|------|------|
| `SeqOf[T](s) Seq[T]` | 从切片构造惰性迭代器 |
| `(Seq).Filter(pred) Seq[T]` | 惰性过滤 |
| `MapSeq[T,U](q, f func(T) U) Seq[U]` | 惰性类型变换 |
| `(Seq).Any / All / Find / First / ForEach / Count / Collect` | 短路消费 |

## 设计说明

- **就地 vs 非破坏**：名称以 `InPlace` 结尾的变体复用入参底层数组并返回其前缀视图；其余均返回独立新切片。就地变体的删除尾部会被 `clear` 清零以避免指针/接口的意外留存。
- **为什么 Chain 每步物化切片**：Go 里多层闭包间接调用的开销往往高于一次切片拷贝，物化既更快也更易调试；需要短路/零分配时用 `Seq`。
- **为什么 key 提取器优先**：`Key[T,K cmp.Ordered]` 覆盖了绝大多数排序场景（按某字段），比手写 `less func(a,b)bool` 直观；多字段/非 key 排序用 `By[T]` 比较器的 `*By` 变体。
- **为什么方法不能加类型参数**：Go 1.21 不允许方法声明类型参数，故类型变换（`MapTo`）、有序排序（`SortChain`）、链式去重（`DeduplicateByChain`）以自由函数提供。
- **OrderedSet 的墓碑删除**：`Remove` 标记槽位为逻辑空而非搬运切片，使删除 O(1)；`ToSlice`/`ForEach` 跳过墓碑，墓碑由 `Clone` 回收。零值作为合法元素不会与墓碑冲突（判定以 `index` map 为准）。

引入路径：`github.com/tenz-io/gokit/functional/v3`（包名 `fn`）。
