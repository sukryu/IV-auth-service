package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlatformAccount(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)

	tests := []struct {
		name             string
		userID           string
		platform         string
		platformUserID   string
		platformUsername string
		accessToken      string
		refreshToken     string
		tokenExpiresAt   *time.Time
		wantErr          bool
		wantIDNotEmpty   bool
		wantCreatedAt    bool
		wantUpdatedAt    bool
	}{
		{
			name:             "Valid platform account creation",
			userID:           "550e8400-e29b-41d4-a716-446655440000",
			platform:         "TWITCH",
			platformUserID:   "twitch123",
			platformUsername: "TestTwitchUser",
			accessToken:      "mock_access_token",
			refreshToken:     "mock_refresh_token",
			tokenExpiresAt:   &expiresAt,
			wantErr:          false,
			wantIDNotEmpty:   true,
			wantCreatedAt:    true,
			wantUpdatedAt:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa, err := NewPlatformAccount(tt.userID, tt.platform, tt.platformUserID, tt.platformUsername, tt.accessToken, tt.refreshToken, tt.tokenExpiresAt)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, pa)

			// ID 검증
			if tt.wantIDNotEmpty {
				assert.NotEmpty(t, pa.ID())
			}

			// 필드 값 검증
			assert.Equal(t, tt.userID, pa.UserID())
			assert.Equal(t, tt.platform, pa.Platform())
			assert.Equal(t, tt.platformUserID, pa.PlatformUserID())
			assert.Equal(t, tt.platformUsername, pa.PlatformUsername())
			assert.Equal(t, tt.accessToken, pa.AccessToken())
			assert.Equal(t, tt.refreshToken, pa.RefreshToken())
			assert.Equal(t, tt.tokenExpiresAt, pa.TokenExpiresAt())

			// 시간 관련 검증
			if tt.wantCreatedAt {
				assert.False(t, pa.CreatedAt().IsZero())
			}
			if tt.wantUpdatedAt {
				assert.False(t, pa.UpdatedAt().IsZero())
			}
		})
	}
}

func TestPlatformAccountSetTokens(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)
	pa, err := NewPlatformAccount("user123", "TWITCH", "twitch123", "TestTwitchUser", "old_access", "old_refresh", &expiresAt)
	assert.NoError(t, err)

	originalUpdatedAt := pa.UpdatedAt()
	time.Sleep(1 * time.Millisecond) // UpdatedAt 변경 확인을 위해 약간 대기

	newExpiresAt := now.Add(2 * time.Hour)
	pa.SetTokens("new_access", "new_refresh", &newExpiresAt)

	assert.Equal(t, "new_access", pa.AccessToken())
	assert.Equal(t, "new_refresh", pa.RefreshToken())
	assert.Equal(t, &newExpiresAt, pa.TokenExpiresAt())
	assert.True(t, pa.UpdatedAt().After(originalUpdatedAt))
}
