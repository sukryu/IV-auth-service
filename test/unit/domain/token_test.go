package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/IV-auth-services/internal/core/domain"
)

func TestNewToken(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)

	tests := []struct {
		name         string
		accessToken  string
		refreshToken string
		jti          string
		expiry       time.Time
		wantErr      bool
	}{
		{
			name:         "Valid token",
			accessToken:  "access_token_123",
			refreshToken: "refresh_token_123",
			jti:          "jti-123",
			expiry:       expiresAt,
			wantErr:      false,
		},
		{
			name:         "Empty access token",
			accessToken:  "",
			refreshToken: "refresh_token_123",
			jti:          "jti-123",
			expiry:       expiresAt,
			wantErr:      true,
		},
		{
			name:         "Empty refresh token",
			accessToken:  "access_token_123",
			refreshToken: "",
			jti:          "jti-123",
			expiry:       expiresAt,
			wantErr:      true,
		},
		{
			name:         "Empty JTI",
			accessToken:  "access_token_123",
			refreshToken: "refresh_token_123",
			jti:          "",
			expiry:       expiresAt,
			wantErr:      true,
		},
		{
			name:         "Expired token",
			accessToken:  "access_token_123",
			refreshToken: "refresh_token_123",
			jti:          "jti-123",
			expiry:       time.Now().Add(-time.Hour),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := domain.NewToken(tt.accessToken, tt.refreshToken, tt.jti, tt.expiry)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.Equal(t, tt.accessToken, token.AccessToken())
				assert.Equal(t, tt.refreshToken, token.RefreshToken())
				assert.Equal(t, tt.jti, token.JTI())
				assert.Equal(t, tt.expiry, token.Expiry())
				assert.False(t, token.IsExpired())
			}
		})
	}
}

func TestTokenIsExpired(t *testing.T) {
	// 미래 만료 토큰
	futureExpiresAt := time.Now().Add(time.Hour)
	futureToken, err := domain.NewToken("access2", "refresh2", "jti-456", futureExpiresAt)
	assert.NoError(t, err)
	assert.NotNil(t, futureToken)
	assert.False(t, futureToken.IsExpired())

	// 과거 만료 토큰 테스트 (직접 생성 대신 미래 토큰으로 테스트 후 시간 조작)
	// 시간 조작을 위해 별도 헬퍼 함수 사용
	expiredToken, err := domain.NewToken("access", "refresh", "jti-123", time.Now().Add(time.Second))
	assert.NoError(t, err)
	assert.NotNil(t, expiredToken)

	// 2초 대기 후 만료 확인 (간단한 테스트 용도로)
	time.Sleep(2 * time.Second)
	assert.True(t, expiredToken.IsExpired())
}
