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
	})

	job2 := async.NewJob(func(ctx context.Context) (string, error) {
		time.Sleep(120 * time.Millisecond)
		return "hello", nil
	})

	start := time.Now()
	async.Wait(context.Background(), job1, job2)
	duration := time.Since(start)
	log.Printf("duration: %v\n", duration)

	log.Printf("job1 result: %v, err: %v", job1.Val, job1.Err)

	log.Printf("job2 result: %v, err: %v", job2.Val, job2.Err)

}
