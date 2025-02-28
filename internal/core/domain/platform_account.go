package domain

import (
	"errors"
	"time"
)

// PlatformType defines the supported external platforms.
type PlatformType string

const (
	PlatformTwitch   PlatformType = "TWITCH"
	PlatformYouTube  PlatformType = "YOUTUBE"
	PlatformFacebook PlatformType = "FACEBOOK"
	PlatformAfreeca  PlatformType = "AFREECA"
)

// PlatformAccount represents an external platform account linked to a user.
type PlatformAccount struct {
	id               string
	userID           string
	platform         PlatformType
	platformUserID   string
	platformUsername string
	accessToken      string
	refreshToken     string
	tokenExpiresAt   *time.Time // nullable
	createdAt        time.Time
	updatedAt        time.Time
}

// NewPlatformAccount creates a new PlatformAccount instance.
func NewPlatformAccount(id, userID string, platform PlatformType, platformUserID, platformUsername, accessToken, refreshToken string, tokenExpiresAt *time.Time) (*PlatformAccount, error) {
	if id == "" {
		return nil, errors.New("platform account id must not be empty")
	}
	if userID == "" {
		return nil, errors.New("user id must not be empty")
	}
	if platformUserID == "" {
		return nil, errors.New("platform user id must not be empty")
	}
	if !isValidPlatform(platform) {
		return nil, errors.New("invalid platform type")
	}

	now := time.Now()
	return &PlatformAccount{
		id:               id,
		userID:           userID,
		platform:         platform,
		platformUserID:   platformUserID,
		platformUsername: platformUsername,
		accessToken:      accessToken,
		refreshToken:     refreshToken,
		tokenExpiresAt:   tokenExpiresAt,
		createdAt:        now,
		updatedAt:        now,
	}, nil
}

// ID returns the platform account's unique identifier.
func (p *PlatformAccount) ID() string {
	return p.id
}

// UserID returns the associated user's identifier.
func (p *PlatformAccount) UserID() string {
	return p.userID
}

// Platform returns the platform type.
func (p *PlatformAccount) Platform() PlatformType {
	return p.platform
}

// PlatformUserID returns the user's ID on the external platform.
func (p *PlatformAccount) PlatformUserID() string {
	return p.platformUserID
}

// PlatformUsername returns the user's username on the external platform.
func (p *PlatformAccount) PlatformUsername() string {
	return p.platformUsername
}

// AccessToken returns the OAuth access token.
func (p *PlatformAccount) AccessToken() string {
	return p.accessToken
}

// RefreshToken returns the OAuth refresh token.
func (p *PlatformAccount) RefreshToken() string {
	return p.refreshToken
}

// TokenExpiresAt returns the token expiration time, or nil if not set.
func (p *PlatformAccount) TokenExpiresAt() *time.Time {
	return p.tokenExpiresAt
}

// CreatedAt returns the creation time.
func (p *PlatformAccount) CreatedAt() time.Time {
	return p.createdAt
}

// UpdatedAt returns the last update time.
func (p *PlatformAccount) UpdatedAt() time.Time {
	return p.updatedAt
}

// isValidPlatform checks if the platform type is supported.
func isValidPlatform(p PlatformType) bool {
	switch p {
	case PlatformTwitch, PlatformYouTube, PlatformFacebook, PlatformAfreeca:
		return true
	}
	return false
}
