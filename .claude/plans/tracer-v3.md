# tracer/v3 — 无兼容重写

> 范围：仅 `tracer/v3` 新模块（与 `tracer/v2` 并存，v2 不动）。本模块对外暴露的 API 变更会引起 ginext/grpcext/httpext/gormext/cache 等调用方编译错误——**本次不修复这些调用方**，按用户要求留到后续逐步迁移。

## 设计目标（对应"更好"四点）

| 用户诉求 | v3 做法 |
| --- | --- |
| 更好的表 | 数据驱动的 `flagTable` 注册表登记每个 flag 的位值/名字/别名；解析、字符串化、遍历都走表，新增 flag 只加一行 |
| 更好的写法 | 按 annotation/v3、logger/v3 的多文件布局拆分；包文档单独成 `doc.go`；gci 导入分组；无注释残留 |
| 更容易使用 | 把"字符串→Flag 解析 + Flag→字符串"这类调用方重复逻辑收进模块：`ParseFlag`/`Flag.String()`；ginext 的 `getTracerFlag` 以后可删 |
| 更好性能 | `Flag` 用 `int8`（小，可内联比较）；`String()` 走预计算表零分配；context key 用类型化空结构体（零碰撞、无字符串）；热路径不做 `strings.Split` 重复分配 |

## 模块布局

```
tracer/v3/
├── go.mod                 module github.com/tenz-io/gokit/tracer/v3 ; go 1.24 ; require google/uuid v1.6.0
├── go.sum
├── Makefile               复用 annotation/v3 的 Makefile（test/cover/vet/fmt/tidy/clean/run-example）
├── README.md              中文（对标 v2 README，更新 API 速查）
├── doc.go                 包文档（package tracer 注释 + 用法示例块）
├── flag.go                Flag 类型 + flagTable 注册表 + 位运算方法
├── parse.go               ParseFlag(string) Flag + (Flag) String() / Names() / Aliases()
├── context.go             flag 的 context 读写：WithFlag / WithFlags / FromContext
├── requestid.go           RequestID 的 context 读写 + NewRequestID 生成
├── tracer_test.go         核心测试（Flag 位运算、context 读写、requestid）
├── *_additional_test.go    拆分测试（parse / 零值 / nil-context 边界）
└── example/
    ├── go.mod             module tracer-v3-example ; replace => ./..
    └── main.go            运行示例（flag 组合、context 串联、requestid 兜底）
```

> 注意：example 模块名用 `tracer-v3-example`（与 `logger-v3-example` 一致），因为 v2 的 example 占用了 `tracer-example` 名（实际上 v2 没有 example，但仍按 v3 系列命名规范用 `tracer-v3-example` 以保持一致）。

## 关键 API 设计

### flag.go — 类型 + 注册表

```go
type Flag int8

const (
    FlagNone   Flag = 0
    FlagDebug  Flag = 1 << iota  // debug
    FlagStress                   // stress test
    FlagShadow                    // shadow traffic (record/replay)
)

type flagDef struct {
    flag    Flag
    name    string   // "debug"
    aliases []string // {"db"} 可选别名，解析时也认
}

// flagTable 是 flag 的单一事实来源；解析/字符串化/遍历都走它。
var flagTable = []flagDef{
    {FlagDebug,  "debug",  nil},
    {FlagStress, "stress", nil},
    {FlagShadow, "shadow", nil},
}

func (f Flag) Is(x Flag) bool        { return f&x == x && x != 0 } // FlagNone 视为恒真? 见下
func (f Flag) HasAll(x Flag) bool    { return f&x == x }
func (f Flag) IsDebug() bool         { return f.Is(FlagDebug) }
func (f Flag) IsStress() bool        { return f.Is(FlagStress) }
func (f Flag) IsShadow() bool        { return f.Is(FlagShadow) }
```

> `Is` 语义决定：v2 的 `Is` 对 `FlagNone` 返回 `(f&0)==0` 恒真。v3 保持 `Is` 只在 `x != 0` 时为"包含全部位"，`Is(FlagNone)` 返回 false（更直观，避免 `FlagNone` 恒真陷阱）。这点会进测试 + README 注明为 v3 行为差异。

### parse.go — 字符串 ↔ Flag（收进模块，替代 ginext getTracerFlag）

```go
// ParseFlag 把 "debug|shadow" 这类字符串解析成 Flag。
// 支持 '|' 分隔、忽略空白与未知项、空串返回 FlagNone。
// 大小写不敏感（trim + ToLower）。
func ParseFlag(s string) Flag

// String 返回 "debug|shadow" 形式；无标志返回 "none"。
// 走 flagTable 预计算，零分配拼接（用 strings.Builder，长度可预算）。
func (f Flag) String() string
```

### context.go — flag 的 context 透传

```go
type flagCtxKey struct{}              // 类型化 key，零碰撞
func WithFlag(ctx context.Context, flag Flag) context.Context
func WithFlags(ctx context.Context, flags ...Flag) context.Context  // 位或组合
func FromContext(ctx context.Context) Flag                            // nil-safe，返回 FlagNone
```

### requestid.go — RequestID 透传 + 生成

```go
type requestIDCtxKey struct{}
func WithRequestID(ctx context.Context, id string) context.Context
func RequestIDFromCtx(ctx context.Context) string      // 未设置则 NewRequestID 兜底
func RequestIDFromCtxOr(ctx context.Context) string    // 未设置返回 ""，不兜底
func NewRequestID() string                             // uuid 去短横线，可被调用方直接用
```

> 命名：v2 是 `WithRequestId`（驼峰 d 小写）。v3 统一为 `WithRequestID`（ID 大写，符合 Go 命名规范与 logger/v3 的 `WithRequestID` 一致）。这是 v3 显式行为差异，README 会注明。同理 `RequestIdFromCtx` → `RequestIDFromCtx`。

## 行为差异（v2 → v3，README "变更说明" 会列）

1. `Is(FlagNone)` 由"恒真"改为"返回 false"（更直观，避免 `FlagNone` 陷阱）。
2. 方法/函数命名 `RequestId` → `RequestID`（统一 ID 大写）。
3. 新增 `ParseFlag` / `Flag.String()` / `Flag.Names()` / `NewRequestID()`，收口调用方重复逻辑。
4. `Flag` 由 `int` 改 `int8`（位掩码 3 个 flag 远够用，更小）。
5. context key 由字符串常量改为类型化空结构体（无碰撞、无字符串比较）。

## 与调用方的关系（不在本次范围）

调用方当前用法与 v3 新 API 的对应（仅记录，**本次不改**）：

| 调用方代码 (v2) | v3 等价（后续迁移用） |
| --- | --- |
| `tracer.FromContext(ctx).IsDebug()` | 不变（签名一致） |
| `tracer.RequestIdFromCtx(ctx)` | `tracer.RequestIDFromCtx(ctx)` |
| `tracer.WithRequestId(ctx, id)` | `tracer.WithRequestID(ctx, id)` |
| `tracer.WithFlag(ctx, flag)` | 不变 |
| ginext `getTracerFlag(string)` | `tracer.ParseFlag(string)` |

## 实施步骤

1. 建目录 `tracer/v3/`，写 `go.mod`、`Makefile`。
2. 写 `doc.go`、`flag.go`、`parse.go`、`context.go`、`requestid.go`。
3. 写测试：`tracer_test.go`（核心）+ `parse_additional_test.go`（ParseFlag/String 边界）+ `context_additional_test.go`（nil ctx、组合、嵌套）。
4. 写 `example/main.go` + `example/go.mod`。
5. 写 `README.md`（中文，对标 v2 但反映新 API + 变更说明）。
6. `go.work` `use` 列表追加 `./tracer/v3` 与 `./tracer/v3/example`。
7. 在 `tracer/v3` 跑 `go mod tidy` → `go vet ./...` → `go test ./... -cover -v`；在 example 跑 `go run .`。
8. `gofmt -w *.go` 收尾。
9. 不动其他模块；预期的调用方编译错误不在本次处理。

## 验证标准

- [x] `tracer/v3` 内 `go test ./... -cover` 全绿，覆盖率 ≥ v2（v2 仅测位运算 + context 基本读写）。
- [x] `go vet ./...` 无告警。
- [x] `example` 可 `go run .` 运行并打印演示输出。
- [x] `go.work` 包含 v3 两个模块且 `go build ./tracer/v3/...` 成功。
- [x] README 的 API 速查表与导出符号一一对应（无文档漂移）。
