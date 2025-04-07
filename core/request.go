package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rskv-p/mini/pkg/x_log"

	"github.com/nats-io/nats.go"
)

// Handler defines the interface for request handlers.
type Handler interface {
	Handle(Request)
}

// HandlerFunc is an adapter to use functions as handlers.
type HandlerFunc func(Request)

func (fn HandlerFunc) Handle(req Request) {
	fn(req)
}

// Request represents a service request.
type Request interface {
	Respond([]byte, ...RespondOpt) error
	RespondJSON(any, ...RespondOpt) error
	Error(code, description string, data []byte, opts ...RespondOpt) error
	Data() []byte
	Headers() Headers
	Subject() string
	Reply() string
}

// Headers wraps nats.Header.
type Headers nats.Header

// RespondOpt configures a response message.
type RespondOpt func(*nats.Msg)

// request is the default Request implementation.
type request struct {
	msg          *nats.Msg
	respondError error
	logger       x_log.Logger
}

type serviceError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// Errors returned by request methods.
var (
	ErrRespond         = errors.New("NATS error when sending response")
	ErrMarshalResponse = errors.New("marshaling response")
	ErrArgRequired     = errors.New("argument required")
)

// ContextHandler allows passing context into a handler.
func ContextHandler(ctx context.Context, fn func(context.Context, Request)) Handler {
	return HandlerFunc(func(req Request) {
		fn(ctx, req)
	})
}

// Respond sends raw data as a response.
func (r *request) Respond(data []byte, opts ...RespondOpt) error {
	respMsg := &nats.Msg{Data: data}
	for _, opt := range opts {
		opt(respMsg)
	}
	if err := r.msg.RespondMsg(respMsg); err != nil {
		r.respondError = fmt.Errorf("%w: %s", ErrRespond, err)
		// Log the error using global logger
		x_log.Error().Str("subject", r.msg.Subject).Err(err).Msg("failed to respond")
		return r.respondError
	}
	return nil
}

// RespondJSON sends a JSON response.
func (r *request) RespondJSON(value any, opts ...RespondOpt) error {
	data, err := json.Marshal(value)
	if err != nil {
		// Log the error using global logger
		x_log.Error().Str("subject", r.msg.Subject).Err(err).Msg("failed to marshal JSON")
		return ErrMarshalResponse
	}
	return r.Respond(data, opts...)
}

// Error sends an error response with code and message.
func (r *request) Error(code, description string, data []byte, opts ...RespondOpt) error {
	if code == "" {
		return fmt.Errorf("%w: error code", ErrArgRequired)
	}
	if description == "" {
		return fmt.Errorf("%w: description", ErrArgRequired)
	}

	msg := &nats.Msg{
		Header: nats.Header{
			ErrorHeader:     []string{description},
			ErrorCodeHeader: []string{code},
		},
		Data: data,
	}

	for _, opt := range opts {
		opt(msg)
	}

	if err := r.msg.RespondMsg(msg); err != nil {
		r.respondError = err
		// Log the error using global logger
		x_log.Error().Str("subject", r.msg.Subject).Str("code", code).
			Str("description", description).Err(err).Msg("failed to send error response")
		return err
	}

	r.respondError = &serviceError{
		Code:        code,
		Description: description,
	}
	return nil
}

// WithHeaders adds headers to a response message.
func WithHeaders(headers Headers) RespondOpt {
	return func(m *nats.Msg) {
		if m.Header == nil {
			m.Header = nats.Header(headers)
			return
		}
		for k, v := range headers {
			m.Header[k] = v
		}
	}
}

// Data returns the request payload.
func (r *request) Data() []byte {
	return r.msg.Data
}

// Headers returns request headers.
func (r *request) Headers() Headers {
	return Headers(r.msg.Header)
}

// Subject returns the request subject.
func (r *request) Subject() string {
	return r.msg.Subject
}

// Reply returns the request reply subject.
func (r *request) Reply() string {
	return r.msg.Reply
}

// Get returns the first header value for a key.
func (h Headers) Get(key string) string {
	return nats.Header(h).Get(key)
}

// Values returns all header values for a key.
func (h Headers) Values(key string) []string {
	return nats.Header(h).Values(key)
}

// Error returns string representation of a serviceError.
func (e *serviceError) Error() string {
	return fmt.Sprintf("%s:%s", e.Code, e.Description)
}
