package core

import (
	"encoding/json"
	"fmt"
)

type Validator interface {
	Validate(data any) error // Validates the given data
}

// DecodeAndValidate decodes JSON data and validates it.
func (s *service) DecodeAndValidate(req *request, v any) error {
	raw := req.Data()
	if err := json.Unmarshal(raw, v); err != nil { // Check if JSON is valid
		if s.Logger != nil {
			s.Logger.Errorw("invalid JSON", "err", err, "raw", string(raw)) // Log error
		}
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if s.Validator != nil { // Validate if a validator is set
		if err := s.Validator.Validate(v); err != nil {
			if s.Logger != nil {
				s.Logger.Warnw("validation failed", "err", err, "raw", string(raw)) // Log validation failure
			}
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
				if s.(*service).Logger != nil {
					s.(*service).Logger.Errorw("invalid request type (not *request)", "actual", fmt.Sprintf("%T", req)) // Log type error
				}
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
