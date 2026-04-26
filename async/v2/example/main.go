package main

import (
	"context"
	"log"
	"time"

	"github.com/tenz-io/gokit/async/v2"
)

func main() {
	// AllOf: collect all results in order
	fns := []async.Fn[string]{
		func(ctx context.Context) (string, error) {
			time.Sleep(50 * time.Millisecond)
			return "hello", nil
		},
		func(ctx context.Context) (string, error) {
			return "world", nil
		},
	}
	results := async.AllOf(context.Background(), fns)
	for _, r := range results {
		log.Printf("result: %v, err: %v", r.Val, r.Err)
	}
}
