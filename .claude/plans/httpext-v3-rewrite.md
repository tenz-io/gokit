# httpext/v3 重写计划

## 目标

对 `httpext`（HTTP 客户端扩展：可组合的 `http.RoundTripper` 拦截器链）进行 v3 干净重写，依赖 `logger/v3`、`monitor/v3`、`tracer/v3`。v2 保留不动，与 v3 并存，消费方本轮不迁移（与 logger/tracer/monitor 的 v3 处理一致）。

核心能力不变：拦截器链一次性 `Apply` 到 `*http.Client`；注入固定 Header；上报指标；记录请求/响应流量流水；简化 JSON 调用。

## v2 的问题（v3 修复）

1. **慢日志绕链 bug**：`slowLogTransport.RoundTrip` 调 `http.DefaultTransport.RoundTrip(req)` 而非 `st.tripper.RoundTrip(req)`，绕过拦截器链其余部分 → **v3 直接移除 slowLogTransport**，慢请求由 `monitor/v3` 的 latency 直方图覆盖（可配告警阈值）。`WithSlowLogFloor` / `Config.SlowLogFloor` 一并删除。
2. **非 200 即 error**：`client.DoSimple` 在 `resp.StatusCode != http.StatusOK` 时返回 error，对 REST（201/204/304 合法、4xx/5xx 需读 body）是错的 → v3 `DoSimple` 返回 `(respBody []byte, status int, err error)`，仅 transport/网络层错误返回 err，状态码交由调用方判断。`JSON` 仍要求成功反序列化，对非 2xx 返回带 status 的 error。
3. **`captureResponse` 的 `ifThen` 泛型 + 不安全类型断言**：`ifThen(cond, tf, ff).(context.Context)` 难读且依赖类型断言 → v3 用普通 `if`/局部变量替换，删除 `ifThen`。
4. **流量采集逻辑散乱、req/resp 重复 content-type 解析** → v3 抽到独立 `capture.go`，合并 request/response 的 content-type 分支为单一 `captureBody(contentType, body)`，统一 JSON/表单/文本/截断处理。
5. **transport 链构建依赖 `active()` + 类型断言** → 保留 `transporter` 接口与 `active()` 机制（v2 测试深度依赖此结构），但顺序固定、语义清晰：`injectHeader → metrics → traffic`（slowLog 已删）。
6. **`captureRequest` 里 `req.PostForm` 在 body 已读后不可靠** → v3 仅对 `x-www-form-urlencoded` 在未读 body 前返回 `req.PostForm`，否则按文本处理；并在读 body 后用 `bytes.NewReader`（非 `NewBuffer`）复位，行为一致。

## v3 API 表面（依赖确认）

**logger/v3**（`github.com/tenz-io/gokit/logger/v3`）：
- `logger.FromContext(ctx) Entry`（无 attached 时回落全局）
- `Entry` 方法：`Debug/Debugf/Debugw`, `Warn/Warnf/Warnw`, `WithError(err) Entry`, `WithFields(Fields) Entry`, `WithField(k,v) Entry`
- 流量：`Entry.StartTraffic(cmd string) *TrafficRec`（traffic 未配置返回 nil，nil-safe）；`(*TrafficRec).WithTyp(TrafficTypSend) *TrafficRec`；`(*TrafficRec).End(resp any, code string, fields ...any)`；`(*TrafficRec).EndWithError(err error, fields ...any)`。`TrafficTypRecv/TrafficTypSend`
- `type Fields map[string]any`
- 配置（example 用）：`logger.ConfigureWithOpts(logger.WithLevel(DebugLevel), logger.WithConsole(true), logger.WithFilePath("log"), logger.WithCaller(true), logger.WithCallerSkip(1), logger.WithTraffic(true))`

**monitor/v3**（`github.com/tenz-io/gokit/monitor/v3`）：
- `monitor.Begin(ctx, dsCmd string) *Recorder`；`(*Recorder).EndWithCode(code string)`（幂等）；`(*Recorder).EndWithError(err error)`；`(*Recorder).End()`。`EndWithCode` 接受 `"200"` 等字符串。无 Exporter 时为 noop，`Begin`/`End` 均安全。
- `monitor.FromContext(ctx) Exporter`、`monitor.Init(ctx, cmd) context.Context`（请求边注入单飞 Exporter；httpext 作为出站 client 不主动 Init，只复用 ctx 里已有的）。

**tracer/v3**（`github.com/tenz-io/gokit/tracer/v3`）：
- `tracer.FromContext(ctx) Flag`；`Flag.IsDebug() bool`；`Flag.Is(Flag) bool`。`tracer.FlagDebug`。nil ctx 返回 `FlagNone`，不 panic。

## 文件结构（多文件拆分，遵循 v3 惯例）

```
httpext/v3/
  doc.go            // package doc + godoc 快速开始
  config.go         // Config + ConfigOption(WithEnableTraffic/WithEnableMetrics/WithHeaders)
  interceptor.go    // Interceptor 接口 + intercept/Apply + newTransporters 顺序
  client.go         // Client 接口 + SimpleRequest + RequestOption + NewSimpleRequest
  transport.go      // metricsTransport + injectHeaderTransport + transporter 接口 + newTransporterFunc
  traffic.go        // trafficTransport（RoundTrip 编排 StartTraffic/End）
  capture.go        // captureRequest/captureResponse/captureBody + limitString（删除 ifThen）
  *_test.go         // 按职责：interceptor_test/client_test/traffic_test
  go.mod / go.sum
  Makefile
  README.md
  example/
    go.mod          // module httpext-v3-example, replace ./..
    main.go
```

`newTransporters` 顺序：`newInjectHeaderTransport`, `newMetricsTransport`, `newTrafficTransport`（slowLog 删除）。

## 关键实现点

### config.go
- `Config{ EnableTraffic bool; EnableMetrics bool; Headers map[string]string }`（删 `SlowLogFloor`）
- `defaultConfig = Config{EnableTraffic: true, EnableMetrics: false, Headers: nil}`（与 v2 默认一致，便于 example）
- 选项：`WithEnableTraffic`, `WithEnableMetrics`, `WithHeaders`（删 `WithSlowLogFloor`）

### client.go
- `Client` 接口：
  - `JSON(ctx, url, method, reqBody, respBody any, reqOpts ...RequestOption) error` —— 内部调 `DoSimple`，对非 2xx status 返回 `fmt.Errorf("unexpected status %d: %w", status, ...)`；2xx 且有 body 才 unmarshal
  - `DoSimple(ctx, req *SimpleRequest) (respBody []byte, status int, err error)` —— 返回 body + status；仅 validate/建请求/网络/读 body 错误返回 err；**不再因 status 返回 error**
  - `Do(req *http.Request) (*http.Response, error)` —— 透传
- `SimpleRequest{Url, Method, Headers, Params, ReqBody}`、`validate()`、`RequestOption`、`WithRequestParams/WithRequestHeaders/WithRequestBody`、`NewSimpleRequest` 保持签名兼容（仅 method 类型 `HttpMethod`、`MethodGet/...` 常量、`Params`/`Headers` 别名、`HeaderNameAuthorization/ContentType` 常量不变）
- `readBody` 不变

### transport.go
- `metricsTransport.RoundTrip`：用 `monitor.Begin(ctx, url)` + defer `rec.EndWithCode(fmt.Sprintf("%d", code))`（code 兜底沿用 v2 的 `1` 当 err 路径）
- `injectHeaderTransport.RoundTrip`：`req.Header.Set` 注入 headers，不变
- 删除 `slowLogTransport`、`newSlowLogTransport`

### traffic.go
- `trafficTransport.RoundTrip`：条件 `if !tt.enable && !tracer.FromContext(ctx).IsDebug() { return tt.tripper.RoundTrip(req) }`
- 编排：
  ```go
  rec := logger.FromContext(ctx).StartTraffic(url).WithTyp(logger.TrafficTypSend)
  defer func() {
      code := 1
      var respHeaders http.Header
      if err == nil && resp != nil {
          code = resp.StatusCode
          respHeaders = resp.Header
      }
      if err != nil {
          rec.EndWithError(err, "method", req.Method, "url", req.URL.String(),
              "query", req.URL.Query(), "req_header", reqHeaders,
              "resp_header", respHeaders)
          return
      }
      rec.End(captureResponse(resp), fmt.Sprintf("%d", code),
          "method", req.Method, "url", req.URL.String(),
          "query", req.URL.Query(), "req_header", reqHeaders,
          "resp_header", respHeaders)
  }()
  // 在调用前先把 request body 抓下来（captureRequest 会读并复位 body）
  reqCapture := captureRequest(req)
  _ = reqCapture // 通过 fields 传入；见下
  return tt.tripper.RoundTrip(req)
  ```
  > 注意：`captureRequest` 会 `io.ReadAll(req.Body)` 并复位。request 抓取放主流程、`captureRequest` 复位 body 后，`tt.tripper.RoundTrip(req)` 读到复位后的 body，正确。
  >
  > 字段传 `End`：v3 `End(resp, code, fields ...any)` 收交替 `key,value`；`logger.Fields.toArgs()` 未导出，不能跨包调用。因此 traffic.go 直接把结构化字段拼成 `[]any{"method", req.Method, "url", req.URL.String(), "query", req.URL.Query(), "req_header", reqHeaders, "req", reqCapture, "resp_header", respHeaders}` 传入（query/Header 是 map/slice，zap 支持结构化输出）。req 抓取结果 `reqCapture` 在主流程算好，defer 闭包捕获使用。
  >
  > 改用闭包显式捕获：主流程 `reqHeaders := req.Header; reqCapture := captureRequest(req)`；defer 引用这两个局部变量。
- `captureRequest`/`captureResponse` 移到 `capture.go`

### capture.go
- `captureBody(contentType string, body []byte) any`：合并 json/text/表单分支；json unmarshal 失败返 `"<json decode failed>"`；其它文本 `limitString(string(body), 128)`；空 body 返 nil
- `captureRequest(req *http.Request) any`：GET/nil body → nil；`x-www-form-urlencoded` → `req.PostForm`（注意读序）；json/text → `captureBody`；不支持 → `"<unsupported content-type>"`。读 body 后 `req.Body = io.NopCloser(bytes.NewReader(bs))` 复位
- `captureResponse(resp *http.Response) any`：从 `resp.Request.Context()` 取 ctx（用普通 if，删 ifThen）；json/text → `captureBody`；读后 `resp.Body = io.NopCloser(bytes.NewReader(bodyCopy))` 复位
- `limitString(s string, n int) string`：不变
- 删除 `ifThen`、`errorMsg`（err msg 由 `EndWithError` 处理）

## 测试

- `interceptor_test.go`：迁移 v2 的 `Test_interceptor_Apply` / `TestInterceptor`（mockTransport + 表/子测试），断言去掉 slowLog 层；transport 链断言改为 `traffic→metrics→injectHeader→mock` 或子集。改用 `logger/v3`、`tracer/v3` 的 setup。
- `client_test.go`：迁移 `Test_client_JSON`；新增 `DoSimple` 返回 status 的用例（2xx 返 body+status、nil err；4xx 返 body+status、nil err；网络 err 返 0+err）。
- `traffic_test.go`：`captureRequest`/`captureResponse`/`captureBody` 单测（json/text/form/empty/unsupported 各路径）。
- setup 用 `logger.ConfigureWithOpts(logger.WithLevel(DebugLevel), logger.WithConsole(true), logger.WithFilePath("log"), logger.WithCaller(true), logger.WithCallerSkip(1), logger.WithTraffic(true))`，teardown `time.Sleep(100ms)` 等 traffic 异步落盘。
- 验证：`GOWORK=off go vet ./...` + `GOWORK=off go test ./... -cover`（模块独立，与 tracer/v3 同款 workspace 规避——但本模块依赖 logger/monitor/tracer v3，GOWORK=off 需 replace 或走 go.work；优先用 `go test ./httpext/v3/...` workspace 模式，若 go.work 有不存在的 example 目录阻断则 GOWORK=off + 本地 replace）。

## go.work

在 `use (...)` 列表追加（v2 行后）：
```
./httpext/v3
./httpext/v3/example
```

## go.mod

```
module github.com/tenz-io/gokit/httpext/v3

go 1.24

require (
	github.com/stretchr/testify v1.9.0
	github.com/tenz-io/gokit/logger/v3 v3.0.0
	github.com/tenz-io/gokit/monitor/v3 v3.0.0
	github.com/tenz-io/gokit/tracer/v3 v3.0.0
)
```
indirect 由 `go mod tidy` 补齐（zap/lumberjack/prometheus/uuid 等）。prometheus 不在 httpext 直接 require（由 monitor/v3 传递），避免内存里 app/v3 提到的 pin 漂移问题。

## README.md（中文）

结构对齐 tracer/v3 README：模块介绍 → 能力清单（表）→ 快速开始 → API 速查（表）→ 变更说明（v2→v3：移除 SlowLogFloor/slowLogTransport；DoSimple 返回 status；流量 API 走 logger/v3 StartTraffic/End）。

## 验收

1. `GOWORK=off go vet ./...`（在 httpext/v3 内）clean
2. `GOWORK=off go test ./... -cover` 通过，覆盖 > 70%
3. `go.work` 列入 v3 两模块，`go build ./...` 在仓库根通过
4. example 可 `go run ./example`（指向 localhost，允许连接失败但编译/运行不 panic）
5. 无 `ifThen`、无 `slowLogTransport`、无 `WithSlowLogFloor`/`SlowLogFloor` 残留

## 不做

- 不迁移 v2 消费方（仅 example 自身用 v3）
- 不改 v2 任何文件
- 不引入新的外部依赖（仅 v3 三件套 + testify）
