# annotation

基于 struct tag 的结构体工具库，提供**声明式默认值注入**与**可插拔的校验引擎**。一次 struct 反射产出缓存的字段计划，默认值注入、校验、以及外部绑定层（如 `ginext`）复用同一份计划，避免每请求重复反射。

```go
import "github.com/tenz-io/gokit/annotation/v3"
```

## 模块介绍

annotation 解决三类问题：

- **默认值注入**（`ApplyDefaults`）：读取字段的 `default` tag，给零值字段填入默认值。已设值的字段保留不动；指针字段按需分配。
- **声明式校验**（`Validate`）：读取字段的 `validate` tag，运行规则并**收集全部错误**一次性返回（而非遇到第一个就停），错误带点路径（如 `Config.Addr.Street`）。
- **可插拔规则**（`Register`）：内置常用规则，也允许运行时注册业务自定义规则。规则参数在字段计划构建期就解析成闭包（编译好的正则、数值阈值），运行期零解析。

核心能力：

- 默认值：标量、`*string`/`*int` 等指针字段、嵌套 struct、`time.Duration`、`[]byte`
- 校验：数值/长度比较、字符串格式、slice 元素逐个校验（dive）、枚举、自定义消息
- 错误模型：`ValidationErrors`（多条 `FieldError`），支持 `errors.As` 提取
- 配置错误暴露：`gt=abc` 这类坏参数在构建期就报错，而非静默当成 `gt=0`

## 快速开始

```go
package main

import (
	"errors"
	"log"

	"github.com/tenz-io/gokit/annotation/v3"
)

type Config struct {
	Host string `default:"localhost" validate:"required"`
	Port int    `default:"8080"      validate:"required,gt=0,lte=65535"`
}

func main() {
	cfg := &Config{Host: "override.example"}
	if err := annotation.ApplyDefaults(cfg); err != nil { // Port 取默认 8080
		log.Fatal(err)
	}

	if err := annotation.Validate(cfg); err != nil {
		var verrs annotation.ValidationErrors
		if errors.As(err, &verrs) {
			for _, e := range verrs {
				log.Println(e) // 点路径 + 规则 + 消息
			}
		}
	}
}
```

默认值注入的语义：

- 零值字段（空字符串、nil 指针、0 数值）填入 `default`；**已设值字段保留不动**
- 指针字段（如 `*string`）有 `default` 时自动分配并设值
- 嵌套 struct 的指针：只要后裔某处带 `default`，指针会被自动分配以便注入后裔默认值

## 规则说明

规则写在 `validate` tag 里，多条用逗号分隔，参数用 `=`（或 `:`）赋值：

```go
type S struct {
	Age   int      `validate:"required,gte=0,lte=150"`
	Email string   `validate:"required,email"`
	Tags  []string `validate:"min_len=1,dive:non_blank"`
	Level string   `validate:"oneof=debug release test,msg=Level invalid"`
}
```

### 存在性

| 规则 | 含义 |
|---|---|
| `required` | 字段非空。空字符串/空 slice/空 map/nil 指针/`false`/零值 `time.Time` 视为缺失。**注意：整数 0 不算缺失**（强调"存在"而非"非零"）。 |

### 数值比较（int/uint/float）

| 规则 | 含义 |
|---|---|
| `gt=N` | 大于 N（`>`） |
| `gte=N` | 大于等于 N（`>=`） |
| `lt=N` | 小于 N（`<`） |
| `lte=N` | 小于等于 N（`<=`） |
| `min=N` | 大于等于 N（`gte` 的别名，`>=`） |
| `max=N` | 小于等于 N（`lte` 的别名，`<=`） |
| `eq=N` | 等于 N（数值或字符串） |
| `ne=N` | 不等于 N（数值或字符串） |

> 用在 slice/array 上时，比较作用于**每一个元素**，任一元素不满足即报错（错误消息带元素下标）。

### 长度比较（string/slice/array/map）

| 规则 | 含义 |
|---|---|
| `len=N` | 长度恰好等于 N |
| `min_len=N` | 长度大于等于 N |
| `max_len=N` | 长度小于等于 N |

> 注意区分：`min/max` 是**数值**比较，`min_len/max_len` 是**长度**比较，两者作用于不同对象。

### 字符串格式

| 规则 | 含义 |
|---|---|
| `non_blank` | 去除首尾空格后非空。用在 `[]string` 上时检查每个元素。 |
| `pattern=<re>` | 匹配正则（编译后缓存，可复用） |
| `pattern=#name` | 使用预定义命名模式（见下表） |
| `email` | 邮箱格式 |
| `url` | URL 格式 |
| `uuid` | UUID 格式 |
| `ipv4` / `ipv6` | IP 地址格式 |
| `alpha` | 仅字母 `[a-zA-Z]` |
| `alphanum` | 字母数字 `[a-zA-Z0-9]` |
| `numeric` | 仅数字 |
| `hex` | 十六进制 `[0-9a-fA-F]` |
| `date` | 日期 `YYYY-MM-DD` |
| `base64` | Base64 字符集 |

预定义命名模式也可用 `pattern=#email` 形式引用；还保留了 v2 兼容别名 `#abc`(=alpha)、`#abc123`(=alphanum)、`#digits`/`#123`(=numeric)。

### 子串匹配（字符串）

| 规则 | 含义 |
|---|---|
| `contains=s` | 包含子串 s |
| `prefix=s` | 以 s 开头 |
| `suffix=s` | 以 s 结尾 |

### 枚举

| 规则 | 含义 |
|---|---|
| `oneof=a b c` | 取值在空格分隔的列表中之一。`in` 是其别名。 |

### slice/map 元素逐个校验

| 规则 | 含义 |
|---|---|
| `dive:R` | 对 slice/array 的每个元素、map 的每个值套用规则 R。R 可为任意已注册规则，如 `dive:non_blank`、`dive:gt=0`。 |

例如 `Tags []string validate:"min_len=1,dive:non_blank"`：slice 至少 1 个元素，且每个元素去除空格后非空。

### 消息覆盖

| 修饰符 | 含义 |
|---|---|
| `msg=文本` | 覆盖**上一条**规则的失败消息。可含中文/空格，写在该规则之后。 |

```go
Age int `validate:"gte=0,lte=150,msg=Age must be between 0 and 150"`
```
上例中 `lte` 失败时消息为"Age must be between 0 and 150"。

### 自定义规则

通过 `Register` 运行时注册业务规则：

```go
annotation.Register("even", func(_ string, _ reflect.StructField) (annotation.Rule, error) {
	return func(rv reflect.Value) (bool, string) {
		if rv.Kind() == reflect.Int && rv.Int()%2 != 0 {
			return false, "must be even"
		}
		return true, ""
	}, nil
})

type Counter struct {
	N int `validate:"required,even"`
}
```

返回的闭包已捕获参数，运行期不再解析 tag。注册后即可像内置规则一样在 `validate` tag 中使用。

## 错误处理

校验有两个入口：

- **`Validate(ptr)`** — 收集**全部**失败，返回 `ValidationErrors`（多条 `FieldError`）。适合需要一次性向用户报告所有问题的场景（如表单校验）。
- **`QuickValidate(ptr)`** — 遇**第一个**失败即返回，返回单条 `FieldError`。适合只关心"是否合法 + 一个原因"的场景（如快速失败、内部断言）。

两者都返回 `nil` 表示通过。

```go
// 全部错误
err := annotation.Validate(&req)
var verrs annotation.ValidationErrors
if errors.As(err, &verrs) {
	for _, e := range verrs {
		// e.Field  点路径，如 "User.Address.City"
		// e.Rule   规则名，如 "required"、"gt"
		// e.Param  规则参数，如 "0"
		// e.Msg    失败消息
	}
}

// 第一个错误
err = annotation.QuickValidate(&req)
var fe annotation.FieldError
if errors.As(err, &fe) {
	// fe.Field / fe.Rule / fe.Msg
}
```

绑定层（如 `ginext`）用 `annotation.Err(field, rule, msg)` 构造单条错误，`annotation.AsErrors(err)` 从 error 中提取收集到的失败项。
