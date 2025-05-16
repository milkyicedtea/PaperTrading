package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// UserClaimsKey is the key used to store user claims in the request context.
	UserClaimsKey ContextKey = "userClaims"
)

// Middleware is the middleware which provides authentication.
type Middleware struct {
	service *AuthService
}

// NewMiddleware creates a new auth middleware instance.
func NewMiddleware(service *AuthService) *Middleware {
	return &Middleware{service: service}
}

// Authenticate is a go-chi middleware that checks for a valid JWT in the `Authorization` header.
// if valid, it stores the claims in the request context.
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			RespondWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			RespondWithError(w, http.StatusUnauthorized, "Authorization header format must be Bearer {token}")
			return
		}

		tokenString := parts[1]
		claims, err := m.service.ValidateToken(tokenString) // ValidateToken already handles "Bearer " prefix if it was there
		if err != nil {
			// ValidateToken returns specific errors like ErrTokenExpired, ErrInvalidToken
			// this means the response can be tailored based on the error type
			if errors.Is(err, ErrTokenExpired) {
				RespondWithError(w, http.StatusUnauthorized, "Token has expired")
			} else {
				RespondWithError(w, http.StatusUnauthorized, "Invalid or malformed token")
			}
			return
		}

		// token is valid, store claims in context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserClaims retrieves user claims from the request context.
// this is a helper function for protected handlers.
func GetUserClaims(ctx context.Context) (*JWTCustomClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*JWTCustomClaims)
	return claims, ok
}
