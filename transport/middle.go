// file: mini/transport/middle.go
package transport

import (
	"context"
	"fmt"
	"time"

	"github.com/rskv-p/mini/codec"
)

// ----------------------------------------------------
// Middleware types
// ----------------------------------------------------

// MiddlewareFunc wraps a TransportHandler with cross-cutting logic.
type MiddlewareFunc func(next TransportHandler) TransportHandler

// ----------------------------------------------------
// Trace middleware
// ----------------------------------------------------

// TraceMiddleware injects trace_id and context_id into the message and logs routing.
func TraceMiddleware() MiddlewareFunc {
	return func(next TransportHandler) TransportHandler {
		return func(ctx context.Context, subject string, data []byte) error {
			msg := codec.NewMessage("")
			if err := codec.Unmarshal(data, msg); err != nil {
				return err
			}

			// Ensure trace_id
			traceID := msg.GetString("trace_id")
			if traceID == "" {
				traceID = generateTraceID()
				msg.Set("trace_id", traceID)
			}

			// Ensure context_id
			if msg.GetContextID() == "" {
				msg.SetContextID(generateTraceID())
			}

			fmt.Printf("[trace] → %s (trace_id=%s, ctx_id=%s)\n",
				subject, traceID, msg.GetContextID())

			data, _ = codec.Marshal(msg)
			return next(ctx, subject, data)
		}
	}
}

// ----------------------------------------------------
// Logger middleware
// ----------------------------------------------------

// LoggerMiddleware logs subject, size, duration and status.
func LoggerMiddleware() MiddlewareFunc {
	return func(next TransportHandler) TransportHandler {
		return func(ctx context.Context, subject string, data []byte) error {
			start := time.Now()
			err := next(ctx, subject, data)
			status := "OK"
			if err != nil {
				status = "ERR"
			}

			fmt.Printf("[log] %s (%d bytes) %s [%s]\n",
				subject, len(data), status, time.Since(start))
			return err
		}
	}
}
