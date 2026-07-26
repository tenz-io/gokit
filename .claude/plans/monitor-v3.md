# monitor V3 重写计划

## 目标

完全重写 monitor 模块为 V3，专注**指标采集**，采用**单飞注入式**（single-flight injection）模式。
- 不考虑向后兼容（V2 保留不动，其他模块引用 v2 不改、可编译报错不管）。
- 修改范围仅限 `monitor/` 目录新增 `monitor/v3/`。

## 三个已确认的设计决策

1. **single-flight 模式 = 单飞注入式**：请求入口 `Init(ctx, cmd)` 创建/复用一个 cmd 作用域的导出器并注入 ctx；下游 `Begin(ctx, "x")` 复用同一份导出器；`End` 同步聚批写一次。
2. **Registry 可注入 + 默认兜底**：`Init(opts...)` 启动期注入自定义 Registry / 命名空间；未调用则用 `prometheus.DefaultRegisterer` 兜底；测试可隔离。
3. **End 同步写 + 聚批**：去掉 v2 的 `go func{}`，Begin/End 同步维护 active gauge 的 Inc/Dec 配对，保证有序。

## 对 v2 的优化点（重写依据）

| # | v2 问题 | V3 做法 |
|---|---|---|
| 1 | `init()` 全局强制注册 4 个指标族，重复导入即 panic | 启动期 `Init(opts...)` 显式注册到可注入 Registry；首次 `Init` 注册，后续幂等；未 `Init` 兜底默认 Registry |
| 2 | `EndWithCodeOpt` 用 `go func{}` 异步写，活跃数 gauge 有序性错配、goroutine 堆积 | 同步写 counter/histogram/active-gauge（prometheus 本身原子操作），Begin=Inc(active)，End=counter+histogram+Dec(active) |
| 3 | Histogram 缺 `opt` 标签、`Sample`/`Observe` 各自做 code 归一，标签基数策略不一致 | 统一标签集 `{cmd, dsCmd, code, opt}`；code 归一逻辑集中在一处；提供 label 规范化 |
| 4 | 每个 API 重复 `if opt=="" {opt=NA}` + 手搓 labels map，每次分配 | 统一 label 规范化函数 + 复用；`opt==""` → `NA`，code 空 → `ok`/`err` 归一在 recorder 层 |
| 5 | "SingleFlight" 命名混乱（既是接口名又是内部概念） | 明确分层：`Exporter`（cmd 作用域指标导出器）/ `Recorder`（单次调用计时器）/ `singleFlight`（飞行器本身，包内）。对外清晰 |
| 6 | 无 Registry 隔离，测试难 | `WithRegistry(r)`，测试用独立 Registry |
| 7 | 无配置入口，namespace/subsystem 写死 | `Init(WithNamespace(...), WithSubsystem(...), WithRegistry(...))` |
| 8 | 测试缺失 | 补全单元测试（覆盖 Init 幂等、Begin/End 配对、code 归一、context 注入/复用、Registry 隔离） |

## 文件结构

```
monitor/v3/
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── monitor.go          # 包文档 + 配置(Init/Options) + Registry 注册
├── exporter.go        # Exporter（cmd 作用域指标导出器）+ 指标族构建
├── recorder.go         # Recorder（单次调用计时器）+ Begin/End
├── context.go         # ctx 注入/读取/复用（WithMonitor/FromContext/HasMonitor/CopyToContext）
├── labels.go          # 标签集 + code/opt 归一化 + 延迟/分桶常量
├── monitor_test.go     # 主测试（Init 幂等、Begin/End、active 配对）
├── context_test.go    # ctx 注入/复用/拷贝
├── labels_test.go      # 标签归一化
└── example/
    ├── go.mod
    ├── main.go         # 单飞注入式端到端示例
    └── go.sum
```

## 核心类型与 API

### labels.go
```go
const (
    defaultNamespace = "gokit"
    defaultSubsystem = "flight"
    labelCmd, labelDsCmd, labelCode, labelOpt = "cmd", "dsCmd", "code", "opt"
    codeOK   = "0"
    codeErr  = "1"
    valNA    = "NA"
)
var (
    latencyBuckets = []float64{ /* 0.1ms..10s，沿用 v2 */ }
    summaryObjectives = map[float64]float64{0.5:0.05, 0.9:0.01, 0.95:0.05, 0.99:0.001}
)
// normalizeOpt: "" → "NA"
// normalizeCode: code!="" && code!="0" → "1"，"" → "0"
// labelsOf(cmd, dsCmd, code, opt) prometheus.Labels
```

### monitor.go（配置 + Registry）
```go
// Init 在启动期调用一次，注入 Registry/Namespace/Subsystem。
// 后续调用幂等（仅首参生效），保证不重复注册。
func Init(opts ...Option)
type Option func(*config)
func WithNamespace(ns string) Option
func WithSubsystem(sub string) Option
func WithRegistry(r prometheus.Registerer) Option

// 内部：pkg 级 metrics 持有 counter/gauge/histogram/summary，首次 Init 时注册。
```

### exporter.go
```go
// Exporter 是 cmd 作用域的指标导出器（一个 cmd 一个实例）。
type Exporter struct{ cmd string }
// Set / Incr / Decr（gauge）；Count / CountDelta（counter）；
// Observe（histogram，毫秒，延迟）；Sample（summary，数据量）
// 全部走 normalizeXxx + labelsOf，同步写。

// NewExporter(cmd string) *Exporter —— cmd 为空 → "NA"
```

### recorder.go
```go
// Recorder 单次调用计时器。Begin→active++，End→counter+histogram+active--。
type Recorder struct {
    exp     *Exporter
    dsCmd   string
    start   time.Time
    // 配对防护：ended bool，防 End 重复调用导致 active 多 Dec
}
func (r *Recorder) End()
func (r *Recorder) EndWithCode(code string)
func (r *Recorder) EndWithOpt(opt string)
func (r *Recorder) EndWithError(err error)
func (r *Recorder) EndWithErrorOpt(err error, opt string)
func (r *Recorder) EndWithCodeOpt(code, opt string)  // 终点：同步写
// Begin(ctx, dsCmd) *Recorder —— 从 ctx 取 Exporter，开始记录
```

### context.go
```go
// 单飞注入式核心：
func Init(ctx context.Context, cmd string) context.Context   // 创建/复用 Exporter 注入 ctx
func WithExporter(ctx, *Exporter) context.Context
func FromContext(ctx) *Exporter          // 不存在返回 nil（调用方判空）或返回 no-op Exporter
func HasExporter(ctx) bool
func CopyToContext(src, dst) context.Context
```

**关键：`FromContext` 返回 no-op Exporter（非 nil），调用方无需判空**（沿用 v2 体验）。

## example/main.go（单飞注入式端到端）

```go
func main() {
    monitor.Init()                                   // 默认兜底
    ctx := context.Background()
    ctx = monitor.Init(ctx, "userService")            // 单飞：创建+注入
    rec := monitor.Begin(ctx, "total")                 // 主记录器
    err := doGetUser(ctx)                             // 下游复用同一 ctx
    rec.EndWithError(err)
}
func doGetUser(ctx context.Context) error {
    rec := monitor.Begin(ctx, "getUser")              // 子记录器，复用单飞 Exporter
    defer rec.EndWithError(nil)
    return nil
}
```

## go.work / 脚本影响

- `go.work` 增加 `./monitor/v3` 和 `./monitor/v3/example` 两行（按字母序插入 monitor/v2 之后）。
- `scripts/version-check.sh` Check 2 当前只匹配 `annotation/v3`——可选扩展为通用 `*/v3`，但**本计划不动 release 脚本**（用户未要求，且范围仅限 monitor）。
- `README.md`/`README-ZH.md` 模块表里 monitor 行加 v3 指引——属文档，**可选**，本计划包含最小更新。

## 范围红线

- ✅ 新建 `monitor/v3/` 全部文件 + 测试 + example + Makefile + README
- ✅ 更新 `go.work`（加入 v3 模块）
- ❌ 不修改 `monitor/v2/`
- ❌ 不修改其他模块（cache/grpcext/httpext/ginext/gormext 仍 import monitor/v2，可编译报错不管）
- ❌ 不动 release/tag/version-check 脚本

## 验证

- `cd monitor/v3 && go vet ./... && go test ./... -cover -v`
- `cd monitor/v3/example && go run .`（端到端不 panic、输出打点路径）
- `cd monitor/v3 && gofmt -l *.go`（无输出）
