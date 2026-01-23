package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name           string
		tokenExpiryHrs int
		wantExpiry     int
	}{
		{"positive expiry", 48, 48},
		{"zero defaults to 24", 0, 24},
		{"negative defaults to 24", -5, 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager("admin", "hash", "secret", tt.tokenExpiryHrs)
			if m.tokenExpiryHrs != tt.wantExpiry {
				t.Errorf("tokenExpiryHrs = %d, want %d", m.tokenExpiryHrs, tt.wantExpiry)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword should not error: %v", err)
	}
	if hash == "" {
		t.Error("HashPassword should return non-empty hash")
	}
	if hash == password {
		t.Error("Hash should not equal original password")
	}

	// Hash should be different each time (bcrypt includes salt)
	hash2, _ := HashPassword(password)
	if hash == hash2 {
		t.Error("Hashes should be different due to random salt")
	}
}

func TestManager_ValidateCredentials(t *testing.T) {
	password := "correctpassword"
	hash, _ := HashPassword(password)
	m := NewManager("admin", hash, "secret", 24)

	tests := []struct {
		name     string
		username string
		password string
		wantErr  error
	}{
		{"valid credentials", "admin", password, nil},
		{"wrong username", "wronguser", password, ErrInvalidCredentials},
		{"wrong password", "admin", "wrongpassword", ErrInvalidCredentials},
		{"both wrong", "wronguser", "wrongpassword", ErrInvalidCredentials},
		{"empty username", "", password, ErrInvalidCredentials},
		{"empty password", "admin", "", ErrInvalidCredentials},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.ValidateCredentials(tt.username, tt.password)
			if err != tt.wantErr {
				t.Errorf("ValidateCredentials() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_GenerateToken(t *testing.T) {
	m := NewManager("admin", "hash", "secret", 24)

	token, expiresAt, err := m.GenerateToken("admin")

	if err != nil {
		t.Fatalf("GenerateToken should not error: %v", err)
	}
	if token == "" {
		t.Error("Token should not be empty")
	}
	if expiresAt == 0 {
		t.Error("ExpiresAt should not be zero")
	}

	// ExpiresAt should be approximately 24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour).Unix()
	if expiresAt < expectedExpiry-60 || expiresAt > expectedExpiry+60 {
		t.Errorf("ExpiresAt should be ~24h from now, got %d, want ~%d", expiresAt, expectedExpiry)
	}
}

func TestManager_ValidateToken(t *testing.T) {
	m := NewManager("admin", "hash", "secret", 24)

	// Generate a valid token
	token, _, err := m.GenerateToken("admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Test valid token
	info, err := m.ValidateToken(token)
	if err != nil {
		t.Errorf("ValidateToken should not error for valid token: %v", err)
	}
	if info == nil {
		t.Fatal("TokenInfo should not be nil")
	}
	if info.Username != "admin" {
		t.Errorf("Username = %s, want 'admin'", info.Username)
	}
	if info.Role != "admin" {
		t.Errorf("Role = %s, want 'admin'", info.Role)
	}
}

func TestManager_ValidateToken_Invalid(t *testing.T) {
	m := NewManager("admin", "hash", "secret", 24)

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{"empty token", "", ErrInvalidToken},
		{"malformed token", "not.a.valid.token", ErrInvalidToken},
		{"random string", "randomstring", ErrInvalidToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := m.ValidateToken(tt.token)
			if err != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_ValidateToken_WrongSecret(t *testing.T) {
	m1 := NewManager("admin", "hash", "secret1", 24)
	m2 := NewManager("admin", "hash", "secret2", 24)

	// Generate token with m1's secret
	token, _, _ := m1.GenerateToken("admin")

	// Validate with m2's different secret
	_, err := m2.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Errorf("ValidateToken with wrong secret should return ErrInvalidToken, got %v", err)
	}
}

func TestManager_GetUser(t *testing.T) {
	m := NewManager("testuser", "hash", "secret", 24)

	user := m.GetUser()

	if user.Username != "testuser" {
		t.Errorf("Username = %s, want 'testuser'", user.Username)
	}
	if user.Role != "admin" {
		t.Errorf("Role = %s, want 'admin'", user.Role)
	}
}

func TestMiddleware_MissingHeader(t *testing.T) {
	m := NewManager("admin", "hash", "secret", 24)
	middleware := Middleware(m)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestMiddleware_InvalidFormat(t *testing.T) {
	m := NewManager("admin", "hash", "secret", 24)
	middleware := Middleware(m)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name   string
		header string
	}{
		{"missing bearer", "token123"},
		{"wrong prefix", "Basic token123"},
		{"no token", "Bearer"},
		{"empty after bearer", "Bearer "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.header)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("Status = %d, want %d", rr.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestMiddleware_ValidToken(t *testing.T) {
	m := NewManager("admin", "hash", "secret", 24)
	token, _, _ := m.GenerateToken("admin")

	var gotUser User
	var gotOK bool

	middleware := Middleware(m)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotOK = GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !gotOK {
		t.Error("User should be in context")
	}
	if gotUser.Username != "admin" {
		t.Errorf("Username = %s, want 'admin'", gotUser.Username)
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	m := NewManager("admin", "hash", "secret", 24)
	middleware := Middleware(m)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestOptionalMiddleware_NoHeader(t *testing.T) {
	m := NewManager("admin", "hash", "secret", 24)

	var called bool
	middleware := OptionalMiddleware(m)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, ok := GetUserFromContext(r.Context())
		if ok {
			t.Error("User should not be in context when no header")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("Handler should be called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestOptionalMiddleware_ValidToken(t *testing.T) {
	m := NewManager("admin", "hash", "secret", 24)
	token, _, _ := m.GenerateToken("admin")

	var gotUser User
	var gotOK bool

	middleware := OptionalMiddleware(m)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotOK = GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !gotOK {
		t.Error("User should be in context")
	}
	if gotUser.Username != "admin" {
		t.Errorf("Username = %s, want 'admin'", gotUser.Username)
	}
}

func TestOptionalMiddleware_InvalidToken(t *testing.T) {
	m := NewManager("admin", "hash", "secret", 24)

	var called bool
	middleware := OptionalMiddleware(m)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, ok := GetUserFromContext(r.Context())
		if ok {
			t.Error("User should not be in context for invalid token")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("Handler should be called even with invalid token")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestOptionalMiddleware_InvalidFormat(t *testing.T) {
	m := NewManager("admin", "hash", "secret", 24)

	var called bool
	middleware := OptionalMiddleware(m)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic sometoken") // Wrong format
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("Handler should be called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestGetUserFromContext_NoUser(t *testing.T) {
	ctx := context.Background()

	user, ok := GetUserFromContext(ctx)

	if ok {
		t.Error("Should return false when no user in context")
	}
	if user.Username != "" {
		t.Error("User should be zero value")
	}
}

func TestGetUserFromContext_WithUser(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserContextKey, User{
		Username: "testuser",
		Role:     "admin",
	})

	user, ok := GetUserFromContext(ctx)

	if !ok {
		t.Error("Should return true when user in context")
	}
	if user.Username != "testuser" {
		t.Errorf("Username = %s, want 'testuser'", user.Username)
	}
	if user.Role != "admin" {
		t.Errorf("Role = %s, want 'admin'", user.Role)
	}
}

func TestMiddleware_BearerCaseInsensitive(t *testing.T) {
	m := NewManager("admin", "hash", "secret", 24)
	token, _, _ := m.GenerateToken("admin")
	middleware := Middleware(m)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name   string
		prefix string
	}{
		{"lowercase", "bearer"},
		{"uppercase", "BEARER"},
		{"mixed case", "BeArEr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.prefix+" "+token)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
			}
		})
	}
}
