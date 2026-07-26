module httpext-v3-example

go 1.24

require (
	github.com/tenz-io/gokit/httpext/v3 v3.0.0
	github.com/tenz-io/gokit/logger/v3 v3.0.0
	github.com/tenz-io/gokit/monitor/v3 v3.0.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/tenz-io/gokit/tracer/v3 v3.0.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)

// The v3 gokit modules are not published yet; resolve httpext/v3 from the
// parent dir and its transitive v3 deps from their sibling dirs (three levels
// up: example -> v3 -> httpext -> repo root), so this example module builds
// standalone (GOWORK=off) as well as in the workspace.
replace (
	github.com/tenz-io/gokit/httpext/v3 => ./..
	github.com/tenz-io/gokit/logger/v3 => ../../../logger/v3
	github.com/tenz-io/gokit/monitor/v3 => ../../../monitor/v3
	github.com/tenz-io/gokit/tracer/v3 => ../../../tracer/v3
)
