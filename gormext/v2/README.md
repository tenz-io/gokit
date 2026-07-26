# gormext

GORM 插件：通过 GORM 回调注入追踪、流量日志、Prometheus 指标、错误日志和慢查询检测。

## 功能特性

- 通过 `Tracker.Apply` 一次性为 `Query`/`Create`/`Update`/`Delete`/`Row` 注册前后置回调，无需手动埋点
- `WithTraffic` 开启流量日志，记录每条 SQL 及其变量，落到 `logger` 的 TrafficRec 中；当 `tracer` 处于 Debug 模式时即使未开启也会自动记录
- `WithMetrics` 开启后基于 `monitor.BeginRecord`/`EndWithError` 按命令类型（`db_query`/`db_create`/`db_update`/`db_delete`/`db_row`）采集 Prometheus 指标
- `WithErrorLog` 开启后对非 `gorm.ErrRecordNotFound` 的数据库错误输出 Error 级别日志，记录未找到则降级为 Debug
- `WithSlowLogFloor` 设置慢查询阈值，超过该耗时的 SQL 会带上 SQL 语句、参数和耗时输出 Warn 日志
- `NewTracker`/`NewTrackerWithOpts` 两种构造方式，支持传入完整 `Config` 或链式 Option 组合

## 能力清单

| 能力 | 含义 |
| --- | --- |
| 一键接入 GORM 回调 | `Tracker.Apply(db)` 在 `Query`/`Create`/`Update`/`Delete`/`Row` 五类操作上统一注册前后置回调，无需在业务代码里逐个埋点 |
| 记录 SQL 流量日志 | 开启 `EnableTraffic` 后，每条 SQL 及其绑定变量、耗时、结果码会作为一条 TrafficRec 落到 `logger` 中，用于排查具体请求执行了哪些 SQL |
| Debug 模式下自动补记流量日志 | 即使未显式开启 `EnableTraffic`，只要 `tracer.FromContext(ctx).IsDebug()` 为真，也会自动记录流量日志，便于线上临时开启调试排障 |
| 采集 Prometheus 指标 | 开启 `EnableMetrics` 后，按命令类型（`db_query`/`db_create`/`db_update`/`db_delete`/`db_row`）调用 `monitor.BeginRecord`/`EndWithError`，统计各类数据库操作的耗时和成功率 |
| 区分记录数据库错误日志 | 开启 `EnableErrorLog` 后，非 `gorm.ErrRecordNotFound` 的错误输出 Error 级别日志，记录未找到则降级为 Debug，避免正常的"查无此记录"淹没真正的错误 |
| 检测并告警慢查询 | 通过 `SlowLogFloor` 设置耗时阈值，超过阈值的 SQL 会带上语句、参数和实际耗时输出 Warn 日志，用于发现慢 SQL |
| 灵活的配置方式 | 提供 `NewTracker(Config)` 直接传入完整配置，或 `NewTrackerWithOpts` 基于默认配置叠加 `With*` Option 链式组合，适配不同初始化风格 |

## 快速开始

```go
import "github.com/tenz-io/gokit/gormext/v2"

tracker := gormext.NewTrackerWithOpts(
    gormext.WithTraffic(true),
    gormext.WithMetrics(true),
    gormext.WithErrorLog(true),
    gormext.WithSlowLogFloor(500*time.Millisecond),
)

if err := tracker.Apply(db); err != nil {
    log.Fatalf("apply gorm tracker error: %v", err)
}

// 之后正常使用 db 执行查询，回调会自动记录追踪、日志与指标
var user User
db.WithContext(ctx).First(&user, 1)
```

## API 速查

| 符号 | 说明 |
| --- | --- |
| `type Tracker` | 接口，`Apply(db *gorm.DB) error` 将回调注册到指定的 `*gorm.DB` 实例 |
| `NewTracker(config Config) Tracker` | 使用完整 `Config` 创建 Tracker |
| `NewTrackerWithOpts(opts ...ConfigOption) Tracker` | 基于默认配置叠加 Option 创建 Tracker |
| `type Config` | 配置项：`EnableTraffic`、`EnableMetrics`、`EnableErrorLog`、`SlowLogFloor` |
| `type ConfigOption` | `func(*Config)`，用于函数式配置 |
| `WithTraffic(enable bool) ConfigOption` | 开关流量日志记录 |
| `WithMetrics(enable bool) ConfigOption` | 开关 Prometheus 指标采集 |
| `WithErrorLog(enable bool) ConfigOption` | 开关数据库错误日志 |
| `WithSlowLogFloor(floor time.Duration) ConfigOption` | 设置慢查询日志阈值 |

引入路径：`import "github.com/tenz-io/gokit/gormext/v2"`
