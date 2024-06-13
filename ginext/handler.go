package ginext

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/tenz-io/gokit/ginext/errcode"
	"github.com/tenz-io/gokit/ginext/metadata"
	"github.com/tenz-io/gokit/logger"
	"github.com/tenz-io/gokit/monitor"
	"github.com/tenz-io/gokit/tracer"
)

type (
	RpcHandler func(ctx context.Context, req any) (resp any, err error)
)

type RpcInterceptor interface {
	Intercept(ctx context.Context, req any, handler RpcHandler) (resp any, err error)
}

type (
	TracerRpcInterceptor struct {
		interceptor RpcInterceptor
	}
	MetricsRpcInterceptor struct {
		interceptor RpcInterceptor
	}
	TrafficRpcInterceptor struct {
		interceptor RpcInterceptor
	}
	SlogLogRpcInterceptor struct {
		interceptor RpcInterceptor
		threshold   time.Duration
	}
	PanicRecoverRpcInterceptor struct {
		interceptor RpcInterceptor
	}
	// handlerRpcInterceptor always used in the last interceptor
	handlerRpcInterceptor struct{}
)

type (
	NewRpcInterceptor func(interceptor RpcInterceptor) RpcInterceptor
)

var (
	NewRpcInterceptors = []NewRpcInterceptor{
		NewTracerRpcInterceptor,
		NewMetricsRpcInterceptor,
		NewTrafficRpcInterceptor,
		NewSlogLogRpcInterceptor,
		NewPanicRecoverRpcInterceptor,
	}
	AllRpcInterceptor = newHandlerRpcInterceptor()
)

func init() {
	for i := len(NewRpcInterceptors) - 1; i >= 0; i-- {
		AllRpcInterceptor = NewRpcInterceptors[i](AllRpcInterceptor)
	}
}

func NewTracerRpcInterceptor(interceptor RpcInterceptor) RpcInterceptor {
	return &TracerRpcInterceptor{interceptor: interceptor}
}

func (t *TracerRpcInterceptor) Intercept(ctx context.Context, req any, handler RpcHandler) (resp any, err error) {
	if t.interceptor == nil {
		return handler(ctx, req)
	}

	meta, ok := metadata.FromContext(ctx)
	if !ok || meta == nil {
		return t.interceptor.Intercept(ctx, req, handler)
	}

	// inject request id into context
	if meta.RequestID == "" {
		meta.RequestID = tracer.RequestIdFromCtx(ctx)
	}
	ctx = tracer.WithRequestId(ctx, meta.RequestID)

	// inject traffic flag into context
	flag := getTracerFlag(meta.RequestFlag)
	ctx = tracer.WithFlag(ctx, flag)

	// inject logger into context
	ctx = logger.WithLogger(ctx,
		logger.WithTracing(meta.RequestID).
			WithFields(logger.Fields{
				"url":          meta.Path,
				"entry_cmd":    meta.Cmd,
				"request_flag": meta.RequestFlag,
			}),
	)

	return t.interceptor.Intercept(ctx, req, handler)

}

func NewMetricsRpcInterceptor(interceptor RpcInterceptor) RpcInterceptor {
	return &MetricsRpcInterceptor{interceptor: interceptor}
}

func (m *MetricsRpcInterceptor) Intercept(ctx context.Context, req any, handler RpcHandler) (resp any, err error) {
	if m.interceptor == nil {
		return handler(ctx, req)
	}

	meta, ok := metadata.FromContext(ctx)
	if !ok || meta == nil {
		return m.interceptor.Intercept(ctx, req, handler)
	}

	ctx = monitor.InitSingleFlight(ctx, meta.Cmd)
	rec := monitor.BeginRecord(ctx, "total")
	defer func() {
		rec.EndWithCode(getErrCode(err))
	}()

	return m.interceptor.Intercept(ctx, req, handler)
}

func NewTrafficRpcInterceptor(interceptor RpcInterceptor) RpcInterceptor {
	return &TrafficRpcInterceptor{interceptor: interceptor}
}

func (t *TrafficRpcInterceptor) Intercept(ctx context.Context, req any, handler RpcHandler) (resp any, err error) {
	if t.interceptor == nil {
		return handler(ctx, req)
	}

	meta, ok := metadata.FromContext(ctx)
	if !ok || meta == nil {
		return t.interceptor.Intercept(ctx, req, handler)
	}

	var (
		cmd   = meta.Cmd
		reqID = meta.RequestID
	)

	// inject traffic logger into context
	ctx = logger.WithTrafficEntry(
		ctx,
		logger.WithTrafficTracing(ctx, reqID).
			WithFields(logger.Fields{
				"url": meta.Path,
			}).WithIgnores(
			"password",
			"Authorization",
		),
	)

	rec := logger.StartTrafficRec(ctx, &logger.ReqEntity{
		Typ: logger.TrafficTypRecv,
		Cmd: cmd,
		Req: req,
		Fields: logger.Fields{
			"method":     meta.Method,
			"client":     meta.ClientIP,
			"trace_flag": tracer.FromContext(ctx),
		},
	})

	defer func() {
		rec.End(&logger.RespEntity{
			Code: getErrCode(err),
			Msg:  getErrMsg(err),
			Resp: resp,
		}, logger.Fields{})
	}()

	return t.interceptor.Intercept(ctx, req, handler)

}

func NewSlogLogRpcInterceptor(interceptor RpcInterceptor) RpcInterceptor {
	return &SlogLogRpcInterceptor{
		interceptor: interceptor,
		threshold:   5 * time.Second,
	}
}

func (s *SlogLogRpcInterceptor) Intercept(ctx context.Context, req any, handler RpcHandler) (resp any, err error) {
	var (
		le    = logger.FromContext(ctx)
		start = time.Now()
	)

	meta, ok := metadata.FromContext(ctx)
	if !ok || meta == nil {
		// should not happen
		// add here for panic free
		meta = &metadata.MD{}
	}

	defer func() {
		if duration := time.Since(start); s.threshold > 0 && duration > s.threshold {
			le.WithFields(logger.Fields{
				"duration":  duration,
				"threshold": s.threshold,
				"method":    meta.Method,
				"client_ip": meta.ClientIP,
				"command":   meta.Cmd,
				"err_code":  getErrCode(err),
				"err_msg":   getErrMsg(err),
			}).Warn("slow log")
		}
	}()

	if s.interceptor == nil {
		return handler(ctx, req)
	}
	return s.interceptor.Intercept(ctx, req, handler)
}

func NewPanicRecoverRpcInterceptor(interceptor RpcInterceptor) RpcInterceptor {
	return &PanicRecoverRpcInterceptor{interceptor: interceptor}
}

func (p *PanicRecoverRpcInterceptor) Intercept(ctx context.Context, req any, handler RpcHandler) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic recovery: %s, stacktrace: %s\n", r, string(debug.Stack()))
			logger.FromContext(ctx).WithFields(logger.Fields{
				"panic": fmt.Sprintf("%s", r),
			}).Error("panic recovery")
			err = errcode.InternalServer(http.StatusInternalServerError, "panic")
		}
	}()

	if p.interceptor == nil {
		return handler(ctx, req)
	}

	return p.interceptor.Intercept(ctx, req, handler)
}

func newHandlerRpcInterceptor() RpcInterceptor {
	return &handlerRpcInterceptor{}
}

func (h *handlerRpcInterceptor) Intercept(ctx context.Context, req any, handler RpcHandler) (resp any, err error) {
	return handler(ctx, req)
}
