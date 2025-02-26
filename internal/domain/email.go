package domain

import (
	"errors"
	"net/mail"
	"strings"
)

// Email은 이메일 주소 관련 기능을 제공하는 값 객체입니다.
type Email struct {
	address string
}

// NewEmail은 새로운 Email 객체를 생성합니다.
func NewEmail(address string) (*Email, error) {
	if address == "" {
		return nil, errors.New("이메일 주소는 필수입니다")
	}

	// 이메일 형식 검증
	if _, err := mail.ParseAddress(address); err != nil {
		return nil, errors.New("잘못된 이메일 형식입니다")
	}

	// 이메일 정규화 (소문자 변환)
	normalizedAddress := strings.ToLower(address)

	return &Email{
		address: normalizedAddress,
	}, nil
}

// Address는 정규화된 이메일 주소를 반환합니다.
func (e *Email) Address() string {
	return e.address
}

// Domain은 이메일 주소에서 도메인 부분을 추출합니다.
func (e *Email) Domain() string {
	parts := strings.Split(e.address, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// Username은 이메일 주소에서 사용자명 부분을 추출합니다.
func (e *Email) Username() string {
	parts := strings.Split(e.address, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}
