# functional V3 全新重写

## 目标
全新实现 `functional/v3`，**不考虑兼容性**。相比 v2：
- 更合理：补齐 v2 缺口（索引版、就地版、按 key 提取的 TopK/Min/Max、Chunk/Window/Zip/FlatMap、OrderedSet、Coalesce/Default）。
- 更高效：提供零分配就地变体复用底层数组；TopK 仍用 min-heap O(n·log k)；成员判定 O(1)。
- 更易用：双轨 API——独立函数（v2 超集）+ 流式 `Chain` 链式 + 轻量惰性 `Seq` 短路读。

包名 `fn`，路径 `github.com/tenz-io/gokit/functional/v3`，`go 1.21`，零第三方依赖（仅 `cmp`/`slices`/`container/heap` 标准库）。

## 设计决策（已与用户确认）
1. **双轨**：独立函数 + 流式 `Chain`；另加惰性 `Seq` 用于短路（Any/All/Find）。
2. **Key 提取器优先**：`TopK/BottomK/Min/Max` 用 `Key[T,K cmp.Ordered]` 提取器；高级版 `*By` 用比较器 `func(a,b T) int`（cmp 风格，非 v2 的 `less bool`）。
3. **就地变体**：`MapInPlace/FilterInPlace/DeduplicateInPlace/ReverseInPlace/SortInPlace` 等。
4. **仅加 OrderedSet**（保序集合，与 `collection/v2.Set` 区隔：那个是无序 map，本版保序+切片访问，聚焦去重/成员判定）。集合代数需求导引到 `collection/v2`。

## API 范围

### 文件布局（`functional/v3/`）
```
go.mod                      module github.com/tenz-io/gokit/functional/v3 ; go 1.21
doc.go                      包注释 + Key/By 提取器类型 + 公共类型
transform.go                Map/MapIdx/MapInPlace, Filter/FilterIdx/FilterInPlace,
                            Reduce/ReduceIdx, ForEach/ForEachIdx, Flatten/FlatMap,
                            Reverse/ReverseInPlace, Chunk, Window, Zip, Concat, Repeat
predicate.go                All/Any/None, Contains/ContainsBy, Count,
                            Find/FindIndex/FindLast/IndexOf/LastIndexOf
aggregate.go                Min/Max/Sum/Avg, MinBy/MaxBy(比较器版),
                            MinByKey/MaxByKey(Key 提取器版), TopK/BottomK(Key), TopKBy/BottomKBy(比较器)
dedupe_group.go             Deduplicate/DeduplicateBy/DeduplicateInPlace/DeduplicateByInPlace,
                            GroupBy/GroupByCount, Partition, Unique/UniqueBy
set.go                      OrderedSet[T comparable] + NewOrderedSet/OrderedSetFrom
                            (Add/Contains/Remove/ToSlice/Len/Has/ForEach/Clone)
conditional.go              If/When/IfElse, Coalesce, Default, Ternary
chain.go                    Chain[T] 流式：Map/Filter/Reduce/FlatMap/Take/Drop/
                            Deduplicate/Reverse/Sort/TopK/BottomK/Chunk/Concat/Collect/ToSlice
seq.go                      Seq[T] 惰性迭代器：Of[T]/Filter/Map/Any/All/Find/ForEach/Count/First
heap.go                     内部 min-heap（泛型，给 TopK 用）
functional_test.go          全量测试（覆盖 v2 全部用例 + 新增能力）
README.md                   能力清单 + 快速开始 + API 速查
Makefile                    复用 monitor/v3 同款
```

### 提取器类型（doc.go）
```go
// Key[T,K] 从 T 提取一个有序 key，用于 TopK/MinByKey/MaxByKey
type Key[T any, K cmp.Ordered] func(T) K
// By[T] 是 cmp 风格比较器 func(a,b T) int，用于 *By 变体
type By[T any] func(a, b T) int
```

### 关键签名示例
```go
func Map[T, U any](s []T, f func(T) U) []U
func MapIdx[T, U any](s []T, f func(int, T) U) []U
func MapInPlace[T any](s []T, f func(T) T)
func Filter[T any](s []T, pred func(T) bool) []T        // 返回新切片
func FilterInPlace[T any](s []T, pred func(T) bool) []T // 紧凑重写 s[:n]，返回 s[:n]
func FlatMap[T, U any](s []T, f func(T) []U) []U
func FindIndex[T any](s []T, pred func(T) bool) (int, bool)
func IndexOf[T comparable](s []T, v T) (int, bool)
func MinByKey[T any, K cmp.Ordered](s []T, key Key[T,K]) (T, bool)
func TopK[T any, K cmp.Ordered](s []T, k int, key Key[T,K]) []T   // 降序
func BottomK[T any, K cmp.Ordered](s []T, k int, key Key[T,K]) []T
func TopKBy[T any](s []T, k int, cmp By[T]) []T
func Chunk[T any](s []T, n int) [][]T
func Window[T any](s []T, n int) [][]T   // 滑动窗口
func Zip[A, B any](a []A, b []B) []struct{ A A; B B }
func Coalesce[T comparable](vs ...T) T  // 取首个非零
func Default[T comparable](v, def T) T
```

### 流式 Chain
```go
type Chain[T any] struct{ s []T }
func ChainOf[T any](s []T) Chain[T]
func (c Chain[T]) Map(...) / Filter(...) / FlatMap(...) / Take(n) / Drop(n) /
     Deduplicate(...) / Reverse(...) / Sort(cmp) / TopK(k, key) / Chunk(n) / Concat(...)
func (c Chain[T]) Collect() []T
```
每步物化为切片（Go 里通常比多层闭包间接调用更快、更易调试）。

### 惰性 Seq（短路读，零分配）
```go
type Seq[T any] func(yield func(T) bool)   // 接口式迭代器
func SeqOf[T any](s []T) Seq[T]
func (q Seq[T]) Filter(pred) Seq[T]; Map(...) Seq[U]
func (q Seq[T]) Any(pred)/All(pred)/Find(pred)/(T,bool)/ForEach(fn)/Count()
```
用 `iter.Seq` 风格回调（go1.21 手写，不依赖 1.23 iter 包，保证 go.mod=1.21 可编译）。

### OrderedSet
```go
type OrderedSet[T comparable] struct{ m map[T]struct{}; order []T }
func NewOrderedSet[T comparable](vs ...T) *OrderedSet[T]
func (s *OrderedSet[T]) Add(v T) bool / Contains(v) bool / Remove(v) bool /
     Len() int / ToSlice() []T / ForEach(f) / Clone() *OrderedSet[T]
```
集合代数（Union/Intersect/Difference）不在此实现，README 指向 collection/v2。

## 实现要点 / 效率
- `FilterInPlace`：双指针紧凑重写 `s[:k]`，复用底层数组，返回 `s[:k]`。
- `MapInPlace`：原地改写，长度不变。
- `DeduplicateInPlace`：map 记 seen + 紧凑重写。
- `TopK`：k≥len 退化为 `slices.SortFunc`；否则 min-heap 维护 k 大。
- `Filter`：预分配 `len/2` 而非 `len`（多数筛选结果更小），溢出时 append 自扩。
- `Map`：`make([]U, len)` 直接下标写（比 append 略快）。
- `Chain` 每步物化切片，避免闭包链。
- `Seq` 回调迭代器，Any/All/Find 短路。

## 测试策略
- `functional_test.go`：表驱动，覆盖 nil/empty/single/多元素/边界。
- 直接迁移 v2 全部测试用例并扩展（索引版、就地版、Chunk/Window/Zip、OrderedSet、Chain、Seq、Coalesce）。
- Benchmark：Map/Filter/TopK/Chain vs v2 对比。
- 运行 `go test ./... -cover -v` 与 `go vet ./...`。

## 落地步骤
1. 建 `functional/v3/go.mod` + `Makefile`（抄 monitor/v3）。
2. 注册到根 `go.work`（`./functional/v3`）。
3. 写 `doc.go` → `transform.go` → `predicate.go` → `aggregate.go` → `dedupe_group.go` → `set.go` → `conditional.go` → `chain.go` → `seq.go` → `heap.go`。
4. 写 `functional_test.go`（全量）。
5. `go test ./... -cover -v` + `go vet ./...` + `gofmt`。
6. 写 `README.md`。
7. （不改其他模块）如引起编译错误，记录在交付说明里留给后续逐个修。

## 不做
- 不动 v2、不修其他模块、不写 example 子目录（可选，先不加）。
- 不引入反射版 ByField（用闭包 Key，更高效）。
- 不实现集合代数（留给 collection/v2）。
