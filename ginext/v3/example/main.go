package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext/v3"
	"github.com/tenz-io/gokit/logger/v3"
)

func init() {
	logger.ConfigureWithOpts(
		logger.WithLevel(logger.DebugLevel),
		logger.WithConsole(true),
		logger.WithFilePath("log"),
		logger.WithCaller(true),
		logger.WithCallerSkip(1),
		logger.WithTraffic(true),
	)
}

func main() {
	engine := gin.New()

	engine.GET("/ping", func(c *gin.Context) {
		ginext.Response(c, gin.H{"message": "pong"})
	})

	// 演示 bind + 校验 + 201 响应。
	engine.POST("/user/:id", func(c *gin.Context) {
		req := RestRequestEntity{}
		if err := ginext.BindAndValidate(c, &req); err != nil {
			ginext.ErrorResponse(c, err)
			return
		}
		ginext.ResponseStatus(c, http.StatusCreated, gin.H{
			"id":     req.ID,
			"name":   req.Name,
			"limit":  req.Limit,
			"offset": req.Offset,
			"auth":   req.Auth,
		})
	})

	// 演示 multipart 文件上传绑定。
	engine.PUT("/search", func(c *gin.Context) {
		req := FileRequestEntity{}
		if err := ginext.BindAndValidate(c, &req); err != nil {
			ginext.ErrorResponse(c, err)
			return
		}
		ginext.Response(c, gin.H{
			"size": len(req.Image),
			"auth": req.Auth,
			"bbox": req.Bbox,
		})
	})

	// 演示 RPC 拦截器链(panic 恢复 + tracer + metrics + traffic + slowlog)。
	engine.POST("/rpc", func(c *gin.Context) {
		handler := ginext.RpcHandler(func(ctx context.Context, req any) (any, error) {
			logger.FromContext(ctx).Infof("handle rpc request")
			return gin.H{"hello": "world"}, nil
		})
		ctx := ginext.ChainContext(c, "rpc")
		resp, err := ginext.DefaultChain().Intercept(ctx, nil, handler)
		if err != nil {
			ginext.ErrorResponse(c, err)
			return
		}
		ginext.Response(c, resp)
	})

	log.Println("server is running on :8080")
	if err := engine.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

type RestRequestEntity struct {
	Name    string `json:"name" validate:"required,min_len=1,max_len=20,pattern=^[a-zA-Z]+$"`
	ID      int64  `json:"id" bind:"uri,name=id"`
	Offset  int32  `json:"offset" bind:"query,name=offset" validate:"required,gte=0"`
	Limit   int32  `json:"limit" bind:"query,name=limit" validate:"required,gt=0,lte=100" default:"10"`
	Auth    string `json:"auth" bind:"header,name=Authorization" validate:"required"`
	Profile string `json:"profile" bind:"form,name=profile" validate:"max_len=100"`
}

type FileRequestEntity struct {
	Auth  string `bind:"header,name=Authorization" validate:"required"`
	Image []byte `bind:"file,name=image" validate:"required,max_len=204800"`
	// Bbox format: x1,y1,x2,y2
	// x1,y1 is the top-left corner, and x2,y2 is the bottom-right corner
	// positive integer
	Bbox string `bind:"form,name=bbox" validate:"pattern=^[0-9]{8}$"`
}
