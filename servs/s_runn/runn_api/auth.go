package runn_api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rskv-p/mini/servs/s_runn/runn_cfg"
	"github.com/rskv-p/mini/servs/s_runn/runn_serv"
	"gorm.io/gorm"
)

type jwtClaims struct {
	Username string `json:"sub"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type contextKey string

const jwtContextKey = contextKey("jwt_claims")

var jwtKey = []byte("default_secret") // Default JWT secret

// InitAuth initializes the JWT secret from configuration
func InitAuth() {
	jwtKey = []byte(runn_cfg.C().JwtSecret)
}

// -------- /auth/login --------
func HandleLogin(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		// Decode login request body
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", 400)
			return
		}

		// Find user and check password
		user, err := runn_serv.FindUserByUsername(db, req.Username)
		if err != nil || !user.CheckPassword(req.Password) {
			http.Error(w, "unauthorized", 401)
			return
		}

		// Generate JWT token
		tokenStr, err := generateToken(user)
		if err != nil {
			http.Error(w, "token error", 500)
			return
		}

		// Respond with token
		_ = json.NewEncoder(w).Encode(map[string]string{"token": tokenStr})
	}
}

// -------- /auth/register --------
func HandleRegister(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		// Decode registration request body
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		// Create new user
		if err := runn_serv.CreateUser(db, req.Username, req.Password, "user"); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

// -------- JWT Token Generation --------
func generateToken(u *runn_serv.User) (string, error) {
	claims := jwtClaims{
		Username: u.Username,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// -------- Middleware: JWT Token Validation --------
func JWTMiddleware(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract JWT token from request
			tokenStr := extractToken(r)
			if tokenStr == "" {
				http.Error(w, "unauthorized", 401)
				return
			}

			// Parse and validate token
			claims := &jwtClaims{}
			_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})
			if err != nil || claims.ExpiresAt.Time.Before(time.Now()) {
				http.Error(w, "invalid or expired token", 401)
				return
			}

			// Check for required role
			if requiredRole != "" && claims.Role != requiredRole {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), jwtContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -------- Utility: Extract Token --------
func extractToken(r *http.Request) string {
	// Extracts Bearer token from Authorization header
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

// -------- Utility: Get User Claims from Context --------
func UserFromContext(ctx context.Context) (username, role string, ok bool) {
	// Retrieves JWT claims (user info) from the context
	claims, ok := ctx.Value(jwtContextKey).(*jwtClaims)
	if !ok {
		return "", "", false
	}
	return claims.Username, claims.Role, true
}
