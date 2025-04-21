// file:mini/pkg/x_bus/middle.go
package x_bus

import (
	"fmt"
	"strings"

	"github.com/rskv-p/mini/pkg/x_req"
)

//---------------------
// JWT Middleware
//---------------------

func JWTMiddleware(secret string) Middleware {
	return func(r *x_req.Request) error {
		auth := r.Headers().Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return fmt.Errorf("missing or invalid Authorization header")
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		claims, err := verifyJWT(tokenStr, secret)
		if err != nil {
			return fmt.Errorf("token verification failed: %w", err)
		}

		// Save subject and other claims into headers for downstream
		if sub, ok := (*claims)["sub"].(string); ok {
			r.SetHeader("jwt.sub", sub)
		}
		if role, ok := (*claims)["role"].(string); ok {
			r.SetHeader("jwt.role", role)
		}

		return nil
	}
}
