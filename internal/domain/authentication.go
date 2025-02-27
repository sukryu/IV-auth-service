package domain

import (
	"context"
	"fmt"
	"time"
)

// TokenRepository defines the interface for token blacklist management
type TokenRepository interface {
	InsertTokenBlacklist(ctx context.Context, tokenID, userID string, expiresAt time.Time, reason string) error
	IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error)
}

// AuthenticationDomainService handles authentication-related business logic
type AuthenticationDomainService struct {
	userRepo    UserRepository
	tokenRepo   TokenRepository
	tokenIssuer TokenIssuer
}

// NewAuthenticationDomainService creates a new AuthenticationDomainService instance
func NewAuthenticationDomainService(userRepo UserRepository, tokenRepo TokenRepository, tokenIssuer TokenIssuer) *AuthenticationDomainService {
	return &AuthenticationDomainService{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		tokenIssuer: tokenIssuer,
	}
}

// Login authenticates a user and issues tokens
func (s *AuthenticationDomainService) Login(ctx context.Context, username, password string) (*User, *TokenPair, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve user: %w", err)
	}
	if user == nil {
		return nil, nil, errUserNotFound
	}

	pwd := Password(user.PasswordHash())
	valid, err := pwd.Verify(password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify password: %w", err)
	}
	if !valid {
		return nil, nil, errInvalidCredentials
	}

	// 토큰 발급 (여기서는 TokenIssuer로 외부 플랫폼 토큰 대신 JWT로 가정)
	accessToken := "jwt_access_" + user.ID() // 실제 JWT 생성은 pkg/auth에서 처리 예정
	refreshToken := "jwt_refresh_" + user.ID()
	tokenPair, err := NewTokenPair(accessToken, refreshToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create token pair: %w", err)
	}

	user.RecordLogin()
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("failed to update user login time: %w", err)
	}

	return user, tokenPair, nil
}

// Logout invalidates a user's tokens
func (s *AuthenticationDomainService) Logout(ctx context.Context, userID, jti string) error {
	err := s.tokenRepo.InsertTokenBlacklist(ctx, jti, userID, time.Now().Add(15*time.Minute), "logout")
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	return nil
}

// ValidateToken checks if a token is valid and not blacklisted
func (s *AuthenticationDomainService) ValidateToken(ctx context.Context, tokenID string) (bool, error) {
	isBlacklisted, err := s.tokenRepo.IsTokenBlacklisted(ctx, tokenID)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if isBlacklisted {
		return false, nil
	}

	// 여기서 JWT 검증 로직 추가 가능 (pkg/auth에서 처리 예정)
	return true, nil
}

var (
	errInvalidCredentials = Error("invalid username or password")
)
