# gokit

Go common kits — a monorepo of Go 1.24+ modules providing application bootstrap, observability, communication, data structures, and code generation for building production services.

## Module Overview

### Foundation

| Module                              | Purpose                                                                                |
|-------------------------------------|----------------------------------------------------------------------------------------|
| [functional](functional/)           | Generic functional programming utilities: `Map`, `Filter`, `Reduce`, `GroupBy`, `TopK` |
| [collection](collection/)           | Generic data structures: `Stack`, `Queue`, `PriorityQueue`, `Set`                      |
| [tracer](tracer/)                   | Context-based request ID propagation and debug/stress/shadow flags                      |
| [annotation](annotation/)           | Struct tag-based binding (`bind`/`json`/`protobuf`), defaults, and validation rules     |

### Observability

| Module                  | Purpose                                                                              |
|-------------------------|--------------------------------------------------------------------------------------|
| [logger](logger/)       | Structured logging wrapping `go.uber.org/zap` with rotation, rate limiting, trimming |
| [monitor](monitor/)     | Prometheus metrics for single-flight patterns: histograms, counters, gauges, summaries |

### Application Bootstrap

| Module            | Purpose                                                                                                                  |
|-------------------|--------------------------------------------------------------------------------------------------------------------------|
| [app](app/)       | Lifecycle framework: config loading (YAML/JSON/dotenv), admin HTTP endpoints (pprof, Prometheus, health), graceful shutdown |

### Communication

| Module                              | Purpose                                                                                       |
|-------------------------------------|-----------------------------------------------------------------------------------------------|
| [ginext](ginext/)                   | Gin extension: request binding + validation, JWT auth, RPC interceptors, structured responses |
| [httpext](httpext/)                 | HTTP client with composable transport interceptor chain: headers, metrics, traffic, slow-log  |

### Data & Persistence

| Module                  | Purpose                                                                                             |
|-------------------------|-----------------------------------------------------------------------------------------------------|
| [gormext](gormext/)     | GORM plugin: tracing, traffic logging, Prometheus metrics, error logging, slow-query detection      |
| [cache](cache/)         | Caching abstraction with pluggable backends: in-memory map, generic LRU (with TTL), and Redis      |

### Concurrency & Resilience

| Module                      | Purpose                                                                                    |
|------------------------------|--------------------------------------------------------------------------------------------|
| [async](async/)              | Generic concurrent job runner: `AllOf`, `AnyOf`, panic-safe, error-group-based termination  |
| [retriever](retriever/)      | Configurable retry with pluggable backoff (exponential + jitter, linear, none)              |

## Dependency Graph

```
        annotation (v3)         functional
              |                      |
            app        ginext ──────────────── logger ── monitor ── tracer
                      |  \
                   httpext / gormext / cache  ── (observability backends)
```

**Internal dependency summary** (direct edges):

| Module               | Internal Dependencies                           |
|----------------------|-------------------------------------------------|
| `app`                | `annotation`, `logger`                          |
| `ginext`             | `annotation`, `logger`, `monitor`, `tracer`     |
| `httpext`            | `logger`, `monitor`, `tracer`                   |
| `gormext`            | `logger`, `monitor`, `tracer`                   |
| `cache`              | `logger`, `monitor`, `tracer`                   |
| `functional`         | _(none)_                                         |
| `collection`         | _(none)_                                         |
| `tracer`             | _(none)_                                         |
| `annotation`         | _(none)_                                         |
| `async`              | _(none)_                                         |
| `retriever`          | _(none)_                                         |

> `annotation` has been upgraded to `v3`; all other modules remain on `v2`.

**Layers:**

1. **Foundation** — `functional`, `collection`, `tracer`, `annotation` have zero internal dependencies.
2. **Observability** — `logger`, `monitor`, `tracer` form the observability backbone used by most mid-tier modules.
3. **Mid-tier** — `app`, `ginext`, `httpext`, `gormext`, `cache` compose foundation + observability.

**Key external dependencies:**

- `go.uber.org/zap` — structured logging (`logger`)
- `github.com/gin-gonic/gin` — HTTP framework (`ginext`)
- `gorm.io/gorm` — ORM (`gormext`)
- `github.com/go-redis/redis/v8` — Redis client (`cache`)
- `github.com/prometheus/client_golang` — metrics (`monitor`, `app`)

## Development

### Requirements

- `wire` for DI generation. Install from https://github.com/google/wire
- `go-enum` for enum generation. Install from https://github.com/abice/go-enum
- `gci` for Go import formatting. Install from https://github.com/daixiang0/gci
- `mockery` for mock generation. Install from https://github.com/vektra/mockery

### Workspace

This repository uses Go workspaces (`go.work`). All submodules are listed in the workspace and can be developed together. Each module maintains its own `go.mod` with independent versioning.
