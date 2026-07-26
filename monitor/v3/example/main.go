// Example: monitor/v3 — single-flight injection end-to-end.
//
// One Exporter is created at the request edge via monitor.Init and injected
// into the context; every downstream Begin along the call chain reuses that
// same Exporter. Run it and point a /metrics scrape at it to see the metrics.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/tenz-io/gokit/monitor/v3"
)

func main() {
	// Use a private registry so the example metrics don't leak into the
	// process default. In production, call monitor.Configure() once with the
	// default registry (or omit Configure entirely for the default fallback).
	registry := prometheus.NewRegistry()
	if err := monitor.Configure(monitor.WithRegistry(registry)); err != nil {
		log.Fatalf("configure: %v", err)
	}

	ctx := context.Background()

	// Simulate a few requests sharing the single-flight model.
	for i := 0; i < 3; i++ {
		handleGetUser(ctx)
	}

	// Dump the collected metrics so the example is self-contained.
	dumpMetrics(registry)
}

// handleGetUser is the request edge: it injects the Exporter and times the
// whole call under the "total" dsCmd. Downstream call sites reuse the ctx.
func handleGetUser(ctx context.Context) {
	// single-flight injection: create (first time) / reuse the Exporter.
	ctx = monitor.Init(ctx, "userService")
	rec := monitor.Begin(ctx, "total")
	defer rec.EndWithError(nil)

	if err := fetchFromCache(ctx); err != nil {
		// A cache miss is an expected outcome, not a failure — the cache
		// layer recorded it with opt="miss" and code=ok below. Fall through
		// to the DB, which is the real work.
		_ = err
	}

	_ = fetchFromDB(ctx)
}

// fetchFromCache always misses here. A miss is a cache-layer *result*, not an
// error, so it records code=ok with opt="miss" (matching the hit/miss/error
// opt convention). A genuine cache failure (e.g. redis down) would instead
// record code=err with opt="error".
func fetchFromCache(ctx context.Context) error {
	rec := monitor.Begin(ctx, "getUser") // reuses the injected Exporter
	defer rec.EndWithOpt("miss")         // code=ok, opt=miss
	return errCacheMiss
}

var errCacheMiss = errors.New("cache miss")

func fetchFromDB(ctx context.Context) error {
	rec := monitor.Begin(ctx, "db_query") // reuses the injected Exporter
	defer func() {
		time.Sleep(time.Millisecond) // pretend to do work
		rec.EndWithError(nil)
	}()
	return nil
}

func dumpMetrics(reg *prometheus.Registry) {
	mfs, err := reg.Gather()
	if err != nil {
		fmt.Println("gather error:", err)
		return
	}
	for _, mf := range mfs {
		fmt.Printf("# %s\n", mf.GetName())
		for _, m := range mf.GetMetric() {
			labels := make(map[string]string, len(m.GetLabel()))
			for _, l := range m.GetLabel() {
				labels[l.GetName()] = l.GetValue()
			}
			switch {
			case m.GetCounter() != nil:
				fmt.Printf("  counter %v = %v\n", labels, m.GetCounter().GetValue())
			case m.GetGauge() != nil:
				fmt.Printf("  gauge   %v = %v\n", labels, m.GetGauge().GetValue())
			case m.GetHistogram() != nil:
				fmt.Printf("  hist    %v samples=%d sum=%v\n",
					labels, m.GetHistogram().GetSampleCount(), m.GetHistogram().GetSampleSum())
			case m.GetSummary() != nil:
				fmt.Printf("  summary %v samples=%d sum=%v\n",
					labels, m.GetSummary().GetSampleCount(), m.GetSummary().GetSampleSum())
			}
		}
	}
}
