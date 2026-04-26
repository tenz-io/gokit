# collection

泛型数据结构库 — 为 Go 1.21+ 提供零外部依赖的通用容器：`Stack`、`Queue`、`PriorityQueue` 和 `Set`。

## 快速开始

```go
import "github.com/tenz-io/gokit/collection"

// Stack — 后进先出
s := collection.NewStack[int]()
s.Push(1)
s.Push(2)
top, _ := s.Pop() // 2

// Queue — 先进先出
q := collection.NewQueue[string]()
q.Enqueue("a")
q.Enqueue("b")
front, _ := q.Dequeue() // "a"

// PriorityQueue — 最小堆
pq := collection.NewPriorityQueue(func(a, b int) bool { return a < b })
pq.Push(3)
pq.Push(1)
pq.Push(2)
min, _, _ := pq.Pop() // 1, 2, 3 ...

// Set — 集合操作
a := collection.NewSet(1, 2, 3)
b := collection.NewSet(2, 3, 4)
intersection := collection.Intersection(a, b) // {2, 3}
```

## API 参考

### Stack[T]

后进先出栈，基于动态切片实现。

| 方法 | 签名 | 说明 |
|------|------|------|
| `NewStack` | `NewStack[T]() *Stack[T]` | 创建栈，默认容量 16 |
| `NewStackWithCap` | `NewStackWithCap[T](cap int) *Stack[T]` | 创建指定容量的栈，cap≤0 时回退为 16 |
| `Push` | `Push(v T)` | 压入栈顶，O(1) 均摊 |
| `Pop` | `Pop() (T, bool)` | 弹出栈顶。空栈返回 `(zero, false)`，O(1) |
| `Peek` | `Peek() (T, bool)` | 查看栈顶但不移除，O(1) |
| `Len` / `Size` | `Len() int` / `Size() int` | 元素个数，O(1) |
| `IsEmpty` | `IsEmpty() bool` | 是否为空，O(1) |
| `Clear` | `Clear()` | 清空所有元素，O(1) |
| `Values` | `Values() []T` | 返回元素副本（栈顶在末尾），O(n) |
| `Clone` | `Clone() *Stack[T]` | 深拷贝，O(n) |

### Queue[T]

先进先出队列，基于动态切片实现。

| 方法 | 签名 | 说明 |
|------|------|------|
| `NewQueue` | `NewQueue[T]() *Queue[T]` | 创建队列，默认容量 16 |
| `NewQueueWithCap` | `NewQueueWithCap[T](cap int) *Queue[T]` | 创建指定容量的队列 |
| `Enqueue` | `Enqueue(v T)` | 入队（尾部），O(1) 均摊 |
| `Dequeue` | `Dequeue() (T, bool)` | 出队（头部）。空队列返回 `(zero, false)`，O(1) |
| `Peek` | `Peek() (T, bool)` | 查看队首但不移除，O(1) |
| `Len` / `Size` | `Len() int` / `Size() int` | 元素个数，O(1) |
| `IsEmpty` | `IsEmpty() bool` | 是否为空，O(1) |
| `Clear` | `Clear()` | 清空所有元素，O(1) |
| `Values` | `Values() []T` | 返回元素副本（队首在第一位），O(n) |
| `Clone` | `Clone() *Queue[T]` | 深拷贝，O(n) |

### PriorityQueue[T]

基于二叉堆实现的最小优先队列。`less(a, b) == true` 表示 a 优先级更高。

| 方法 | 签名 | 说明 |
|------|------|------|
| `NewPriorityQueue` | `NewPriorityQueue[T](less func(a, b T) bool) *PriorityQueue[T]` | 创建优先队列，默认容量 16 |
| `NewPriorityQueueWithCap` | `NewPriorityQueueWithCap[T](cap int, less func(a, b T) bool) *PriorityQueue[T]` | 创建指定容量的优先队列 |
| `Push` | `Push(v T)` | 插入元素，O(log n) |
| `Pop` | `Pop() (T, bool)` | 弹出最高优先级元素。空队列返回 `(zero, false)`，O(log n) |
| `Peek` | `Peek() (T, bool)` | 查看最高优先级元素但不移除，O(1) |
| `Len` / `Size` | `Len() int` / `Size() int` | 元素个数，O(1) |
| `IsEmpty` | `IsEmpty() bool` | 是否为空，O(1) |
| `Clear` | `Clear()` | 清空所有元素，O(1) |
| `Values` | `Values() []T` | 返回底层切片副本（堆序，不保证全局排序），O(n) |
| `Clone` | `Clone() *PriorityQueue[T]` | 深拷贝，O(n) |

### Set[T]

基于 `map[T]struct{}` 的集合，`T` 必须满足 `comparable`。

**构造与方法：**

| 方法 | 签名 | 说明 |
|------|------|------|
| `NewSet` | `NewSet[T](values ...T) Set[T]` | 创建集合并可选预填充元素 |
| `NewSetWithCap` | `NewSetWithCap[T](cap int) Set[T]` | 创建空集合并预分配容量 |
| `Add` | `Add(values ...T)` | 插入一个或多个元素，O(1) |
| `Remove` | `Remove(v T)` | 移除元素，不存在时不报错，O(1) |
| `Contains` | `Contains(v T) bool` | 是否包含元素，O(1) |
| `Len` / `Size` | `Len() int` / `Size() int` | 元素个数，O(1) |
| `IsEmpty` | `IsEmpty() bool` | 是否为空，O(1) |
| `Clear` | `Clear()` | 清空所有元素（使用 Go 1.21+ `clear` 内建函数），O(1) |
| `Values` | `Values() []T` | 返回所有元素（顺序非确定性），O(n) |

**集合运算（包级函数）：**

| 函数 | 签名 | 说明 |
|------|------|------|
| `Clone` | `Clone[T](s Set[T]) Set[T]` | 浅拷贝，O(n) |
| `Intersection` | `Intersection[T](a, b Set[T]) Set[T]` | 交集 a ∩ b。自动遍历较小集合，O(min(|a|,|b|)) |
| `Union` | `Union[T](a, b Set[T]) Set[T]` | 并集 a ∪ b，O(|a|+|b|) |
| `Difference` | `Difference[T](a, b Set[T]) Set[T]` | 差集 a − b，O(|a|) |
| `SymmetricDifference` | `SymmetricDifference[T](a, b Set[T]) Set[T]` | 对称差 a ∆ b，O(|a|+|b|) |
| `Equal` | `Equal[T](a, b Set[T]) bool` | 相等 a = b。O(n)，先比较长度再检查子集 |
| `IsSubset` | `IsSubset[T](a, b Set[T]) bool` | a ⊆ b。先比较长度快速排除，O(|a|) |
| `IsSuperset` | `IsSuperset[T](a, b Set[T]) bool` | a ⊇ b，O(|b|) |
| `IsDisjoint` | `IsDisjoint[T](a, b Set[T]) bool` | a ∩ b = ∅，自动遍历较小集合 |

**函数式操作（包级函数）：**

| 函数 | 签名 | 说明 |
|------|------|------|
| `Find` | `Find[T](s Set[T], predicate func(T) bool) (T, bool)` | 找出任意一个满足条件的元素 |
| `FindAll` | `FindAll[T](s Set[T], predicate func(T) bool) Set[T]` | 找出所有满足条件的元素 |
| `Partition` | `Partition[T](s Set[T], predicate func(T) bool) (matched, unmatched Set[T])` | 按条件拆分为两个集合 |
| `Map` | `Map[T, U comparable](s Set[T], fn func(T) U) Set[U]` | 变换每个元素，重复键自动合并 |
| `Reduce` | `Reduce[T, U any](s Set[T], reducer func(acc U, elem T) U, initial U) U` | 折叠为单个值（U 无需 comparable） |
| `ForEach` | `ForEach[T](s Set[T], fn func(T))` | 对每个元素执行副作用 |
| `Any` | `Any[T](s Set[T], predicate func(T) bool) bool` | 是否有任意元素满足条件 |
| `All` | `All[T](s Set[T], predicate func(T) bool) bool` | 是否所有元素满足条件。空集返回 `true` |
| `None` | `None[T](s Set[T], predicate func(T) bool) bool` | 是否没有元素满足条件。空集返回 `true` |

## 最佳实践

### 选择合适的构造函数

大多数场景使用无参构造函数即可，容量会自动按需增长：

```go
s := collection.NewStack[int]()        // 通用
q := collection.NewQueue[string]()     // 通用
```

当已知预期元素数量时，使用 `WithCap` 变体避免扩容：

```go
// 已知要处理 10000 个元素
s := collection.NewStackWithCap[int](10000)
```

### 善用 `(T, bool)` 返回值

Pop/Peek/Dequeue 使用 Go 惯用的 comma-ok 模式，无需提前检查 `IsEmpty`：

```go
if v, ok := s.Pop(); ok {
    // 处理 v
} else {
    // 栈空
}
```

### Set 相交运算的性能优化

`Intersection` 和 `IsDisjoint` 自动选择较小的集合遍历，将时间复杂度从 O(n²) 降至 O(min(|a|,|b|))。但请注意 — 如果已知集合大小差异很大，将小集合作为第一个参数传入差分/对称差运算也可以受益：

```go
large := collection.NewSet[int]() // 10000 个元素
small := collection.NewSet(1, 2)  // 2 个元素

// 自动优化：遍历 small
r := collection.Intersection(small, large)
```

### Set.Map 的合并语义

`Map` 在多个元素映射到同一键时会自动去重。这符合集合语义，但需注意可能减少元素数量：

```go
s := collection.NewSet(1, -1, 2)
r := collection.Map(s, func(x int) int { return x * x })
// r 包含 {1, 4}，而非 {1, 1, 4}
// 因为 1² = 1 且 (-1)² = 1
```

### 避免在热路径上使用 Values()

`Values()` 每次调用都会分配新切片。如需反复遍历，使用 `ForEach` 或在外部缓存：

```go
// 避免
for _, v := range s.Values() { ... } // 每次分配

// 推荐
collection.ForEach(s, func(v int) { ... }) // 零分配
```

### Queue 无内存泄漏

`Dequeue` 会在元素出队后将底层引用置零（`*new(T)`），确保 GC 可以回收。这对存放指针或大型结构体的长生命周期队列很重要。

## 与 v1 的区别（破坏性变更）

| 变更项 | v1 | v2 |
|--------|----|----|
| 构造函数 | `NewStack[T](initSize)` | `NewStack[T]()` / `NewStackWithCap[T](n)` |
| `AddAll` | `s.AddAll(vals...)` | 合并至 `s.Add(vals...)` |
| `Filter` (Set) | 存在 | 移除，使用 `FindAll`（功能完全相同） |
| `Reduce` (Set) | `U comparable` 约束 | 修正为 `U any` |
| `String()` | Stack/Queue/PQ 均有 | 移除 |
| `NewSetWithValues` | 存在 | 合并至 `NewSet(values...)` |
| Set 构造函数可变参 | 仅 `values ...T` | `NewSet[T](values ...T)` |
| `Clear` (Set) | for-range delete | Go 1.21+ `clear()` |
| `Equal` (Set) | IsSubset+IsSuperset | `len(a)==len(b) && IsSubset(a,b)` |
| `Intersection` | 始终遍历 a | 遍历较小集合 |
| `IsDisjoint` | 始终遍历 a | 遍历较小集合 |
| `IsEmpty` (Set) | 不存在 | 新增 |
| GC 安全 | Queue 不清理引用 | Dequeue/Pop 将引用置零 |

## 性能特征

| 操作 | 复杂度 | 分配 |
|------|--------|------|
| Stack Push/Pop | O(1) 均摊 | 仅在扩容时分配 |
| Queue Enqueue/Dequeue | O(1) 均摊 | 仅在扩容时分配 |
| PriorityQueue Push/Pop | O(log n) | 仅在扩容时分配 |
| Set Add/Remove/Contains | O(1) | map 自动管理 |
| Set Intersection | O(min(|a|,|b|)) | 新 set 分配 |
| Set Union | O(|a|+|b|) | 新 set 分配 |
| Set Difference | O(|a|) | 新 set 分配 |
| 所有 Clone/Values | O(n) | 新切片/新 set 分配 |
