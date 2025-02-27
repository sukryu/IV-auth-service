package domain

import (
	"time"

	"github.com/google/uuid"
)

// PlatformAccount represents an external platform account linked to a User
type PlatformAccount struct {
	id               string
	userID           string
	platform         string
	platformUserID   string
	platformUsername string
	accessToken      string
	refreshToken     string
	tokenExpiresAt   *time.Time
	createdAt        time.Time
	updatedAt        time.Time
}

// NewPlatformAccount creates a new PlatformAccount instance
func NewPlatformAccount(userID, platform, platformUserID, platformUsername, accessToken, refreshToken string, tokenExpiresAt *time.Time) (*PlatformAccount, error) {
	id := uuid.New().String()
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

// Getters
func (p *PlatformAccount) ID() string                 { return p.id }
func (p *PlatformAccount) UserID() string             { return p.userID }
func (p *PlatformAccount) Platform() string           { return p.platform }
func (p *PlatformAccount) PlatformUserID() string     { return p.platformUserID }
func (p *PlatformAccount) PlatformUsername() string   { return p.platformUsername }
func (p *PlatformAccount) AccessToken() string        { return p.accessToken }
func (p *PlatformAccount) RefreshToken() string       { return p.refreshToken }
func (p *PlatformAccount) TokenExpiresAt() *time.Time { return p.tokenExpiresAt }
func (p *PlatformAccount) CreatedAt() time.Time       { return p.createdAt }
func (p *PlatformAccount) UpdatedAt() time.Time       { return p.updatedAt }

// Setters
func (p *PlatformAccount) SetTokens(accessToken, refreshToken string, tokenExpiresAt *time.Time) {
	p.accessToken = accessToken
	p.refreshToken = refreshToken
	p.tokenExpiresAt = tokenExpiresAt
	p.updatedAt = time.Now()
}
