// Package collection 提供 generic、对分配感知的数据结构：
// [Stack]、[Queue]、[Heap] 和 [Set]。
//
// V3 是 collection 模块的从头重写,与 v2 不兼容。设计遵循三个目标 —— 合理、高效、易用 —— 通过以下方式实现：
//
//   - 更好的数据结构：[Queue] 是 ring buffer(修复了 v2 的 slice[1:] 内存
//     泄漏);[Heap] 为 [cmp.Ordered] 类型暴露 [NewMinHeap]/[NewMaxHeap];
//     [Set] 将其 backing map 隐藏在 struct 之后,使布局可演进而不破坏调用方。
//   - 更好的惯用法：每个容器都暴露 [Stack.All]/[Queue.All]/
//     [Heap.All]/[Set.All] 返回 [iter.Seq],因此 `for v := range s.All()` 可用,
//     且容器可与标准库的 [slices]/[maps] 包组合使用。[Stack.Pop]/[Queue.Dequeue]/[Heap.Peek]
//     返回 (T, bool);零值会被清零,以便 GC 回收引用。
//   - 更易使用：[Set] 支持方法链式代数
//     (`a.Union(b).Intersect(c)`) 并搭配自由函数别名,[Heap] 有 cmp.Ordered 构造器,因此调用方几乎不必手写 `less` 函数。
//
// 模块边界:
//
//	保序去重 / 成员判定 → functional/v3.OrderedSet。
//	无序集合代数(union/intersect/diff)→ 本包的 Set。
//	slice 级 FP(Map/Filter/Reduce on []T)→ functional/v3。
//
// 所有操作都是指针接收者方法;若需共享状态,请在修改前用 Clone 复制容器。
package collection
