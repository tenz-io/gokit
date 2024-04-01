package logger

import (
	"strings"

	"go.uber.org/zap"
)

type LogTrafficEntry struct {
	dataLogger *zap.Logger
	sep        string
	requestId  string
	ignores    []string
	allow      bool // for policy use, init true
}

func (le *LogTrafficEntry) Start(req *ReqEntity, fields Fields) *TrafficRec {
	if !le.validate() || req == nil {
		return nil
	}

	if fields == nil {
		fields = make(Fields)
	}

	le.DataWith(&Traffic{
		Typ: TrafficTypRecv,
		Cmd: req.Cmd,
		Req: req.Req,
	}, fields)
	return newTrafficRec(le, req.Cmd)
}

// Data Log a request
func (le *LogTrafficEntry) Data(tc *Traffic) {
	le.DataWith(tc, nil)
}

// DataWith Log a request with fields
func (le *LogTrafficEntry) DataWith(tc *Traffic, fields Fields) {
	if tc == nil || !le.validate() {
		return
	}

	newFields := copyFields(fields)

	if tc.Req != nil {
		newFields[defaultReqFieldName] = tc.Req
		if reqLen, ok := lenIfArrayType(tc.Req); ok {
			newFields[arrFieldPrefix+defaultReqFieldName] = reqLen
		}

	}
	if tc.Resp != nil {
		newFields[defaultRespFieldName] = tc.Resp
		if respLen, ok := lenIfArrayType(tc.Resp); ok {
			newFields[arrFieldPrefix+defaultRespFieldName] = respLen
		}
	}

	// async log
	go func() {
		le.dataLogger.Info(
			le.withHead(tc.headString(le.sep)),
			toZapFields(newFields, le.ignores...)...,
		)
	}()
}

// WithFields modifies an existing dataLogger with new fields (cannot be removed)
func (le *LogTrafficEntry) WithFields(fields Fields) TrafficEntry {
	if !le.validate() {
		return le
	}
	args := toZapFields(fields)
	return &LogTrafficEntry{
		dataLogger: le.dataLogger.With(args...),
		sep:        le.sep,
		requestId:  le.requestId,
		ignores:    le.ignores,
		allow:      le.allow,
	}
}

// WithTracing create copy of LogEntry with tracing.Span
func (le *LogTrafficEntry) WithTracing(requestId string) TrafficEntry {
	if !le.validate() {
		return le
	}
	return &LogTrafficEntry{
		dataLogger: le.dataLogger,
		sep:        le.sep,
		ignores:    le.ignores,
		requestId:  requestId,
		allow:      le.allow,
	}
}

func (le *LogTrafficEntry) WithIgnores(ignores ...string) TrafficEntry {
	if !le.validate() {
		return le
	}
	return &LogTrafficEntry{
		dataLogger: le.dataLogger,
		sep:        le.sep,
		requestId:  le.requestId,
		ignores:    ignores,
		allow:      le.allow,
	}
}

// WithPolicy create copy of LogEntry with policy
// disable: true: disable policy, false: enable policy
func (le *LogTrafficEntry) WithPolicy(policy Policy) TrafficEntry {
	if !le.validate() || policy == nil {
		return le
	}

	return &LogTrafficEntry{
		dataLogger: le.dataLogger,
		sep:        le.sep,
		requestId:  le.requestId,
		ignores:    le.ignores,
		allow:      policy.Allow(),
	}
}

func (le *LogTrafficEntry) withHead(msg string) string {
	if !le.validate() {
		return msg
	}

	infos := append([]string{defaultDataLevelName})
	if le.requestId == "" {
		infos = append(infos, defaultLogEmpty)
	} else {
		infos = append(infos, le.requestId)
	}
	return strings.Join(append(infos, msg), le.sep)
}

// clone a log entry
func (le *LogTrafficEntry) clone() *LogTrafficEntry {
	if !le.validate() {
		return nil
	}
	return &LogTrafficEntry{
		dataLogger: le.dataLogger,
		sep:        le.sep,
		requestId:  le.requestId,
		allow:      le.allow,
	}
}

func (le *LogTrafficEntry) validate() bool {
	if le == nil || le.dataLogger == nil || !le.allow {
		return false
	}
	return true
}
