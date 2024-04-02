package dbtracker

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/tenz-io/gokit/logger"
	"github.com/tenz-io/gokit/monitor"
)

type trackingKeyType string

const (
	trackingMetaCtxKey trackingKeyType = "_tracking_db_meta_ctx_key"
)

type meta struct {
	startTime  time.Time
	metricsRec *monitor.Recorder
	trafficRec *logger.TrafficRec
}

type Tracker interface {
	Start(cmd string) func(db *gorm.DB)
	End() func(db *gorm.DB)
}

// NewTrackerWithOpts creates a new tracker with the given options.
func NewTrackerWithOpts(opts ...ConfigOption) Tracker {
	config := defaultConfig
	for _, opt := range opts {
		opt(&config)
	}
	return NewTracker(config)
}

func NewTracker(config Config) Tracker {
	return &tracker{
		config: config,
	}
}

type tracker struct {
	config Config
}

func (t *tracker) Start(cmd string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		var (
			ctx = db.Statement.Context
			m   = &meta{
				startTime: time.Now(),
			}
		)

		if t.config.EnableMetrics {
			m.metricsRec = monitor.BeginRecord(ctx, cmd)
		}

		if t.config.EnableTraffic {
			m.trafficRec = logger.StartTrafficRec(ctx, &logger.ReqEntity{
				Typ: logger.TrafficTypSend,
				Cmd: cmd,
			})
		}

		db.Statement.Context = context.WithValue(ctx, trackingMetaCtxKey, m)
	}
}

func (t *tracker) End() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		var (
			ctx = db.Statement.Context
		)

		m, ok := ctx.Value(trackingMetaCtxKey).(*meta)
		if !ok || m == nil {
			return
		}

		if t.config.EnableMetrics && m.metricsRec != nil {
			m.metricsRec.EndWithError(db.Error)
		}

		if t.config.EnableTraffic && m.trafficRec != nil {
			m.trafficRec.End(&logger.RespEntity{
				Code: errorCode(db.Error),
				Msg:  errorMsg(db.Error),
			}, logger.Fields{
				"sql":  db.Statement.SQL.String(),
				"vars": db.Statement.Vars,
			})
		}

		if t.config.SlowLogFloor > 0 && m.startTime != (time.Time{}) {
			duration := time.Since(m.startTime)
			if duration > t.config.SlowLogFloor {
				logger.FromContext(ctx).
					WithError(db.Error).
					WithFields(logger.Fields{
						"sql":      db.Statement.SQL.String(),
						"vars":     db.Statement.Vars,
						"duration": duration.String(),
					}).Warn("slow query")
			}
		}
	}
}

func errorMsg(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func errorCode(err error) string {
	if err == nil {
		return "ok"
	}
	return "error"
}
