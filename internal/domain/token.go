package domain

import (
	"errors"
	"time"
)

// TokenPair는 액세스 토큰과 리프레시 토큰의 쌍을 표현합니다.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// NewTokenPair는 새로운 토큰 쌍을 생성합니다.
func NewTokenPair(accessToken, refreshToken string, expiresAt time.Time) (*TokenPair, error) {
	if accessToken == "" {
		return nil, errors.New("액세스 토큰은 필수입니다")
	}
	if refreshToken == "" {
		return nil, errors.New("리프레시 토큰은 필수입니다")
	}
	if expiresAt.IsZero() {
		return nil, errors.New("만료 시각은 필수입니다")
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// TokenBlacklist는 무효화된 토큰 정보를 표현합니다.
type TokenBlacklist struct {
	TokenID       string
	UserID        string
	ExpiresAt     time.Time
	Reason        string
	BlacklistedAt time.Time
}

// NewTokenBlacklist는 토큰 블랙리스트 항목을 생성합니다.
func NewTokenBlacklist(tokenID, userID string, expiresAt time.Time, reason string) (*TokenBlacklist, error) {
	if tokenID == "" {
		return nil, errors.New("토큰 ID는 필수입니다")
	}
	if userID == "" {
		return nil, errors.New("사용자 ID는 필수입니다")
	}
	if expiresAt.IsZero() {
		return nil, errors.New("만료 시각은 필수입니다")
	}

	return &TokenBlacklist{
		TokenID:       tokenID,
		UserID:        userID,
		ExpiresAt:     expiresAt,
		Reason:        reason,
		BlacklistedAt: time.Now(),
	}, nil
}
