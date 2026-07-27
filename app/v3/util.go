package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tenz-io/gokit/logger/v3"
)

// wait 阻塞,直到以下之一发生:收到 signal、app context done、或
// Run goroutine 在 errC 上报告。hook 会执行一次 graceful cleanup,
// 在每个分支返回之前都会被调用。返回的 ExitCode 反映
// 原因:非 nil errC 为 ExitRunError、nil errC(干净
// 完成)为 ExitOK、被中断为 ExitSignal。
//
// 与 v2 的 WaitSignal 不同,它从不调用 os.Exit —— Run 返回 code,由
// 调用方决定。
func wait(ctx context.Context, errC <-chan error, hook func()) ExitCode {
	// SIGINT(Ctrl-C)与 SIGTERM(container/k8s shutdown)。os.Kill 不可
	// 捕获,因此刻意省略。
	signC := make(chan os.Signal, 1)
	signal.Notify(signC, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signC)

	for {
		select {
		case <-signC:
			logger.Infow("received interrupt signal")
			hook()
			return ExitSignal
		case <-ctx.Done():
			logger.Infow("context done")
			hook()
			return ExitOK
		case err := <-errC:
			if err != nil {
				logger.Errorf("run error: %+v", err)
				hook()
				return ExitRunError
			}
			logger.Infow("run completed")
			hook()
			return ExitOK
		}
	}
}

// PrettyString 将 v 渲染为紧凑的 JSON 字符串,失败时回退到 %+v。适合
// 在 verbose 模式下打印解码后的 config。
func PrettyString(v any) string {
	if v == nil {
		return "nil"
	}
	if j, err := json.Marshal(v); err == nil {
		return string(j)
	}
	return fmt.Sprintf("%+v", v)
}
