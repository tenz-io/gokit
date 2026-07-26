package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/tenz-io/gokit/async/v3"
)

func main() {
	ctx := context.Background()

	// Run: concurrent, first error wins, siblings keep running.
	if err := async.Run(ctx,
		slowTask("a", 10*time.Millisecond),
		slowTask("b", 5*time.Millisecond),
	); err != nil {
		log.Fatal(err)
	}

	// AllOf: one Result per task, in input order, errors isolated per slot.
	for i, r := range async.AllOf(ctx,
		slowTask("hello", 0),
		slowTask("world", 0),
		failingTask("oops"),
	) {
		fmt.Printf("AllOf[%d] value=%q err=%v\n", i, r.Value, r.Err)
	}

	// AnyOf: first success wins, losers are cancelled.
	val, err := async.AnyOf(ctx,
		slowTask("slow", 50*time.Millisecond),
		slowTask("fast", 0),
	)
	fmt.Printf("AnyOf value=%q err=%v\n", val, err)

	// Group: limit concurrency, cancel on first error.
	g := async.New[int](ctx, async.WithLimit[int](2), async.WithCancelOnError[int]())
	for i := 0; i < 5; i++ {
		i := i
		g.Go(func(ctx context.Context) (int, error) {
			if i == 3 {
				return 0, errors.New("bad index")
			}
			time.Sleep(5 * time.Millisecond)
			return i, nil
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Printf("Group err=%v (first failure wins)\n", err)
	}
	fmt.Printf("Group successes: %d\n", len(g.Results()))
}

func slowTask(v string, d time.Duration) async.Task[string] {
	return func(context.Context) (string, error) {
		time.Sleep(d)
		return v, nil
	}
}

func failingTask(v string) async.Task[string] {
	return func(context.Context) (string, error) {
		return "", fmt.Errorf("fail:%s", v)
	}
}
