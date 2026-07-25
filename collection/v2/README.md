# collection

泛型数据结构：Stack、Queue、PriorityQueue、Set，附带集合运算。

## 功能特性

- `Stack[T]`：LIFO 栈，支持 `Push`/`Pop`/`Peek`，可指定初始容量（`NewStackWithCap`）
- `Queue[T]`：FIFO 队列，支持 `Enqueue`/`Dequeue`/`Peek`
- `PriorityQueue[T]`：基于二叉堆实现，通过自定义 `less` 函数决定优先级，`Push`/`Pop` 均为 O(log n)
- `Set[T]`：基于 map 的集合，支持 `Add`/`Remove`/`Contains`，以及 `Intersection`/`Union`/`Difference`/`SymmetricDifference` 等集合运算
- 集合关系判断：`IsSubset`/`IsSuperset`/`IsDisjoint`/`Equal`
- 集合上的函数式操作：`Find`/`FindAll`/`Partition`/`Map`/`Reduce`/`ForEach`/`Any`/`All`/`None`
- 所有结构均提供 `Len`/`Size`/`IsEmpty`/`Clear`/`Values`/`Clone`，行为一致、可预期

## 快速开始

```go
import "github.com/tenz-io/gokit/collection/v2"

func main() {
	// Stack
	s := collection.NewStack[int]()
	s.Push(1)
	s.Push(2)
	top, _ := s.Pop() // 2

	// Queue
	q := collection.NewQueue[string]()
	q.Enqueue("a")
	q.Enqueue("b")
	front, _ := q.Dequeue() // "a"

	// PriorityQueue（小顶堆，数值越小优先级越高）
	pq := collection.NewPriorityQueue(func(a, b int) bool { return a < b })
	pq.Push(3)
	pq.Push(1)
	pq.Push(2)
	min, _ := pq.Pop() // 1

	// Set 与集合运算
	a := collection.NewSet(1, 2, 3)
	b := collection.NewSet(2, 3, 4)
	inter := collection.Intersection(a, b) // {2, 3}
	union := collection.Union(a, b)        // {1, 2, 3, 4}

	_ = top
	_ = front
	_ = min
	_ = inter
	_ = union
}
```

## API 速查

| 符号 | 说明 |
|---|---|
| `Stack[T]` / `NewStack[T]` / `NewStackWithCap[T]` | LIFO 栈及构造函数 |
| `(*Stack[T]).Push/Pop/Peek` | 入栈、出栈、查看栈顶 |
| `Queue[T]` / `NewQueue[T]` / `NewQueueWithCap[T]` | FIFO 队列及构造函数 |
| `(*Queue[T]).Enqueue/Dequeue/Peek` | 入队、出队、查看队首 |
| `PriorityQueue[T]` / `NewPriorityQueue` / `NewPriorityQueueWithCap` | 基于二叉堆的优先队列及构造函数 |
| `(*PriorityQueue[T]).Push/Pop/Peek` | 插入、弹出最高优先级元素、查看堆顶 |
| `Set[T]` / `NewSet[T]` / `NewSetWithCap[T]` | 基于 map 的集合及构造函数 |
| `(Set[T]).Add/Remove/Contains` | 添加、删除、判断元素是否存在 |
| `Clone[T]` | 复制一个集合 |
| `Intersection/Union/Difference/SymmetricDifference` | 交集、并集、差集、对称差集 |
| `IsSubset/IsSuperset/IsDisjoint/Equal` | 子集、超集、不相交、相等判断 |
| `Find/FindAll/Partition` | 按条件查找单个/全部元素、按条件分组 |
| `Map[T,U]/Reduce[T,U]/ForEach` | 映射转换、归约、遍历 |
| `Any/All/None` | 存在性、全量、无匹配的谓词判断 |
| `(*Stack/Queue/PriorityQueue[T]).Len/Size/IsEmpty/Clear/Values/Clone` | 三类容器通用的长度、清空、导出、克隆方法 |
| `(Set[T]).Len/Size/IsEmpty/Clear/Values` | Set 上的等价方法 |

引用路径：`github.com/tenz-io/gokit/collection/v2`
