package domain

import (
	"context"
	"errors"
	"time"
)

// Token represents an authentication token pair issued to a user.
type Token struct {
	accessToken  string
	refreshToken string
	jti          string
	expiry       time.Time
}

// TokenRepository defines the interface for token data access.
type TokenRepository interface {
	BlacklistToken(ctx context.Context, tokenID, userID, reason string, expiresAt time.Time) error
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
}

// NewToken creates a new Token instance.
func NewToken(accessToken, refreshToken, jti string, expiry time.Time) (*Token, error) {
	if accessToken == "" {
		return nil, errors.New("access token must not be empty")
	}
	if refreshToken == "" {
		return nil, errors.New("refresh token must not be empty")
	}
	if jti == "" {
		return nil, errors.New("jti must not be empty")
	}
	if expiry.Before(time.Now()) {
		return nil, errors.New("token expiry must be in the future")
	}

	return &Token{
		accessToken:  accessToken,
		refreshToken: refreshToken,
		jti:          jti,
		expiry:       expiry,
	}, nil
}

// AccessToken returns the access token string.
func (t *Token) AccessToken() string {
	return t.accessToken
}

// RefreshToken returns the refresh token string.
func (t *Token) RefreshToken() string {
	return t.refreshToken
}

// JTI returns the JWT token identifier.
func (t *Token) JTI() string {
	return t.jti
}

// Expiry returns the token's expiration time.
func (t *Token) Expiry() time.Time {
	return t.expiry
}

// IsExpired checks if the token has expired.
func (t *Token) IsExpired() bool {
	return time.Now().After(t.expiry)
}
