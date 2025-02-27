package domain

import (
	"time"
)

// TokenPair represents an immutable pair of access and refresh tokens
type TokenPair struct {
	accessToken      string
	refreshToken     string
	accessExpiresAt  time.Time
	refreshExpiresAt time.Time
}

// NewTokenPair creates a new TokenPair with specified durations
func NewTokenPair(accessToken, refreshToken string) (*TokenPair, error) {
	now := time.Now()
	accessDuration := 15 * time.Minute    // 액세스 토큰 기본 15분
	refreshDuration := 7 * 24 * time.Hour // 리프레시 토큰 기본 7일

	return &TokenPair{
		accessToken:      accessToken,
		refreshToken:     refreshToken,
		accessExpiresAt:  now.Add(accessDuration),
		refreshExpiresAt: now.Add(refreshDuration),
	}, nil
}

// Getters
func (t *TokenPair) AccessToken() string         { return t.accessToken }
func (t *TokenPair) RefreshToken() string        { return t.refreshToken }
func (t *TokenPair) AccessExpiresAt() time.Time  { return t.accessExpiresAt }
func (t *TokenPair) RefreshExpiresAt() time.Time { return t.refreshExpiresAt }

// IsAccessExpired checks if the access token has expired
func (t *TokenPair) IsAccessExpired() bool {
	return time.Now().After(t.accessExpiresAt)
}

// IsRefreshExpired checks if the refresh token has expired
func (t *TokenPair) IsRefreshExpired() bool {
	return time.Now().After(t.refreshExpiresAt)
}
