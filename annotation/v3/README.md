# annotation

struct-tag 驱动的工具：用声明式 `default`/`validate`/`bind` 标签描述结构体，一次反射构建可缓存的字段 `Plan`，供默认值填充、校验、请求绑定复用，避免每次请求重复反射。

## 功能特性

- `ApplyDefaults` 按 `default` 标签为零值字段填充默认值，不会为无默认值的可选指针分配内存。
- `Validate` 一次性收集所有字段的全部校验失败（而非遇到第一个错误就返回），失败信息带点号路径（如 `Config.Addr.Street`）。
- 内置常用规则：`required`、`min`/`max`/`gt`/`lt`/`gte`/`lte`、`len`/`min_len`/`max_len`、`oneof`（别名 `in`）、`email`/`url`/`uuid`/`ipv4`/`alpha`/`alphanum`/`date` 等命名正则、`pattern=<re>`、`contains`/`prefix`/`suffix`、`dive`（对切片/map 元素逐个校验）等。
- `Register` 支持注册自定义校验规则，配置错误（如非法参数）会在构建 Plan 时暴露，而不是被吞掉。
- `PlanFor` 缓存结构体类型的字段计划（路径、绑定来源/名称、编译后的规则），`Plan.Walk`/`Fields`/`FieldsBySource` 供上层（如 HTTP 绑定层）遍历复用。
- `SetString`/`Set` 提供从字符串或已具类型的值写入字段的通用 setter，支持指针穿透、`time.Duration`、`[]byte`。

## 快速开始

```go
import "github.com/tenz-io/gokit/annotation/v3"
```

```go
type Config struct {
    Host string `default:"localhost" validate:"required"`
    Port int    `default:"8080" validate:"required,gt=0,lte=65535"`
}

cfg := &Config{}
if err := annotation.ApplyDefaults(cfg); err != nil {
    // 处理默认值填充错误
}
if err := annotation.Validate(cfg); err != nil {
    if verrs, ok := annotation.AsErrors(err); ok {
        for _, fe := range verrs {
            fmt.Println(fe.Field, fe.Rule, fe.Msg)
        }
    }
}
```

## API 速查

| 符号 | 说明 |
| --- | --- |
| `ApplyDefaults(ptr any) error` | 为结构体零值字段填充 `default` 标签指定的默认值 |
| `Validate(ptr any) error` | 校验结构体全部字段，返回全部失败（`ValidationErrors`） |
| `AsErrors(err error) (ValidationErrors, bool)` | 从 `Validate` 返回的 error 中提取失败列表 |
| `PlanFor(ptr any) (*Plan, error)` | 构建/取缓存的结构体字段计划 |
| `Plan.Fields() []*Field` | 计划的顶层字段（按声明顺序） |
| `Plan.Walk(fn func(*Field) bool)` | 深度遍历计划中的全部字段（含嵌套） |
| `Plan.FieldsBySource(src BindSource) []*Field` | 按绑定来源筛选顶层字段，供绑定器使用 |
| `Field` | 单个字段的缓存信息：`Path`/`Name`/`Index`/`Type`/`BindName`/`BindSource`/`Default` 等 |
| `Field.IsRequired() bool` | 判断字段是否带 `required` 规则 |
| `BindSource` / `BindNone`、`BindURI`、`BindQuery`、`BindHeader`、`BindForm`、`BindFile` | 绑定来源类型及其取值 |
| `Validator` | 自定义规则的编译函数签名 |
| `Register(name string, v Validator)` | 注册/覆盖一个校验规则 |
| `FieldError` | 单条字段校验失败（`Field`/`Rule`/`Param`/`Msg`），实现 `error` |
| `ValidationErrors` | 校验失败的有序集合，实现 `error`，`Has()` 判断是否非空 |
| `NewFieldError(field, rule, param, msg string) FieldError` | 构造一条字段错误 |
| `Err(field, rule, msg string) ValidationErrors` | 构造单条 ad-hoc 校验失败 |
| `Errf(field, rule, format string, args ...any) ValidationErrors` | 带格式化消息的 `Err` |
| `SetString(rv reflect.Value, s string) error` | 将字符串写入字段（支持指针穿透、`time.Duration`、`[]byte`） |
| `Set(rv reflect.Value, v any) error` | 将已具类型的值写入字段（直接赋值或可转换时转换） |

引入路径：`github.com/tenz-io/gokit/annotation/v3`
