// file:mini/pkg/x_bus/middle_test.go
package x_bus

import (
	"testing"
	"time"

	"github.com/rskv-p/mini/pkg/x_req"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

//---------------------
// JWT Middleware
//---------------------

func TestJWTMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"

	// Generate valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "user123",
		"role": "admin",
		"exp":  time.Now().Add(time.Minute).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	// Create request with Authorization header
	req := &x_req.Request{}
	req.SetHeader("Authorization", "Bearer "+tokenStr)

	// Run middleware
	mw := JWTMiddleware(secret)
	err = mw(req)

	// Assertions
	require.NoError(t, err)
	require.Equal(t, "user123", req.Headers().Get("jwt.sub"))
	require.Equal(t, "admin", req.Headers().Get("jwt.role"))
}
