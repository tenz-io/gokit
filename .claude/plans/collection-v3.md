# collection/v3 完全重写计划

## 目标
对 `collection/` 模块做**完全重写**(v3),不兼容 v2。只动本模块。对标仓库 v3 规范(annotation/v3、async/v3、logger/v3 的布局与文档风格):分文件、中文 README、`example/` 子模块、`Makefile`、注册进 `go.work`。

四条主线:**更好的表(数据结构)**、**更好的写法(idiom)**、**更容易的使用方式**、**更好的性能**。

## V2 现状与缺陷(重写的依据)

- `Queue[T]`:`[]T` + `Dequeue` 用 `q.data = q.data[1:]`。每次出队让底层数组头指针前移,**永远不回收已出队槽位**,内存只增不减 → 真实内存泄漏 + 性能随队列老化退化。`TestQueue_MemoryLeak` 自己都注明"无法验证"。
- `Set[T]`:`type Set[T comparable] map[T]struct{}` 别名,实现完全裸露;集合代数全为自由函数,**无法链式**;`Size()` 与 `Len()` 重复。
- `PriorityQueue[T]`:手写二叉堆,`less func` 必填,无 `cmp.Ordered` 便捷构造;`*new(T)` 置零写法偏旧(应 `var zero T`)。
- 三类容器 API 表面不统一(命名、`Clone` 返回类型、`Values` 语义)。
- 全部缺 `iter.Seq[T]` 集成,无法 range-over-func,无法直接喂给 `slices`/`maps` 标准库。

## 模块边界(与相邻模块划清职责,避免重叠)

- `functional/v3` 已有 `OrderedSet`(保序、去重、成员判定)与切片上的 `Partition/Find/Any/All`。collection/v3 的 `Set` 聚焦**无序集合代数**(交/并/差/对称差/子集超集),不保序、不与 `OrderedSet` 竞争。README 显式指引:保序去重用 `functional/v3.OrderedSet`,集合代数用本模块。
- 不引入 `iter` 标准库之外的依赖,`go.mod` 仅 `testify`(与 v2 一致)。

## 文件布局

```
collection/v3/
  doc.go              # 包文档:设计原则、与 v2/functional 的边界
  stack.go            # Stack[T]
  queue.go            # Queue[T] 环形缓冲
  priority_queue.go   # Heap[T] + NewMinHeap/NewMaxHeap 便捷构造
  set.go              # Set[T] struct 封装 + 方法链 + 自由函数
  iter.go             # 各容器的 All() iter.Seq[T]
  collection_test.go  # 全量测试 + bench
  example/
    go.mod            # module collection-v3-example
    main.go
  go.mod              # module github.com/tenz-io/gokit/collection/v3
  Makefile            # 沿用 v3 模板
  README.md           # 中文
```

`go.work` 在 `use (...)` 块加 `./collection/v3` 与 `./collection/v3/example`。

## 各结构设计

### Stack[T](stack.go)— 切片 LIFO,小幅现代化
- `type Stack[T any] struct{ data []T }`,`*Stack[T]` 指针接收器(与 v2 一致,值拷贝会 alias 底层数组)。
- `New[T]()`(默认 cap 16)、`NewWithCap[T](n int)`(n<=0 回落 16)。
- `Push/Pop/Peek`,`Pop/Peek` 返回 `(T, bool)`,空返回 `(zero,false)`。
- 置零改 `var zero T; s.data[n]=zero`(替掉 `*new(T)`)。
- `Len/IsEmpty/Clear/Values/Clone`,`All() iter.Seq[T]`(栈底→栈顶)。
- **去掉 `Size()`**(冗余别名,`Len` 是 Go idiom)。

### Queue[T](queue.go)— **环形缓冲,修泄漏**
- `type Queue[T any] struct{ buf []T; head, tail, count int }`,cap=0 时按需扩容。
- `Enqueue`:满则 `grow`(新 buf 2x,把 `[head..tail)` 线性拷回,重置 head=0/tail=count);写入 `buf[tail]`,`tail=(tail+1)%cap`。
- `Dequeue`:空返回 `(zero,false)`;否则 `v=buf[head]`,`buf[head]=zero`(置零助 GC),`head=(head+1)%cap`,`count--`。
- `Peek` 看队首 `buf[head]`。
- 均摊 O(1) Enqueue/Dequeue,**内存稳定**(扩容时旧 buf 整体释放)。
- `Len/IsEmpty/Clear/Values`(队首→队尾有序)/`Clone`/`All() iter.Seq[T]`。
- `Clear` 不只清元素,把 `head=tail=count=0` 复位(但保留 buf 容量,可复用)。
- **去掉 `Size()`**。

### Heap[T](priority_queue.go)— 二叉堆 + 便捷构造
- 重命名 `PriorityQueue[T]` → **`Heap[T]`**(更短、更通用;`PriorityQueue` 保留为 type alias 以兼容直觉?——不兼容 v2 是明确要求,**不加 alias**,直接用 `Heap`,README 说明)。
- `type Heap[T any] struct{ data []T; less func(a,b T) bool }`。
- `NewHeap[T](less)`(cap 16)、`NewHeapWithCap[T](n, less)`。
- **便捷构造**(用户已选):`NewMinHeap[T cmp.Ordered]()` = `NewHeap(cmp.Less[T] # cmp.Compare≤0)`;`NewMaxHeap[T cmp.Ordered]()` 反向。实现用 `func(a,b T) bool { return cmp.Compare(a,b) < 0 }`。
- `Push`(append+bubbleUp)、`Pop`(swap 根尾、bubbleDown、置零)、`Peek`。
- `Len/IsEmpty/Clear/Values/Clone`,`All() iter.Seq[T]`(无序,文档注明)。
- **去掉 `Size()`**。

### Set[T](set.go)— **struct 封装 + 方法链 + 自由函数**
- `type Set[T comparable] struct{ m map[T]struct{} }`(隐藏 map,留演进空间)。
- `NewSet[T](vs...)`、`NewSetWithCap[T](n)`。
- 方法:`Add(...T)`、`Remove(T)`、`Contains(T) bool`、`Len()`、`IsEmpty()`、`Clear()`、`Values() []T`、`Clone() Set[T]`、`All() iter.Seq[T]`。
- **方法链代数**(返回新 `Set`,不改接收方):`Union(other) Set[T]`、`Intersect(other)`、`Subtract(other)`、`SymmetricDifference(other)`。
- 关系判断(方法,返 bool):`IsSubset(other)`、`IsSuperset(other)`、`IsDisjoint(other)`、`Equal(other)`。
- **自由函数别名**(等价,供函数式风格 / 不可变场景):`Union(a,b)`、`Intersect(a,b)`、`Difference(a,b)`(注意:自由函数叫 `Difference` 对齐 v2 名称;方法叫 `Subtract` 链式更顺)、`SymmetricDifference(a,b)`、`IsSubset/IsSuperset/IsDisjoint/Equal`。
- FP(仅自由函数,避免堆砌方法):`Find/FindAll/Partition/Map/Reduce/ForEach/Any/All/None`——保留 v2 的这套,但基于 struct Set 适配,`Map` 支持类型变换 `Set[U]`。
- 性能:`Intersect` 迭代较小集合;`Union` 预分配 `len(a)+len(b)`;`IsSubset` 先比长度短路。

### iter.go — All() iter.Seq[T]
- 每个容器一个 `All() iter.Seq[T]`:`func(yield func(T) bool)`。
- Stack:栈底→栈顶;Queue:队首→队尾;Heap:无序(底层切片序);Set:无序(map 迭代序)。
- 用法:`for v := range s.All() { ... }`,可直接 `slices.Sorted(s.All())`、`slices.Collect(s.All())`。

## 统一的容器表面(写法一致性)
所有容器:`Len()/IsEmpty()/Clear()/Values() []T/Clone()/All() iter.Seq[T]`。**无 `Size()`**。Pop/Peek/Dequeue 统一 `(T, bool)` 语义。

## 测试与基准(collection_test.go)
- 移植 v2 全部用例并适配新 API(改名、去 `Size`、Queue 链式)。
- **新增回归**:`TestQueue_RingBuffer_NoMemoryLeak`——循环 Enqueue/Dequeue N 次后 `Len()==0` 且底层 buf 容量稳定(不再无限增长)。这是 v2 的真正修复点,必须有测试守住。
- **新增**:`TestSet_Chained` 验证 `a.Union(b).Intersect(c)` 链式。
- **新增**:`TestIter_All`——range over `s.All()` 收集、`slices.Sorted(minHeap.All())`。
- **新增**:`TestMinHeap/TestMaxHeap`——cmp.Ordered 便捷构造。
- bench:`Stack_PushPop`、`Queue_EnqueueDequeue`(对比 v2 应明显胜出,尤其长寿命队列)、`Heap_PushPop`、`Set_Add`、`Set_Contains`、`Set_Intersect`。

## example/(main.go)
- Stack:括号匹配小演示。
- Queue:任务缓冲,演示环形缓冲长期使用不泄漏。
- Heap:`NewMinHeap` 做合并 K 个有序流 / Top-K。
- Set:链式集合代数 + range `All()`。

## 执行步骤(逐步落地,每步可独立验证)
1. 建目录 + `go.mod`(module collection/v3,go 1.24,require testify)+ `Makefile`(抄模板)+ 注册 `go.work`。
2. `doc.go` + `stack.go`(最简单,先打通 toolchain:`go build ./...` 过)。
3. `queue.go`(环形缓冲,重点)+ 回归测试。
4. `priority_queue.go`(Heap + Min/Max 便捷)+ 测试。
5. `set.go`(struct + 方法链 + 自由函数 + FP)+ 测试。
6. `iter.go`(`All()`)+ iter 测试。
7. `collection_test.go` 补全 + bench。
8. `example/` 子模块。
9. `README.md`(中文,对齐 annotation/v3 风格:模块介绍/能力清单表/快速开始/API 速查表)。
10. `make test vet fmt` 全绿;`make run-example` 跑通。

## 不做的事(边界)
- 不动 v2(collection/v2 原地保留)。
- 不迁移 functional/v3 的 `OrderedSet`(各管各的)。
- 不加 `FromSlice` heapify(用户未选;保持构造面精简)。
- 不引入第三方依赖。
- 即便本模块引起其他模块编译错误也不处理(明确指示)。
