# 🧩 `mini` — Core Utilities for Modular Microservices

The `mini/` package is a lightweight yet powerful foundation for building modular, event-driven microservices in Go. It provides standardized messaging, in-process transport over NSQ, dynamic routing, structured logging, service discovery, and resilient panic-safe execution.

---

## 📦 Directory Overview

```txt
mini/
├── codec/       # Typed messages (Message, IMessage)
├── config/      # JSON+ENV config loader with fallbacks
├── constant/    # Shared constants and error types
├── context/     # Request lifecycle and response tracking
├── logger/      # Structured and contextual logger
├── recover/     # Safe execution utilities
├── registry/    # In-memory service registry
├── router/      # Declarative message routing
├── selector/    # Service node selection strategies
└── transport/   # NSQ-based message transport with file support
```

---

## 📡 `transport/` — Message Transport Layer

Built on top of NSQ and abstracted via `ITransport`. Includes:

* `Publish`, `Request`, `Respond`, `Broadcast`
* Topic and prefix subscriptions (`SubscribeTopic`, `SubscribePrefix`)
* Retry policies per topic/subject
* Middleware support (context-aware)
* File chunking (`SendFile`, `ReceiveFileWithHooks`)

Backed by a flexible `Conn` layer for producer/consumer + reply channels.

---

## 📨 `codec/` — Typed Messages

Standard message format used throughout the system:

* Core fields: `Type`, `Node`, `ContextID`, `ReplyTo`, `Headers`, `Body`
* Type-safe accessors: `GetString`, `GetInt`, `GetBool`, etc.
* `SetError`, `SetResult`, `Validate`, `Copy`
* `RawBody` support for low-level access
* Interface: `IMessage`

---

## 🔒 `constant/` — Standard Constants

* Message types: `request`, `response`, `stream`, `event`, `health`
* Errors: `ErrNotFound`, `ErrEmptyMessage`, `ErrInvalidPath`
* Special keys: `error`, `result`, `input`
* Limits: `MaxFileChunkSize = 2MB`
* Health status: `StatusOK`, `StatusWarning`, `StatusCritical`

---

## 🧠 `context/` — Request Lifecycle Manager

Maps `ContextID → Conversation` with support for:

* Response tracking for `Request/Respond` pairs
* Automatic TTL cleanup and lifecycle hooks
* Methods: `Add`, `Get`, `Done`, `WaitTimeout`, `Range`

Used internally to ensure responses are correctly routed.

---

## 🔧 `config/` — Config Loader

* Supports JSON files with `${ENV_VAR}` interpolation
* Env variable fallbacks (e.g. `SRV_LOG_LEVEL`)
* Methods: `MustString`, `MustInt`, `Has`, `Dump`
* Automatically injects defaults for missing values

---

## 🪵 `logger/` — Structured Logger

* Contextual and leveled: `Debug`, `Info`, `Warn`, `Error`
* Add metadata: `With(key, value)`, `WithContext(traceID)`
* Interfaces: `ILogger`, `LoggerEntry`
* Configurable log level: `SetLevel("warn")`

---

## 🔁 `registry/` — Service Registry

In-memory registry with optional plugin support:

* Register/Deregister services and nodes
* TTL-based cleanup
* Watchers for live updates
* Introspectable service state

---

## 🎯 `selector/` — Service Node Selector

* Node selection strategies: `RoundRobin`, `Random`, `First`
* Metadata-based filtering
* Internal caching (`cacheTTL`) for faster resolution

---

## 🚦 `router/` — Dynamic Routing

* Declarative routing via `IAction` and `IRouter`
* Register handlers dynamically
* Middleware support: `HandlerWrapper`
* Input validation: required fields, type checks, custom rules
* Hooks: `OnErrorHook`, `OnNotFound`

---

## 🛡️ `recover/` — Panic Protection

* `RecoverWithContext()` for panic-resilient routing
* `Safe("label", fn)` for safe goroutines
* Global panic hook
* Keeps the system alive even on handler failure

---

## 📈 Metrics (in `service/`)

* Built-in counters: `IncMetric`, `AddMetric`, `SetMetric`
* Snapshot: `ExportMetrics()` as `map[string]float64`
* Scoped recording: `.WithMetricPrefix("db.")`

---

## 🩺 Health Monitoring

* Uses `gopsutil` to monitor:

  * Memory (free %, thresholds)
  * CPU load (load5 per core)
* Thresholds configurable via `config`
* Register custom health probes with `RegisterHealthProbe`

---

## 💡 Example

```go
log := logger.NewLogger("auth", "debug")

bus := transport.New(
    transport.Addrs("127.0.0.1:4150"),
    transport.Subject("auth.v1"),
    transport.WithLogger(log),
    transport.WithDebug(),
)

svc := service.NewService("auth", "v1",
    service.Transport(bus),
    service.Logger(log),
)

svc.RegisterAction("auth.login", []service.InputSchemaField{
    {Name: "email", Type: "string", Required: true},
    {Name: "password", Type: "string", Required: true},
}, func(ctx context.Context, input map[string]any) (any, error) {
    return map[string]any{"token": "jwt123"}, nil
})

_ = svc.Init()
_ = svc.Run()
```

---

## ✅ Features

* In-process, type-safe transport using NSQ
* Schema-based request validation and introspection
* Middleware chaining for actions and handlers
* File transfer over pub/sub (chunked)
* Dynamic service discovery and routing
* Built-in metrics, health checks, and error recovery

---

## 🧪 Ideal Use Cases

* Internal service-to-service communication
* Background workers (async `Request`/`Respond`)
* File processing pipelines
* Gateway → Bus → Worker architecture
* Auth, logging, streaming, task queues