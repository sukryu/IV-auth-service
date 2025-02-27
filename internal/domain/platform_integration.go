package domain

import (
	"context"
	"time"
)

// PlatformAccountRepository defines the interface for platform account data access
type PlatformAccountRepository interface {
	InsertPlatformAccount(ctx context.Context, account *PlatformAccount) error
	GetPlatformAccountsByUserID(ctx context.Context, userID string) ([]*PlatformAccount, error)
	UpdatePlatformAccount(ctx context.Context, account *PlatformAccount) error
}

// TokenIssuer defines the interface for issuing external platform tokens
type TokenIssuer interface {
	IssueTokens(ctx context.Context, authCode string) (string, string, time.Time, error)       // 이름 제거
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, time.Time, error) // 이름 제거
}

// PlatformIntegrationDomainService handles platform account integration logic
type PlatformIntegrationDomainService struct {
	repo        PlatformAccountRepository
	tokenIssuer TokenIssuer
}

// NewPlatformIntegrationDomainService creates a new instance
func NewPlatformIntegrationDomainService(repo PlatformAccountRepository, tokenIssuer TokenIssuer) *PlatformIntegrationDomainService {
	return &PlatformIntegrationDomainService{
		repo:        repo,
		tokenIssuer: tokenIssuer,
	}
}

// ConnectPlatformAccount connects a user's external platform account
func (s *PlatformIntegrationDomainService) ConnectPlatformAccount(ctx context.Context, userID, platform, platformUserID, platformUsername, authCode string) (*PlatformAccount, error) {
	accessToken, refreshToken, expiresAt, err := s.tokenIssuer.IssueTokens(ctx, authCode)
	if err != nil {
		return nil, err
	}

	account, err := NewPlatformAccount(userID, platform, platformUserID, platformUsername, accessToken, refreshToken, &expiresAt)
	if err != nil {
		return nil, err
	}

	if err := s.repo.InsertPlatformAccount(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// RefreshPlatformToken refreshes the tokens for an existing platform account
func (s *PlatformIntegrationDomainService) RefreshPlatformToken(ctx context.Context, userID, platformAccountID string) error {
	accounts, err := s.repo.GetPlatformAccountsByUserID(ctx, userID)
	if err != nil {
		return err
	}

	var targetAccount *PlatformAccount
	for _, acc := range accounts {
		if acc.ID() == platformAccountID {
			targetAccount = acc
			break
		}
	}
	if targetAccount == nil {
		return errPlatformAccountNotFound
	}

	accessToken, refreshToken, expiresAt, err := s.tokenIssuer.RefreshTokens(ctx, targetAccount.RefreshToken())
	if err != nil {
		return err
	}

	targetAccount.SetTokens(accessToken, refreshToken, &expiresAt)
	return s.repo.UpdatePlatformAccount(ctx, targetAccount)
}

var (
	errPlatformAccountNotFound = Error("platform account not found")
)
