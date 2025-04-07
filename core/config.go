package core

import "github.com/rskv-p/mini/pkg/x_log"

// Middleware wraps a Handler function.
type Middleware func(Handler) Handler

// JWTVerifier verifies JWT tokens and returns claims if valid.
type JWTVerifier interface {
	VerifyJWT(token string) (map[string]any, error) // returns claims if ok
}

// Config holds service configuration.
type Config struct {
	Name               string            `json:"name"`                 // Service name
	Endpoint           *EndpointConfig   `json:"endpoint"`             // Endpoint config
	Version            string            `json:"version"`              // Service version
	Description        string            `json:"description"`          // Service description
	Metadata           map[string]string `json:"metadata,omitempty"`   // Optional metadata
	QueueGroup         string            `json:"queue_group"`          // Queue group
	QueueGroupDisabled bool              `json:"queue_group_disabled"` // Disable queue group

	StatsHandler StatsHandler // Optional custom stats handler
	DoneHandler  DoneHandler  // Handler called after Stop()
	ErrorHandler ErrHandler   // Optional error handler
	Middleware   []Middleware // Global middlewares
	Logger       x_log.Logger // Optional structured logger
	Validator    Validator    // Optional JSON payload validator
	JWTVerifier  JWTVerifier  // Optional JWT verifier

	// Lifecycle hooks
	OnStart func(Service)        // Called after Start()
	OnStop  func(Service)        // Called before DoneHandler
	OnError func(Service, error) // Called on internal error
}

// RequireJWT ensures the request contains a valid JWT token.
func RequireJWT(verifier JWTVerifier) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(req Request) {
			token := req.Headers().Get("Authorization") // Get token from header
			if token == "" {
				if l := reqLogger(req); l != nil {
					l.Warnw("JWT rejected: missing Authorization header", "subject", req.Subject())
				}
				_ = req.Error("401", "missing Authorization header", nil) // Error if no token
				return
			}
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:] // Strip Bearer prefix
			}

			if _, err := verifier.VerifyJWT(token); err != nil {
				if l := reqLogger(req); l != nil {
					l.Warnw("JWT rejected: invalid token", "subject", req.Subject(), "err", err)
				}
				_ = req.Error("401", "invalid token: "+err.Error(), nil) // Error if invalid token
				return
			}

			next.Handle(req) // Call next handler if valid
		})
	}
}

// reqLogger retrieves the logger from the request.
func reqLogger(req Request) x_log.Logger {
	if r, ok := req.(*request); ok {
		return r.logger
	}
	return nil
}

// RequireRole ensures the request has the required role in the JWT.
func RequireRole(role string, verifier JWTVerifier) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(req Request) {
			token := req.Headers().Get("Authorization") // Get token from header
			if token == "" {
				if l := reqLogger(req); l != nil {
					l.Warnw("role check rejected: missing Authorization header", "subject", req.Subject(), "required_role", role)
				}
				_ = req.Error("401", "missing Authorization header", nil) // Error if no token
				return
			}
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:] // Strip Bearer prefix
			}

			claims, err := verifier.VerifyJWT(token) // Verify JWT token
			if err != nil {
				if l := reqLogger(req); l != nil {
					l.Warnw("role check rejected: invalid token", "subject", req.Subject(), "required_role", role, "err", err)
				}
				_ = req.Error("401", "invalid token: "+err.Error(), nil) // Error if invalid token
				return
			}

			roles, ok := claims["roles"] // Extract roles claim
			if !ok {
				if l := reqLogger(req); l != nil {
					l.Warnw("role check rejected: missing roles claim", "subject", req.Subject(), "required_role", role)
				}
				_ = req.Error("403", "missing roles claim", nil) // Error if no roles claim
				return
			}

			switch v := roles.(type) {
			case []any:
				for _, r := range v { // Check for matching role in array
					if str, ok := r.(string); ok && str == role {
						next.Handle(req) // Call next handler if role matches
						return
					}
				}
			case []string:
				for _, r := range v { // Check for matching role in string array
					if r == role {
						next.Handle(req) // Call next handler if role matches
						return
					}
				}
			default:
				if l := reqLogger(req); l != nil {
					l.Warnw("role check rejected: invalid roles format", "subject", req.Subject(), "required_role", role)
				}
				_ = req.Error("403", "invalid roles format", nil) // Error if roles format is invalid
				return
			}

			if l := reqLogger(req); l != nil {
				l.Warnw("role check rejected: forbidden", "subject", req.Subject(), "required_role", role)
			}
			_ = req.Error("403", "forbidden", nil) // Forbidden if role doesn't match
		})
	}
}
