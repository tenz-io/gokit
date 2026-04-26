# grpcetcd

etcd 服务注册与 gRPC 客户端发现（轮询负载均衡）。

```go
import "github.com/tenz-io/gokit/grpcetcd/v2"
```

## 快速开始

```go
registry := grpcetcd.NewRegistry(etcdClient, "/services/myapp", 10, logger.L())
revoke, _ := registry.Register(ctx, "127.0.0.1:9090")
defer revoke()

discovery := grpcetcd.NewDiscovery(etcdClient, "/services/myapp", logger.L())
conn, _ := discovery.Dial(ctx)
```
