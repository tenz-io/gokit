package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"

	"github.com/tenz-io/gokit/annotation/v3"
	"github.com/tenz-io/gokit/logger/v3"
)

// UnmarshalFunc 将原始 config 字节解码到 config 值。yaml.Unmarshal
// 与 json.Unmarshal 满足该签名。
type UnmarshalFunc func([]byte, any) error

// WithYamlConfig 将由 `config` flag 命名的 YAML config 文件解码到
// conf,先应用 annotation default 再校验。不返回
// cleanup(文件 I/O 为一次性)。
func WithYamlConfig() InitFunc { return decodeConfig(yaml.Unmarshal) }

// WithJsonConfig 将由 `config` flag 命名的 JSON config 文件解码到
// conf,先应用 annotation default 再校验。
func WithJsonConfig() InitFunc { return decodeConfig(json.Unmarshal) }

func decodeConfig(unmarshal UnmarshalFunc) InitFunc {
	return func(c *Context, conf any) (CleanFunc, error) {
		path := c.Flags().String(FlagNameConfig)
		if path == "" {
			return nil, fmt.Errorf("config file is empty")
		}
		if err := ReadConfig(path, conf, unmarshal); err != nil {
			return nil, fmt.Errorf("read config %s: %w", path, err)
		}
		if c.Flags().Bool(FlagNameVerbose) {
			logger.Debugf("config: %s", PrettyString(conf))
		}
		return nil, nil
	}
}

// WithDotEnvConfig 将给定 .env 文件加载到进程环境。
// 无文件名时加载默认 ".env"。
func WithDotEnvConfig(filenames ...string) InitFunc {
	return func(_ *Context, _ any) (CleanFunc, error) {
		if len(filenames) == 0 {
			filenames = []string{".env"}
		}
		if err := godotenv.Load(filenames...); err != nil {
			return nil, fmt.Errorf("load .env %v: %w", filenames, err)
		}
		return nil, nil
	}
}

// WithTraffic 在运行期覆盖 traffic logger 开关。Run 已按
// -traffic flag 配置过基础 logger,因此多数情况下无需本 init;
// 仅当调用方希望无条件开启 traffic(忽略 flag)时使用。
func WithTraffic() InitFunc {
	return func(_ *Context, _ any) (CleanFunc, error) {
		logger.ConfigureWithOpts(logger.WithTraffic(true))
		return nil, nil
	}
}

// WithLogger 保留以兼容旧调用方;它现在等价于 -traffic flag
// 被设为 trafficEnabled 后的 configureLogger。新代码请用
// WithTraffic 或直接传 -traffic flag。
//
// Deprecated: 改用 WithTraffic() 或 -traffic 命令行 flag。
func WithLogger(trafficEnabled bool) InitFunc {
	return func(_ *Context, _ any) (CleanFunc, error) {
		logger.ConfigureWithOpts(logger.WithTraffic(trafficEnabled))
		return nil, nil
	}
}

// WithAdminHTTPServer 在 `admin-port` flag 上启动 admin HTTP server,并将
// /debug/pprof、/metrics 与 /ping 挂载到其自身 ServeMux 上(绝非全局
// DefaultServeMux,v2 曾污染它)。返回的 CleanFunc 执行
// 带 timeout 的 graceful Shutdown。
func WithAdminHTTPServer() InitFunc {
	return func(c *Context, _ any) (CleanFunc, error) {
		port := c.Flags().Int(FlagNameAdminPort)
		if port <= 0 || port > 65535 {
			return nil, fmt.Errorf("invalid admin port: %d", port)
		}

		mux := http.NewServeMux()
		AddProfilingHandler(mux)
		AddPingHandler(mux)
		AddPrometheusHandler(mux)

		return startHTTPServer("admin http server", port, mux)
	}
}

// WithHTTPServer 在 `port` flag 上启动业务 HTTP server,handler 由调用方提供。
// 与 admin server 一样使用专属 *http.Server + 带 timeout 的 graceful
// Shutdown,使多数 HTTP 服务的 main 无需重复编写 listen/shutdown 样板。
// 返回的 CleanFunc 在退出时优雅关闭。
func WithHTTPServer(handler http.Handler) InitFunc {
	return func(c *Context, _ any) (CleanFunc, error) {
		port := c.Flags().Int(FlagNamePort)
		if port <= 0 || port > 65535 {
			return nil, fmt.Errorf("invalid http port: %d", port)
		}
		return startHTTPServer("http server", port, handler)
	}
}

// startHTTPServer 是 admin 与业务 server 的共享实现:装配带保守 timeout 的
// *http.Server,在 goroutine 中 ListenAndServe,并返回一个带 5s 超时
// graceful Shutdown 的 CleanFunc。它在返回前探一次 errC,使 listen
// 期错误能以 init 失败形式呈现(而非等到 Run 才报)。
func startHTTPServer(label string, port int, handler http.Handler) (CleanFunc, error) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
		// 保守的 timeout,使卡住的请求不会阻塞 shutdown。
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errC := make(chan error, 1)
	go func() {
		logger.Infow("starting "+label, "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errC <- err
		}
	}()

	clean := func(ctx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		logger.Infow("shutting down "+label, "addr", srv.Addr)
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("%s shutdown: %w", label, err)
		}
		return nil
	}

	// 呈现一个在 Run 装配 errC 之前到达的 listen 错误。
	select {
	case err := <-errC:
		_ = clean(context.Background())
		return nil, fmt.Errorf("%s listen: %w", label, err)
	default:
		return clean, nil
	}
}

// ReadConfig 读取 path 处的 config 文件,应用 annotation default,
// 针对进程环境展开 ${VAR} 占位符,然后将
// 字节 unmarshal 到 conf,再校验。顺序为:
//
//	ApplyDefaults  — 在 conf 上应用 struct-tag default(置字段零值)
//	Expand         — 使用 os.LookupEnv 替换原始字节中的
//	                ${VAR} / ${VAR:-x} / ${VAR:?m}
//	                (从而使 WithDotEnvConfig 加载的 .env 值
//	                可见;将该 Init 放在最前)
//	unmarshal      — bytes → conf
//	Validate       — 运行 annotation 规则
//
// 引用未设置变量且无 :-default 或 :?error 的占位符视为
// 错误,因此缺失敏感值会启动失败,而不会将
// 字面量 "${VAR}" 泄漏到解码后的 config。
func ReadConfig(path string, conf any, unmarshal UnmarshalFunc) error {
	bs, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := annotation.ApplyDefaults(conf); err != nil {
		return fmt.Errorf("apply defaults: %w", err)
	}
	if bs, err = Expand(bs, os.LookupEnv); err != nil {
		return fmt.Errorf("expand placeholders: %w", err)
	}
	if err := unmarshal(bs, conf); err != nil {
		return fmt.Errorf("unmarshal %s: %w", path, err)
	}
	if err := annotation.Validate(conf); err != nil {
		return fmt.Errorf("validate %s: %w", path, err)
	}
	return nil
}
