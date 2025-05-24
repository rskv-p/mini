// file: arc/service/transport/io.go
package transport

import (
	"context"
	"fmt"
	"time"

	"github.com/rskv-p/mini/service/codec"
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

	msg := codec.NewMessage("")
	if err := codec.Unmarshal(req, msg); err != nil {
		return err
	}

	traceID := msg.GetString("trace_id")
	if traceID == "" {
		traceID = TraceIDFromContext(ctx)
		if traceID == "" {
			traceID = generateTraceID()
		}
		msg.Set("trace_id", traceID)
		req, _ = codec.Marshal(msg)
	}

	policy := t.opts.RetryPolicies[subject]
	max := policy.MaxAttempts
	if max == 0 {
		max = defaultRetryAttempts
	}
	delay := policy.Delay
	if delay == 0 {
		delay = defaultRetryDelay
	}

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
	call := t.wrapChain(base)

	var lastErr error
	currDelay := delay
	for i := 0; i <= max; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := call(subject, req)
		if err == nil {
			return nil
		}
		lastErr = err
		if t.opts.OnRetry != nil {
			t.opts.OnRetry(subject, i, err)
		}
		if !t.opts.AutoReconnect || i == max {
			break
		}
		if t.opts.Logger != nil {
			t.opts.Logger.Warn(
				"retry Request [%d/%d] after %v: %v (trace_id=%s)",
				i+1, max, currDelay, err, traceID,
			)
		}
		time.Sleep(currDelay)
		currDelay *= 2
		_ = t.reconnect()
	}

	if t.opts.OnFailure != nil {
		t.opts.OnFailure(subject, lastErr)
	}
	if t.opts.DeadLetterHandler != nil {
		t.opts.DeadLetterHandler(subject, req, lastErr)
	}
	return lastErr
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

	traceID := msg.GetString("trace_id")
	if traceID == "" {
		traceID = generateTraceID()
		msg.Set("trace_id", traceID)
		data, _ = codec.Marshal(msg)
	}

	policy := t.opts.RetryPolicies[subject]
	max := policy.MaxAttempts
	if max == 0 {
		max = defaultRetryAttempts
	}
	delay := policy.Delay
	if delay == 0 {
		delay = defaultRetryDelay
	}

	call := t.wrapChain(t.conn.Publish)
	var lastErr error
	currDelay := delay

	for i := 0; i <= max; i++ {
		err := call(subject, data)
		if err == nil {
			if t.opts.Metrics != nil {
				t.opts.Metrics.IncCounter("transport_publish_total")
			}
			return nil
		}
		lastErr = err
		if t.opts.Metrics != nil {
			t.opts.Metrics.IncCounter("transport_publish_failed")
		}
		if t.opts.OnRetry != nil {
			t.opts.OnRetry(subject, i, err)
		}
		if !t.opts.AutoReconnect || i == max {
			break
		}
		if t.opts.Logger != nil {
			t.opts.Logger.Warn(
				"retry Publish [%d/%d] after %v: %v (trace_id=%s)",
				i+1, max, currDelay, err, traceID,
			)
		}
		time.Sleep(currDelay)
		currDelay *= 2
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
	if msg.GetString("trace_id") == "" {
		msg.Set("trace_id", generateTraceID())
	}
	traceID := msg.GetString("trace_id")
	if t.opts.Debug {
		fmt.Printf("[trace] respond → %s (trace_id=%s, ctx=%s)\n",
			replyTo, traceID, msg.GetContextID(),
		)
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
	if msg.GetString("trace_id") == "" {
		msg.Set("trace_id", generateTraceID())
		data, _ = codec.Marshal(msg)
	}
	traceID := msg.GetString("trace_id")
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
// Trace ID helpers
// ----------------------------------------------------

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

type traceKey struct{}
