// Package fn 提供面向 Go slice 的通用函数式编程工具。
//
// V3 是 functional 工具包的从头重写版本,与 v2 不兼容。设计遵循三个目标——
// 合理、高效、易用——通过双轨 API 实现:
//
//   - Standalone 函数:v2 接口的严格超集(Map、Filter、Reduce、TopK……),
//     附带 index-aware 变体(MapIdx、FindIndex)、in-place/zero-allocation 变体
//     (MapInPlace、FilterInPlace),以及基于 key extractor 的聚合
//     (带 Key extractor 的 TopK/BottomK)。
//   - Fluent Chain:ChainOf(s).Map(...).Filter(...).TopK(k, key).Collect()
//     每一步都物化出一个 slice——通常比 Go 中的 closure chain 更快,
//     也更容易调试。
//   - Lazy Seq:callback 风格的迭代器(在理念上与未来的 iter.Seq 兼容),
//     用于在大型或生成的输入上做 short-circuit 读取(Any/All/Find),
//     无需分配物化的 slice。
//
// 排序模型:
//
//	TopK 返回 k 个最大的元素,按降序排列。
//	BottomK 返回 k 个最小的元素,按升序排列。
//
// Key extractor 与 comparator:
//
//	Key[T, K cmp.Ordered] 提取一个有序 key——常见场景(按 score 取 top-k、
//	按 id 取 min/max)。用于 TopK/BottomK/MinByKey/MaxByKey。
//	By[T] 是 cmp 风格的 comparator func(a, b T) int,供 *By 变体使用,
//	当你需要多字段或非 key 排序时使用。当单个有序字段即可表达排序时优先用 Key。
//
// 除非名字以 InPlace 结尾,所有操作都是 pure 的。in-place 变体会复用输入
// slice 的 backing array,并返回其(可能缩短的)视图;它们是 zero-allocation
// 的热路径。
package fn

import "cmp"

// Key 是从类型 T 的值中提取有序 key 的函数。
//
// 它是 key 聚合函数(TopK/BottomK/MinByKey/MaxByKey)的主要输入。当排序
// 来源于单个 cmp.Ordered 字段时,优先用它而非 By。
//
// 示例:
//
//	fn.TopK(users, 10, fn.Key[User, int](func(u User) int { return u.Score }))
type Key[T any, K cmp.Ordered] func(T) K

// By 是 cmp 风格的 comparator:当 a < b 时返回负数,相等时返回零,
// 当 a > b 时返回正数——与 [cmp.Compare] 和 [slices.SortFunc] 约定一致。
//
// 当排序无法用单个有序 key 表达时(多字段、自定义排名),用于 *By 变体
// (TopKBy/BottomKBy/MinBy/MaxBy)。
type By[T any] func(a, b T) int

// Pair 把两个(可能不同类型的)值组合在一起。
//
// 它由 Zip 返回,也是 zipped slice 的元素类型。
type Pair[A, B any] struct {
	A A
	B B
}
