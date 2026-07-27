// Package httpext 为标准库的 *http.Client 提供一个可组合的 transport 层
// interceptor chain:header injection、per-request metrics,以及
// request/response traffic 日志。
//
// v3 是一次干净的重写,不带任何向后兼容 shim,构建于
// logger/v3、monitor/v3 与 tracer/v3 之上。它与未改动的 httpext/v2 并存;
// 使用方不会被自动迁移。
//
// httpext 仅提供 Interceptor —— 不提供 Client wrapper。通过
// Interceptor.Apply 将 chain 接入你自己的 *http.Client,然后照常使用
// 标准库 (http.Client.Do、Get、Post)。激活的层会为每个 request 透明地运行。
//
// chain 通过 Interceptor.Apply 一次性接入 *http.Client,此后每个出站
// request 都按 newTransporters 顺序流经激活的层
// (injectHeader → metrics → traffic)。禁用的层会被剔除,因此 chain 仅包含
// 已启用的层。
//
// 快速开始:
//
//	interceptor := httpext.NewInterceptorWithOpts(
//	    httpext.WithEnableTraffic(true),
//	    httpext.WithEnableMetrics(true),
//	    httpext.WithHeaders(map[string]string{
//	        httpext.HeaderNameAuthorization: "Bearer token",
//	    }),
//	)
//	httpCli := &http.Client{}
//	interceptor.Apply(httpCli)
//
//	// 照常使用标准库 client;chain 会为每次调用运行。
//	resp, err := httpCli.Get("https://example.com/items")
//
// 行为说明 (与 v2 不同):
//   - slow-log transport 与 Config.SlowLogFloor 已移除。慢请求由 monitor/v3
//     的 latency histogram 暴露 (基于其阈值告警)。
//   - 不提供 Client/SimpleRequest/RequestOption 表面 —— 直接使用
//     *http.Client。v2 的 client.go (JSON/DoSimple/Get/Post/...) 已移除;
//     标准库已经覆盖这些动词。
//   - traffic 日志仅记录 cmd/cost/code/method/url/query 以及 request/
//     response header —— 它不会读取或记录 request/response body,因此开销
//     很低且绝不干扰 body 流。v2 会捕获并解码 body (JSON/form 解析、文本
//     截断);该 capture.go 已移除。traffic 使用 logger/v3 的
//     StartTraffic/End span API,而非 v2 的 ReqEntity/RespEntity 表面。
package httpext
