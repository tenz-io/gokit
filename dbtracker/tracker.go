package dbtracker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/tenz-io/gokit/logger"
	"github.com/tenz-io/gokit/monitor"
	"github.com/tenz-io/gokit/tracer"
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
	Apply(db *gorm.DB) error
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

func (t *tracker) begin(cmd string) func(db *gorm.DB) {
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

		if t.config.EnableTraffic || tracer.FromContext(ctx).IsDebug() {
			m.trafficRec = logger.StartTrafficRec(ctx, &logger.ReqEntity{
				Typ: logger.TrafficTypSend,
				Cmd: cmd,
			})
		}

		db.Statement.Context = context.WithValue(ctx, trackingMetaCtxKey, m)
	}
}

func (t *tracker) stop() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		var (
			ctx = db.Statement.Context
			le  = logger.FromContext(ctx).WithFields(logger.Fields{
				"sql":  db.Statement.SQL.String(),
				"vars": db.Statement.Vars,
			})
		)

		m, ok := ctx.Value(trackingMetaCtxKey).(*meta)
		if !ok || m == nil {
			le.Warn("tracker meta not found")
			return
		}

		if t.config.EnableMetrics && m.metricsRec != nil {
			m.metricsRec.EndWithError(db.Error)
		}

		if (t.config.EnableTraffic || tracer.FromContext(ctx).IsDebug()) &&
			m.trafficRec != nil {
			m.trafficRec.End(&logger.RespEntity{
				Code: errorCode(db.Error),
				Msg:  errorMsg(db.Error),
			}, logger.Fields{
				"sql":  db.Statement.SQL.String(),
				"vars": db.Statement.Vars,
			})
		}

		if t.config.EnableErrorLog && db.Error != nil {
			if errors.Is(db.Error, gorm.ErrRecordNotFound) {
				le.WithError(db.Error).Debugf("record not found")
			} else {
				le.WithError(db.Error).Error("db error")
			}
		}

		if t.config.SlowLogFloor > 0 && m.startTime != (time.Time{}) {
			duration := time.Since(m.startTime)
			if duration > t.config.SlowLogFloor {
				le.WithError(db.Error).
					WithFields(logger.Fields{
						"duration": duration.String(),
					}).Warn("slow query")
			}
		}
	}
}

func (t *tracker) Apply(db *gorm.DB) (err error) {
	var (
		callback = db.Callback()
	)
	if err = callback.Query().Before("*").Register("start_query", t.begin("db_query")); err != nil {
		return fmt.Errorf("register start_query error: %w", err)
	}

	if err = callback.Query().After("*").Register("end_query", t.stop()); err != nil {
		return fmt.Errorf("register end_query error: %w", err)
	}

	if err = callback.Create().Before("*").Register("start_create", t.begin("db_create")); err != nil {
		return fmt.Errorf("register start_create error: %w", err)
	}

	if err = callback.Create().After("*").Register("end_create", t.stop()); err != nil {
		return fmt.Errorf("register end_create error: %w", err)
	}

	if err = callback.Update().Before("*").Register("start_update", t.begin("db_update")); err != nil {
		return fmt.Errorf("register start_update error: %w", err)
	}

	if err = callback.Update().After("*").Register("end_update", t.stop()); err != nil {
		return fmt.Errorf("register end_update error: %w", err)
	}

	if err = callback.Delete().Before("*").Register("start_delete", t.begin("db_delete")); err != nil {
		return fmt.Errorf("register start_delete error: %w", err)
	}

	if err = callback.Delete().After("*").Register("end_delete", t.stop()); err != nil {
		return fmt.Errorf("register end_delete error: %w", err)
	}

	if err = callback.Row().Before("*").Register("start_row", t.begin("db_row")); err != nil {
		return fmt.Errorf("register start_row error: %w", err)
	}

	if err = callback.Row().After("*").Register("end_row", t.stop()); err != nil {
		return fmt.Errorf("register end_row error: %w", err)
	}

	return nil
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
	return "err"
}
