package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// TrafficTyp 区分 traffic 记录的方向:从调用方接收到的请求(recv)
// 或发往下游的请求(send)。
type TrafficTyp string

const (
	TrafficTypRecv TrafficTyp = "recv"
	TrafficTypSend TrafficTyp = "send"
)

// TrafficRec 记录单个请求/响应 span。
//
// 通过 Entry.StartTraffic(cmd) 开启一个;用 End 或 EndWithError 完成它。
// 所有方法都是 nil 安全的:当 traffic 日志被禁用时返回的 nil *TrafficRec
// 会让 End/EndWithError 成为 no-op,因此调用方可以直接写
// `rec := logger.StartTraffic(cmd); defer func(){ rec.End(resp, code) }()`
// 而无需做 nil 守卫。
type TrafficRec struct {
	mu        sync.Mutex
	logger    *zap.SugaredLogger
	trimmer   *OutputTrimmer
	cmd       string
	typ       TrafficTyp
	startTime time.Time
	ended     bool
}

func startTrafficRec(logger *zap.SugaredLogger, cmd string, trimmer *OutputTrimmer) *TrafficRec {
	return &TrafficRec{
		logger:    logger,
		trimmer:   trimmer,
		cmd:       cmd,
		typ:       TrafficTypRecv,
		startTime: time.Now(),
	}
}

// WithTyp 设置 traffic 记录的方向。返回 receiver 以便链式调用。
// 默认为 recv。
func (tr *TrafficRec) WithTyp(t TrafficTyp) *TrafficRec {
	if tr == nil {
		return tr
	}
	tr.mu.Lock()
	defer tr.mu.Unlock()
	if tr.ended {
		return tr
	}
	tr.typ = t
	return tr
}

// End 完成该 traffic span,记录响应体、状态码以及任意额外的结构化
// field。仅第一次 End/EndWithError 调用会写入记录;后续调用为 no-op。
//
//	resp   any     要记录的响应负载(可为 nil)
//	code   string  状态码(例如 "200"、"0"、"error")
//	fields ...any  追加到记录中的额外键值对
func (tr *TrafficRec) End(resp any, code string, fields ...any) {
	tr.endWith(resp, code, "", fields...)
}

// EndWithError 以一个错误完成该 traffic span:code 被置为
// "error",msg 被置为该错误的消息。
func (tr *TrafficRec) EndWithError(err error, fields ...any) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	tr.endWith(nil, "error", msg, fields...)
}

func (tr *TrafficRec) endWith(resp any, code, msg string, fields ...any) {
	if tr == nil || tr.logger == nil {
		return
	}
	tr.mu.Lock()
	defer tr.mu.Unlock()
	if tr.ended {
		return
	}
	tr.ended = true

	cost := time.Since(tr.startTime)
	trimmed := tr.trimmer.TrimFields(fields)
	cmd := tr.trimmer.trimString(tr.cmd)
	msg = tr.trimmer.trimString(msg)

	allFields := []any{
		"cmd", cmd,
		"typ", string(tr.typ),
		"code", code,
		"msg", msg,
		"cost", cost.String(),
	}
	if resp != nil && !tr.trimmer.ignores["resp"] {
		allFields = append(allFields, "resp", tr.trimmer.trimAny(resp, tr.trimmer.deepLimit))
	}
	allFields = append(allFields, trimmed...)

	logMsg := strings.Join([]string{
		string(tr.typ),
		cmd,
		cost.String(),
		code,
		msg,
	}, "|")

	tr.logger.Infow(logMsg, allFields...)
}

// newTrafficCore 构建写入 traffic.log 的专用 zap 实例。
// 它刻意与主 core 分离,以免 traffic 记录混入业务日志文件。
func newTrafficCore(cfg Config) *zap.SugaredLogger {
	trafficDir := cfg.TrafficPath
	if trafficDir == "" {
		trafficDir = cfg.FilePath
	}
	if trafficDir == "" {
		trafficDir = "log"
	}

	if err := os.MkdirAll(trafficDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "logger: failed to create traffic directory %s: %v\n", trafficDir, err)
		return nil
	}

	encCfg := encoderConfig()
	encCfg.LevelKey = ""
	encCfg.CallerKey = ""
	// traffic 记录已有一个用于错误详情的结构化 "msg" field。此处用
	// 不同的 key 来承载以竖线分隔的人类可读摘要,使 JSON 输出永远不会
	// 出现重复 key。
	encCfg.MessageKey = "summary"
	var enc zapcore.Encoder
	if cfg.Encoding == JSONEncoding {
		enc = zapcore.NewJSONEncoder(encCfg)
	} else {
		enc = zapcore.NewConsoleEncoder(encCfg)
	}

	core := zapcore.NewCore(
		enc,
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   filepath.Join(trafficDir, "traffic.log"),
			MaxSize:    cfg.TrafficMaxSize,
			MaxAge:     cfg.TrafficMaxAge,
			MaxBackups: cfg.TrafficMaxBackups,
			Compress:   true,
			LocalTime:  true,
		}),
		zapcore.InfoLevel,
	)

	return zap.New(core).Sugar()
}
