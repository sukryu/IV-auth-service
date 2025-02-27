package domain

import (
	"net/mail"
	"strings"
)

// Email represents an immutable email value object with validation and normalization
type Email string

// NewEmail validates and normalizes an email address
func NewEmail(email string) (Email, error) {
	// 공백 제거 및 소문자 변환
	normalized := strings.ToLower(strings.TrimSpace(email))

	// 이메일 형식 검증
	if _, err := mail.ParseAddress(normalized); err != nil {
		return "", err
	}

	// 추가 검증: 길이 제한 (최대 255자)
	if len(normalized) > 255 {
		return "", errEmailTooLong
	}

	return Email(normalized), nil
}

// String returns the normalized email address
func (e Email) String() string {
	return string(e)
}

var (
	errEmailTooLong = Error("email address exceeds 255 characters")
)

// Error is a custom error type for domain errors
type Error string

func (e Error) Error() string {
	return string(e)
}
