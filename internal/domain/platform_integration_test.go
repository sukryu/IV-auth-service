package domain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockPlatformAccountRepository (기존 코드 유지)
type mockPlatformAccountRepository struct {
	accounts map[string]*PlatformAccount
}

func (m *mockPlatformAccountRepository) InsertPlatformAccount(ctx context.Context, account *PlatformAccount) error {
	m.accounts[account.ID()] = account
	return nil
}

func (m *mockPlatformAccountRepository) GetPlatformAccountsByUserID(ctx context.Context, userID string) ([]*PlatformAccount, error) {
	var result []*PlatformAccount
	for _, acc := range m.accounts {
		if acc.UserID() == userID {
			result = append(result, acc)
		}
	}
	return result, nil
}

func (m *mockPlatformAccountRepository) UpdatePlatformAccount(ctx context.Context, account *PlatformAccount) error {
	m.accounts[account.ID()] = account
	return nil
}

// mockTokenIssuer 수정: 반환값 이름 제거
type mockTokenIssuer struct {
	issueFails   bool
	refreshFails bool
}

func (m *mockTokenIssuer) IssueTokens(ctx context.Context, authCode string) (string, string, time.Time, error) {
	if m.issueFails {
		return "", "", time.Time{}, errTokenIssueFailed
	}
	expiresAt := time.Now().Add(1 * time.Hour)
	return "mock_access", "mock_refresh", expiresAt, nil
}

func (m *mockTokenIssuer) RefreshTokens(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
	if m.refreshFails {
		return "", "", time.Time{}, errTokenRefreshFailed
	}
	expiresAt := time.Now().Add(1 * time.Hour)
	return "new_access", "new_refresh", expiresAt, nil
}

var (
	errTokenIssueFailed   = Error("failed to issue tokens")
	errTokenRefreshFailed = Error("failed to refresh tokens")
)

// 테스트 코드 (기존 로직 유지)
func TestPlatformIntegrationDomainServiceConnectPlatformAccount(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		platform         string
		platformUserID   string
		platformUsername string
		authCode         string
		issueFails       bool
		wantErr          bool
		wantAccount      bool
	}{
		{
			name:             "Valid platform account connection",
			userID:           "user123",
			platform:         "TWITCH",
			platformUserID:   "twitch123",
			platformUsername: "TwitchUser",
			authCode:         "mock_code",
			issueFails:       false,
			wantErr:          false,
			wantAccount:      true,
		},
		{
			name:             "Token issue failure",
			userID:           "user123",
			platform:         "TWITCH",
			platformUserID:   "twitch123",
			platformUsername: "TwitchUser",
			authCode:         "mock_code",
			issueFails:       true,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPlatformAccountRepository{accounts: make(map[string]*PlatformAccount)}
			issuer := &mockTokenIssuer{issueFails: tt.issueFails}
			s := NewPlatformIntegrationDomainService(repo, issuer)

			account, err := s.ConnectPlatformAccount(context.Background(), tt.userID, tt.platform, tt.platformUserID, tt.platformUsername, tt.authCode)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantAccount {
				assert.NotNil(t, account)
				assert.Equal(t, tt.userID, account.UserID())
				assert.Equal(t, tt.platform, account.Platform())
				assert.Equal(t, tt.platformUserID, account.PlatformUserID())
				assert.Equal(t, tt.platformUsername, account.PlatformUsername())
				assert.Equal(t, "mock_access", account.AccessToken())
				assert.Equal(t, "mock_refresh", account.RefreshToken())
			}
		})
	}
}

func TestPlatformIntegrationDomainServiceRefreshPlatformToken(t *testing.T) {
	repo := &mockPlatformAccountRepository{accounts: make(map[string]*PlatformAccount)}
	issuer := &mockTokenIssuer{}

	s := NewPlatformIntegrationDomainService(repo, issuer)

	expiresAt := time.Now().Add(1 * time.Hour)
	account, err := NewPlatformAccount("user123", "TWITCH", "twitch123", "TwitchUser", "old_access", "old_refresh", &expiresAt)
	assert.NoError(t, err)
	repo.accounts[account.ID()] = account

	err = s.RefreshPlatformToken(context.Background(), "user123", account.ID())
	assert.NoError(t, err)
	assert.Equal(t, "new_access", account.AccessToken())
	assert.Equal(t, "new_refresh", account.RefreshToken())

	err = s.RefreshPlatformToken(context.Background(), "user123", "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, errPlatformAccountNotFound, err)

	issuer.refreshFails = true
	err = s.RefreshPlatformToken(context.Background(), "user123", account.ID())
	assert.Error(t, err)
	assert.Equal(t, errTokenRefreshFailed, err)
}
