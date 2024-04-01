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

func (te *LogTrafficEntry) Start(req *ReqEntity) *TrafficRec {
	if !te.validate() || req == nil {
		return nil
	}

	return newTrafficRec(te, req)
}

// Data Log a request
func (te *LogTrafficEntry) Data(tc *Traffic) {
	te.DataWith(tc, nil)
}

// DataWith Log a request with fields
func (te *LogTrafficEntry) DataWith(tc *Traffic, fields Fields) {
	if tc == nil || !te.validate() {
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
		te.dataLogger.Info(
			te.withHead(tc.headString(te.sep)),
			toZapFields(newFields, te.ignores...)...,
		)
	}()
}

// WithFields modifies an existing dataLogger with new fields (cannot be removed)
func (te *LogTrafficEntry) WithFields(fields Fields) TrafficEntry {
	if !te.validate() {
		return te
	}
	args := toZapFields(fields)
	return &LogTrafficEntry{
		dataLogger: te.dataLogger.With(args...),
		sep:        te.sep,
		requestId:  te.requestId,
		ignores:    te.ignores,
		allow:      te.allow,
	}
}

// WithTracing create copy of LogEntry with tracing.Span
func (te *LogTrafficEntry) WithTracing(requestId string) TrafficEntry {
	if !te.validate() {
		return te
	}
	return &LogTrafficEntry{
		dataLogger: te.dataLogger,
		sep:        te.sep,
		ignores:    te.ignores,
		requestId:  requestId,
		allow:      te.allow,
	}
}

func (te *LogTrafficEntry) WithIgnores(ignores ...string) TrafficEntry {
	if !te.validate() {
		return te
	}
	return &LogTrafficEntry{
		dataLogger: te.dataLogger,
		sep:        te.sep,
		requestId:  te.requestId,
		ignores:    ignores,
		allow:      te.allow,
	}
}

// WithPolicy create copy of LogEntry with policy
// disable: true: disable policy, false: enable policy
func (te *LogTrafficEntry) WithPolicy(policy Policy) TrafficEntry {
	if !te.validate() || policy == nil {
		return te
	}

	return &LogTrafficEntry{
		dataLogger: te.dataLogger,
		sep:        te.sep,
		requestId:  te.requestId,
		ignores:    te.ignores,
		allow:      policy.Allow(),
	}
}

func (te *LogTrafficEntry) withHead(msg string) string {
	if !te.validate() {
		return msg
	}

	infos := append([]string{defaultDataLevelName})
	if te.requestId == "" {
		infos = append(infos, defaultLogEmpty)
	} else {
		infos = append(infos, te.requestId)
	}
	return strings.Join(append(infos, msg), te.sep)
}

// clone a log entry
func (te *LogTrafficEntry) clone() *LogTrafficEntry {
	if !te.validate() {
		return nil
	}
	return &LogTrafficEntry{
		dataLogger: te.dataLogger,
		sep:        te.sep,
		requestId:  te.requestId,
		allow:      te.allow,
	}
}

func (te *LogTrafficEntry) validate() bool {
	if te == nil || te.dataLogger == nil || !te.allow {
		return false
	}
	return true
}
