package domain

import (
	"errors"
	"time"
)

// PlatformAccount는 외부 플랫폼 계정과의 연동 정보를 표현합니다.
type PlatformAccount struct {
	ID               string
	UserID           string
	Platform         PlatformType
	PlatformUserID   string
	PlatformUsername string
	AccessToken      string
	RefreshToken     string
	TokenExpiresAt   time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// PlatformType은 연동 가능한 외부 플랫폼 종류를 정의하는 타입입니다.
// 문자열 기반으로 플랫폼을 식별하며, 유효한 값만 허용됩니다.
type PlatformType string

// PlatformType 상수 정의
const (
	PlatformTypeTwitch   PlatformType = "TWITCH"   // Twitch 플랫폼
	PlatformTypeYouTube  PlatformType = "YOUTUBE"  // YouTube 플랫폼
	PlatformTypeFacebook PlatformType = "FACEBOOK" // Facebook 플랫폼
	PlatformTypeAfreeca  PlatformType = "AFREECA"  // AfreecaTV 플랫폼
)

// validPlatforms는 허용된 PlatformType 값의 집합입니다.
// 유효성 검증에 사용되며, 새로운 플랫폼 추가 시 이 맵을 업데이트해야 합니다.
var validPlatforms = map[PlatformType]struct{}{
	PlatformTypeTwitch:   {},
	PlatformTypeYouTube:  {},
	PlatformTypeFacebook: {},
	PlatformTypeAfreeca:  {},
}

// IsValid는 PlatformType이 유효한지 확인합니다.
// 허용된 플랫폼 목록에 포함되지 않은 경우 false를 반환합니다.
func (p PlatformType) IsValid() bool {
	_, ok := validPlatforms[p]
	return ok
}

// String은 PlatformType의 문자열 표현을 반환합니다.
// K8s 스타일에 따라 간결한 변환을 제공합니다.
func (p PlatformType) String() string {
	return string(p)
}

// NewPlatformAccount는 새로운 외부 플랫폼 계정 연동 정보를 생성합니다.
// PlatformType의 유효성을 검증하며, 모든 필수 필드가 제공되어야 합니다.
func NewPlatformAccount(id, userID string, platform PlatformType,
	platformUserID, platformUsername, accessToken, refreshToken string,
	tokenExpiresAt time.Time) (*PlatformAccount, error) {

	if id == "" {
		return nil, errors.New("account id is required")
	}
	if userID == "" {
		return nil, errors.New("user id is required")
	}
	if !platform.IsValid() {
		return nil, ErrInvalidPlatform
	}
	if platformUserID == "" {
		return nil, errors.New("platform user id is required")
	}
	if platformUsername == "" {
		return nil, errors.New("platform username is required")
	}
	if accessToken == "" {
		return nil, errors.New("access token is required")
	}
	if tokenExpiresAt.IsZero() {
		return nil, errors.New("token expiration time is required")
	}

	now := time.Now()
	return &PlatformAccount{
		ID:               id,
		UserID:           userID,
		Platform:         platform,
		PlatformUserID:   platformUserID,
		PlatformUsername: platformUsername,
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenExpiresAt:   tokenExpiresAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// IsTokenExpired는 액세스 토큰이 만료되었는지 확인합니다.
// 현재 시각과 TokenExpiresAt을 비교하여 실시간 상태를 반환합니다.
func (pa *PlatformAccount) IsTokenExpired() bool {
	return time.Now().After(pa.TokenExpiresAt)
}

// UpdateTokens는 액세스 토큰과 리프레시 토큰을 업데이트합니다.
// 새로운 토큰 정보로 계정을 갱신하며, UpdatedAt을 현재 시각으로 설정합니다.
func (pa *PlatformAccount) UpdateTokens(accessToken, refreshToken string, expiresAt time.Time) error {
	if accessToken == "" {
		return errors.New("access token is required")
	}
	if expiresAt.IsZero() {
		return errors.New("token expiration time is required")
	}

	pa.AccessToken = accessToken
	pa.RefreshToken = refreshToken
	pa.TokenExpiresAt = expiresAt
	pa.UpdatedAt = time.Now()
	return nil
}

// UpdatePlatformUsername은 플랫폼 사용자명을 업데이트합니다.
// 새로운 사용자명으로 갱신하며, UpdatedAt을 현재 시각으로 설정합니다.
func (pa *PlatformAccount) UpdatePlatformUsername(username string) error {
	if username == "" {
		return errors.New("platform username is required")
	}

	pa.PlatformUsername = username
	pa.UpdatedAt = time.Now()
	return nil
}
