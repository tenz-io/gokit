# httpext

HTTP 客户端扩展，提供可组合的传输层拦截器链：请求头注入、流量采集、指标和慢请求日志。

## 功能特性

- `Interceptor`：将拦截器链（注入 Header、指标、流量、慢日志）一次性 `Apply` 到 `*http.Client`，无需手动包装 `Transport`
- `WithHeaders`：为所有出站请求统一注入固定 Header（如鉴权 Token）
- `WithEnableMetrics`：开启后自动上报每次请求的耗时与状态码到 monitor
- `WithEnableTraffic`：开启后（或链路处于 debug 态）自动记录请求/响应体流水日志，JSON/表单/文本自动解析，非文本类型截断展示
- `WithSlowLogFloor`：设置慢请求阈值，超过阈值的请求自动打印带耗时、URL、Query 的 Warn 日志
- `Client` 接口：封装 `JSON`/`DoSimple`/`Do` 三种粒度的调用方式，简化 JSON 请求的编解码样板代码
- `NewSimpleRequest` + `RequestOption`：以选项模式构造带 Query 参数、Header、Body 的简单请求

## 能力清单

| 能力 | 含义 |
|---|---|
| 一键装配拦截器链 | `Interceptor.Apply` 按顺序叠加注入 Header、指标、流量、慢日志四类 `http.RoundTripper`，直接作用于 `*http.Client`，无需手动包装 `Transport` |
| 统一注入固定请求头 | `WithHeaders` 配置的 Header（如 `Authorization`）在每次出站请求前自动 `Set`，避免调用方重复设置鉴权信息 |
| 上报请求指标 | `WithEnableMetrics` 开启后，`metricsTransport` 用 `monitor.BeginRecord` 记录每次请求的 URL 与响应状态码，用于监控接口耗时与成功率 |
| 记录请求/响应流量日志 | `WithEnableTraffic` 开启（或链路处于 debug 态时自动触发）后，`trafficTransport` 采集请求/响应 Header 与 Body，JSON/表单自动解析为结构化字段，其他文本截断展示，便于排查线上问题 |
| 慢请求告警 | `WithSlowLogFloor` 设置阈值后，`slowLogTransport` 对超过阈值的请求打印带耗时、URL、Query 的 Warn 日志，用于发现异常慢接口 |
| 简化 JSON 请求调用 | `Client.JSON` 自动将请求体序列化为 JSON、设置 `Content-Type`，并把响应体反序列化到目标结构体，省去手写编解码样板代码 |
| 选项式构造简单请求 | `NewSimpleRequest` + `WithRequestParams`/`WithRequestHeaders`/`WithRequestBody` 以选项模式拼装 Query 参数、Header、Body，适合非 JSON 场景 |
| 请求合法性校验 | `SimpleRequest.validate` 在发起请求前检查 URL 是否为空、Method 是否为支持的 HTTP 方法，避免无效请求发出 |

## 快速开始

```go
import "github.com/tenz-io/gokit/httpext/v2"
```

```go
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/tenz-io/gokit/httpext/v2"
)

func main() {
	// 1. 构建拦截器链并应用到 http.Client
	interceptor := httpext.NewInterceptorWithOpts(
		httpext.WithEnableTraffic(true),
		httpext.WithEnableMetrics(true),
		httpext.WithHeaders(map[string]string{
			httpext.HeaderNameAuthorization: "Bearer token",
		}),
		httpext.WithSlowLogFloor(500*time.Millisecond),
	)
	httpCli := &http.Client{}
	interceptor.Apply(httpCli)

	// 2. 用封装的 Client 发起 JSON 请求
	cli := httpext.NewClient(httpCli)

	type Req struct {
		Name string `json:"name"`
	}
	type Resp struct {
		Ok bool `json:"ok"`
	}

	var resp Resp
	err := cli.JSON(context.Background(), "https://example.com/api", httpext.MethodPost, &Req{Name: "gokit"}, &resp)
	if err != nil {
		panic(err)
	}
}
```

## API 速查

| 符号 | 说明 |
|---|---|
| `Interceptor` | 拦截器接口，`Intercept` 包装 `http.RoundTripper`，`Apply` 直接作用于 `http.Client` |
| `NewInterceptor(config Config) Interceptor` | 使用完整 `Config` 创建拦截器链 |
| `NewInterceptorWithOpts(opts ...ConfigOption) Interceptor` | 使用选项模式创建拦截器链，基于默认配置叠加 |
| `Config` | 拦截器配置：`EnableTraffic`、`EnableMetrics`、`Headers`、`SlowLogFloor` |
| `WithEnableTraffic(bool) ConfigOption` | 开启/关闭流量日志记录 |
| `WithEnableMetrics(bool) ConfigOption` | 开启/关闭 Prometheus 指标上报 |
| `WithHeaders(map[string]string) ConfigOption` | 设置统一注入的请求头 |
| `WithSlowLogFloor(time.Duration) ConfigOption` | 设置慢请求日志阈值 |
| `Client` | 封装接口：`JSON`（JSON 编解码请求）、`DoSimple`（简单请求返回原始 body）、`Do`（等价 `http.Client.Do`） |
| `NewClient(cli *http.Client) Client` | 基于标准 `http.Client` 构造 `Client` |
| `SimpleRequest` | 简单请求描述：`Url`、`Method`、`Headers`、`Params`、`ReqBody` |
| `NewSimpleRequest(url string, method HttpMethod, opts ...RequestOption) *SimpleRequest` | 构造 `SimpleRequest` |
| `WithRequestParams(Params) RequestOption` | 设置请求的 Query 参数 |
| `WithRequestHeaders(Headers) RequestOption` | 设置请求的 Header |
| `WithRequestBody([]byte) RequestOption` | 设置请求 Body |
| `HttpMethod` / `MethodGet` `MethodPost` `MethodPut` `MethodDelete` `MethodHead` `MethodPatch` | HTTP 方法类型及常量 |
| `Params` / `Headers` | Query 参数与 Header 的别名类型，均为 `map[string][]string` / `map[string]string` |
| `HeaderNameAuthorization` / `HeaderNameContentType` | 常用 Header 名称常量 |

import path: `github.com/tenz-io/gokit/httpext/v2`
