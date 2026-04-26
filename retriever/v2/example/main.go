package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tenz-io/gokit/retriever/v2"
)

func main() {
	r := retriever.New(
		retriever.WithMaxAttempt(3),
		retriever.WithBackoff(retriever.NewLinearBackoff(50)),
	)

	var count int
	result, err := r.DoAlwaysRetry(context.Background(), func(ctx context.Context) (any, error) {
		count++
		log.Println("attempt:", count)
		if count < 3 {
			return nil, fmt.Errorf("error at attempt %d", count)
		}
		return "success", nil
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("result:", result)
}
