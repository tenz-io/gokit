package logger

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TrafficRec records a single request/response span.
type TrafficRec struct {
	mu        sync.Mutex
	logger    *zap.SugaredLogger
	trimmer   *OutputTrimmer
	cmd       string
	startTime time.Time
	typ       string
}

func startTrafficRec(logger *zap.SugaredLogger, cmd string, trimmer *OutputTrimmer) *TrafficRec {
	return &TrafficRec{
		logger:    logger,
		trimmer:   trimmer,
		cmd:       cmd,
		startTime: time.Now(),
		typ:       "recv",
	}
}

// End completes the traffic span. Supports two calling conventions:
//   New: End(resp any, code string)
//   Old: End(resp *RespEntity, fields Fields)
func (tr *TrafficRec) End(resp any, codeOrFields any, fields ...any) {
	if flds, ok := codeOrFields.(Fields); ok {
		// Old API compat: End(respEntity, fields)
		var re *RespEntity
		if r, ok := resp.(*RespEntity); ok {
			re = r
		}
		tr.endWithRespEntity(re, flds)
		return
	}
	code, _ := codeOrFields.(string)
	tr.endWith(resp, code, "", fields...)
}

func (tr *TrafficRec) endWithRespEntity(resp *RespEntity, fields Fields) {
	if resp == nil {
		tr.endWith(nil, "", "", fields.toArgs()...)
	} else {
		tr.endWith(resp.Resp, resp.Code, resp.Msg, fields.toArgs()...)
	}
}

// EndWithError completes the traffic span with an error response.
func (tr *TrafficRec) EndWithError(err error, fields ...any) {
	if err != nil {
		tr.endWith(nil, "error", err.Error(), fields...)
	} else {
		tr.endWith(nil, "error", "", fields...)
	}
}

// EndWithIgnores is the compat alias for consumers with ignore params.
func (tr *TrafficRec) EndWithIgnores(resp *RespEntity, fields Fields, ignores ...string) {
	tr.endWithRespEntity(resp, fields)
}

func (tr *TrafficRec) endWith(resp any, code, msg string, fields ...any) {
	if tr == nil || tr.logger == nil {
		return
	}
	tr.mu.Lock()
	defer tr.mu.Unlock()

	cost := time.Since(tr.startTime)

	trimmed := tr.trimmer.TrimFields(fields)

	allFields := []any{
		"cmd", tr.cmd,
		"code", code,
		"msg", msg,
		"cost", cost.String(),
	}
	if resp != nil {
		allFields = append(allFields, "resp", resp)
	}
	allFields = append(allFields, trimmed...)

	logMsg := strings.Join([]string{
		tr.typ,
		tr.cmd,
		cost.String(),
		code,
		msg,
	}, "|")

	tr.logger.Infow(logMsg, allFields...)
}

// --- Older ginext-compatible API ---

// TrafficTyp distinguishes request vs response direction.
type TrafficTyp string

const (
	TrafficTypRecv TrafficTyp = "recv"
	TrafficTypSend TrafficTyp = "send"
)

// Traffic holds a complete traffic record (pre-built, not using start/end).
type Traffic struct {
	Typ  TrafficTyp
	Cmd  string
	Cost time.Duration
	Code string
	Msg  string
	Req  any
	Resp any
}

// ReqEntity is request metadata for a traffic log.
type ReqEntity struct {
	Typ    TrafficTyp
	Cmd    string
	Req    any
	Fields Fields
}

// RespEntity is response metadata for a traffic log.
type RespEntity struct {
	Code string
	Msg  string
	Resp any
}

// Data logs a pre-built traffic record synchronously.
func Data(t *Traffic) {
	if t == nil {
		return
	}
	if global.traffic == nil {
		return
	}
	global.traffic.Infow(
		fmt.Sprintf("DATA|%s|%s|%s|%s", t.Cmd, t.Cost, t.Code, t.Msg),
		"cmd", t.Cmd,
		"code", t.Code,
		"cost", t.Cost.String(),
		"msg", t.Msg,
	)
}
