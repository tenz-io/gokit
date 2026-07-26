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

// TrafficTyp distinguishes the direction of a traffic record: a request
// received from a caller (recv) or a request sent to a downstream (send).
type TrafficTyp string

const (
	TrafficTypRecv TrafficTyp = "recv"
	TrafficTypSend TrafficTyp = "send"
)

// TrafficRec records a single request/response span.
//
// Start one with Entry.StartTraffic(cmd); complete it with End or
// EndWithError. All methods are nil-safe: a nil *TrafficRec (returned when
// traffic logging is disabled) makes End/EndWithError no-ops, so callers can
// write `rec := logger.StartTraffic(cmd); defer func(){ rec.End(resp, code) }()`
// without guarding against nil.
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

// WithTyp sets the direction of the traffic record. Returns the receiver for
// chaining. Defaults to recv.
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

// End completes the traffic span, recording the response body, a status code,
// and any extra structured fields. Only the first End/EndWithError call writes
// a record; later calls are no-ops.
//
//	resp   any     the response payload to log (may be nil)
//	code   string  a status code (e.g. "200", "0", "error")
//	fields ...any  extra key-value pairs appended to the record
func (tr *TrafficRec) End(resp any, code string, fields ...any) {
	tr.endWith(resp, code, "", fields...)
}

// EndWithError completes the traffic span with an error: code is set to
// "error" and msg to the error's message.
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

// newTrafficCore builds the dedicated zap instance that writes traffic.log.
// It is intentionally separate from the main core so traffic records never
// mix into the business log files.
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
	// Traffic records already have a structured "msg" field for the error
	// detail. Use a different key for the pipe-delimited human summary so JSON
	// output never contains duplicate keys.
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
