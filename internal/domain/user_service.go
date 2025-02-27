package domain

import (
	"context"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	InsertUser(ctx context.Context, user *User) error
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

// UserDomainService handles user-related business logic
type UserDomainService struct {
	repo UserRepository
}

// NewUserDomainService creates a new UserDomainService instance
func NewUserDomainService(repo UserRepository) *UserDomainService {
	return &UserDomainService{repo: repo}
}

// CreateUser creates a new user with the given details
func (s *UserDomainService) CreateUser(ctx context.Context, username, email, password string) (*User, error) {
	// 중복 username 검증
	exists, err := s.repo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errUserAlreadyExists
	}

	// 비밀번호 해싱
	pwd, err := NewPassword(password)
	if err != nil {
		return nil, err
	}

	// 사용자 생성
	user, err := NewUser(username, email, pwd.String())
	if err != nil {
		return nil, err
	}

	// 저장소에 저장
	if err := s.repo.InsertUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangeUserStatus updates the status of an existing user
func (s *UserDomainService) ChangeUserStatus(ctx context.Context, username, newStatus string) error {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	if user == nil {
		return errUserNotFound
	}

	user.SetStatus(newStatus)
	return s.repo.UpdateUser(ctx, user)
}

// RecordLogin updates the last login time for a user
func (s *UserDomainService) RecordLogin(ctx context.Context, username string) error {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	if user == nil {
		return errUserNotFound
	}

	user.RecordLogin()
	return s.repo.UpdateUser(ctx, user)
}

var (
	errUserAlreadyExists = Error("user already exists")
	errUserNotFound      = Error("user not found")
)
