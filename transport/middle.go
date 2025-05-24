// file: mini/transport/middle.go
package transport

import (
	"fmt"
	"time"

	"github.com/rskv-p/mini/codec"
)

// ----------------------------------------------------
// Middleware type
// ----------------------------------------------------

// MiddlewareFunc wraps transport calls with pre/post processing.
type MiddlewareFunc func(subject string, data []byte, next func(string, []byte) error) error

// ----------------------------------------------------
// Trace middleware
// ----------------------------------------------------

// TraceMiddleware adds trace_id and logs trace information.
func TraceMiddleware() MiddlewareFunc {
	return func(subject string, data []byte, next func(string, []byte) error) error {
		msg := codec.NewMessage("")
		if err := codec.Unmarshal(data, msg); err != nil {
			return err
		}

		// ensure trace_id
		traceID := msg.GetString("trace_id")
		if traceID == "" {
			traceID = generateTraceID()
			msg.Set("trace_id", traceID)
		}

		// ensure contextID
		if msg.GetContextID() == "" {
			msg.SetContextID(generateTraceID())
		}

		data, _ = codec.Marshal(msg)

		fmt.Printf("[trace] → %s (trace_id=%s, ctx=%s)\n",
			subject, traceID, msg.GetContextID(),
		)

		return next(subject, data)
	}
}

// ----------------------------------------------------
// Logger middleware
// ----------------------------------------------------

// LoggerMiddleware logs subject, size, status and duration.
func LoggerMiddleware() MiddlewareFunc {
	return func(subject string, data []byte, next func(string, []byte) error) error {
		start := time.Now()
		err := next(subject, data)
		status := "OK"
		if err != nil {
			status = "ERR"
		}
		fmt.Printf("[log] %s (%d bytes) %s [%s]\n",
			subject, len(data), status, time.Since(start),
		)
		return err
	}
}
