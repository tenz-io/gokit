# httpext

HTTP 出站客户端的传输层拦截器链：请求头注入、指标上报、流量采集。v3 是一次无兼容包袱的干净重写，基于 `logger/v3`、`monitor/v3`、`tracer/v3`，与保留不动的 `httpext/v2` 并存。

httpext **只提供 `Interceptor`**——把它一次性 `Apply` 到标准库 `*http.Client` 上，之后照常用 `cli.Get`/`cli.Post`/`cli.Do`，所有出站请求自动经过注入头、指标、流量采集三层。不提供 client 封装：标准库的 `*http.Client` 已经够用，再包一层只是多一层间接和约定。

```go
import "github.com/tenz-io/gokit/httpext/v3"
```

## 模块介绍

httpext 解决出站 HTTP 调用的三类横切需求：

- **统一注入固定请求头**（`WithHeaders`）：鉴权 Token 等 Header 在每次出站请求前自动 `Set`，避免调用方重复设置。
- **上报请求指标**（`WithEnableMetrics`）：用 `monitor/v3` 记录每次请求的 URL 与响应状态码，用于监控接口耗时与成功率。
- **记录请求/响应流量日志**（`WithEnableTraffic`）：用 `logger/v3` 的 traffic logger 记录每次请求的 cmd（URL path）、耗时、状态码、method、url、query 及请求/响应 Header；链路处于 `tracer.FlagDebug` 态时自动触发。流量日志**不读、不记请求/响应 body**，只记 Header 与元数据，开销小且不打扰 body 流。

核心能力：

- `Interceptor`：将拦截器链一次性 `Apply` 到 `*http.Client`，无需手动包装 `Transport`。
- 拦截器链按序叠加（注入 Header → 指标 → 流量），关闭的层自动跳过，链里只剩启用的层。

## 能力清单

| 能力 | 含义 |
|---|---|
| 一键装配拦截器链 | `Interceptor.Apply` 按顺序叠加注入 Header、指标、流量三类 `http.RoundTripper`，直接作用于 `*http.Client`，无需手动包装 `Transport` |
| 统一注入固定请求头 | `WithHeaders` 配置的 Header（如 `Authorization`）在每次出站请求前自动 `Set` |
| 上报请求指标 | `WithEnableMetrics` 开启后，`metricsTransport` 用 `monitor.Begin` 记录每次请求的 URL 与响应状态码 |
| 记录请求/响应流量日志 | `WithEnableTraffic` 开启（或链路处于 debug 态时自动触发）后，`trafficTransport` 经 `logger.StartTraffic` 记录每次请求的 cmd/cost/code/method/url/query 及请求/响应 Header，不读不记 body |
| 透明叠加标准库 | Apply 之后照常用 `*http.Client` 的 `Get`/`Post`/`Do`，拦截器链对每次出站请求透明生效 |

## 快速开始

```go
package main

import (
	"net/http"

	"github.com/tenz-io/gokit/httpext/v3"
)

func main() {
	// 1. 构建拦截器链并应用到标准库 http.Client
	interceptor := httpext.NewInterceptorWithOpts(
		httpext.WithEnableTraffic(true),
		httpext.WithEnableMetrics(true),
		httpext.WithHeaders(map[string]string{
			httpext.HeaderNameAuthorization: "Bearer token",
		}),
	)
	httpCli := &http.Client{}
	interceptor.Apply(httpCli)

	// 2. 照常用标准库 client 发请求，拦截器链对每次调用透明生效
	resp, err := httpCli.Get("https://example.com/api")
	_ = resp
	_ = err
}
```

## API 速查

| 符号 | 说明 |
|---|---|
| `Interceptor` | 拦截器接口，`Intercept` 包装 `http.RoundTripper`，`Apply` 直接作用于 `http.Client` |
| `NewInterceptor(config Config) Interceptor` | 使用完整 `Config` 创建拦截器链 |
| `NewInterceptorWithOpts(opts ...ConfigOption) Interceptor` | 使用选项模式创建拦截器链，基于默认配置叠加 |
| `Config` | 拦截器配置：`EnableTraffic`、`EnableMetrics`、`Headers` |
| `WithEnableTraffic(bool) ConfigOption` | 开启/关闭流量日志记录 |
| `WithEnableMetrics(bool) ConfigOption` | 开启/关闭 Prometheus 指标上报 |
| `WithHeaders(map[string]string) ConfigOption` | 设置统一注入的请求头 |
| `HeaderNameAuthorization` / `HeaderNameContentType` | 常用 Header 名称常量，配合 `WithHeaders` 使用 |

引入路径：`import "github.com/tenz-io/gokit/httpext/v3"`

## 与 v2 的行为差异

v3 不保证与 v2 兼容，以下是显式的行为差异：

| 差异点 | v2 | v3 |
| --- | --- | --- |
| 慢请求日志 | `WithSlowLogFloor` + `slowLogTransport`，超过阈值打印 Warn | **移除**：慢请求由 `monitor/v3` 的 latency 直方图覆盖（按阈值告警） |
| Client 封装 | `Client` 接口（`JSON`/`DoSimple`/`Do`）+ `SimpleRequest`/`RequestOption`/`HttpMethod` | **全删**：直接用标准库 `*http.Client` 的 `Get`/`Post`/`Do`，JSON 编解码就是 `json.Marshal`/`Unmarshal` 三行 |
| `Client.DoSimple` 返回 | `(respBody []byte, err error)`，非 200 即 error | 无此方法；标准库 `http.Client.Do` 自带 `(resp, err)`，状态码从 `resp.StatusCode` 读 |
| 流量日志 API | `logger.StartTrafficRec` + `ReqEntity`/`RespEntity` | `logger.FromContext(ctx).StartTraffic(cmd).WithTyp(...)` + `rec.End(resp, code, fields...)` / `rec.EndWithError(err, fields...)` |
| 流量日志 body 采集 | 读出请求/响应 body，JSON/表单解析为结构化字段、文本截断展示，读后复位 body | **不记 body**：只记 cmd/cost/code/method/url/query 及请求/响应 Header，不读不打扰 body 流 |
| 指标上报 | `monitor.BeginRecord` + `rec.EndWithCode` | `monitor.Begin` + `rec.EndWithCode`（幂等） |
| 文件结构 | 单文件偏多 | 多文件拆分（`doc`/`config`/`interceptor`/`transport`/`traffic`），与 `logger/tracer/annotation v3` 一致；v3 不再有 `capture.go` |

调用方迁移速查（**不在本次范围，留待后续逐步改**）：

| 调用方代码（v2） | v3 等价 |
| --- | --- |
| `httpext.WithSlowLogFloor(d)` | 删除：改用 `monitor/v3` 直方图阈值告警 |
| `httpext.NewClient(cli)` + `cli.JSON/DoSimple/Get/Post/...` | 删除：直接用 `cli.Get`/`cli.Post`/`cli.Do`（标准库） |
| `httpext.NewSimpleRequest(url, method, opts...)` | 删除：用 `http.NewRequest(method, url, body)` |
| `httpext.WithRequestParams/WithRequestHeaders/WithRequestBody` | 删除：用 `req.URL.Query()` / `req.Header.Set` / `req.Body` |
| `httpext.MethodGet/MethodPost/...` | 删除：用 `http.MethodGet`/`http.MethodPost`/...（标准库） |
| `logger.StartTrafficRec(ctx, &ReqEntity{...})` | `logger.FromContext(ctx).StartTraffic(url).WithTyp(logger.TrafficTypSend)` |
| `rec.End(&RespEntity{...}, logger.Fields{...})` | `rec.End(captureResponse(resp), code, "k", v, ...)`（交替 kv）或 `rec.EndWithError(err, "k", v, ...)` |
