package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	tests := []struct {
		name            string
		accessToken     string
		refreshToken    string
		duration        time.Duration
		wantErr         bool
		wantJtiNotEmpty bool
		wantIssuedAt    bool
		wantExpiresAt   bool
	}{
		{
			name:            "Valid token creation",
			accessToken:     "mock_access_token",
			refreshToken:    "mock_refresh_token",
			duration:        15 * time.Minute,
			wantErr:         false,
			wantJtiNotEmpty: true,
			wantIssuedAt:    true,
			wantExpiresAt:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := NewToken(tt.accessToken, tt.refreshToken, tt.duration)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, token)

			// JTI 검증
			if tt.wantJtiNotEmpty {
				assert.NotEmpty(t, token.JTI())
			}

			// 필드 값 검증
			assert.Equal(t, tt.accessToken, token.AccessToken())
			assert.Equal(t, tt.refreshToken, token.RefreshToken())

			// 시간 관련 검증
			if tt.wantIssuedAt {
				assert.False(t, token.IssuedAt().IsZero())
			}
			if tt.wantExpiresAt {
				assert.False(t, token.ExpiresAt().IsZero())
				assert.True(t, token.ExpiresAt().After(token.IssuedAt()))
			}
		})
	}
}

func TestTokenSetTokens(t *testing.T) {
	token, err := NewToken("old_access", "old_refresh", 15*time.Minute)
	assert.NoError(t, err)

	originalIssuedAt := token.IssuedAt()
	time.Sleep(1 * time.Millisecond)

	newDuration := 30 * time.Minute
	token.SetTokens("new_access", "new_refresh", newDuration)

	assert.Equal(t, "new_access", token.AccessToken())
	assert.Equal(t, "new_refresh", token.RefreshToken())
	assert.True(t, token.IssuedAt().After(originalIssuedAt))
	assert.True(t, token.ExpiresAt().After(token.IssuedAt()))
}
