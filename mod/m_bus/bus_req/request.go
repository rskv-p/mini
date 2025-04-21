package bus_req

import (
	"encoding/json"
	"fmt"
)

//---------------------
// Request
//---------------------

// Request represents an incoming request with associated handlers and metadata.
type Request struct {
	Subject string
	Reply   string
	Data    []byte
	headers Headers

	Respond  func([]byte) error
	RespJSON func(any) error
	Err      func(code, msg string, data any) error
}

// RespondJSON sends a JSON-encoded response to the Reply address.
func (r *Request) RespondJSON(v any) error {
	// Marshal the value into JSON
	data, err := json.Marshal(v)
	if err != nil {
		// x_log.RootLogger().Structured().Error("failed to marshal JSON",
		// 	x_log.FString("subject", r.Subject),
		// 	x_log.FError(err),
		// )
		return fmt.Errorf("marshal error: %w", err)
	}

	// Send the response
	if err := r.Respond(data); err != nil {
		// x_log.RootLogger().Structured().Error("failed to send response",
		// 	x_log.FString("subject", r.Subject),
		// 	x_log.FError(err),
		// )
		return err
	}

	// Log success
	//	x_log.RootLogger().Structured().Info("response sent successfully",
	//		x_log.FString("subject", r.Subject),
	//	)
	return nil
}

// Error sends a standard error payload as JSON.
func (r *Request) Error(code, description string, data []byte) error {
	if r.Err != nil {
		return r.Err(code, description, data)
	}

	// Send a default error response
	errorResponse := map[string]string{
		"error":       code,
		"description": description,
	}
	return r.RespondJSON(errorResponse)
}

// NewTestRequest creates a mock Request for testing purposes.
func NewTestRequest(subject string, data []byte) *Request {
	return &Request{
		Subject:  subject,
		Data:     data,
		headers:  make(Headers),
		Respond:  func(_ []byte) error { return nil },
		RespJSON: func(_ any) error { return nil },
		Err: func(code, msg string, _ any) error {
			return nil
		},
	}
}

// SetErrorHandler sets a custom error handler for the request.
func (r *Request) SetErrorHandler(f func(code, msg string, data any) error) {
	r.Err = f
}

//---------------------
// Headers
//---------------------

// Headers represents HTTP-like headers for the request.
type Headers map[string][]string

// Get retrieves the first value for the given key from the headers.
func (h Headers) Get(key string) string {
	if values, exists := h[key]; exists && len(values) > 0 {
		return values[0]
	}
	return ""
}

// Set sets the value for a header key.
func (h Headers) Set(key, value string) {
	h[key] = []string{value}
}

// SetHeader sets a header for the request.
func (r *Request) SetHeader(key, value string) {
	if r.headers == nil {
		r.headers = make(Headers)
	}
	r.headers[key] = []string{value}
}

// Headers returns the headers associated with the request.
func (r *Request) Headers() Headers {
	return r.headers
}
