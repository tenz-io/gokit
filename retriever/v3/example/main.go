// Command retriever-example demonstrates the retriever/v3 API: typed return
// values, default retry-everything behavior, NonRetryable opt-out, error
// classification, and composable backoff with jitter.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/tenz-io/gokit/retriever/v3"
)

func main() {
	ctx := context.Background()

	// 1) 默认配置:3 次尝试、100ms 起步的指数退避。零 option 即可用。
	r := retriever.New[string]()

	// 直接拿到 string,无需类型断言 —— 这是 v3 相对 v2 的核心易用性改进。
	var count int
	result, err := r.Do(ctx, func(ctx context.Context) (string, error) {
		count++
		if count < 3 {
			return "", fmt.Errorf("fail %d", count)
		}
		return "ok", nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("default retry:", result, "attempts:", count)

	// 2) NonRetryable:标记不可重试的错误,Do 立即返回,不浪费重试次数。
	//    且把 fn 返回的 result 原样带回(本例的 42),便于调用方释放资源。
	perm := retriever.New[int](retriever.WithMaxAttempts(5))
	attempts := 0
	result2, err := perm.Do(ctx, func(ctx context.Context) (int, error) {
		attempts++
		return 42, retriever.NonRetryable(errors.New("bad request"))
	})
	fmt.Printf("non-retryable: err=%v result=%d attempts=%d (expected 1)\n", err, result2, attempts)

	// 3) 错误分类器:仅对特定错误重试。
	classified := retriever.New[int](
		retriever.WithMaxAttempts(4),
		retriever.WithBackoff(retriever.Constant(0)),
		retriever.WithRetryable(func(err error) bool {
			return errors.Is(err, errTransient)
		}),
	)
	i := 0
	_, err = classified.Do(ctx, func(ctx context.Context) (int, error) {
		i++
		if i < 3 {
			return 0, errTransient // 重试
		}
		return 0, errPerm // 不重试,立即返回
	})
	fmt.Printf("classifier: err=%v attempts=%d (transient retried, perm stopped)\n", err, i)

	// 4) 可组合退避 + 抖动 + 全局截止时间。
	flaky := retriever.New[string](
		retriever.WithMaxAttempts(5),
		retriever.WithTimeout(500*time.Millisecond),
		retriever.WithBackoff(retriever.Jitter{
			Backoff: retriever.Exponential{Base: 20 * time.Millisecond, Factor: 2},
			Factor:  0.3,
		}),
	)
	n := 0
	_, err = flaky.Do(ctx, func(ctx context.Context) (string, error) {
		n++
		if n < 4 {
			return "", fmt.Errorf("flaky %d", n)
		}
		return "recovered", nil
	})
	fmt.Printf("flaky backoff: err=%v attempts=%d\n", err, n)
}

var (
	errTransient = errors.New("transient")
	errPerm      = errors.New("permanent")
)
