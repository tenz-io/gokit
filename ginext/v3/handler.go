package ginext

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext/v3/errcode"
	"github.com/tenz-io/gokit/logger/v3"
	"github.com/tenz-io/gokit/monitor/v3"
	"github.com/tenz-io/gokit/tracer/v3"
)

// RpcHandler 是 RPC 处理函数:给定 context 与请求,返回响应与 error。
// 它是传输无关的 —— gin context 在请求边界处由 [ChainContext] 折叠进
// context,因此拦截器与 handler 都只见 context.Context。
type RpcHandler func(ctx context.Context, req any) (resp any, err error)

// Interceptor 是一个包裹 [RpcHandler] 的函数。它先于 next 运行自己的
// 前置/后置逻辑,再调用 next(ctx, req) 并可改写其返回。作为函数类型
// (而非接口),它比 v2 的 RpcInterceptor 接口 + 多个 struct 更简洁,
// 链编排就是一个普通的 reduce。
type Interceptor func(ctx context.Context, req any, next RpcHandler) (resp any, err error)

// Chain 是一组有序的 [Interceptor]。[Intercept] 从外到内依次包裹
// handler,使最外层的拦截器最先进入、最后退出。
type Chain struct {
	interceptors []Interceptor
}

// NewChain 用显式给定的拦截器构造一条链。顺序即外→内顺序:第一个
// 拦截器最外(最先进入、最后退出),最后一个最内。
func NewChain(interceptors ...Interceptor) Chain {
	return Chain{interceptors: append([]Interceptor(nil), interceptors...)}
}

// DefaultChain 返回 ginext 内置的常用链(外→内):
//
//	Tracer → Metrics → Traffic → SlowLog → PanicRecover → handler
//
// PanicRecover 位于**最内层**(紧贴 handler):当 handler panic 时,恢复器
// 先把 panic 转成一个 500 [errcode.Error] 返回,再向外展开到 SlowLog/
// Traffic/Metrics 的 defer —— 此时它们观察到的 err 已非 nil,从而把
// 这次调用记为失败(而非误记为成功)。这与"先恢复再观测"的直觉相反,但
// 正是这个顺序让成功率/流量/慢日志在 panic 时反映真实结果。
//
// Tracer 在最外层,确保 request ID/flag/logger 已注入,供内层观测复用。
// 可按需用 [NewChain] 自行组合子集。
func DefaultChain() Chain {
	return NewChain(
		TracerInterceptor(),
		MetricsInterceptor(),
		TrafficInterceptor(),
		SlowLogInterceptor(0), // 0 取默认阈值
		PanicRecoverInterceptor(),
	)
}

// With 返回一条追加了给定拦截器的新链(原链不变)。追加的拦截器位于
// 更内层。
func (c Chain) With(interceptors ...Interceptor) Chain {
	return NewChain(append(append([]Interceptor{}, c.interceptors...), interceptors...)...)
}

// Intercept 依次用链中的拦截器从外到内包裹 handler 并执行,返回其结果。
// 空链直接调用 handler。
func (c Chain) Intercept(ctx context.Context, req any, handler RpcHandler) (any, error) {
	h := handler
	// 从内到外包裹:最后一个拦截器最先直接包到 handler 上,使它成为最内层;
	// 第一个拦截器最终位于最外层。
	for i := len(c.interceptors) - 1; i >= 0; i-- {
		icp := c.interceptors[i]
		next := h
		h = func(ctx context.Context, req any) (any, error) {
			return icp(ctx, req, next)
		}
	}
	return h(ctx, req)
}

// ChainContext 把一个 gin 请求的请求边界信息折叠进 context,供
// [Chain.Intercept] 使用。它确保 request ID(从 X-Request-Id 头或新生成)、
// request flag(从 X-Request-Flag 头经 [tracer.ParseFlag] 解析)与单飞
// [monitor.Exporter] 都已注入,并绑定了带 request_id 的 logger。这一步
// 在请求边界做一次,使下游所有拦截器与 handler 都能从 ctx 读到。
//
// cmd 是这条 RPC 的逻辑名(通常取路由路径或服务方法),用于 metrics
// 与 traffic 记录的标签。
func ChainContext(c *gin.Context, cmd string) context.Context {
	ctx := c.Request.Context()

	// request ID:信任客户端传入的 X-Request-Id 头,缺省则生成并回写头。
	reqID := c.GetHeader("X-Request-Id")
	if reqID == "" {
		reqID = tracer.RequestIDFromCtxOr(ctx)
	}
	if reqID == "" {
		ctx, reqID = tracer.EnsureRequestID(ctx)
	}
	ctx = tracer.WithRequestID(ctx, reqID)
	c.Writer.Header().Set("X-Request-Id", reqID)

	// request flag:解析 X-Request-Flag(lenient,未知 token 忽略)。
	flag := tracer.ParseFlag(c.GetHeader("X-Request-Flag"))
	ctx = tracer.WithFlag(ctx, flag)

	// logger:绑定一个带 request_id 的 entry,供下游复用。
	ctx = logger.WithLogger(ctx,
		logger.FromContext(ctx).WithRequestID(reqID).With(
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"cmd", cmd,
			"flag", flag.String(),
		))

	// monitor:请求边界单飞 Exporter,使下游 Begin 共享同一个。
	ctx = monitor.Init(ctx, cmd)

	// cmd 通过 ctx key 透传,供 SlowLog/Traffic 复用。
	ctx = withCmd(ctx, cmd)

	return ctx
}

// cmdCtxKey 携带 RPC 的逻辑名,供拦截器从 ctx 读取而无需额外传参。
type cmdCtxKey struct{}

func withCmd(ctx context.Context, cmd string) context.Context {
	if ctx == nil || cmd == "" {
		return ctx
	}
	return context.WithValue(ctx, cmdCtxKey{}, cmd)
}

func cmdFromCtx(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if cmd, ok := ctx.Value(cmdCtxKey{}).(string); ok {
		return cmd
	}
	return ""
}

// TracerInterceptor 返回一个注入 request ID、flag 与带 request_id 的
// logger 的拦截器。它从 ctx 读 request ID(由 [ChainContext] 注入),
// 若缺失则 [tracer.EnsureRequestID] 补一个。它取代了 v2 的
// TracerRpcInterceptor,且不再依赖 metadata.MD。
func TracerInterceptor() Interceptor {
	return func(ctx context.Context, req any, next RpcHandler) (any, error) {
		ctx, _ = tracer.EnsureRequestID(ctx)
		if flag := tracer.FromContext(ctx); flag != tracer.FlagNone {
			ctx = tracer.WithFlag(ctx, flag)
		}
		if cmd := cmdFromCtx(ctx); cmd != "" {
			ctx = logger.WithLogger(ctx, logger.FromContext(ctx).With("cmd", cmd))
		}
		return next(ctx, req)
	}
}

// MetricsInterceptor 返回一个用 [monitor.Begin]/[EndWithError] 为单次
// 调用计时的拦截器。它假设 [ChainContext] 已在请求边界做过单飞
// [monitor.Init];若没有,Begin/End 退化为 no-op。它取代了 v2 的
// MetricsRpcInterceptor。
func MetricsInterceptor() Interceptor {
	return func(ctx context.Context, req any, next RpcHandler) (resp any, err error) {
		rec := monitor.Begin(ctx, cmdFromCtx(ctx))
		defer func() { rec.EndWithError(err) }()
		return next(ctx, req)
	}
}

// TrafficInterceptor 返回一个记录请求/响应流水的拦截器。它用
// [logger.Entry.StartTraffic] 开启一个 traffic span(nil-safe:traffic
// 未配置时 StartTraffic 返回 nil,End 为 no-op),并在结束时用响应与
// 错误码结束它。它取代了 v2 的 TrafficRpcInterceptor。
func TrafficInterceptor() Interceptor {
	return func(ctx context.Context, req any, next RpcHandler) (resp any, err error) {
		cmd := cmdFromCtx(ctx)
		rec := logger.FromContext(ctx).StartTraffic(cmd)
		if rec != nil {
			defer func() {
				if err != nil {
					rec.EndWithError(err, "cmd", cmd)
					return
				}
				rec.End(resp, getErrCode(err), "cmd", cmd)
			}()
		}
		return next(ctx, req)
	}
}

// SlowLogInterceptor 返回一个在耗时超过 threshold 时打印慢日志告警的
// 拦截器。threshold <= 0 时取默认值 5s。它取代了 v2 的 SlogLogRpcInterceptor。
func SlowLogInterceptor(threshold time.Duration) Interceptor {
	if threshold <= 0 {
		threshold = 5 * time.Second
	}
	return func(ctx context.Context, req any, next RpcHandler) (resp any, err error) {
		start := time.Now()
		defer func() {
			if dur := time.Since(start); dur > threshold {
				logger.FromContext(ctx).Warnw("slow log",
					"duration", dur.String(),
					"threshold", threshold.String(),
					"cmd", cmdFromCtx(ctx),
					"err_code", getErrCode(err),
				)
			}
		}()
		return next(ctx, req)
	}
}

// PanicRecoverInterceptor 返回一个在最外层兜住 panic 的拦截器。发生
// panic 时记录一条 error 日志(含堆栈),并把错误翻译成 500
// [errcode.InternalServer],使下游能像普通错误一样处理。它取代了 v2 的
// PanicRecoverRpcInterceptor。
func PanicRecoverInterceptor() Interceptor {
	return func(ctx context.Context, req any, next RpcHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.FromContext(ctx).Errorw("panic recovery",
					"panic", fmt.Sprintf("%v", r),
					"stack", string(debug.Stack()),
				)
				err = errcode.InternalServer(http.StatusInternalServerError, "panic")
			}
		}()
		return next(ctx, req)
	}
}

// getErrCode 把 err 映射为面向 metrics/traffic 的错误码字符串:nil → "0",
// [errcode.Error] → 其 Code,其余 → "1"。沿用 v2 的编码约定。
func getErrCode(err error) string {
	if err == nil {
		return "0"
	}
	if e := new(errcode.Error); errors.As(err, &e) {
		return fmt.Sprintf("%d", e.Code)
	}
	return "1"
}
