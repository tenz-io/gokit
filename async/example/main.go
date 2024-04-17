package main

import (
	"context"
	"log"
	"time"

	"github.com/tenz-io/gokit/async"
)

func main() {
	job1 := async.NewJob(func(ctx context.Context) (int, error) {
		time.Sleep(100 * time.Millisecond)
		panic("oops")
	}, "job1 failed")

	job2 := async.NewJob(func(ctx context.Context) (string, error) {
		time.Sleep(120 * time.Millisecond)
		return "hello", nil
	}, "job1 failed")

	start := time.Now()
	async.Submit(context.Background(), job1, job2)
	duration := time.Since(start)
	log.Printf("duration: %v\n", duration)

	result1, err := job1.Result()
	log.Printf("job1 result: %v, err: %v", result1, err)

	result2, err := job2.Result()
	log.Printf("job2 result: %v, err: %v", result2, err)

}
