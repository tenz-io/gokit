# collection

泛型数据结构:`Stack`、`Queue`(环形缓冲)、`Heap`(二叉堆)、`Set`,附带链式集合代数与 `iter.Seq` 集成。V3 是对 collection 模块的**完全重写**,不兼容 V2。

```go
import "github.com/tenz-io/gokit/collection/v3"
```

## 模块介绍

collection 解决四类日常数据结构需求,并以"更好的表 / 更好的写法 / 更容易用 / 更好的性能"为重写主线:

- **Stack[T]**:切片实现 LIFO 栈,`Push`/`Pop`/`Peek`,出栈槽位置零助 GC。
- **Queue[T]**:**环形缓冲**实现 FIFO 队列,均摊 O(1),修复了 V2 `slice[1:]` 的内存泄漏(已出队槽位永不回收、底层数组只增不减)。
- **Heap[T]**:二叉堆优先队列,`Push`/`Pop` 均 O(log n);提供 `NewMinHeap`/`NewMaxHeap` 对 `cmp.Ordered` 的便捷构造,免去手写 `less`。
- **Set[T]**:`struct` 封装 map(隐藏实现),支持**方法链**集合代数 `a.Union(b).Intersect(c).Subtract(d)`,同时提供自由函数别名与函数式操作。

V3 相对 V2 的核心变化:

- **Queue 改环形缓冲**:彻底修掉 V2 `q.data = q.data[1:]` 的内存泄漏与队列老化性能退化。回归测试 `TestQueue_RingBuffer_NoMemoryLeak` 守住。
- **Heap 便捷构造**:`NewMinHeap[T cmp.Ordered]()` / `NewMaxHeap[T cmp.Ordered]()`,自定义排序仍走 `NewHeap(less)`。
- **Set 方法链 + struct 封装**:代数运算返回新 `Set` 不改接收方,`a.Union(b).Subtract(c)` 一气呵成;隐藏 map 留演进空间。
- **`iter.Seq` 集成**:四类容器都暴露 `All() iter.Seq[T]`,直接 `for v := range s.All()`,可与标准库 `slices`/`maps` 组合(`slices.Collect`、`slices.Sorted`)。
- **统一容器表面**:全部容器统一 `Len/IsEmpty/Clear/Values/Clone/All`,出/入操作统一 `(T, bool)` 语义;去掉冗余的 `Size()`;置零用 `var zero T`。

模块边界(与相邻库划清职责,避免重叠):

- **保序去重 / 成员判定** → `functional/v3.OrderedSet`(按插入序遍历、去重)。
- **无序集合代数**(交/并/差/对称差/子集超集)→ 本模块 `Set`。
- **切片级函数式变换**(`Map`/`Filter`/`Reduce` on `[]T`)→ `functional/v3`。

## 能力清单

| 能力 | 含义 |
|---|---|
| 后进先出存取 | `Stack[T]` 的 `Push`/`Pop`/`Peek`,用于回溯、括号匹配、DFS |
| 先进先出存取(环形缓冲) | `Queue[T]` 的 `Enqueue`/`Dequeue`/`Peek`,均摊 O(1),内存稳定不泄漏,用于任务缓冲、BFS、消息顺序处理 |
| 按优先级出队 | `Heap[T]` 基于二叉堆,`less` 决定优先级,`Push`/`Pop` 均 O(log n);`NewMinHeap`/`NewMaxHeap` 免手写 `less` |
| 集合成员维护 | `Set[T]` 的 `Add`/`Remove`/`Contains`/`Clear`,用于去重、快速存在性判断 |
| 链式集合代数 | `Set.Union`/`Intersect`/`Subtract`/`SymmetricDifference` 返回新 `Set`,支持 `a.Union(b).Subtract(c)` |
| 集合关系判断 | `Set.IsSubset`/`IsSuperset`/`IsDisjoint`/`Equal` 判断包含、超集、互斥、相等 |
| 集合上的函数式操作 | `Find`/`FindAll`/`Partition`/`Map`/`Reduce`/`ForEach`/`Any`/`All`/`None` |
| range-over-func | 四类容器的 `All()` 返回 `iter.Seq[T]`,直接 `for v := range s.All()`,支持提前 break,可与 `slices`/`maps` 组合 |
| 预分配容量构造 | `NewStackWithCap`/`NewQueueWithCap`/`NewHeapWithCap`/`NewSetWithCap` 按已知元素数预分配 |

## 快速开始

```go
package main

import (
	"fmt"
	"slices"

	"github.com/tenz-io/gokit/collection/v3"
)

func main() {
	// Stack:LIFO
	s := collection.NewStack[int]()
	s.Push(1)
	s.Push(2)
	top, _ := s.Pop() // 2

	// Queue:环形缓冲,长期 enqueue/dequeue 不泄漏
	q := collection.NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	front, _ := q.Dequeue() // 1

	// Heap:小顶堆,数值越小越先出
	h := collection.NewMinHeap[int]()
	h.Push(3)
	h.Push(1)
	h.Push(2)
	min, _ := h.Pop() // 1

	// Set:链式集合代数
	a := collection.NewSet(1, 2, 3)
	b := collection.NewSet(2, 3, 4)
	r := a.Union(b).Subtract(collection.NewSet(4)) // {1,2,3}

	// iter.Seq:range + 标准库组合
	doubled := slices.Collect(func(yield func(int) bool) {
		for v := range a.All() {
			if !yield(v * 2) {
				return
			}
		}
	})
	_ = top
	_ = front
	_ = min
	_ = r
	_ = doubled
	fmt.Println("ok")
}
```

## API 速查

| 符号 | 说明 |
|---|---|
| `Stack[T]` / `NewStack[T]` / `NewStackWithCap[T]` | LIFO 栈及构造函数 |
| `(*Stack[T]).Push/Pop/Peek` | 入栈、出栈、查看栈顶(返回 `(T, bool)`) |
| `Queue[T]` / `NewQueue[T]` / `NewQueueWithCap[T]` | 环形缓冲 FIFO 队列及构造函数 |
| `(*Queue[T]).Enqueue/Dequeue/Peek` | 入队、出队、查看队首(返回 `(T, bool)`) |
| `Heap[T]` / `NewHeap[T]` / `NewHeapWithCap[T]` | 二叉堆及自定义 `less` 构造 |
| `NewMinHeap[T cmp.Ordered]` / `NewMaxHeap[T cmp.Ordered]` | 有序类型的便捷构造(小顶/大顶) |
| `(*Heap[T]).Push/Pop/Peek` | 插入、弹出最高优先级、查看堆顶(返回 `(T, bool)`) |
| `Set[T]` / `NewSet[T]` / `NewSetWithCap[T]` | struct 封装的 map 集合及构造函数 |
| `(Set[T]).Add/Remove/Contains` | 添加(`Add` 返回 `Set[T]` 可链)、删除、判断存在 |
| `(Set[T]).Union/Intersect/Subtract/SymmetricDifference` | 链式代数,返回新 `Set`,不改接收方 |
| `(Set[T]).IsSubset/IsSuperset/IsDisjoint/Equal` | 子集、超集、不相交、相等 |
| `UnionOf/IntersectOf/Difference/SymmetricDifference` | 链式代数的自由函数别名(与方法等价) |
| `IsSubset/IsSuperset/IsDisjoint/Equal` | 关系判断的自由函数别名 |
| `Clone[T]` | 复制集合(自由函数,等价 `Set.Clone`) |
| `Find/FindAll/Partition/Map/Reduce/ForEach/Any/All/None` | 集合上的函数式操作 |
| `(*Stack/Queue/Heap[T]).All()` / `(Set[T]).All()` | 返回 `iter.Seq[T]`,支持 `for v := range s.All()` 与 `slices`/`maps` 组合 |
| `Len/IsEmpty/Clear/Values/Clone` | 四类容器统一的长度、判空、清空、导出副本、克隆 |

## 队列环形缓冲:为什么重写

V2 的 `Queue` 用 `[]T`,`Dequeue` 执行 `q.data = q.data[1:]`。这会把切片窗口整体后移,**已出队的槽位永远不被回收**,底层数组只增不减 → 真实内存泄漏,且队列寿命越长性能越差(V2 的 `TestQueue_MemoryLeak` 自己都注明"无法验证")。

V3 改用环形缓冲(`buf` + `head`/`tail`/`count`):出队时置零并推进 `head`,缓冲填满时整体倍增扩容并线性化回前端。均摊 O(1) 的同时,长期稳态使用的内存**有界**(见 `TestQueue_RingBuffer_NoMemoryLeak` 与 `BenchmarkQueue_LongLived`,稳态下 `0 allocs/op`)。

引用路径:`github.com/tenz-io/gokit/collection/v3`
