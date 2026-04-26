# ginext

Gin 框架扩展。请求绑定、JWT 鉴权、RPC 拦截器、结构化错误响应。

```go
import "github.com/tenz-io/gokit/ginext/v2"
```

## 快速开始

```go
r := gin.Default()
r.POST("/api/user", ginext.RpcHandler(func(ctx context.Context, req *CreateUserReq) (*CreateUserResp, error) {
    return &CreateUserResp{ID: 1}, nil
}))
r.Run(":8080")
```
