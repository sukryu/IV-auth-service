package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/sukryu/IV-auth-services/internal/domain"
	"github.com/sukryu/IV-auth-services/internal/domain/repo"
)

// UserService는 사용자 관리와 관련된 핵심 비즈니스 로직을 처리하는 도메인 서비스입니다.
// 사용자 생성, 조회, 수정, 삭제(CRUD) 및 상태/역할 관리 기능을 제공하며, 외부 의존성은 인터페이스를 통해 주입받습니다.
type UserService struct {
	userRepo repo.UserRepository // 사용자 데이터 접근 인터페이스
	eventPub repo.EventPublisher // Kafka 이벤트 발행 인터페이스
}

// UserServiceConfig는 UserService 초기화를 위한 설정을 정의합니다.
type UserServiceConfig struct {
	UserRepo repo.UserRepository
	EventPub repo.EventPublisher
}

// NewUserService는 새로운 UserService 인스턴스를 생성하여 반환합니다.
// 모든 의존성은 UserServiceConfig를 통해 주입되며, 필수 필드 검증 후 초기화합니다.
func NewUserService(cfg UserServiceConfig) *UserService {
	if cfg.UserRepo == nil || cfg.EventPub == nil {
		log.Fatal("NewUserService: 모든 저장소 및 이벤트 발행기는 필수입니다")
	}

	return &UserService{
		userRepo: cfg.UserRepo,
		eventPub: cfg.EventPub,
	}
}

// CreateUser는 새로운 사용자를 생성합니다.
// 사용자명 중복 검사를 수행하고, 비밀번호를 해싱하며, 생성 후 이벤트를 발행합니다.
// 목표: 실시간 처리 (100ms 이내), 데이터 무결성 유지.
func (s *UserService) CreateUser(ctx context.Context, username, email, password string) (*domain.User, error) {
	// 사용자명 중복 체크
	exists, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if exists {
		return nil, domain.ErrUserAlreadyExists
	}

	// 비밀번호 해싱
	pwd, err := domain.NewPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to create password: %w", err)
	}
	if err := pwd.GenerateHash(); err != nil {
		return nil, fmt.Errorf("failed to generate password hash: %w", err)
	}

	// 사용자 엔티티 생성
	user, err := domain.NewUser(uuid.New().String(), username, email, pwd.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to create user entity: %w", err)
	}

	// 데이터베이스에 저장
	if err := s.userRepo.Insert(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	// UserCreated 이벤트 발행
	s.eventPub.Publish(ctx, domain.UserCreated{
		UserID:    user.ID,
		Username:  user.Username,
		Timestamp: time.Now().UTC(),
	})
	return user, nil
}

// GetUserByID는 사용자 ID를 기준으로 사용자 정보를 조회합니다.
// 사용자가 존재하지 않을 경우 nil과 에러를 반환하며, 조회 실패 시 에러를 처리합니다.
func (s *UserService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

// GetUserByUsername은 사용자명을 기준으로 사용자 정보를 조회합니다.
// 사용자가 존재하지 않을 경우 nil과 에러를 반환하며, 조회 실패 시 에러를 처리합니다.
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

// UpdateUser는 사용자 정보를 수정합니다.
// 이메일, 비밀번호, 상태, 역할 등을 갱신하며, 수정 후 이벤트를 발행합니다.
func (s *UserService) UpdateUser(ctx context.Context, id string, email, password *string, status *domain.UserStatus, roles []string) (*domain.User, error) {
	// 사용자 조회
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user for update: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// 변경 전 값 기록 (감사 로그용)
	oldValues := map[string]interface{}{
		"email":  user.Email,
		"status": string(user.Status),
		"roles":  user.Roles,
	}

	// 이메일 업데이트 (선택적)
	if email != nil {
		if err := user.UpdateEmail(*email); err != nil {
			return nil, fmt.Errorf("failed to update email: %w", err)
		}
	}

	// 비밀번호 업데이트 (선택적)
	if password != nil {
		pwd, err := domain.NewPassword(*password)
		if err != nil {
			return nil, fmt.Errorf("failed to create new password: %w", err)
		}
		if err := pwd.GenerateHash(); err != nil {
			return nil, fmt.Errorf("failed to generate new password hash: %w", err)
		}
		if err := user.UpdatePassword(pwd.Hash()); err != nil {
			return nil, fmt.Errorf("failed to update password: %w", err)
		}
	}

	// 상태 업데이트 (선택적)
	if status != nil {
		if err := user.UpdateStatus(*status); err != nil {
			return nil, fmt.Errorf("failed to update status: %w", err)
		}
	}

	// 역할 업데이트 (전체 덮어쓰기)
	if roles != nil {
		user.Roles = nil // 기존 역할 초기화
		for _, role := range roles {
			user.AddRole(role)
		}
	}

	// 데이터베이스에 반영
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// UserUpdated 이벤트 발행
	s.eventPub.Publish(ctx, domain.UserUpdated{
		UserID:    user.ID,
		Timestamp: time.Now().UTC(),
		OldValues: oldValues,
		NewValues: map[string]interface{}{
			"email":  user.Email,
			"status": string(user.Status),
			"roles":  user.Roles,
		},
	})
	return user, nil
}

// DeleteUser는 사용자 ID를 기준으로 사용자를 삭제합니다.
// 논리적 삭제(상태를 DELETED로 변경)로 구현하며, 삭제 후 이벤트를 발행합니다.
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	// 사용자 조회
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user for deletion: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// 상태를 DELETED로 변경
	if err := user.UpdateStatus(domain.UserStatusDeleted); err != nil {
		return fmt.Errorf("failed to set user status to deleted: %w", err)
	}

	// 데이터베이스 업데이트
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// UserDeleted 이벤트 발행
	s.eventPub.Publish(ctx, domain.UserDeleted{
		UserID:    user.ID,
		Timestamp: time.Now().UTC(),
	})
	return nil
}
