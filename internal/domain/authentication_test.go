package domain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockTokenRepository is a mock implementation of TokenRepository
type mockTokenRepository struct {
	blacklist map[string]struct{}
}

func (m *mockTokenRepository) InsertTokenBlacklist(ctx context.Context, tokenID string, userID string, expiresAt time.Time, reason string) error {
	m.blacklist[tokenID] = struct{}{}
	return nil
}

func (m *mockTokenRepository) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	_, exists := m.blacklist[tokenID]
	return exists, nil
}

func TestAuthenticationDomainServiceLogin(t *testing.T) {
	tests := []struct {
		name       string
		username   string
		password   string
		wantErr    bool
		wantUser   bool
		wantTokens bool
	}{
		{
			name:       "Valid login",
			username:   "testuser",
			password:   "StrongP@ssw0rd!",
			wantErr:    false,
			wantUser:   true,
			wantTokens: true,
		},
		{
			name:     "Invalid password",
			username: "testuser",
			password: "wrongpassword",
			wantErr:  true,
		},
		{
			name:     "Non-existent user",
			username: "nonexistent",
			password: "StrongP@ssw0rd!",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{users: make(map[string]*User)}
			tokenRepo := &mockTokenRepository{blacklist: make(map[string]struct{})}
			issuer := &mockTokenIssuer{} // 더미 issuer, 실제 토큰 발급은 생략
			s := NewAuthenticationDomainService(repo, tokenRepo, issuer)

			// 초기 사용자 설정
			if tt.name != "Non-existent user" {
				pwd, _ := NewPassword("StrongP@ssw0rd!")
				user, _ := NewUser("testuser", "test@example.com", pwd.String())
				repo.users["testuser"] = user
			}

			user, tokens, err := s.Login(context.Background(), tt.username, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantUser {
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username())
				assert.NotNil(t, user.LastLoginAt())
			}
			if tt.wantTokens {
				assert.NotNil(t, tokens)
				assert.NotEmpty(t, tokens.AccessToken())
				assert.NotEmpty(t, tokens.RefreshToken())
			}
		})
	}
}

func TestAuthenticationDomainServiceLogout(t *testing.T) {
	repo := &mockUserRepository{users: make(map[string]*User)}
	tokenRepo := &mockTokenRepository{blacklist: make(map[string]struct{})}
	issuer := &mockTokenIssuer{}
	s := NewAuthenticationDomainService(repo, tokenRepo, issuer)

	err := s.Logout(context.Background(), "user123", "token123")
	assert.NoError(t, err)

	blacklisted, err := tokenRepo.IsTokenBlacklisted(context.Background(), "token123")
	assert.NoError(t, err)
	assert.True(t, blacklisted)
}

func TestAuthenticationDomainServiceValidateToken(t *testing.T) {
	repo := &mockUserRepository{users: make(map[string]*User)}
	tokenRepo := &mockTokenRepository{blacklist: make(map[string]struct{})}
	issuer := &mockTokenIssuer{}
	s := NewAuthenticationDomainService(repo, tokenRepo, issuer)

	// 유효한 토큰
	valid, err := s.ValidateToken(context.Background(), "valid_token")
	assert.NoError(t, err)
	assert.True(t, valid)

	// 블랙리스트에 추가
	tokenRepo.blacklist["blacklisted_token"] = struct{}{}
	valid, err = s.ValidateToken(context.Background(), "blacklisted_token")
	assert.NoError(t, err)
	assert.False(t, valid)
}
