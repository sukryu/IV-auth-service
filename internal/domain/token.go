package domain

import (
	"time"

	"github.com/google/uuid"
)

// Token represents a JWT token entity
type Token struct {
	accessToken  string
	refreshToken string
	issuedAt     time.Time
	expiresAt    time.Time
	jti          string // JWT Token Identifier
}

// NewToken creates a new Token instance
func NewToken(accessToken, refreshToken string, duration time.Duration) (*Token, error) {
	now := time.Now()
	jti := uuid.New().String()

	return &Token{
		accessToken:  accessToken,
		refreshToken: refreshToken,
		issuedAt:     now,
		expiresAt:    now.Add(duration),
		jti:          jti,
	}, nil
}

// Getters
func (t *Token) AccessToken() string  { return t.accessToken }
func (t *Token) RefreshToken() string { return t.refreshToken }
func (t *Token) IssuedAt() time.Time  { return t.issuedAt }
func (t *Token) ExpiresAt() time.Time { return t.expiresAt }
func (t *Token) JTI() string          { return t.jti }

// Setters
func (t *Token) SetTokens(accessToken, refreshToken string, duration time.Duration) {
	now := time.Now()
	t.accessToken = accessToken
	t.refreshToken = refreshToken
	t.issuedAt = now
	t.expiresAt = now.Add(duration)
}
