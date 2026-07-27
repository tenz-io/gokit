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

// WithLogger 根据 flag 重新配置全局 logger,可选启用
// traffic logger。它是 logger/v3 的薄封装,使希望
// 开启 traffic 日志的调用方一行即可启用;包已在 Run 中
// 配置了基础 logger,因此仅当需要 Traffic flag 时才用到。
func WithLogger(trafficEnabled bool) InitFunc {
	return func(c *Context, _ any) (CleanFunc, error) {
		logDir := c.Flags().String(FlagNameLog)
		if logDir == "" {
			logDir = "log"
		}
		verbose := c.Flags().Bool(FlagNameVerbose)
		loggingFile := c.Flags().Bool(FlagNameLoggingFile)
		loggingConsole := c.Flags().Bool(FlagNameLoggingConsole)

		lvl := logger.InfoLevel
		if verbose {
			lvl = logger.DebugLevel
		}

		opts := []logger.ConfigOption{
			logger.WithLevel(lvl),
			logger.WithConsole(loggingConsole),
			logger.WithCaller(true),
			logger.WithTraffic(trafficEnabled),
		}
		if loggingFile {
			opts = append(opts, logger.WithFilePath(logDir))
		}
		logger.ConfigureWithOpts(opts...)
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

		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
			// 保守的 timeout,使卡住的 admin 请求不会阻塞 shutdown。
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		}

		errC := make(chan error, 1)
		go func() {
			logger.Infow("starting admin http server", "addr", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errC <- err
			}
		}()

		clean := func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			logger.Infow("shutting down admin http server", "addr", srv.Addr)
			if err := srv.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("admin http shutdown: %w", err)
			}
			return nil
		}

		// 呈现一个在 Run 装配 errC 之前到达的 listen 错误。
		select {
		case err := <-errC:
			_ = clean(context.Background())
			return nil, fmt.Errorf("admin http listen: %w", err)
		default:
			return clean, nil
		}
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
