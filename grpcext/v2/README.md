# grpcext

gRPC 一元/流式拦截器，为服务端和客户端提供统一的请求追踪、流量日志和指标采集能力。

## 功能特性

- `NewInterceptorWithOpts` / `NewInterceptor` 通过配置项组装出可直接注册到 gRPC 的拦截器集合
- 服务端一元拦截器自动注入请求 ID 与结构化日志上下文（基于 `logger`/`tracer`），无需业务代码手动埋点
- 可选的服务端/客户端流量日志（`WithServerTraffic`/`WithClientTraffic`），记录请求、响应及对端地址，并自动忽略 `password`、`state`、`sizeCache`、`unknownFields` 等字段
- 可选的服务端/客户端指标采集（`WithServerMetrics`/`WithClientMetrics`），基于 `monitor` 统计调用耗时与错误率
- `Interceptor` 接口统一暴露一元/流式、服务端/客户端四类拦截器的应用方法，可与业务已有拦截器链式组合（传入已有拦截器时会被包裹增强，传 `nil` 时使用默认透传实现）
- 所有增强能力均可通过 `Option`/`Config` 按需开启，默认全部关闭，不影响现有服务

## 快速开始

```go
import "github.com/tenz-io/gokit/grpcext/v2"
```

```go
package main

import (
	"google.golang.org/grpc"

	"github.com/tenz-io/gokit/grpcext/v2"
)

func main() {
	ic := grpcext.NewInterceptorWithOpts(
		grpcext.WithServerTraffic(true),
		grpcext.WithServerMetrics(true),
		grpcext.WithClientTraffic(true),
		grpcext.WithClientMetrics(true),
	)

	// 服务端：套上流量日志、指标采集与请求追踪
	unaryServer := ic.ApplyUnaryServerInterceptor(nil)
	server := grpc.NewServer(grpc.UnaryInterceptor(unaryServer))
	_ = server

	// 客户端：套上流量日志与指标采集
	unaryClient := ic.ApplyUnaryClientInterceptor(nil)
	conn, err := grpc.NewClient("localhost:9000", grpc.WithUnaryInterceptor(unaryClient))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
}
```

## API 速查

| 符号 | 说明 |
| --- | --- |
| `type Interceptor` | 拦截器组装接口，暴露一元/流式、服务端/客户端四种 `Apply*Interceptor` 方法 |
| `NewInterceptor(config Config) Interceptor` | 用显式 `Config` 创建拦截器组装器 |
| `NewInterceptorWithOpts(opts ...Option) Interceptor` | 用一组 `Option` 创建拦截器组装器，未设置项默认关闭 |
| `type Config` | 拦截器开关配置：`EnabledServerTraffic`/`EnabledClientTraffic`/`EnabledServerMetrics`/`EnabledClientMetrics` |
| `type Option` | 修改 `Config` 的函数式选项 |
| `WithServerTraffic(enabled bool) Option` | 开启/关闭服务端流量日志 |
| `WithClientTraffic(enabled bool) Option` | 开启/关闭客户端流量日志 |
| `WithServerMetrics(enabled bool) Option` | 开启/关闭服务端指标采集 |
| `WithClientMetrics(enabled bool) Option` | 开启/关闭客户端指标采集 |
| `Interceptor.ApplyUnaryServerInterceptor` | 在给定一元服务端拦截器基础上叠加追踪/流量/指标能力 |
| `Interceptor.ApplyStreamServerInterceptor` | 应用流式服务端拦截器（当前为透传/直接返回传入实现） |
| `Interceptor.ApplyUnaryClientInterceptor` | 在给定一元客户端拦截器基础上叠加流量/指标能力 |
| `Interceptor.ApplyStreamClientInterceptor` | 应用流式客户端拦截器（当前为透传/直接返回传入实现） |

导入路径：`github.com/tenz-io/gokit/grpcext/v2`
