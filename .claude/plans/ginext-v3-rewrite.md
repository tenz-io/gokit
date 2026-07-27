# ginext/v3 重写计划

## 目标

对 `ginext`（Gin 扩展：请求绑定+校验、JWT 鉴权、统一响应、RPC 观测拦截器链）进行 v3 干净重写，依赖 `annotation/v3`、`logger/v3`、`monitor/v3`、`tracer/v3`。v2 保留不动，与 v3 并存，消费方本轮不迁移（与 logger/tracer/monitor/httpext 的 v3 处理一致）。

核心能力不变：`bind`/`validate`/`default` 标签驱动多来源请求绑定与统一校验；`{code,message,data}` 统一响应体 + 文件流直传；JWT 签发/校验与基于角色的鉴权中间件；RPC 拦截器链（Tracer→Metrics→Traffic→SlowLog→PanicRecover）。

目标：更简洁、更易用、消除 v2 的历史包袱。

## v2 的问题（v3 修复）

1. **`metadata.MD` 持有 `*gin.Context`**：观测拦截器从 context 取 `*MD`，`MD` 又回指 `*gin.Context`，形成 gin↔拦截器层的耦合，且 `MD` 里的 `RequestID`/`Userid` 等字段本质是请求上下文的一次性快照，不该作为跨包传递载体 → **v3 删除 `metadata` 子包**，拦截器直接从 `tracer`/`logger`/`gin.Context` 取所需，不再有中间 `MD` 层。
2. **`init()` 反向装配拦截器链 + `AllRpcInterceptor` 全局变量**：`init()` 倒序包裹 5 个构造函数得到一个包级 `AllRpcInterceptor`，不可配、不可测、顺序硬编码 → **v3 改为 `NewChain(opts ...Option) Chain`**，默认构造常用链，可按需裁剪；不再用 `init()` 做装配。
3. **每个拦截器重复 `meta, ok := metadata.FromContext(ctx); if !ok { return next }` 短路**：5 个拦截器各写一遍 → **v3 拦截器不再依赖 `MD`**，各自直接、廉价地读 `tracer.FromContext(ctx)` / `monitor.FromContext(ctx)` / `logger.FromContext(ctx)`，nil-safe，无短路样板。
4. **JWT 全局可变 `jwtKey = []byte("my_secret_key")` + `InitJWT`**：包级 `var jwtKey` 非并发安全（`InitJWT` 写、`GenerateToken`/`VerifyToken` 读，无锁），默认硬编码弱密钥 → **v3 改为 `Auth{secret []byte}` 结构体**，`NewAuth(secret)` 构造，`(*Auth).GenerateToken` / `(*Auth).VerifyToken` 方法；包级零值兜底 `defaultAuth`（仅便于 example/测试，密钥仍硬编码但不再被竞写）。
5. **bit-mask 角色判定 `role & claims.Role > 0`**：`RoleType=int32` + `RoleAnonymous=0` 的魔法值让 `0 & x` 恒为 0，"匿名跳过" 靠 `role == RoleAnonymous` 早返回而非掩码，掩码仅对 admin/user 两位巧合成立，语义脆弱 → **v3 用显式 `[]string` roles**：`Claims{UserID int64; Roles []string}`，`RequireRole(roles ...string)` 中间件做集合包含判定，无掩码、无魔法值。
6. **`AuthType` 枚举（Web=cookie / Rest=header）**：枚举本身只是函数选择，却要穿透 `Authenticate(role, authType)` → `AuthenticateCookie`/`AuthenticateRest` 三层 → **v3 直接两个函数** `RequireRoleHeader(roles...)` / `RequireRoleCookie(roles...)`（及主动判定 `IsAuthedHeader`/`IsAuthedCookie`），删 `AuthType` 枚举与 `Authenticate(role, authType)` 调度层。
7. **`Response` 恒 200，`ResponseFrame` 无 HTTP status 维度**：成功 201/204 无法表达 → **v3 `Response(c, data)`（默认 200）+ `ResponseStatus(c, status, data)`**；`ResponseFrame{code,message,data}` 不变，HTTP status 由 `c.Status()` 设定，JSON 体仍是 `{code,message,data}`。
8. **`metadata.New` 里 `c.Request.WithContext(...)` 未生效**：`c.Request = c.Request.WithContext(...)` 才会替换，仅调 `WithContext` 返回新 request 却不赋值，requestID 写回 header 但 context 未传播 → v3 删除 `metadata` 后此 bug 自然消失；requestID 由拦截器在请求边 `tracer.WithRequestID` 注入并 `c.Set`/header 回写。

## v3 API 表面（依赖确认）

**annotation/v3**（`github.com/tenz-io/gokit/annotation/v3`）—— 与 v2 ginext 已用一致：
- `annotation.ApplyDefaults(ptr) error`、`annotation.PlanFor(ptr) (*Plan, error)`、`plan.FieldsBySource(src) []*Field`、`annotation.SetString/Set(rv, v) error`、`annotation.Validate(ptr) error`、`annotation.AsErrors(err) (ValidationErrors, bool)`、`annotation.Err(field, rule, msg)`。
- `BindSource`：`BindURI/BindQuery/BindHeader/BindForm/BindFile`。`Field.BindName/BindSource/IsRequired/Index`。v3 无需改动 bind 逻辑。

**logger/v3**（`github.com/tenz-io/gokit/logger/v3`）：
- `logger.FromContext(ctx) Entry`、`logger.WithLogger(ctx, entry) context.Context`。
- `Entry.With(args ...any) Entry`、`WithError(err) Entry`、`WithRequestID(id) Entry`、`Warnf/Warnw/Debugf/Infof/Errorf`。
- 流量：`Entry.StartTraffic(cmd string) *TrafficRec`（traffic 未配置返回 nil，nil-safe）；`(*TrafficRec).End(resp any, code string, fields ...any)`；`(*TrafficRec).EndWithError(err error, fields ...any)`。
- 配置（example 用）：`logger.ConfigureWithOpts(logger.WithLevel(DebugLevel), logger.WithConsole(true), logger.WithFilePath("log"), logger.WithCaller(true), logger.WithCallerSkip(1), logger.WithTraffic(true))`。

**monitor/v3**（`github.com/tenz-io/gokit/monitor/v3`）：
- `monitor.Init(ctx, cmd string) context.Context`（请求边注入单飞 Exporter，幂等）；`monitor.Begin(ctx, dsCmd string) *Recorder`（无 Exporter 时 noop，nil-safe）；`(*Recorder).EndWithError(err error)`（幂等）/`EndWithCode(code string)`/`End()`。
- `monitor.FromContext(ctx) Exporter`、`monitor.HasExporter(ctx) bool`。

**tracer/v3**（`github.com/tenz-io/gokit/tracer/v3`）：
- `tracer.EnsureRequestID(ctx) (context.Context, string)`（无则生成，返回带 id 的 ctx + id，用于请求边）；`tracer.RequestIDFromCtx(ctx) string`（无则空串）；`tracer.WithRequestID(ctx, id) context.Context`。
- `tracer.ParseFlag(s string) Flag`（"debug|shadow" → 掩码，未知 token 忽略，替代 v2 的 `getTracerFlag`）；`tracer.WithFlag(ctx, flag) context.Context`；`tracer.FromContext(ctx) Flag`；`Flag.Is(Flag) bool` / `Flag.IsDebug() bool`。`FlagDebug/FlagShadow/FlagStress`、`FlagNone`。

## 文件结构（多文件拆分，遵循 v3 惯例）

```
ginext/v3/
  doc.go          // package doc + godoc 快速开始
  bind.go         // BindAndValidate / Validate + tryBind* (uri/query/header/form/multipart/json)
  response.go     // ResponseFrame / FileResponse / Response / ResponseStatus / ErrorResponse
  auth.go         // Auth struct + NewAuth + GenerateToken/VerifyToken + Claims + RequireRoleHeader/Cookie + IsAuthedHeader/Cookie
  handler.go      // RpcHandler / Interceptor / Chain / 默认链构造 + 5 个内置拦截器
  errcode/error.go // 错误码（保留子包，与 v2 一致）
  go.mod / go.sum
  Makefile
  README.md
  example/        // 绑定+响应示例
    go.mod / main.go
  example-auth/   // 鉴权示例（token 签发 + 中间件 + 角色保护）
    go.mod / main.go
```

> `metadata` 子包**删除**（理由见上）。`utils.go` 里 `getStructFieldNames`/`getFieldNames`/`getProtoName`/`getJSONName`/`getPtrElem` 在 v2 中是死代码（grep 无调用方），v3 不移植；仅保留 `getErrCode`/`getErrMsg` 移入 `handler.go`。`getTracerFlag` 删除（改用 `tracer.ParseFlag`）。`validate.go` 的 `warpError` 合并进 `bind.go`（私有 `warpError`）。

## 关键实现点

### bind.go
- `BindAndValidate(c *gin.Context, ptr any) (err error)`：签名不变；逻辑与 v2 一致（`ApplyDefaults` → uri/query/header/multipart/form/json 顺序 tryBind → `Validate`），失败 defer 包 `warpError`。
- `Validate(c, ptr)`：不变。
- `tryBindURI/Query/Header/Form/JSON/Multipart`、`readFileAndSetField`/`readFormAndSetField`：逻辑不变（已是 v3 风格，注释改中文化与 v3 惯例一致）。
- `warpError(c, err)`：与 v2 一致（`annotation.AsErrors` → `errcode.New(StatusBadRequest, ...)`；json 错误 → `errcode.New`；其余 → `errcode.New`）。
- `rootValue`/`fieldByIndex`：不变。

### response.go
- `ResponseFrame{Code int; Message string; Data any}` 不变。
- `FileResponse` 接口 `GetFile() []byte` 不变。
- `Response(c, data)`：默认 200；data 为 `FileResponse` → 直传文件（`http.DetectContentType`）；nil data → `gin.H{}`。
- **新增 `ResponseStatus(c *gin.Context, status int, data any)`**：与 `Response` 同逻辑，但 `c.Status(status)` 在写 JSON 前设 HTTP status（FileResponse 路径仍用 `c.Data(200, ...)`，文件流 status 固定 200，符合文件直传语义）。
- `ErrorResponse(c, err, data...)`：与 v2 一致（`errors.As` `errcode.Error` → 用其 `Status` 写 JSON；否则 500）。`c.Error(err)` + `c.Abort()`。

### auth.go
```go
type Auth struct {
    secret []byte
    now    func() time.Time // 注入便于测试，默认 time.Now
}

func NewAuth(secret string) *Auth
func (a *Auth) GenerateToken(userid int64, roles []string, tokenType TokenType, expiredAt time.Time) (string, error)
func (a *Auth) VerifyToken(tokenString string) (*Claims, error)

type Claims struct {
    UserID int64    `json:"user_id"`
    Roles  []string `json:"roles"`
    Type   TokenType `json:"type"`
    jwt.RegisteredClaims
}

type TokenType int // 0 access, 1 refresh —— 仅 token 类型，不含角色/鉴权方式枚举
const (TokenTypeAccess TokenType = 0; TokenTypeRefresh TokenType = 1)

// 角色常量（字符串，显式、可扩展）
const (RoleAdmin = "admin"; RoleUser = "user"; RoleAnonymous = "")
```
- 中间件（gin.HandlerFunc）：
  - `RequireRoleHeader(a *Auth, roles ...string) gin.HandlerFunc`：从 `Authorization`（去 `Bearer `）取 token；`a.VerifyToken` → 校 `TokenType == Access` → `RequireRole` 集合包含判定 → `c.Set("user_id",...)`/`c.Set("roles",...)`/`c.Next()`；失败 `ErrorResponse(c, errcode.Unauthorized(...))` + `c.Abort()`。
  - `RequireRoleCookie(a *Auth, roles ...string) gin.HandlerFunc`：从 cookie `token` 取；同上判定；**鉴权成功后用 `a.GenerateToken` 续期**并 `c.SetCookie`（沿用 v2 行为）。
  - `roles` 含 `RoleAnonymous`("") 或为空 → 不鉴权（`c.Next()`），与 v2 "匿名跳过" 语义一致。
- 主动判定：
  - `IsAuthedHeader(a *Auth, c *gin.Context, roles ...string) bool`
  - `IsAuthedCookie(a *Auth, c *gin.Context, roles ...string) bool`（含续期）
- 私有 `verifyAndCheck(a, token, roles) (*Claims, bool)`：单一校验路径，`VerifyToken` + type + role 集合判定，无 bit-mask。
- `IsUnauthorizedError(err)`：保留，判 `errors.Is(err, ErrUnauthorized)` / `ErrInvalidToken`。`ErrInvalidToken`/`ErrUnauthorized` 为包级 sentinel。
- 包级零值兜底：`var defaultAuth = NewAuth("my_secret_key")`，仅 example/test 用；生产必须 `NewAuth`。`GenerateToken`/`VerifyToken` 的包级便捷函数转发到 `defaultAuth`（保持 example 简短），但中间件必须传显式 `*Auth`。

### handler.go（拦截器链）
```go
type RpcHandler func(ctx context.Context, req any) (resp any, err error)
type Interceptor func(ctx context.Context, req any, next RpcHandler) (resp any, err error)
type Chain struct{ interceptors []Interceptor }

func NewChain(interceptors ...Interceptor) Chain   // 显式给定链
func DefaultChain() Chain                           // Tracer→Metrics→Traffic→SlowLog→PanicRecover

func (c Chain) Intercept(ctx context.Context, req any, h RpcHandler) (any, error)
func (c Chain) With(interceptors ...Interceptor) Chain // 追加/前置
```
- **`Interceptor` 是函数类型**（非接口），比 v2 的 `RpcInterceptor` 接口 + 5 个 struct 简洁；链编排是普通 reduce。
- 内置 5 个 `Interceptor` 工厂（均返回 `Interceptor`）：
  - `TracerInterceptor`：`tracer.EnsureRequestID(ctx)` 注入 id + 回写 `X-Request-Id` header；`tracer.ParseFlag(flagHeader)` → `tracer.WithFlag`；`logger.WithLogger(ctx, logger.WithRequestID(id).With("path",..,"cmd",..,"flag",..))`。直接从 `ctx`/`*gin.Context` 取，不依赖 `MD`。
  - `MetricsInterceptor`：`monitor.Init(ctx, cmd)`（请求边单飞）+ `rec := monitor.Begin(ctx, "total")` + defer `rec.EndWithError(err)`。
  - `TrafficInterceptor`：`rec := logger.FromContext(ctx).StartTraffic(cmd)`（nil-safe）+ defer `rec.End(resp, getErrCode(err), "path",..,"cmd",..)` / `rec.EndWithError(err)`。
  - `SlowLogInterceptor`（默认阈值 5s，可 `WithSlowLogThreshold` 覆盖）：`time.Since(start)` > 阈值 → `logger.FromContext(ctx).Warnw("slow log", ...)`。
  - `PanicRecoverInterceptor`：defer `recover()` → `logger.FromContext(ctx).Errorw("panic recovery", "panic", r, "stack", string(debug.Stack()))` + `err = errcode.InternalServer(500, "panic")`。
- `getErrCode(err)` / `getErrMsg(err)` 移入此文件。
- `cmd` 来源：拦截器从 ctx 取 `"rpc_cmd"`（或 gin context 的 path）。请求边由 `TracerInterceptor` 设 `cmd = c.Request.URL.Path`，后续拦截器复用。
- 链顺序（`DefaultChain`）：`PanicRecover` 最外（先 recover 再记录），`Tracer`→`Metrics`→`Traffic`→`SlowLog` 内层，业务 handler 最内。与 v2 `init()` 倒序结果等价，但显式可读。

### errcode/error.go
- 与 v2 一致（`Error{Code,Message,Status}` + `New/BadRequest/Unauthorized/Forbidden/NotFound/InternalServer/FromError`）。v3 仅注释中文化。**不改 API**（`response.go`/`warpError`/中间件均依赖）。

### example/main.go（绑定+响应）
- 用 `logger/v3` ConfigureWithOpts 初始化。
- gin engine：`POST /user/:id`（`BindAndValidate` + `ResponseStatus` 示例，演示 201）、`PUT /search`（multipart 文件 + form）、`GET /ping`。
- 请求 struct 沿用 v2 example 的 `bind`/`validate`/`default` 标签。

### example-auth/main.go
- `a := ginext.NewAuth("change-me")`。
- `POST /login` → `a.GenerateToken(...)` → `Response(c, gin.H{"token": t})`。
- `GET /admin` + `a.RequireRoleHeader(ginext.RoleAdmin)` → `Response(c, gin.H{"ok": true})`。
- `GET /me` + `a.RequireRoleCookie(ginext.RoleUser)` → 演示 cookie 续期。

## 测试

迁移 v2 测试到 v3（`annotation/v3` API 未变，断言基本可平移）：
- `bind_test.go`：`Test_BindAndValidate_file/_form/_json` + `Test_tryBindMultipart`（`TestResponseFrame[T]` 泛型壳不变）。
- `response_test.go`：`TestResponse`（nil/normal/file 三路径）；新增 `TestResponseStatus`（验证 `ResponseStatus(c, 201, data)` 产出 HTTP 201 + `{code,message,data}`）。
- `auth_test.go`：`TestRequireRoleHeader/_Cookie`（替换 v2 `TestAuthenticate`/`TestIsAuthenticated/_web`）；token 经 `NewAuth(...).GenerateToken`；断言 401/200 与 cookie 续期。
- `handler_test.go`：`TestChain`（替换 `TestAllRpcInterceptor`）；构造 `DefaultChain()`，验证 panic 被恢复为 500、traffic/metrics 不 panic、requestID 回写 header。
- setup：`logger.ConfigureWithOpts(logger.WithLevel(DebugLevel), logger.WithConsole(true), logger.WithFilePath("log"), logger.WithCaller(true), logger.WithCallerSkip(1), logger.WithTraffic(true))` + `teardown time.Sleep(100ms)` 等 traffic 异步落盘（与 httpext/v3 同款）。
- 验证：`GOWORK=off go vet ./...` + `GOWORK=off go test ./... -cover`（模块独立，依赖 v3 三件套，用本地 replace）。

## go.work

在 `use (...)` 列表追加（v2 行后）：
```
./ginext/v3
./ginext/v3/example
./ginext/v3/example-auth
```

## go.mod

```
module github.com/tenz-io/gokit/ginext/v3

go 1.24

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/golang-jwt/jwt/v5 v5.2.1
    github.com/stretchr/testify v1.9.0
    github.com/tenz-io/gokit/annotation/v3 v3.0.0
    github.com/tenz-io/gokit/logger/v3 v3.0.0
    github.com/tenz-io/gokit/monitor/v3 v3.0.0
    github.com/tenz-io/gokit/tracer/v3 v3.0.0
)

// v3 gokit 模块尚未发布，从 workspace 同级目录解析。example 模块同款 replace。
replace (
    github.com/tenz-io/gokit/annotation/v3 => ../../annotation/v3
    github.com/tenz-io/gokit/logger/v3 => ../../logger/v3
    github.com/tenz-io/gokit/monitor/v3 => ../../monitor/v3
    github.com/tenz-io/gokit/tracer/v3 => ../../tracer/v3
)
```
indirect 由 `go mod tidy` 补齐（zap/lumberjack/prometheus/uuid/sonic 等）。gin v1.9.1 与 v2 一致（example 用得到，v3 仍绑 gin）。

## README.md（中文）

结构对齐 tracer/v3/httpext/v3 README：模块介绍 → 能力清单（表）→ 快速开始（绑定/响应/鉴权三段代码）→ API 速查（表）→ 变更说明（v2→v3：删 `metadata` 子包；拦截器链改 `Chain`/`Interceptor` 函数类型 + `DefaultChain`，去 `init()`；`Auth` 结构体替代全局 `jwtKey`；显式 `[]string` 角色替代 bit-mask；`ResponseStatus` 支持 201/204；`tracer.ParseFlag` 替代 `getTracerFlag`）。

## 验收

1. `GOWORK=off go vet ./...`（在 ginext/v3 内）clean
2. `GOWORK=off go test ./... -cover` 通过，覆盖 > 70%
3. `go.work` 列入 v3 三个模块，仓库根 `go build ./...` 通过
4. example 可 `go run ./example`（监听 :8080，允许连接失败但编译/运行不 panic）
5. 无 `metadata` 子包、无 `init()` 装配、无全局可变 `jwtKey`、无 bit-mask 角色判定残留
6. memory：完成后写 `ginext/v3 rewrite` 记忆（删 metadata、Auth 结构、Interceptor 函数链、ResponseStatus、ParseFlag）

## 不做

- 不迁移 v2 消费方（仅 example 自身用 v3）
- 不改 v2 任何文件
- 不改 `errcode` API（仅注释中文化）
- 不引入新的外部依赖（仅 v3 四件套 + gin + jwt + testify）
