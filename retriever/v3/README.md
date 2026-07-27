# retriever

泛型重试执行器:`Do` 返回类型化的 `(T, error)`,内置可组合的退避策略(常量/线性/指数 + 抖动装饰器),默认对任何错误重试、不可重试错误用 `NonRetryable` 标记后立即返回。

```go
import "github.com/tenz-io/gokit/retriever/v3"
```

## 模块介绍

retriever 是一个泛型重试执行器,围绕三条设计原则展开:

- **类型化结果**:`Retriever[T].Do` 返回 `(T, error)`,调用方直接拿到 `T`,无需任何类型断言。
- **单一重试模型**:只有一个 `Do`,默认对任何错误重试;不可重试的错误用 `NonRetryable` 标记后立即返回,更精细的"哪些错误值得重试"交给 `WithRetryable` 分类器。
- **可组合退避**:`Backoff` 全部以 `time.Duration` 表达,是可直接用复合字面量构造的值类型;`Jitter` 作为装饰器叠加在任意 `Backoff` 之上,不与任何具体策略耦合。

核心特性:

- **泛型接口**:`Retriever[T]` / `DoFunc[T]`,`Do` 返回 `(T, error)`;无 `any`、无类型断言。
- **单一 `Do`**:默认"出错即重试";不可重试错误用 `NonRetryable(err)` 包装后 `Do` 立即返回;更精细的"哪些错误值得重试"用 `WithRetryable` 分类器。
- **退避全 `time.Duration`**:所有 `Backoff` 以 `time.Duration` 表达,均为可比较字段、可直接用复合字面量构造的值类型(`Constant`/`Linear`/`Exponential`)。
- **抖动可组合**:`Jitter{Backoff: ..., Factor: 0.3}` 装饰任意 `Backoff`,不绑定任何具体退避。
- **单一构造器**:仅 `New[T](opts...)`,函数式 `Option`,无导出的配置结构体。
- **无第三方依赖**:测试仅用标准库。

## 能力清单

| 能力 | 含义 |
| --- | --- |
| 类型化返回值 | `Retriever[T].Do` 返回 `(T, error)`,调用方直接拿到 `T`,无需类型断言 |
| 出错即重试直到成功 | 默认对任何非 `NonRetryable` 错误重试,直到成功、达到上限或 ctx 取消 |
| 标记不可重试错误 | `NonRetryable(err)` 包装的错误,`Do` 立即原样返回;标记对 `Error()` 文本完全透明 |
| 错误分类器 | `WithRetryable(func(error) bool)` 决定哪些错误值得重试,`NonRetryable` 标记优先于它 |
| 常量退避 | `Constant(d)` 每次重试前等待相同时长,零值即立即重试 |
| 线性退避 | `Linear{Base, Step}` 按 `Base + Step*attempt` 线性增长 |
| 指数退避 | `Exponential{Base, Factor}` 按 `Base * Factor^attempt` 指数增长,溢出钳制为正 |
| 抖动装饰器 | `Jitter{Backoff, Factor}` 在任意退避上叠加 `[0, Factor*wait)` 随机抖动,避免重试风暴 |
| 限制最大尝试次数 | `WithMaxAttempts(n)` 设置含首次调用的总尝试上限,耗尽后返回 `ErrMaxAttempts` 包裹的错误 |
| 全局截止时间 | `WithTimeout(d)` 为全部尝试设置统一截止时间,超时后立即停止 |
| 及时响应 ctx 取消 | 退避等待期间监听 `ctx.Done()`,外部取消或超时时立即返回,不会多等一次退避时长 |
| 并发安全 | `Retriever` 构造后自身字段不可变,`Do` 可被多 goroutine 并发调用;内置 `Backoff` 策略并发安全,自定义 `Backoff` 或分类器闭包须自行保证 |

## 默认值

零个 `Option` 即可用。`New[T]()` 的默认行为:

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| 最大尝试次数 | `3` | 含首次调用;`WithMaxAttempts(n<=0)` 视为无效、不修改当前值(默认 3) |
| 退避策略 | `Exponential{Base: 100ms, Factor: 2}` | 100ms 起步,每次失败翻倍 |
| 全局截止时间 | 无 | 仅受传入 `ctx` 自身约束;`WithTimeout(d<=0)` 表示不设 |
| 错误分类器 | 无 | 对所有非 `NonRetryable` 错误一律重试 |

## 快速开始

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/tenz-io/gokit/retriever/v3"
)

func main() {
	ctx := context.Background()

	// 1) 默认配置:3 次尝试、100ms 起步的指数退避。零 option 即可用。
	//    出错即重试,直到成功或达到上限;直接拿到 *http.Response,无需类型断言。
	r := retriever.New[*http.Response]()
	resp, err := r.Do(ctx, func(ctx context.Context) (*http.Response, error) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com", nil)
		return http.DefaultClient.Do(req)
	})
	if err != nil {
		fmt.Println(err)
	}
	_ = resp

	// 2) 区分可重试与不可重试的响应:
	//    - 5xx(服务端错误)可重试:重试前关闭响应体,避免连接泄漏。
	//    - 4xx(客户端错误)不可重试:用 NonRetryable 标记后,Do 立即返回,
	//      并把 resp 带回给调用方释放(见下方 err 分支)。
	r = retriever.New[*http.Response](retriever.WithMaxAttempts(5))
	resp, err = r.Do(ctx, func(ctx context.Context) (*http.Response, error) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err // 网络错误,重试
		}
		switch {
		case resp.StatusCode >= 500: // 服务端错误:可重试
			_ = resp.Body.Close() // 重试前释放本次响应体
			return nil, fmt.Errorf("server error: %d", resp.StatusCode)
		case resp.StatusCode >= 400: // 客户端错误:不可重试
			return resp, retriever.NonRetryable(fmt.Errorf("client error: %d", resp.StatusCode))
		}
		return resp, nil
	})
	// 4xx 时 resp 非 nil(由 Do 带回),调用方负责关闭。
	if resp != nil {
		defer resp.Body.Close()
	}
	_ = err

	// 3) 更精细的错误分类器:仅对 net.Error 的 Timeout 重试。
	r = retriever.New[*http.Response](retriever.WithRetryable(func(err error) bool {
		var netErr net.Error
		return errors.As(err, &netErr) && netErr.Timeout()
	}))
	_, _ = r.Do(ctx, func(ctx context.Context) (*http.Response, error) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com", nil)
		return http.DefaultClient.Do(req)
	})

	// 4) 可组合退避 + 抖动 + 全局截止时间。
	r = retriever.New[*http.Response](
		retriever.WithMaxAttempts(5),
		retriever.WithTimeout(5*time.Second),
		retriever.WithBackoff(retriever.Jitter{
			Backoff: retriever.Exponential{Base: 100 * time.Millisecond, Factor: 2},
			Factor:  0.3, // 叠加 [0, 30%) 随机抖动,避免重试风暴
		}),
	)
	_, _ = r.Do(ctx, func(ctx context.Context) (*http.Response, error) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com", nil)
		return http.DefaultClient.Do(req)
	})
}
```

### 错误处理

`Do` 的返回错误可按语义判别。建议按"特殊到一般"的顺序判断:先区分不可重试与
context 终止,最后才落到尝试耗尽:

```go
resp, err := r.Do(ctx, fn)

switch {
case err == nil:
	// 成功

case retriever.IsNonRetryable(err):
	// 你用 NonRetryable 标记的不可重试错误(err 仍是原错误)
	var bizErr *MyBizError
	_ = errors.As(err, &bizErr)

case errors.Is(err, context.Canceled):
	// ctx 被外部取消(含退避等待期间的取消)

case errors.Is(err, context.DeadlineExceeded):
	// 全局截止时间 (WithTimeout) 或 ctx 自身超时

case errors.Is(err, retriever.ErrMaxAttempts):
	// 重试到上限仍失败:err 同时包裹最后一次失败原因,
	// errors.Is(err, io.EOF) 可穿透判断最后一次失败是不是 EOF
}
```

判别优先级:

- **NonRetryable 最先判**:用 `IsNonRetryable(err)`,它在 ctx 终止检查之前判定。
- **ctx 终止优先于尝试耗尽**:fn 调用后若 ctx 已超时/取消,`Do` 直接返回 `ctx.Err()`,
  不会再用 `ErrMaxAttempts` 包裹 —— 因此同一超时不论发生在哪次尝试都统一归类。
- **ErrMaxAttempts 仅在真正"用完所有尝试仍失败"时返回**,用 `%w` 包裹最后一次失败错误,
  故 `errors.Is(err, ErrMaxAttempts)` 与 `errors.Is(err, <lastError>)` 同时成立。

返回值约定:`Do` 的所有终止路径都携带 fn 最后一次返回的 `result`(连同 `err`),
便于调用方释放 fn 已创建但未消费的资源(如 `resp.Body`)。fn 自行负责在中间失败、
即将重试时释放本次应丢弃的资源 —— `Do` 不会在重试之间代为回收。

- `NonRetryable(err)` 是**字符串透明**的:被标记错误的 `Error()` 文本与原错误完全一致,不引入 "retriever: ..." 前缀;`errors.Is` 仍能穿透到内层原因。
- ctx 取消/超时在退避等待期间被监听,`Do` 会立即返回 `ctx.Err()`,不会多等一次退避时长。

## 退避策略

所有 `Backoff` 均为值类型,可直接用复合字面量构造,`Next(attempt)` 的 `attempt` 从首次重试前的失败按 `0,1,2,...` 编号:

```go
// 常量:每次重试前等待 50ms(零值 = 立即重试)。
retriever.Constant(50 * time.Millisecond)

// 线性:Base + Step*attempt => 100,150,200,250...ms
retriever.Linear{Base: 100 * time.Millisecond, Step: 50 * time.Millisecond}

// 指数:Base * Factor^attempt => 100,200,400,800...ms;溢出钳制为正,绝不回绕为负。
retriever.Exponential{Base: 100 * time.Millisecond, Factor: 2}

// 抖动装饰器:在任意 Backoff 上叠加 [0, Factor*wait) 随机抖动。
// 下面在指数退避基础上叠加最多 30% 抖动。
retriever.Jitter{
    Backoff: retriever.Exponential{Base: 100 * time.Millisecond, Factor: 2},
    Factor:  0.3,
}

// Jitter 可装饰任意 Backoff(包括嵌套):
retriever.Jitter{
    Backoff: retriever.Linear{Base: 100 * time.Millisecond, Step: 50 * time.Millisecond},
    Factor:  0.25,
}
```

## API 速查

| 符号 | 说明 |
| --- | --- |
| `type Retriever[T any] interface{ Do(ctx, fn) (T, error) }` | 泛型重试执行器接口 |
| `func New[T any](opts ...Option) Retriever[T]` | 通过函数式选项创建 `Retriever`,零 option 即可用 |
| `type DoFunc[T any] func(context.Context) (T, error)` | 可重试函数类型 |
| `func NonRetryable(err error) error` | 标记错误不可重试;`Do` 立即返回,标记对 `Error()` 文本透明 |
| `func IsNonRetryable(err error) bool` | 报告 err 是否被 `NonRetryable` 标记(穿透包装) |
| `func WithMaxAttempts(n int) Option` | 设置最大尝试次数(含首次),`n<=0` 视为无效、不修改当前值(默认 3) |
| `func WithTimeout(d time.Duration) Option` | 为全部尝试设置统一截止时间,`d<=0` 表示不设 |
| `func WithBackoff(b Backoff) Option` | 设置退避策略,nil 被忽略 |
| `func WithRetryable(fn func(error) bool) Option` | 设置错误分类器,返回 false 时立即返回;nil 表示对所有错误重试 |
| `type Backoff interface{ Next(attempt int) time.Duration }` | 退避策略接口 |
| `type Constant time.Duration` | 常量退避,值类型 |
| `type Linear struct{ Base, Step time.Duration }` | 线性退避,值类型 |
| `type Exponential struct{ Base time.Duration; Factor float64 }` | 指数退避,值类型,溢出钳制为正 |
| `type Jitter struct{ Backoff Backoff; Factor float64 }` | 抖动装饰器,值类型,可叠加任意 `Backoff` |
| `var ErrNonRetryable error` | 不可重试标记哨兵,供 `errors.Is` 判别 |
| `var ErrMaxAttempts error` | 耗尽尝试次数哨兵,包裹最后一次错误,供 `errors.Is` 判别 |

引入路径:`github.com/tenz-io/gokit/retriever/v3`
