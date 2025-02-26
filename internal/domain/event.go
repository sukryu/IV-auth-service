package domain

import (
	"time"
)

// LoginSucceeded는 사용자가 성공적으로 로그인했을 때 발생하는 도메인 이벤트입니다.
// 로그인 성공 시각, 사용자 ID, 사용자명 등의 정보를 포함하여 다른 서비스(예: 분석, 모니터링)로 전달됩니다.
type LoginSucceeded struct {
	UserID    string    // 로그인한 사용자의 고유 식별자 (UUID)
	Username  string    // 로그인에 사용된 사용자명
	Timestamp time.Time // 로그인 성공 시각 (UTC)
	IPAddress string    // 로그인 요청의 클라이언트 IP 주소 (IPv4/IPv6)
	UserAgent string    // 로그인 요청의 User-Agent 헤더
}

// NewLoginSucceeded는 새로운 LoginSucceeded 이벤트를 생성합니다.
// 필수 필드(UserID, Username)를 검증하고, 현재 시각으로 Timestamp를 설정합니다.
func NewLoginSucceeded(userID, username, ipAddress, userAgent string) (*LoginSucceeded, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if username == "" {
		return nil, ErrInvalidUsername
	}

	return &LoginSucceeded{
		UserID:    userID,
		Username:  username,
		Timestamp: time.Now().UTC(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}, nil
}

// LoginFailed는 사용자의 로그인 시도가 실패했을 때 발생하는 도메인 이벤트입니다.
// 실패 이유와 함께 사용자 정보를 기록하여 보안 모니터링 및 감사에 활용됩니다.
type LoginFailed struct {
	UserID    string    // 로그인 시도 대상 사용자의 ID (없을 경우 빈 문자열 가능)
	Username  string    // 시도에 사용된 사용자명
	Reason    string    // 실패 사유 (예: "invalid password", "user not found")
	Timestamp time.Time // 로그인 실패 시각 (UTC)
	IPAddress string    // 실패 요청의 클라이언트 IP 주소
	UserAgent string    // 실패 요청의 User-Agent 헤더
}

// NewLoginFailed는 새로운 LoginFailed 이벤트를 생성합니다.
// Username은 필수이며, 실패 이유를 명확히 기록하여 이벤트 발행 시 유효성을 보장합니다.
func NewLoginFailed(userID, username, reason, ipAddress, userAgent string) (*LoginFailed, error) {
	if username == "" {
		return nil, ErrInvalidUsername
	}
	if reason == "" {
		return nil, ErrInvalidEventReason
	}

	return &LoginFailed{
		UserID:    userID,
		Username:  username,
		Reason:    reason,
		Timestamp: time.Now().UTC(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}, nil
}

// TokenBlacklisted는 토큰이 블랙리스트에 추가되었을 때 발생하는 도메인 이벤트입니다.
// 보안 문제나 로그아웃 시 발생하며, 토큰 무효화 정보를 다른 시스템으로 전달합니다.
type TokenBlacklisted struct {
	TokenID   string    // 블랙리스트에 추가된 토큰의 고유 식별자 (JTI)
	UserID    string    // 토큰 소유자의 사용자 ID
	Reason    string    // 블랙리스트 사유 (예: "logout", "compromised")
	Timestamp time.Time // 블랙리스트 등록 시각 (UTC)
	ExpiresAt time.Time // 토큰 원래 만료 시각
	IPAddress string    // 요청 IP 주소 (옵션, 로그아웃 시 사용)
}

// NewTokenBlacklisted는 새로운 TokenBlacklisted 이벤트를 생성합니다.
// TokenID와 UserID는 필수이며, ExpiresAt으로 토큰의 유효성을 추적합니다.
func NewTokenBlacklisted(tokenID, userID, reason string, expiresAt time.Time, ipAddress string) (*TokenBlacklisted, error) {
	if tokenID == "" {
		return nil, ErrInvalidTokenID
	}
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if reason == "" {
		return nil, ErrInvalidEventReason
	}
	if expiresAt.IsZero() {
		return nil, ErrInvalidExpiresAt
	}

	return &TokenBlacklisted{
		TokenID:   tokenID,
		UserID:    userID,
		Reason:    reason,
		Timestamp: time.Now().UTC(),
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
	}, nil
}

// UserCreated는 새로운 사용자가 생성되었을 때 발생하는 도메인 이벤트입니다.
// 사용자 생성 시각과 기본 정보를 기록하여 다른 시스템(예: 분석, 알림)으로 전달됩니다.
type UserCreated struct {
	UserID    string    // 생성된 사용자의 고유 식별자 (UUID)
	Username  string    // 생성된 사용자명
	Timestamp time.Time // 사용자 생성 시각 (UTC)
}

// NewUserCreated는 새로운 UserCreated 이벤트를 생성합니다.
// UserID와 Username은 필수이며, 현재 시각으로 Timestamp를 설정합니다.
func NewUserCreated(userID, username string) (*UserCreated, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if username == "" {
		return nil, ErrInvalidUsername
	}

	return &UserCreated{
		UserID:    userID,
		Username:  username,
		Timestamp: time.Now().UTC(),
	}, nil
}

// UserUpdated는 사용자 정보가 수정되었을 때 발생하는 도메인 이벤트입니다.
// 변경 전후 값을 기록하여 감사 및 동기화에 활용됩니다.
type UserUpdated struct {
	UserID    string                 // 수정된 사용자의 고유 식별자 (UUID)
	Timestamp time.Time              // 수정 시각 (UTC)
	OldValues map[string]interface{} // 변경 전 값 (예: email, status, roles)
	NewValues map[string]interface{} // 변경 후 값
}

// NewUserUpdated는 새로운 UserUpdated 이벤트를 생성합니다.
// UserID와 변경 값은 필수이며, 현재 시각으로 Timestamp를 설정합니다.
func NewUserUpdated(userID string, oldValues, newValues map[string]interface{}) (*UserUpdated, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if oldValues == nil || newValues == nil {
		return nil, ErrInvalidEventValues
	}

	return &UserUpdated{
		UserID:    userID,
		Timestamp: time.Now().UTC(),
		OldValues: oldValues,
		NewValues: newValues,
	}, nil
}

// UserDeleted는 사용자가 삭제되었을 때 발생하는 도메인 이벤트입니다.
// 삭제 시각과 사용자 ID를 기록하여 다른 시스템과의 동기화를 지원합니다.
type UserDeleted struct {
	UserID    string    // 삭제된 사용자의 고유 식별자 (UUID)
	Timestamp time.Time // 삭제 시각 (UTC)
}

// NewUserDeleted는 새로운 UserDeleted 이벤트를 생성합니다.
// UserID는 필수이며, 현재 시각으로 Timestamp를 설정합니다.
func NewUserDeleted(userID string) (*UserDeleted, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	return &UserDeleted{
		UserID:    userID,
		Timestamp: time.Now().UTC(),
	}, nil
}

// PlatformConnected는 플랫폼 계정이 성공적으로 연동되었을 때 발생하는 도메인 이벤트입니다.
// 연동된 계정 정보를 기록하여 동기화 및 알림에 활용됩니다.
type PlatformConnected struct {
	UserID            string       // 연동된 사용자의 ID
	PlatformAccountID string       // 연동된 플랫폼 계정의 ID
	Platform          PlatformType // 연동된 플랫폼 종류
	Timestamp         time.Time    // 연동 시각 (UTC)
}

// NewPlatformConnected는 새로운 PlatformConnected 이벤트를 생성합니다.
// UserID와 PlatformAccountID는 필수입니다.
func NewPlatformConnected(userID, platformAccountID string, platform PlatformType) (*PlatformConnected, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if platformAccountID == "" {
		return nil, ErrInvalidPlatformAccountID
	}
	if platform == "" {
		return nil, ErrInvalidPlatform
	}

	return &PlatformConnected{
		UserID:            userID,
		PlatformAccountID: platformAccountID,
		Platform:          platform,
		Timestamp:         time.Now().UTC(),
	}, nil
}

// PlatformDisconnected는 플랫폼 계정이 해제되었을 때 발생하는 도메인 이벤트입니다.
// 해제된 계정 정보를 기록하여 동기화에 활용됩니다.
type PlatformDisconnected struct {
	UserID            string       // 해제된 사용자의 ID
	PlatformAccountID string       // 해제된 플랫폼 계정의 ID
	Platform          PlatformType // 해제된 플랫폼 종류
	Timestamp         time.Time    // 해제 시각 (UTC)
}

// NewPlatformDisconnected는 새로운 PlatformDisconnected 이벤트를 생성합니다.
// UserID와 PlatformAccountID는 필수입니다.
func NewPlatformDisconnected(userID, platformAccountID string, platform PlatformType) (*PlatformDisconnected, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if platformAccountID == "" {
		return nil, ErrInvalidPlatformAccountID
	}
	if platform == "" {
		return nil, ErrInvalidPlatform
	}

	return &PlatformDisconnected{
		UserID:            userID,
		PlatformAccountID: platformAccountID,
		Platform:          platform,
		Timestamp:         time.Now().UTC(),
	}, nil
}

// PlatformTokenRefreshed는 플랫폼 계정의 토큰이 갱신되었을 때 발생하는 도메인 이벤트입니다.
// 토큰 갱신 시각과 계정 정보를 기록합니다.
type PlatformTokenRefreshed struct {
	UserID            string       // 사용자의 ID
	PlatformAccountID string       // 갱신된 플랫폼 계정의 ID
	Platform          PlatformType // 갱신된 플랫폼 종류
	Timestamp         time.Time    // 갱신 시각 (UTC)
}

// NewPlatformTokenRefreshed는 새로운 PlatformTokenRefreshed 이벤트를 생성합니다.
// UserID와 PlatformAccountID는 필수입니다.
func NewPlatformTokenRefreshed(userID, platformAccountID string, platform PlatformType) (*PlatformTokenRefreshed, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if platformAccountID == "" {
		return nil, ErrInvalidPlatformAccountID
	}
	if platform == "" {
		return nil, ErrInvalidPlatform
	}

	return &PlatformTokenRefreshed{
		UserID:            userID,
		PlatformAccountID: platformAccountID,
		Platform:          platform,
		Timestamp:         time.Now().UTC(),
	}, nil
}

// PlatformConnectionFailed는 플랫폼 계정 연동이 실패했을 때 발생하는 도메인 이벤트입니다.
// 실패 사유와 사용자 정보를 기록하여 보안 및 디버깅에 활용됩니다.
type PlatformConnectionFailed struct {
	UserID    string       // 연동 시도한 사용자의 ID
	Platform  PlatformType // 연동 시도한 플랫폼 종류
	Reason    string       // 실패 사유 (예: "oauth failure")
	Timestamp time.Time    // 실패 시각 (UTC)
}

// NewPlatformConnectionFailed는 새로운 PlatformConnectionFailed 이벤트를 생성합니다.
// UserID와 Reason은 필수입니다.
func NewPlatformConnectionFailed(userID string, platform PlatformType, reason string) (*PlatformConnectionFailed, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if platform == "" {
		return nil, ErrInvalidPlatform
	}
	if reason == "" {
		return nil, ErrInvalidEventReason
	}

	return &PlatformConnectionFailed{
		UserID:    userID,
		Platform:  platform,
		Reason:    reason,
		Timestamp: time.Now().UTC(),
	}, nil
}

// PlatformTokenRefreshFailed는 플랫폼 계정의 토큰 갱신이 실패했을 때 발생하는 도메인 이벤트입니다.
// 실패 사유와 계정 정보를 기록하여 보안 및 디버깅에 활용됩니다.
type PlatformTokenRefreshFailed struct {
	UserID            string    // 사용자의 ID
	PlatformAccountID string    // 갱신 실패한 플랫폼 계정의 ID
	Reason            string    // 실패 사유 (예: "invalid refresh token")
	Timestamp         time.Time // 실패 시각 (UTC)
}

// NewPlatformTokenRefreshFailed는 새로운 PlatformTokenRefreshFailed 이벤트를 생성합니다.
// UserID, PlatformAccountID, Reason은 필수입니다.
func NewPlatformTokenRefreshFailed(userID, platformAccountID, reason string) (*PlatformTokenRefreshFailed, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if platformAccountID == "" {
		return nil, ErrInvalidPlatformAccountID
	}
	if reason == "" {
		return nil, ErrInvalidEventReason
	}

	return &PlatformTokenRefreshFailed{
		UserID:            userID,
		PlatformAccountID: platformAccountID,
		Reason:            reason,
		Timestamp:         time.Now().UTC(),
	}, nil
}

// ErrInvalidEventValues는 이벤트 값이 유효하지 않을 때 반환되는 에러입니다.
// UserUpdated 등에서 OldValues/NewValues가 누락된 경우 사용됩니다.
var ErrInvalidEventValues = NewError("invalid event values")

// internal/domain/errors.go (추가된 부분)
// ErrPlatformAccountNotFound는 플랫폼 계정이 존재하지 않을 때 반환되는 에러입니다.
var ErrPlatformAccountNotFound = NewError("platform account not found")

// ErrInvalidPlatform는 플랫폼 종류가 유효하지 않을 때 반환되는 에러입니다.
var ErrInvalidPlatform = NewError("invalid platform")

// ErrInvalidPlatformAccountID는 플랫폼 계정 ID가 유효하지 않을 때 반환되는 에러입니다.
var ErrInvalidPlatformAccountID = NewError("invalid platform account id")
