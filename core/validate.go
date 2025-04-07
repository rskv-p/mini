package core

import (
	"encoding/json"
	"fmt"

	"github.com/rskv-p/mini/pkg/x_log"
)

type Validator interface {
	Validate(data any) error // Validates the given data
}

// DecodeAndValidate decodes JSON data and validates it.
func (s *service) DecodeAndValidate(req *request, v any) error {
	raw := req.Data()
	if err := json.Unmarshal(raw, v); err != nil { // Check if JSON is valid
		// Log invalid JSON using global logger
		x_log.Error().Err(err).Str("raw", string(raw)).Msg("invalid JSON")
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if s.Validator != nil { // Validate if a validator is set
		if err := s.Validator.Validate(v); err != nil {
			// Log validation failure using global logger
			x_log.Warn().Err(err).Str("raw", string(raw)).Msg("validation failed")
			return fmt.Errorf("validation failed: %w", err)
		}
	}
	return nil
}

// ValidateJSON creates a middleware to validate JSON data.
func ValidateJSON[T any](s Service) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(req Request) {
			var v T
			r, ok := req.(*request)
			if !ok { // Check if request is of type *request
				_ = req.Error("500", "internal error", nil)
				// Log invalid request type using global logger
				x_log.Error().Str("actual", fmt.Sprintf("%T", req)).Msg("invalid request type (not *request)")
				return
			}
			if err := s.(*service).DecodeAndValidate(r, &v); err != nil { // Decode and validate
				_ = req.Error("400", err.Error(), nil) // Return validation error
				return
			}
			next.Handle(req) // Proceed to the next handler
		})
	}
}
