// file: mini/transport/io.go
package transport

import (
	"context"
	"fmt"
	"time"

	"github.com/rskv-p/mini/codec"
)

// ----------------------------------------------------
// Response handler type
// ----------------------------------------------------

type ResponseHandler func(codec.IMessage) error

// ----------------------------------------------------
// Request with retry and tracing
// ----------------------------------------------------

func (t *Transport) Request(subject string, req []byte, handler ResponseHandler) error {
	return t.RequestWithContext(context.Background(), subject, req, handler)
}

func (t *Transport) RequestWithContext(
	ctx context.Context,
	subject string,
	req []byte,
	handler ResponseHandler,
) error {
	if t.conn == nil {
		return ErrDisconnected
	}

	// ensure trace
	msg := codec.NewMessage("")
	if err := codec.Unmarshal(req, msg); err != nil {
		return err
	}
	setDefaultTrace(ctx, msg)
	traceID := msg.GetString("trace_id")
	req, _ = codec.Marshal(msg)

	base := func(subj string, data []byte) error {
		start := time.Now()
		respMsg, err := t.conn.Request(subj, data, t.opts.Timeout)

		if t.opts.Metrics != nil {
			t.opts.Metrics.IncCounter("transport_requests_total")
			t.opts.Metrics.AddLatency("transport_request_latency_ms", time.Since(start).Milliseconds())
			if err != nil {
				t.opts.Metrics.IncCounter("transport_requests_failed")
			}
		}

		if err == nil && handler != nil {
			return handler(respMsg)
		}
		return err
	}

	return t.retry("Request", subject, traceID, req, base)
}

// ----------------------------------------------------
// Publish with retry and tracing
// ----------------------------------------------------

func (t *Transport) Publish(subject string, data []byte) error {
	if t.conn == nil {
		return ErrDisconnected
	}

	msg := codec.NewMessage("")
	_ = codec.Unmarshal(data, msg)
	setDefaultTrace(context.Background(), msg)
	traceID := msg.GetString("trace_id")
	data, _ = codec.Marshal(msg)

	return t.retry("Publish", subject, traceID, data, t.conn.Publish)
}

// ----------------------------------------------------
// Retry logic
// ----------------------------------------------------

func (t *Transport) retry(
	label string,
	subject string,
	traceID string,
	data []byte,
	fn func(string, []byte) error,
) error {
	policy := t.opts.RetryPolicies[subject]
	if policy.MaxAttempts == 0 {
		policy.MaxAttempts = defaultRetryAttempts
	}
	if policy.Delay == 0 {
		policy.Delay = defaultRetryDelay
	}

	call := t.wrapChain(fn)
	var lastErr error
	delay := policy.Delay

	for attempt := 0; attempt <= policy.MaxAttempts; attempt++ {
		err := call(subject, data)
		if err == nil {
			if label == "Publish" && t.opts.Metrics != nil {
				t.opts.Metrics.IncCounter("transport_publish_total")
			}
			return nil
		}

		lastErr = err
		if t.opts.Metrics != nil && label == "Publish" {
			t.opts.Metrics.IncCounter("transport_publish_failed")
		}
		if t.opts.OnRetry != nil {
			t.opts.OnRetry(subject, attempt, err)
		}
		if !t.opts.AutoReconnect || attempt == policy.MaxAttempts {
			break
		}
		if t.opts.Logger != nil {
			t.opts.Logger.Warn("retry %s [%d/%d] after %v: %v (trace_id=%s)",
				label, attempt+1, policy.MaxAttempts, delay, err, traceID)
		}
		time.Sleep(delay)
		delay *= 2
		_ = t.reconnect()
	}

	if t.opts.OnFailure != nil {
		t.opts.OnFailure(subject, lastErr)
	}
	if t.opts.DeadLetterHandler != nil {
		t.opts.DeadLetterHandler(subject, data, lastErr)
	}
	return lastErr
}

// ----------------------------------------------------
// Respond and broadcast utilities
// ----------------------------------------------------

func (t *Transport) Respond(replyTo string, msg codec.IMessage) error {
	if t.conn == nil {
		return ErrDisconnected
	}
	setDefaultTrace(context.Background(), msg)
	traceID := msg.GetString("trace_id")
	if t.opts.Debug {
		fmt.Printf("[trace] respond → %s (trace_id=%s, ctx=%s)\n",
			replyTo, traceID, msg.GetContextID())
	}
	data, err := codec.Marshal(msg)
	if err != nil {
		return err
	}
	return t.Publish(replyTo, data)
}

func (t *Transport) Broadcast(subjects []string, data []byte) error {
	msg := codec.NewMessage("")
	if err := codec.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("broadcast unmarshal: %w", err)
	}
	setDefaultTrace(context.Background(), msg)
	traceID := msg.GetString("trace_id")
	data, _ = codec.Marshal(msg)

	if t.opts.Debug {
		fmt.Printf("[trace] broadcast → %d targets (trace_id=%s)\n", len(subjects), traceID)
	}
	for _, subj := range subjects {
		if err := t.Publish(subj, data); err != nil {
			return fmt.Errorf("broadcast to %s: %w", subj, err)
		}
	}
	return nil
}

func (t *Transport) Broadcastf(format string, keys ...string) error {
	for _, k := range keys {
		subj := fmt.Sprintf(format, k)
		if t.opts.Debug {
			fmt.Printf("[broadcastf] → %s\n", subj)
		}
		if err := t.Publish(subj, nil); err != nil {
			return fmt.Errorf("broadcastf to %s: %w", subj, err)
		}
	}
	return nil
}

// ----------------------------------------------------
// Trace helpers
// ----------------------------------------------------

func setDefaultTrace(ctx context.Context, msg codec.IMessage) {
	traceID := msg.GetString("trace_id")
	if traceID == "" {
		traceID = TraceIDFromContext(ctx)
		if traceID == "" {
			traceID = generateTraceID()
		}
		msg.Set("trace_id", traceID)
	}
	if msg.GetContextID() == "" {
		msg.SetContextID(traceID)
	}
}

func generateTraceID() string {
	return fmt.Sprintf("trace-%d", time.Now().UnixNano())
}

func WithTrace(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceKey{}, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	if v := ctx.Value(traceKey{}); v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func (t *Transport) wrapChain(fn func(string, []byte) error) func(string, []byte) error {
	return func(subject string, data []byte) error {
		handler := func(ctx context.Context, subject string, data []byte) error {
			return fn(subject, data)
		}
		for i := len(t.middlewares) - 1; i >= 0; i-- {
			handler = t.middlewares[i](handler)
		}
		return handler(context.Background(), subject, data)
	}
}

type traceKey struct{}
