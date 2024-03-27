package logger

import (
	"fmt"
	"strings"
	"time"
)

type TrafficTyp string

const (
	TrafficTypReq  TrafficTyp = "req_to"
	TrafficTypResp TrafficTyp = "resp_from"
)

// Traffic is provided by user when logging
type Traffic struct {
	Typ  TrafficTyp    // Typ: type of traffic, access or request
	Cmd  string        // Cmd: command
	Code int           // Code: error code
	Msg  string        // Msg: error message if you have
	Cost time.Duration // Cost: elapse of processing
	Req  any
	Resp any
}

// TrafficReq is provided by user when logging
type TrafficReq struct {
	Cmd string // Cmd: command
	Req any
}

type TrafficResp struct {
	Code int    // Code: error code
	Msg  string // Msg: error message if you have
	Resp any
}

type TrafficRec struct {
	te        TrafficEntry
	startTime time.Time
	pairId    string
	cmd       string
}

func newTrafficRec(te TrafficEntry, cmd, pairId string) *TrafficRec {
	return &TrafficRec{
		te:        te,
		startTime: time.Now(),
		pairId:    pairId,
		cmd:       cmd,
	}
}

func (t *TrafficRec) End(resp *TrafficResp, fields Fields) {
	if t == nil || t.te == nil || resp == nil {
		return
	}

	if fields == nil {
		fields = make(Fields)
	}

	fields[defaultPairFieldName] = t.pairId

	t.te.DataWith(&Traffic{
		Typ:  TrafficTypResp,
		Cmd:  t.cmd,
		Code: resp.Code,
		Msg:  resp.Msg,
		Cost: time.Since(t.startTime),
		Resp: resp.Resp,
	}, fields)

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

	Start(req *TrafficReq, fields Fields) *TrafficRec
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

// convertToMessage converts a Traffic to a string
func convertToMessage(tb *Traffic, separator string) string {
	if tb == nil {
		return ""
	}
	if tb.Typ == "" {
		tb.Typ = defaultFieldOccupied
	}
	if tb.Msg == "" {
		tb.Msg = defaultFieldOccupied
	}
	if tb.Cmd == "" {
		tb.Cmd = defaultFieldOccupied
	}

	var reqTyp = tb.Typ == TrafficTypReq

	return strings.Join(append([]string{
		string(tb.Typ),
		tb.Cmd,
		ifThen(reqTyp, defaultFieldOccupied, fmt.Sprintf("%s", tb.Cost)).(string),
		ifThen(reqTyp, defaultFieldOccupied, fmt.Sprintf("%d", tb.Code)).(string),
		tb.Msg,
	}), separator)
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

func (et *emptyTrafficEntry) Start(req *TrafficReq, fields Fields) *TrafficRec {
	return nil
}
