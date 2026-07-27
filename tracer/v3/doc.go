// Package tracer 提供基于 context 的 request ID 传播与流量模式
// (debug/stress/shadow) flag 管理。
//
// v3 是一次干净重写,不带任何向后兼容垫片。它围绕数据驱动的 flag 注册表
// 设计,因此解析 flag 字符串("debug|shadow")、把 flag 渲染回字符串形式、
// 遍历已知 flag 都从同一张表读取 —— 调用方不再需要为每个 transport 手写
// switch。
//
// 快速上手:
//
//	ctx := tracer.WithRequestID(context.Background(), "req-123")
//	ctx = tracer.WithFlag(ctx, tracer.FlagDebug|tracer.FlagShadow)
//
//	if tracer.FromContext(ctx).IsDebug() {
//	    // 详细路径
//	}
//
//	// 把 header 值解析成 flag 集合,再存入 context。
//	f := tracer.ParseFlag("debug|shadow")
//	ctx = tracer.WithFlag(ctx, f)
//
//	fmt.Println(tracer.FromContext(ctx)) // "debug|shadow"
//
// ID 约定:
//   - EnsureRequestID: inbound 边界原语 —— 保证 ctx 携带一个 id(缺失时
//     生成并存入),使整个请求的日志、响应 header 与下游调用都看到同一
//     id。middleware 中优先使用它。
//   - WithRequestID / RequestIDFromCtx:显式设置,或读取已存 id 并在缺失
//     时回退生成。注意 RequestIDFromCtx 不会把生成的 id 写回,因此对
//     无 id 的 ctx 反复读取会得到不同 id;用 EnsureRequestID 来固定一个。
//   - RequestIDFromCtxOr:不自动生成地读取;缺失时返回 "",适用于
//     "是否已存在?" 的判断。
//
// 行为说明(与 v2 不同):
//   - Flag.Is(FlagNone) 返回 false(v2 返回 true)。FlagNone 不是真实
//     flag;改用 f == FlagNone 来测试它。
//   - request-id 符号拼写为 RequestID(大写 ID),以契合 Go 命名约定与
//     logger/v3。
//   - Flag 为 uint8(8 个可用 flag 位,无符号位溢出),context key 为带类型
//     的空 struct(零碰撞,无需字符串比较)。
package tracer
