
# 🧩 `mini` — Core Utilities for Modular Microservices

The `mini/` package provides a suite of reusable building blocks for message-based, modular microservices. It includes NSQ-powered transport, dynamic routing, in-memory service registry, error recovery, and structured logging.

---

## 📦 Directory Structure

```text
mini/
├── codec/       # Standardized message format
├── config/      # Configuration loading and defaults
├── constant/    # Shared constants and error definitions
├── context/     # Request lifecycle manager (conversation context)
├── logger/      # Structured leveled logger
├── recover/     # Safe function execution and panic recovery
├── registry/    # In-memory service registry
├── router/      # Declarative message routing with validation
├── selector/    # Service node selection strategies
└── transport/   # NSQ-based transport: publish, request, file streams
````

---

## 📡 `transport/`

Message transport layer over NSQ.

* `Transport` (implements `ITransport`) supports:

  * `Publish`, `Request`, `Respond`, `Broadcast`
  * `Subscribe`, `SubscribeTopic`, `SubscribePrefix`
  * Retry policies, tracing, middleware
* File chunking support via `SendFile` and `ReceiveFileRouter`
* Underlying `Conn` abstracts NSQ producer/consumer + reply channels

---

## 📨 `codec/`

Unified `Message` structure for all communications:

* Standard fields: `Type`, `Node`, `ContextID`, `ReplyTo`
* `Header`: for metadata (e.g., trace IDs, message type)
* `Body`: dynamic JSON content with typed accessors (`GetString`, `GetInt`, etc.)
* `RawBody`: raw JSON fallback
* Helpers: `SetResult`, `SetError`, `Validate`, `Copy`, etc.

---

## 🔒 `constant/`

Shared constants and standardized error definitions:

* Message types: `request`, `response`, `stream`, `event`
* Common errors: `ErrNotFound`, `ErrEmptyMessage`, `ErrInvalidPath`
* Standard keys: `result`, `error`, `input`
* Max file chunk size, health status flags, and internal status codes

---

## 🧠 `context/`

Manages conversation lifecycle (`ContextID` → `Conversation`):

* Methods: `Add`, `Get`, `Done`, `WaitTimeout`, `Range`
* Supports auto-deletion and custom hooks (`onAdd`, `onDelete`)
* Used for managing response expectations via `ContextID`

---

## 🔧 `config/`

Configuration loader with support for:

* JSON config files with env var injection (`${VAR}`)
* Environment variable fallbacks (`SRV_*`)
* `Config` structure includes:

  * Global options: `ServiceName`, `LogLevel`, `Port`, etc.
  * NSQ settings: TCP/HTTP address, queue size, buffer, etc.
* Includes `Validate()`, `Dump()`, and default loaders

---

## 🪵 `logger/`

Structured logging engine with levels and context:

* Interface: `ILogger`

  * `Debug`, `Info`, `Warn`, `Error`
  * `WithContext(traceID)`, `With(key, value)`
  * Level filtering (`SetLevel("warn")`)
* Built-in formatting with sorted metadata

---

## 🔁 `registry/`

In-memory service registry for discovery:

* `Register`, `Deregister`, `GetService`, `ListServices`
* TTL-based cleanup of stale nodes
* Support for watchers (reactive updates on registry changes)
* Dumps current state for diagnostics

---

## 🎯 `selector/`

Picks a service node using strategies and filters:

* Strategies: `RoundRobin`, `Random`, `First`
* Filters: match metadata (`MatchMeta`)
* Uses internal cache (`cacheTTL`) to reduce registry calls

---

## 🚦 `router/`

Declarative routing engine for incoming messages:

* Interfaces:

  * `IRouter` for dynamic route registration
  * `IAction` for declarative actions with validation
* Features:

  * `RegisterActions`, `Dispatch`, `Add`, `Deregister`
  * Middleware support (`HandlerWrapper`)
  * Input validation: `required`, `min`, `max` with custom error messages
* Error hook (`OnErrorHook`) and not-found handler (`OnNotFound`)

---

## 🛡️ `recover/`

Utilities for safe execution and panic recovery:

* `RecoverHandler()` for safe routing
* `Safe(label, fn)` for safe goroutines
* `WrapRecover()` for context-aware functions
* Custom `OnPanic` hook for global crash tracking

---

## 🔗 Example

```go
import (
    "github.com/rskv-p/mini/service/transport"
    "github.com/rskv-p/mini/service/logger"
)

func main() {
    log := logger.NewLogger("auth", "debug")
    t := transport.New(
        transport.Addrs("127.0.0.1:4150"),
        transport.Subject("auth"),
        transport.WithLogger(log),
        transport.WithDebug(),
    )
    _ = t.Init()
    // ...
}
```

---

## ✅ Features

* In-process transport with NSQ backend
* Fully typed messages with traceable metadata
* Retry support and circuit-breaker-style hooks
* File streaming and context-bound responses
* Extensible registry/selector/router abstraction layers