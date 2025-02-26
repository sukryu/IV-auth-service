package grpcsvc

import (
	"context"

	"github.com/sukryu/IV-auth-services/internal/domain"
	authv1 "github.com/sukryu/IV-auth-services/internal/proto/auth/v1"
	"github.com/sukryu/IV-auth-services/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status" // 추가
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserService는 gRPC 기반 사용자 관리 서비스 어댑터입니다.
// proto/auth/v1.UserService 인터페이스를 구현하며, 도메인 서비스를 호출합니다.
type UserService struct {
	authv1.UnimplementedUserServiceServer
	domainSvc *services.UserService
}

// NewUserService는 새로운 gRPC UserService 인스턴스를 생성하여 반환합니다.
// 도메인 UserService를 주입받아 초기화합니다.
func NewUserService(domainSvc *services.UserService) *UserService {
	return &UserService{domainSvc: domainSvc}
}

// CreateUser는 새로운 사용자를 생성합니다.
// proto/auth/v1.CreateUser RPC를 구현합니다.
func (s *UserService) CreateUser(ctx context.Context, req *authv1.CreateUserRequest) (*authv1.UserResponse, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username, email, and password are required")
	}

	user, err := s.domainSvc.CreateUser(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		switch err {
		case domain.ErrUserAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
		}
	}

	return &authv1.UserResponse{
		User: &authv1.User{
			Id:               user.ID,
			Username:         user.Username,
			Email:            user.Email,
			Status:           string(user.Status),
			SubscriptionTier: user.SubscriptionTier,
			CreatedAt:        timestamppb.New(user.CreatedAt),
			UpdatedAt:        timestamppb.New(user.UpdatedAt),
			LastLoginAt:      timestamppb.New(user.LastLoginAt),
			Roles:            user.Roles,
		},
	}, nil
}

// GetUserByID는 사용자 ID로 사용자 정보를 조회합니다.
// proto/auth/v1.GetUserByID RPC를 구현합니다.
func (s *UserService) GetUserByID(ctx context.Context, req *authv1.GetUserByIDRequest) (*authv1.UserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := s.domainSvc.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user by id: %v", err)
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, domain.ErrUserNotFound.Error())
	}

	return &authv1.UserResponse{
		User: &authv1.User{
			Id:               user.ID,
			Username:         user.Username,
			Email:            user.Email,
			Status:           string(user.Status),
			SubscriptionTier: user.SubscriptionTier,
			CreatedAt:        timestamppb.New(user.CreatedAt),
			UpdatedAt:        timestamppb.New(user.UpdatedAt),
			LastLoginAt:      timestamppb.New(user.LastLoginAt),
			Roles:            user.Roles,
		},
	}, nil
}

// GetUserByUsername은 사용자명으로 사용자 정보를 조회합니다.
// proto/auth/v1.GetUserByUsername RPC를 구현합니다.
func (s *UserService) GetUserByUsername(ctx context.Context, req *authv1.GetUserByUsernameRequest) (*authv1.UserResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	user, err := s.domainSvc.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user by username: %v", err)
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, domain.ErrUserNotFound.Error())
	}

	return &authv1.UserResponse{
		User: &authv1.User{
			Id:               user.ID,
			Username:         user.Username,
			Email:            user.Email,
			Status:           string(user.Status),
			SubscriptionTier: user.SubscriptionTier,
			CreatedAt:        timestamppb.New(user.CreatedAt),
			UpdatedAt:        timestamppb.New(user.UpdatedAt),
			LastLoginAt:      timestamppb.New(user.LastLoginAt),
			Roles:            user.Roles,
		},
	}, nil
}

// UpdateUser는 사용자 정보를 수정합니다.
// proto/auth/v1.UpdateUser RPC를 구현합니다.
func (s *UserService) UpdateUser(ctx context.Context, req *authv1.UpdateUserRequest) (*authv1.UserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	var email, password *string
	var statusVal *domain.UserStatus
	if req.Email != "" {
		email = &req.Email
	}
	if req.Password != "" {
		password = &req.Password
	}
	if req.Status != "" {
		s := domain.UserStatus(req.Status)
		statusVal = &s
	}

	user, err := s.domainSvc.UpdateUser(ctx, req.UserId, email, password, statusVal, req.Roles)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
		}
	}

	return &authv1.UserResponse{
		User: &authv1.User{
			Id:               user.ID,
			Username:         user.Username,
			Email:            user.Email,
			Status:           string(user.Status),
			SubscriptionTier: user.SubscriptionTier,
			CreatedAt:        timestamppb.New(user.CreatedAt),
			UpdatedAt:        timestamppb.New(user.UpdatedAt),
			LastLoginAt:      timestamppb.New(user.LastLoginAt),
			Roles:            user.Roles,
		},
	}, nil
}

// DeleteUser는 사용자 ID로 사용자를 삭제합니다.
// proto/auth/v1.DeleteUser RPC를 구현합니다.
func (s *UserService) DeleteUser(ctx context.Context, req *authv1.DeleteUserRequest) (*authv1.DeleteUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.domainSvc.DeleteUser(ctx, req.UserId)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
		}
	}

	return &authv1.DeleteUserResponse{Success: true}, nil
}
