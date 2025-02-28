package domain

import (
	"errors"
	"net/mail"
	"strings"
)

// Email represents an email address with validation and normalization logic.
type Email string

// NewEmail creates a new Email instance after validation and normalization.
func NewEmail(email string) (Email, error) {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return "", errors.New("email must not be empty")
	}

	// 유효성 검사: net/mail 패키지로 기본 형식 확인
	addr, err := mail.ParseAddress(trimmed)
	if err != nil || addr.Address != trimmed {
		return "", errors.New("invalid email format")
	}

	// 길이 제한 (DB 스키마 기준: 255자)
	if len(trimmed) > 255 {
		return "", errors.New("email exceeds maximum length of 255 characters")
	}

	// 정규화: 소문자로 변환
	normalized := strings.ToLower(trimmed)
	return Email(normalized), nil
}

// String returns the email as a string.
func (e Email) String() string {
	return string(e)
}

// IsValid checks if the email format is valid.
func (e Email) IsValid() bool {
	_, err := mail.ParseAddress(e.String())
	return err == nil && len(e) <= 255
}

// Domain returns the domain part of the email address.
func (e Email) Domain() string {
	parts := strings.SplitN(e.String(), "@", 2)
	if len(parts) < 2 {
		return "" // 유효하지 않은 경우 빈 문자열 반환
	}
	return parts[1]
}

// Normalize ensures the email is in lowercase (already normalized in NewEmail).
func (e Email) Normalize() Email {
	return Email(strings.ToLower(e.String()))
}
