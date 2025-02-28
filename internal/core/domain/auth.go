package domain

import (
	"context"
	"crypto/rand"
	"errors"
	"time"
)

// AuthService defines the authentication-related operations.
type AuthService interface {
	Authenticate(ctx context.Context, username, password string) (*Token, error)
	GenerateTokenPair(userID string) (*Token, error)
	Logout(tokenID string) error
	ValidateToken(tokenStr string) (string, error) // Returns userID
	RefreshToken(refreshTokenStr string) (*Token, error)
}

// authService implements AuthService with domain logic.
type authService struct {
	userRepo UserRepository // 사용자 조회를 위한 저장소 (미구현)
	tokenGen TokenGenerator // 토큰 생성기 (미구현)
	eventPub EventPublisher // 이벤트 발행기 (미구현)
}

// UserRepository defines the interface for user data access (placeholder).
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	// 추가 메서드 필요 시 정의
}

// TokenGenerator defines the interface for generating tokens (placeholder).
type TokenGenerator interface {
	GenerateAccessToken(userID string, expiry time.Time) (string, error)
	GenerateRefreshToken(userID string, expiry time.Time) (string, error)
	ValidateToken(tokenStr string) (string, error) // Returns userID
}

// EventPublisher defines the interface for publishing domain events (placeholder).
type EventPublisher interface {
	Publish(event interface{}) error
}

// NewAuthService creates a new instance of authService.
func NewAuthService(userRepo UserRepository, tokenGen TokenGenerator, eventPub EventPublisher) AuthService {
	return &authService{
		userRepo: userRepo,
		tokenGen: tokenGen,
		eventPub: eventPub,
	}
}

// Authenticate verifies user credentials and returns a token pair.
func (s *authService) Authenticate(ctx context.Context, username, password string) (*Token, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password must not be empty")
	}

	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("failed to find user: " + err.Error())
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if !user.PasswordHash().Verify(password) {
		_ = s.eventPub.Publish(&LoginFailed{userID: user.ID(), timestamp: time.Now()})
		return nil, errors.New("invalid password")
	}

	token, err := s.GenerateTokenPair(user.ID())
	if err != nil {
		return nil, err
	}

	user.SetLastLoginAt(time.Now())
	_ = s.eventPub.Publish(&LoginSucceeded{userID: user.ID(), timestamp: time.Now()})
	return token, nil
}

// GenerateTokenPair generates a new access and refresh token pair for a user.
func (s *authService) GenerateTokenPair(userID string) (*Token, error) {
	if userID == "" {
		return nil, errors.New("user id must not be empty")
	}

	// JTI 생성
	jti := generateRandomString(16)
	accessExpiry := time.Now().Add(15 * time.Minute)    // 15분 만료
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour) // 7일 만료

	accessToken, err := s.tokenGen.GenerateAccessToken(userID, accessExpiry)
	if err != nil {
		return nil, errors.New("failed to generate access token: " + err.Error())
	}
	refreshToken, err := s.tokenGen.GenerateRefreshToken(userID, refreshExpiry)
	if err != nil {
		return nil, errors.New("failed to generate refresh token: " + err.Error())
	}

	return NewToken(accessToken, refreshToken, jti, accessExpiry)
}

// Logout adds a token to the blacklist (implementation pending).
func (s *authService) Logout(tokenID string) error {
	if tokenID == "" {
		return errors.New("token id must not be empty")
	}
	// 블랙리스트 추가 로직은 이후 저장소 구현 시 추가
	return nil
}

// ValidateToken verifies the validity of a token and returns the user ID.
func (s *authService) ValidateToken(tokenStr string) (string, error) {
	if tokenStr == "" {
		return "", errors.New("token must not be empty")
	}

	userID, err := s.tokenGen.ValidateToken(tokenStr)
	if err != nil {
		return "", errors.New("invalid token: " + err.Error())
	}
	return userID, nil
}

// RefreshToken generates a new token pair using a valid refresh token.
func (s *authService) RefreshToken(refreshTokenStr string) (*Token, error) {
	if refreshTokenStr == "" {
		return nil, errors.New("refresh token must not be empty")
	}

	userID, err := s.tokenGen.ValidateToken(refreshTokenStr)
	if err != nil {
		return nil, errors.New("invalid refresh token: " + err.Error())
	}

	return s.GenerateTokenPair(userID)
}

// generateRandomString generates a random string of given length.
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	_, _ = rand.Read(b) // 에러 무시 (테스트 용도로 단순화)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}
