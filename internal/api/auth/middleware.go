package auth

import (
	"context"
	"net/http"
	"strings"
)

// Middleware creates an authentication middleware for chi router
func Middleware(manager *Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			// Check Bearer prefix
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error": "invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate token
			tokenInfo, err := manager.ValidateToken(tokenString)
			if err != nil {
				switch err {
				case ErrTokenExpired:
					http.Error(w, `{"error": "token has expired"}`, http.StatusUnauthorized)
				default:
					http.Error(w, `{"error": "invalid token"}`, http.StatusUnauthorized)
				}
				return
			}

			// Add user info to context
			user := User{
				Username: tokenInfo.Username,
				Role:     tokenInfo.Role,
			}
			ctx := context.WithValue(r.Context(), UserContextKey, user)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(UserContextKey).(User)
	return user, ok
}

// OptionalMiddleware creates a middleware that extracts user info if token is present
// but doesn't require authentication
func OptionalMiddleware(manager *Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				next.ServeHTTP(w, r)
				return
			}

			tokenInfo, err := manager.ValidateToken(parts[1])
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			user := User{
				Username: tokenInfo.Username,
				Role:     tokenInfo.Role,
			}
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
