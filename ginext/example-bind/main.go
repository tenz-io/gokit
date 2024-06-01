package main

import (
	"log"
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

	engine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	engine.GET("/panic", func(c *gin.Context) {
		// test panic recovery
		panic("something went wrong")
	})

	log.Println("server is running on :8080")
	err := engine.Run(":8080")
	if err != nil {
		log.Fatal(err)
		return
	}

}
