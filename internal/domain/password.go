package domain

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Password represents an immutable password value object with hashing and validation
type Password string

// NewPassword hashes a plain-text password using Argon2id
func NewPassword(plainText string) (Password, error) {
	if len(plainText) < 12 {
		return "", fmt.Errorf("password must be at least 12 characters long")
	}

	// Argon2id 파라미터: time=3, memory=64MB, threads=1, key length=32
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(plainText), salt, 3, 64*1024, 1, 32)
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	// PHC 형식으로 저장: $argon2id$v=19$m=65536,t=3,p=1$salt$hash
	return Password(fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=1$%s$%s", encodedSalt, encodedHash)), nil
}

// Verify checks if a plain-text password matches the hashed password
func (p Password) Verify(plainText string) (bool, error) {
	parts := strings.Split(string(p), "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, fmt.Errorf("invalid password hash format")
	}

	// 파라미터 파싱 (m=65536,t=3,p=1)
	params := strings.Split(parts[3], ",")
	if len(params) != 3 || params[0] != "m=65536" || params[1] != "t=3" || params[2] != "p=1" {
		return false, fmt.Errorf("invalid argon2id parameters")
	}

	// Salt와 Hash 디코딩
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// 입력 비밀번호로 동일 조건에서 해시 생성
	computedHash := argon2.IDKey([]byte(plainText), salt, 3, 64*1024, 1, 32)

	// 상수 시간 비교
	return subtle.ConstantTimeCompare(hash, computedHash) == 1, nil
}

// String returns the hashed password as a string
func (p Password) String() string {
	return string(p)
}
