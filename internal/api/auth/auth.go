package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials is returned when credentials are invalid
	ErrInvalidCredentials = errors.New("invalid username or password")
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired is returned when token has expired
	ErrTokenExpired = errors.New("token has expired")
)

// JWTClaims represents the claims in the JWT token
type JWTClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Manager handles authentication operations
type Manager struct {
	username       string
	passwordHash   string
	jwtSecret      []byte
	tokenExpiryHrs int
}

// NewManager creates a new auth manager
func NewManager(username, passwordHash, jwtSecret string, tokenExpiryHrs int) *Manager {
	if tokenExpiryHrs <= 0 {
		tokenExpiryHrs = 24 // Default 24 hours
	}
	return &Manager{
		username:       username,
		passwordHash:   passwordHash,
		jwtSecret:      []byte(jwtSecret),
		tokenExpiryHrs: tokenExpiryHrs,
	}
}

// ValidateCredentials checks if the provided credentials are valid
func (m *Manager) ValidateCredentials(username, password string) error {
	if username != m.username {
		return ErrInvalidCredentials
	}

	// Check password against bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(m.passwordHash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}

	return nil
}

// GenerateToken creates a new JWT token for the user
func (m *Manager) GenerateToken(username string) (string, int64, error) {
	expiresAt := time.Now().Add(time.Duration(m.tokenExpiryHrs) * time.Hour)

	claims := &JWTClaims{
		Username: username,
		Role:     "admin", // Single-user mode, always admin
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "outb",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.jwtSecret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

// ValidateToken validates a JWT token and returns the token info
func (m *Manager) ValidateToken(tokenString string) (*TokenInfo, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return &TokenInfo{
		Username:  claims.Username,
		Role:      claims.Role,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

// HashPassword generates a bcrypt hash from a plain password
// This is a utility function for generating password hashes
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// GetUser returns the configured user
func (m *Manager) GetUser() User {
	return User{
		Username: m.username,
		Role:     "admin",
	}
}
