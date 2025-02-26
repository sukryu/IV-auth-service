package domain

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Password는 비밀번호 해싱 및 검증을 위한 값 객체입니다.
type Password struct {
	plaintext *string
	hash      string
}

// NewPassword는 새 비밀번호 객체를 생성합니다(평문 비밀번호로부터).
func NewPassword(plaintext string) (*Password, error) {
	if plaintext == "" {
		return nil, errors.New("비밀번호는 필수입니다")
	}

	p := &Password{
		plaintext: &plaintext,
	}

	return p, nil
}

// NewPasswordFromHash는 해시된 비밀번호로부터 Password 객체를 생성합니다.
func NewPasswordFromHash(hash string) (*Password, error) {
	if hash == "" {
		return nil, errors.New("해시는 필수입니다")
	}

	// 해시 형식 검증은 생략 (실제 구현 시 추가)

	return &Password{
		hash: hash,
	}, nil
}

// GenerateHash는 비밀번호의 해시 값을 생성합니다 (Argon2id 사용).
func (p *Password) GenerateHash() error {
	if p.plaintext == nil {
		return errors.New("비밀번호가 설정되지 않았습니다")
	}

	// 사용할 Argon2id 파라미터 (실제 구현 시 조정)
	iterations := uint32(3)
	memory := uint32(64 * 1024)
	parallelism := uint8(2)
	saltLength := uint32(16)
	keyLength := uint32(32)

	// 솔트 생성 로직 (실제 구현 시 추가)
	salt := make([]byte, saltLength)
	// crypto/rand.Read(salt) 로직 추가 필요

	// Argon2id 해싱
	hash := argon2.IDKey(
		[]byte(*p.plaintext),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	// PHC 문자열 형식으로 해시 포맷
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	p.hash = fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory,
		iterations,
		parallelism,
		b64Salt,
		b64Hash,
	)

	// 이제 평문 비밀번호 메모리에서 제거
	p.plaintext = nil

	return nil
}

// Matches는 제공된 평문 비밀번호가 저장된 해시와 일치하는지 확인합니다.
func (p *Password) Matches(plaintext string) (bool, error) {
	if p.hash == "" {
		return false, errors.New("해시 값이 설정되지 않았습니다")
	}

	// PHC 문자열 구문 분석
	vals := strings.Split(p.hash, "$")
	if len(vals) != 6 {
		return false, errors.New("잘못된 해시 형식")
	}

	var variant, _, params, b64Salt, b64Hash string
	if len(vals) == 6 {
		variant, _, params, b64Salt, b64Hash = vals[1], vals[2], vals[3], vals[4], vals[5]
	}

	if variant != "argon2id" {
		return false, errors.New("지원되지 않는 해시 알고리즘")
	}

	// 파라미터 파싱
	var memory, iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(params, "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, fmt.Errorf("파라미터 파싱 실패: %w", err)
	}

	// 솔트와 해시 디코딩
	salt, err := base64.RawStdEncoding.DecodeString(b64Salt)
	if err != nil {
		return false, fmt.Errorf("솔트 디코딩 실패: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(b64Hash)
	if err != nil {
		return false, fmt.Errorf("해시 디코딩 실패: %w", err)
	}

	// 제공된 비밀번호 해싱
	compareHash := argon2.IDKey(
		[]byte(plaintext),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(hash)),
	)

	// 일정 시간 비교로 타이밍 공격 방지
	return subtle.ConstantTimeCompare(hash, compareHash) == 1, nil
}

// Hash는 비밀번호의 해시 값을 반환합니다.
func (p *Password) Hash() string {
	return p.hash
}
