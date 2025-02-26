package domain

import "fmt"

// ErrInvalidUserID는 사용자 ID가 유효하지 않을 때 반환되는 에러입니다.
var ErrInvalidUserID = NewError("invalid user id")

// ErrInvalidUsername는 사용자명이 유효하지 않을 때 반환되는 에러입니다.
var ErrInvalidUsername = NewError("invalid username")

// ErrInvalidTokenID는 토큰 ID가 유효하지 않을 때 반환되는 에러입니다.
var ErrInvalidTokenID = NewError("invalid token id")

// ErrInvalidEventReason는 이벤트 사유가 유효하지 않을 때 반환되는 에러입니다.
var ErrInvalidEventReason = NewError("invalid event reason")

// ErrInvalidExpiresAt는 만료 시각이 유효하지 않을 때 반환되는 에러입니다.
var ErrInvalidExpiresAt = NewError("invalid expires at time")

// ErrInvalidCredentials는 사용자 자격 증명(예: 비밀번호)이 잘못되었을 때 반환되는 에러입니다.
// 로그인 시 비밀번호 불일치나 사용자 미존재 상황에서 사용됩니다.
var ErrInvalidCredentials = NewError("invalid credentials")

// ErrUserNotActive는 사용자가 활성 상태가 아닐 때 반환되는 에러입니다.
// 계정이 SUSPENDED, DELETED 등 비활성 상태일 경우 발생합니다.
var ErrUserNotActive = NewError("user not active")

// ErrInvalidToken는 토큰이 유효하지 않을 때 반환되는 에러입니다.
// JWT 서명 오류, 만료, 형식 문제 등에서 사용됩니다.
var ErrInvalidToken = NewError("invalid token")

// ErrTokenBlacklisted는 토큰이 블랙리스트에 포함되어 있을 때 반환되는 에러입니다.
// 로그아웃이나 보안 문제로 무효화된 토큰에 대해 발생합니다.
var ErrTokenBlacklisted = NewError("token blacklisted")

// ErrUserAlreadyExists는 사용자 생성 시 이미 존재하는 사용자명일 때 반환되는 에러입니다.
var ErrUserAlreadyExists = NewError("user already exists")

// ErrUserNotFound는 사용자 조회 시 사용자가 존재하지 않을 때 반환되는 에러입니다.
var ErrUserNotFound = NewError("user not found")

// NewError는 도메인 수준의 사용자 정의 에러를 생성합니다.
// K8s 스타일에 따라 소문자 메시지를 사용하며, 필요 시 wrapping으로 디테일을 추가할 수 있습니다.
func NewError(msg string) error {
	return &domainError{msg: msg}
}

// domainError는 도메인 관련 에러를 표현하는 내부 구조체입니다.
type domainError struct {
	msg string
}

// Error는 domainError의 문자열 표현을 반환합니다.
// K8s 스타일에 맞춰 간결한 메시지를 제공합니다.
func (e *domainError) Error() string {
	return e.msg
}

// Wrap은 기존 에러에 추가 컨텍스트를 덧붙여 새로운 에러를 반환합니다.
// fmt.Errorf와 유사하며, 디버깅 시 유용한 세부 정보를 제공합니다.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &domainError{msg: fmt.Sprintf("%s: %v", msg, err)}
}
