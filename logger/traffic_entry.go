package logger

import (
	"fmt"
	"strings"
	"time"
)

type TrafficTyp string

const (
	TrafficTypRecv TrafficTyp = "recv"
	TrafficTypSend TrafficTyp = "send"
)

// Traffic is provided by user when logging
type Traffic struct {
	Typ  TrafficTyp    // Typ: type of traffic, receive request or send request
	Cmd  string        // Cmd: command
	Cost time.Duration // Cost: elapse of processing
	Code string        // Code: error code
	Msg  string        // Msg: error message if you have
	Req  any
	Resp any
}

// ReqEntity is provided by user when logging
type ReqEntity struct {
	Typ    TrafficTyp // Typ: type of traffic, receive request or send request
	Cmd    string     // Cmd: command
	Req    any
	Fields Fields // Fields: request additional fields
}

type RespEntity struct {
	Code string // Code: error code
	Msg  string // Msg: error message if you have
	Resp any
}

type TrafficRec struct {
	te        TrafficEntry
	startTime time.Time
	req       *ReqEntity
}

func newTrafficRec(te TrafficEntry, req *ReqEntity) *TrafficRec {
	return &TrafficRec{
		te:        te,
		startTime: time.Now(),
		req:       req,
	}
}

func (t *Traffic) headString(sep string) string {
	if t == nil {
		return ""
	}

	return strings.Join(append([]string{
		ifThenFunc(t.Typ == "", func() string {
			return defaultFieldOccupied
		}, func() string {
			return string(t.Typ)
		}),
		ifThen(t.Cmd == "", defaultFieldOccupied, t.Cmd),
		fmt.Sprintf("%s", t.Cost),
		t.Code,
		t.Msg,
	}), sep)
}

func (t *TrafficRec) End(resp *RespEntity, fields Fields) {
	t.EndWithIgnores(resp, fields)
}

func (t *TrafficRec) EndWithIgnores(resp *RespEntity, fields Fields, ignores ...string) {
	if t == nil || t.te == nil || resp == nil {
		return
	}

	newFields := copyFields(t.req.Fields)
	for k, v := range fields {
		newFields[k] = v
	}

	t.te.WithIgnores(ignores...).
		DataWith(&Traffic{
			Typ:  t.req.Typ,
			Cmd:  t.req.Cmd,
			Code: resp.Code,
			Msg:  resp.Msg,
			Cost: time.Since(t.startTime),
			Req:  t.req.Req,
			Resp: resp.Resp,
		}, newFields)
}

type TrafficEntry interface {
	// Data logs traffic
	Data(traffic *Traffic)
	// DataWith logs traffic with fields
	DataWith(traffic *Traffic, fields Fields)
	// WithFields adds fields to traffic dataLogger
	WithFields(fields Fields) TrafficEntry
	// WithTracing adds requestId to traffic dataLogger
	WithTracing(requestId string) TrafficEntry
	// WithIgnores adds ignores to traffic dataLogger
	WithIgnores(ignores ...string) TrafficEntry
	// WithPolicy adds policy to traffic dataLogger
	// disable: true: disable policy, false: enable policy
	WithPolicy(policy Policy) TrafficEntry

	Start(req *ReqEntity) *TrafficRec
}

func copyFields(fields Fields) Fields {
	if len(fields) == 0 {
		return map[string]any{}
	}
	mapCopy := make(map[string]any, len(fields))
	for k, v := range fields {
		mapCopy[k] = v
	}
	return mapCopy
}

type emptyTrafficEntry struct{}

func (et *emptyTrafficEntry) Data(traffic *Traffic) {
}

func (et *emptyTrafficEntry) DataWith(traffic *Traffic, fields Fields) {
}

func (et *emptyTrafficEntry) WithFields(fields Fields) TrafficEntry {
	return et
}

func (et *emptyTrafficEntry) WithTracing(requestId string) TrafficEntry {
	return et
}

func (et *emptyTrafficEntry) WithIgnores(ignores ...string) TrafficEntry {
	return et
}

func (et *emptyTrafficEntry) WithPolicy(policy Policy) TrafficEntry {
	return et
}

func (et *emptyTrafficEntry) Start(req *ReqEntity) *TrafficRec {
	return nil
}
