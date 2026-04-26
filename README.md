# gokit

Go common kits — a monorepo of Go 1.21+ modules providing application bootstrap, observability, communication, data structures, and code generation for building production services.

## Module Overview

### Foundation

| Module                              | Purpose                                                                                |
|-------------------------------------|----------------------------------------------------------------------------------------|
| [functional](functional/)           | Generic functional programming utilities: `Map`, `Filter`, `Reduce`, `GroupBy`, `TopK` |
| [collection](collection/)           | Generic data structures: `Stack`, `Queue`, `PriorityQueue`, `Set`                      |
| [tracer](tracer/)                   | Context-based request ID propagation and debug/stress/shadow flags                      |
| [bloom](bloom/)                     | Probabilistic Bloom filter with Murmur3 hashing, tunable false-positive rate            |
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
| [cmd](cmd/)       | CLI-first bootstrap wrapping `urfave/cli/v2` — same primitives as `app`, exposed via CLI commands |

### Communication

| Module                              | Purpose                                                                                       |
|-------------------------------------|-----------------------------------------------------------------------------------------------|
| [ginext](ginext/)                   | Gin extension: request binding + validation, JWT auth, RPC interceptors, structured responses |
| [grpcext](grpcext/)                 | gRPC unary/stream interceptors for request tracing, traffic logging, and metrics              |
| [grpcetcd](grpcetcd/)               | etcd-based service registration (lease + keep-alive) and client-side resolver for gRPC        |
| [httpext](httpext/)                 | HTTP client with composable transport interceptor chain: headers, metrics, traffic, slow-log  |
| [genproto](genproto/)               | Shared protobuf types: `Auth`, `RequestHeader`, `ResponseHeader`                              |

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

### Code Generation

| Module                                    | Purpose                                                                                                   |
|-------------------------------------------|-----------------------------------------------------------------------------------------------------------|
| [protoc-gen-go-gin](protoc-gen-go-gin/)   | `protoc` plugin generating Gin HTTP handlers from protobuf services with `google.api.http` annotations    |

### Integrations

| Module                  | Purpose                                                                                     |
|-------------------------|---------------------------------------------------------------------------------------------|
| [notionx](notionx/)     | Markdown to Notion API block conversion for programmatic page creation                      |

## Dependency Graph

```
                        protoc-gen-go-gin
                           /         \
                       ginext       genproto
                      /  |   \
        annotation  functional  logger -- monitor -- tracer
           |                        |
          app        ┌─────────────┼─────────────┐
          cmd   grpcetcd         grpcext       cache
         (CLI)  (etcd gRPC)   (gRPC metrics)  (Redis/LRU)
                                 httpext
                              (HTTP client)
                                 gormext
                               (GORM plugin)
```

**Internal dependency summary** (direct edges):

| Module               | Internal Dependencies                           |
|----------------------|-------------------------------------------------|
| `app`                | `annotation`, `logger`                          |
| `cmd`                | `annotation`, `logger`                          |
| `ginext`             | `annotation`, `functional`, `logger`, `monitor`, `tracer` |
| `grpcext`            | `logger`, `monitor`, `tracer`                   |
| `grpcetcd`           | `logger`                                         |
| `httpext`            | `logger`, `monitor`, `tracer`                   |
| `gormext`            | `logger`, `monitor`, `tracer`                   |
| `cache`              | `logger`, `monitor`, `tracer`                   |
| `protoc-gen-go-gin`  | `ginext`, `genproto`                            |
| `functional`         | _(none)_                                         |
| `collection`         | _(none)_                                         |
| `tracer`             | _(none)_                                         |
| `bloom`              | _(none)_                                         |
| `annotation`         | _(none)_                                         |
| `async`              | _(none)_                                         |
| `retriever`          | _(none)_                                         |
| `notionx`            | _(none)_                                         |

**Layers:**

1. **Foundation** — `functional`, `collection`, `tracer`, `bloom`, `annotation` have zero internal dependencies.
2. **Observability** — `logger`, `monitor`, `tracer` form the observability backbone used by most mid-tier modules.
3. **Mid-tier** — `app`, `cmd`, `ginext`, `grpcext`, `grpcetcd`, `httpext`, `gormext`, `cache` compose foundation + observability.
4. **Top** — `protoc-gen-go-gin` sits at the highest level, depending on `ginext` and `genproto`.

**Key external dependencies:**

- `go.uber.org/zap` — structured logging (`logger`)
- `github.com/gin-gonic/gin` — HTTP framework (`ginext`)
- `google.golang.org/grpc` — gRPC (`grpcext`, `grpcetcd`)
- `go.etcd.io/etcd/client/v3` — service discovery (`grpcetcd`)
- `gorm.io/gorm` — ORM (`gormext`)
- `github.com/go-redis/redis/v8` — Redis client (`cache`)
- `github.com/prometheus/client_golang` — metrics (`monitor`, `app`, `cmd`)
- `github.com/urfave/cli/v2` — CLI framework (`cmd`)
- `google.golang.org/protobuf` — protobuf runtime (`genproto`, `grpcext`, `protoc-gen-go-gin`)

## Development

### Requirements

- `wire` for DI generation. Install from https://github.com/google/wire
- `go-enum` for enum generation. Install from https://github.com/abice/go-enum
- `gci` for Go import formatting. Install from https://github.com/daixiang0/gci
- `mockery` for mock generation. Install from https://github.com/vektra/mockery

### Workspace

This repository uses Go workspaces (`go.work`). All submodules are listed in the workspace and can be developed together. Each module maintains its own `go.mod` with independent versioning.
