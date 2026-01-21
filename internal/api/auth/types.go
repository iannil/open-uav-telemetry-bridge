// Package auth provides authentication functionality for the API server
package auth

import "time"

// User represents an authenticated user
type User struct {
	Username string `json:"username"`
	Role     string `json:"role"` // "admin" for single-user mode
}

// LoginRequest is the request body for login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is the response for successful login
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	User      User   `json:"user"`
}

// Claims represents JWT claims
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// TokenInfo contains parsed token information
type TokenInfo struct {
	Username  string
	Role      string
	ExpiresAt time.Time
}

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserContextKey is the context key for storing user information
	UserContextKey ContextKey = "user"
)
