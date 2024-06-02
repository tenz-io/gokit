package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext"
	"github.com/tenz-io/gokit/logger"
)

func init() {
	logger.ConfigureWithOpts(
		logger.WithLoggerLevel(logger.DebugLevel),
		logger.WithConsoleEnabled(true),
		logger.WithFileEnabled(true),
		logger.WithCallerEnabled(true),
		logger.WithCallerSkip(1),
	)

	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficEnabled(true),
	)
}

func main() {
	interceptor := ginext.NewInterceptorWithOpts(
		ginext.WithTracking(true),
		ginext.WithTraffic(true),
		ginext.WithMetrics(true),
		ginext.WithEnableAccessLog(true),
		ginext.WithTimeout(20*time.Millisecond),
	)

	engine := gin.New()
	interceptor.Apply(engine)

	engine.POST("/user/:id", func(c *gin.Context) {
		req := RestRequestEntity{}
		err := ginext.ShouldBind(c, &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, req)
	})

	engine.PUT("/search", func(c *gin.Context) {
		req := FileRequestEntity{}
		if err := ginext.ShouldBind(c, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"size": len(req.Image),
			"auth": req.Auth,
			"bbox": req.Bbox,
		})
	})

	log.Println("server is running on :8080")
	err := engine.Run(":8080")
	if err != nil {
		log.Fatal(err)
		return
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
