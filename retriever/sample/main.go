package main

import (
	"context"
	"fmt"
	"github.com/tenz-io/gokit/retriever"
	"log"
	"time"
)

func main() {
	retr := retriever.NewRetrieverWithConfig(
		retriever.WithBackoff(retriever.NewLinearBackoff(50)),
		retriever.WithMaxAttempt(3),
		retriever.WithMaxTotalAttemptTime(350*time.Millisecond),
	)

	var count int
	result, err := retr.DoAlwaysRetry(context.Background(), func(ctx context.Context) (any, error) {
		defer func() {
			count++
		}()
		log.Println("count:", count)
		if count < 3 {
			return nil, fmt.Errorf("error in count-%d", count)
		}

		return "success", nil
	})
	if err != nil {
		log.Println("error:", err)
		return
	}

	log.Println("result:", result)

}
