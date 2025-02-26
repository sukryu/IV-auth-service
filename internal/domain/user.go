package domain

import (
	"errors"
	"time"
)

// User는 사용자 엔티티를 표현합니다.
type User struct {
	ID               string
	Username         string
	Email            string
	PasswordHash     string
	Status           UserStatus
	SubscriptionTier string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	LastLoginAt      time.Time
	Roles            []string
}

// UserStatus는 사용자 계정의 상태를 표현합니다.
type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
	UserStatusInactive  UserStatus = "INACTIVE"
	UserStatusDeleted   UserStatus = "DELETED"
)

// NewUser는 새로운 User 엔티티를 생성합니다.
func NewUser(id, username, email, passwordHash string) (*User, error) {
	if id == "" {
		return nil, errors.New("사용자 ID는 필수입니다")
	}
	if username == "" {
		return nil, errors.New("사용자명은 필수입니다")
	}
	if email == "" {
		return nil, errors.New("이메일은 필수입니다")
	}
	if passwordHash == "" {
		return nil, errors.New("비밀번호 해시는 필수입니다")
	}

	now := time.Now()
	return &User{
		ID:               id,
		Username:         username,
		Email:            email,
		PasswordHash:     passwordHash,
		Status:           UserStatusActive,
		SubscriptionTier: "FREE", // 기본 구독 티어
		CreatedAt:        now,
		UpdatedAt:        now,
		Roles:            []string{"USER"}, // 기본 역할
	}, nil
}

// IsActive는 사용자 계정이 활성 상태인지 확인합니다.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// UpdatePassword는 사용자의 비밀번호 해시를 업데이트합니다.
func (u *User) UpdatePassword(passwordHash string) error {
	if passwordHash == "" {
		return errors.New("비밀번호 해시는 필수입니다")
	}

	u.PasswordHash = passwordHash
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateStatus는 사용자의 상태를 변경합니다.
func (u *User) UpdateStatus(status UserStatus) error {
	if status == "" {
		return errors.New("상태 값은 필수입니다")
	}

	u.Status = status
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateEmail은 사용자의 이메일을 업데이트합니다.
func (u *User) UpdateEmail(email string) error {
	if email == "" {
		return errors.New("이메일은 필수입니다")
	}

	u.Email = email
	u.UpdatedAt = time.Now()
	return nil
}

// RecordLogin은 사용자의 마지막 로그인 시간을 업데이트합니다.
func (u *User) RecordLogin() {
	u.LastLoginAt = time.Now()
}

// HasRole은 사용자가 특정 역할을 가지고 있는지 확인합니다.
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// AddRole은 사용자에게 역할을 추가합니다.
func (u *User) AddRole(role string) {
	// 이미 역할이 있는지 확인
	if u.HasRole(role) {
		return
	}
	u.Roles = append(u.Roles, role)
	u.UpdatedAt = time.Now()
}

// RemoveRole은 사용자에서 역할을 제거합니다.
func (u *User) RemoveRole(role string) {
	newRoles := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		if r != role {
			newRoles = append(newRoles, r)
		}
	}
	u.Roles = newRoles
	u.UpdatedAt = time.Now()
}
