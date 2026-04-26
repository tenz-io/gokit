# annotation

基于 struct tag 的 Go 结构体注解库。提供默认值注入、请求字段绑定及声明式校验。

```go
import "github.com/tenz-io/gokit/annotation/v2"
```

## 快速开始

```go
type Config struct {
    Host string `default:"localhost"`
    Port int    `default:"8080" validate:"required,gt=0,lte=65535"`
}
cfg := &Config{}
annotation.ParseDefault(cfg)
annotation.ValidateStruct(cfg)
```
