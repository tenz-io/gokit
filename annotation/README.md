# annotation

基于 struct tag 的 Go 结构体注解库。提供默认值注入、请求字段绑定及声明式校验。纯反射实现，零依赖。

## 快速开始

```go
import "github.com/tenz-io/gokit/annotation"

type Config struct {
    Host    string `default:"localhost"`
    Port    int    `default:"8080" validate:"required,gt=0,lte=65535"`
    Timeout int    `default:"30"`
}

cfg := &Config{}
// 注入默认值
if err := annotation.ParseDefault(cfg); err != nil {
    panic(err)
}
// 校验
if err := annotation.ValidateStruct(cfg); err != nil {
    panic(err) // 包含所有校验失败信息
}
```

## API 参考

### 注解类型

| 常量 | 标签名 | 用途 |
|------|--------|------|
| `Bind` | `bind` | 请求字段绑定：`bind:"uri,name=id"` |
| `Default` | `default` | 默认值：`default:"localhost"` |
| `Protobuf` | `protobuf` | protobuf 字段名 |
| `JSON` | `json` | JSON 字段名 |
| `YAML` | `yaml` | YAML 字段名 |
| `Validate` | `validate` | 校验规则 |

### 默认值

| 函数 | 签名 | 说明 |
|------|------|------|
| `ParseDefault` | `func ParseDefault(structPtr any) error` | 根据 `default` 标签设置字段默认值。支持递归初始化嵌套 struct 和指针字段。支持 string/int/uint/float/bool/[]byte 类型 |

### 校验

| 函数 | 签名 | 说明 |
|------|------|------|
| `ValidateStruct` | `func ValidateStruct(structPtr any) error` | 根据 `validate` 标签校验所有字段 |

校验规则：

| 规则 | 语法 | 适用类型 | 说明 |
|------|------|----------|------|
| 必填 | `required` | string, array, slice, map, ptr | 非空值 |
| 小于 | `lt=N` | 数值, []数值 | 每个元素均满足 |
| 小于等于 | `lte=N` | 数值, []数值 | |
| 大于 | `gt=N` | 数值, []数值 | |
| 大于等于 | `gte=N` | 数值, []数值 | |
| 精确长度 | `len=N` | string, slice | |
| 最小长度 | `min_len=N` | string, slice | |
| 最大长度 | `max_len=N` | string, slice | |
| 非空串 | `non_blank` | string, []string | 不能是纯空白 |
| 正则 | `pattern=RE` | string | 必须以 `^` 开头 `$` 结尾 |
| 预定义 | `pattern=#email` | string | `#email`, `#url`, `#abc`, `#123`, `#abc123`, `#hex`, `#base64`, `#date` |

组合示例：`validate:"required,gt=0,lte=300"`

### 请求字段绑定

| 函数 | 签名 | 说明 |
|------|------|------|
| `GetRequestFields` | `func GetRequestFields(structPtr any) RequestFields` | 提取所有可绑定字段及其绑定元信息 |
| `GetAnnotations` | `func GetAnnotations(field reflect.StructField) []Annotation` | 获取单个字段上的所有注解标签 |

`RequestField` 核心字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `FieldName` | string | 结构体字段名 |
| `TagName` | string | 标签名称（用于匹配请求参数名） |
| `IsRequired` | bool | 是否 mandatory |
| `IsUri` / `IsQuery` / `IsHeader` / `IsForm` / `IsFile` | bool | 绑定来源 |
| `Set(value)` | func | 写入值，自动类型转换 |
| `SetString(value)` | func | 写入字符串，自动类型转换 |

### 错误类型

| 类型 | 说明 |
|------|------|
| `ProtoError` | 协议错误（如字段类型与绑定类型不匹配） |
| `ValidationError` | 单字段校验错误 |
| `ValidationErrors` | 多字段校验错误集合 |

## 最佳实践

### 校验与默认值配合

先注入默认值，再校验——确保默认值本身符合校验规则：

```go
if err := annotation.ParseDefault(cfg); err != nil {
    return err
}
if err := annotation.ValidateStruct(cfg); err != nil {
    return fmt.Errorf("config validation: %w", err)
}
```

### 使用预定义正则

优先使用预定义 pattern 名称而非手写正则：

```go
type Form struct {
    Email string `validate:"required,pattern=#email"`
    URL   string `validate:"pattern=#url"`
}
```

### 自定义错误处理

校验返回的 `ValidationErrors` 支持聚合多个字段的错误，无需提前返回：

```go
var errs annotation.ValidationErrors
for _, result := range results {
    if err := result.Err; err != nil {
        var vErrs annotation.ValidationErrors
        if errors.As(err, &vErrs) {
            errs = append(errs, vErrs...)
        }
    }
}
```

## 与 v1 的区别

| 变更项 | v1 | v2 |
|--------|----|----|
| 正则编译 | 每次校验重新 `regexp.Compile` | 预编译（`precompiledPatterns`） |
| 预定义模式 | 字符串常量存储 | `*regexp.Regexp` 直接匹配 |
| `matchString` | 存在 | 重命名为 `matchPattern` |
| 数值比较 | 4 个独立函数 | 合并为 `cmpNumeric` |
| 错误分隔符 | `", "` | `"; "`（更利于日志解析） |
