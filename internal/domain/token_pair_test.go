package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenPair(t *testing.T) {
	tests := []struct {
		name             string
		accessToken      string
		refreshToken     string
		wantErr          bool
		wantAccessToken  string
		wantRefreshToken string
		wantExpires      bool
	}{
		{
			name:             "Valid token pair creation",
			accessToken:      "mock_access_token",
			refreshToken:     "mock_refresh_token",
			wantErr:          false,
			wantAccessToken:  "mock_access_token",
			wantRefreshToken: "mock_refresh_token",
			wantExpires:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp, err := NewTokenPair(tt.accessToken, tt.refreshToken)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, tp)

			// 필드 값 검증
			assert.Equal(t, tt.wantAccessToken, tp.AccessToken())
			assert.Equal(t, tt.wantRefreshToken, tp.RefreshToken())

			// 만료 시간 검증
			if tt.wantExpires {
				assert.False(t, tp.AccessExpiresAt().IsZero())
				assert.False(t, tp.RefreshExpiresAt().IsZero())
				assert.True(t, tp.AccessExpiresAt().Before(tp.RefreshExpiresAt()))
			}
		})
	}
}

func TestTokenPairExpiration(t *testing.T) {
	// 짧은 만료 시간으로 테스트
	tp, err := NewTokenPair("access", "refresh")
	assert.NoError(t, err)

	// 기본 만료 시간 확인
	assert.False(t, tp.IsAccessExpired())
	assert.False(t, tp.IsRefreshExpired())

	// 시간 조작 대신 수동으로 만료 시간 설정 (테스트 용이성)
	tp.accessExpiresAt = time.Now().Add(-1 * time.Minute)
	tp.refreshExpiresAt = time.Now().Add(-1 * time.Minute)
	assert.True(t, tp.IsAccessExpired())
	assert.True(t, tp.IsRefreshExpired())
}
