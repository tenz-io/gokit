package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginterceptor"
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
		logger.WithTrafficConsoleEnabled(true),
		logger.WithTrafficFileEnabled(true),
	)
}

func main() {
	interceptor := ginterceptor.NewInterceptorWithOpts(
		ginterceptor.WithTracking(true),
		ginterceptor.WithTraffic(true),
		ginterceptor.WithMetrics(true),
		ginterceptor.WithEnableAccessLog(true),
		ginterceptor.WithTimeout(20*time.Millisecond),
	)

	engine := gin.New()
	interceptor.Apply(engine)

	engine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	log.Println("server is running on :8080")
	err := engine.Run(":8080")
	if err != nil {
		log.Fatal(err)
		return
	}

}
