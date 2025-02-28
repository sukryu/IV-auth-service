package domain

import (
	"context"
	"errors"
	"time"
)

// UserManagementService defines operations for managing users.
type UserManagementService interface {
	CreateUser(ctx context.Context, username, email, password, subscriptionTier string) (*User, error)
	UpdateUserRole(ctx context.Context, userID, roleID string) error
}

// userManagementService implements UserManagementService with domain logic.
type userManagementService struct {
	userRepo UserRepository
	eventPub EventPublisher
}

// NewUserManagementService creates a new instance of userManagementService.
func NewUserManagementService(userRepo UserRepository, eventPub EventPublisher) UserManagementService {
	return &userManagementService{
		userRepo: userRepo,
		eventPub: eventPub,
	}
}

// CreateUser creates a new user with the given attributes.
func (s *userManagementService) CreateUser(ctx context.Context, username, emailStr, password, subscriptionTier string) (*User, error) {
	if username == "" || emailStr == "" || password == "" {
		return nil, errors.New("username, email, and password must not be empty")
	}

	email, err := NewEmail(emailStr)
	if err != nil {
		return nil, err
	}

	pwd, err := NewPassword(password)
	if err != nil {
		return nil, err
	}

	// 임시 UUID 생성 (실제 저장소에서 생성될 수 있음)
	userID := generateRandomString(36)
	user, err := NewUser(userID, username, email, pwd, nil, subscriptionTier)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.SaveUser(ctx, user); err != nil {
		return nil, errors.New("failed to save user: " + err.Error())
	}

	_ = s.eventPub.Publish(&UserCreated{userID: user.ID(), timestamp: user.CreatedAt()})
	return user, nil
}

// UpdateUserRole assigns a role to a user.
func (s *userManagementService) UpdateUserRole(ctx context.Context, userID, roleID string) error {
	if userID == "" || roleID == "" {
		return errors.New("user id and role id must not be empty")
	}

	// 사용자 조회
	user, err := s.userRepo.FindByUsername(ctx, "") // 임시로 빈 username (FindByID 필요 시 추가)
	if err != nil {
		return errors.New("failed to find user: " + err.Error())
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 역할 추가 (RoleIDs 업데이트)
	user.roleIDs = append(user.roleIDs, roleID)
	if err := s.userRepo.SaveUser(ctx, user); err != nil {
		return errors.New("failed to update user roles: " + err.Error())
	}

	_ = s.eventPub.Publish(&UserUpdated{userID: userID, timestamp: time.Now()})
	return nil
}
