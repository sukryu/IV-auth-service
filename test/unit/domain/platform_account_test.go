package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/IV-auth-services/internal/core/domain"
)

func TestNewPlatformAccount(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)

	tests := []struct {
		name             string
		id               string
		userID           string
		platform         domain.PlatformType
		platformUserID   string
		platformUsername string
		accessToken      string
		refreshToken     string
		tokenExpiresAt   *time.Time
		wantErr          bool
	}{
		{
			name:             "Valid Twitch account",
			id:               "pa-123",
			userID:           "user-123",
			platform:         domain.PlatformTwitch,
			platformUserID:   "twitch123",
			platformUsername: "TwitchUser",
			accessToken:      "access_token_123",
			refreshToken:     "refresh_token_123",
			tokenExpiresAt:   &expiresAt,
			wantErr:          false,
		},
		{
			name:             "Empty ID",
			id:               "",
			userID:           "user-123",
			platform:         domain.PlatformYouTube,
			platformUserID:   "yt123",
			platformUsername: "YTUser",
			accessToken:      "access_token_123",
			refreshToken:     "refresh_token_123",
			tokenExpiresAt:   &expiresAt,
			wantErr:          true,
		},
		{
			name:             "Empty userID",
			id:               "pa-123",
			userID:           "",
			platform:         domain.PlatformTwitch,
			platformUserID:   "twitch123",
			platformUsername: "TwitchUser",
			accessToken:      "access_token_123",
			refreshToken:     "refresh_token_123",
			tokenExpiresAt:   &expiresAt,
			wantErr:          true,
		},
		{
			name:             "Empty platformUserID",
			id:               "pa-123",
			userID:           "user-123",
			platform:         domain.PlatformTwitch,
			platformUserID:   "",
			platformUsername: "TwitchUser",
			accessToken:      "access_token_123",
			refreshToken:     "refresh_token_123",
			tokenExpiresAt:   &expiresAt,
			wantErr:          true,
		},
		{
			name:             "Invalid platform",
			id:               "pa-123",
			userID:           "user-123",
			platform:         "INVALID",
			platformUserID:   "twitch123",
			platformUsername: "TwitchUser",
			accessToken:      "access_token_123",
			refreshToken:     "refresh_token_123",
			tokenExpiresAt:   &expiresAt,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa, err := domain.NewPlatformAccount(tt.id, tt.userID, tt.platform, tt.platformUserID, tt.platformUsername, tt.accessToken, tt.refreshToken, tt.tokenExpiresAt)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pa)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pa)
				assert.Equal(t, tt.id, pa.ID())
				assert.Equal(t, tt.userID, pa.UserID())
				assert.Equal(t, tt.platform, pa.Platform())
				assert.Equal(t, tt.platformUserID, pa.PlatformUserID())
				assert.Equal(t, tt.platformUsername, pa.PlatformUsername())
				assert.Equal(t, tt.accessToken, pa.AccessToken())
				assert.Equal(t, tt.refreshToken, pa.RefreshToken())
				if tt.tokenExpiresAt != nil {
					assert.WithinDuration(t, *tt.tokenExpiresAt, *pa.TokenExpiresAt(), time.Second)
				}
				assert.WithinDuration(t, time.Now(), pa.CreatedAt(), time.Second)
			}
		})
	}
}
