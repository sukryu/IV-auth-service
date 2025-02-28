package domain

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"

	"golang.org/x/crypto/argon2"
)

// Password represents a securely hashed password.
type Password struct {
	hash []byte // Argon2id로 해싱된 값
	salt []byte // 해싱에 사용된 솔트
}

// NewPassword creates a new Password instance by hashing the raw password.
func NewPassword(rawPassword string) (Password, error) {
	if rawPassword == "" {
		return Password{}, errors.New("password must not be empty")
	}
	if len(rawPassword) < 8 {
		return Password{}, errors.New("password must be at least 8 characters")
	}

	// 솔트 생성
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return Password{}, errors.New("failed to generate salt: " + err.Error())
	}

	// Argon2id로 해싱 (메모리: 64MB, 반복: 3, 병렬: 1)
	hash := argon2.IDKey([]byte(rawPassword), salt, 3, 64*1024, 1, 32)

	return Password{
		hash: hash,
		salt: salt,
	}, nil
}

// NewPasswordFromHash creates a Password from an existing hash and salt (for DB retrieval).
func NewPasswordFromHash(hash, salt []byte) (Password, error) {
	if len(hash) == 0 {
		return Password{}, errors.New("hash must not be empty")
	}
	return Password{
		hash: hash,
		salt: salt, // salt가 없는 경우 nil 허용
	}, nil
}

// Hash returns the hashed password bytes.
func (p Password) Hash() []byte {
	return p.hash
}

// Salt returns the salt used for hashing.
func (p Password) Salt() []byte {
	return p.salt
}

// Verify checks if the provided raw password matches the stored hash.
func (p Password) Verify(rawPassword string) bool {
	hashToCompare := argon2.IDKey([]byte(rawPassword), p.salt, 3, 64*1024, 1, 32)
	return subtle.ConstantTimeCompare(p.hash, hashToCompare) == 1
}

// Change creates a new Password instance with a new raw password.
func (p Password) Change(rawPassword string) (Password, error) {
	return NewPassword(rawPassword)
}
