// Package httpext provides a composable transport-layer interceptor chain for
// the standard library's *http.Client: header injection, per-request metrics,
// and request/response traffic logging.
//
// v3 is a clean rewrite with no backwards-compatibility shims, built on
// logger/v3, monitor/v3, and tracer/v3. It sits alongside the untouched
// httpext/v2; consumers are not migrated automatically.
//
// httpext only provides the Interceptor — there is no Client wrapper. Wire
// the chain onto your own *http.Client via Interceptor.Apply, then keep using
// the standard library (http.Client.Do, Get, Post) as normal. The active
// layers run for every request transparently.
//
// The chain is wired once onto an *http.Client via Interceptor.Apply, after
// which every outbound request flows through the active layers in
// newTransporters order (injectHeader → metrics → traffic). Disabled layers
// are dropped, so the chain contains only the enabled ones.
//
// Quick start:
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
//	// Use the standard library client as normal; the chain runs for every call.
//	resp, err := httpCli.Get("https://example.com/items")
//
// Behavior notes (differ from v2):
//   - The slow-log transport and Config.SlowLogFloor are removed. Slow requests
//     are surfaced by monitor/v3's latency histogram (alert on its threshold).
//   - There is no Client/SimpleRequest/RequestOption surface — use *http.Client
//     directly. v2's client.go (JSON/DoSimple/Get/Post/...) is removed; the
//     standard library already covers those verbs.
//   - Traffic logging records cmd/cost/code/method/url/query and request/
//     response headers only — it does NOT read or record the request/response
//     bodies, so it is cheap and never disturbs the body streams. v2 captured
//     and decoded bodies (JSON/form parse, text truncation); that capture.go
//     is removed. Traffic uses logger/v3's StartTraffic/End span API instead of
//     v2's ReqEntity/RespEntity surface.
package httpext
