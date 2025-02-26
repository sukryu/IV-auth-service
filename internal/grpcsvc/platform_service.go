package grpcsvc

import (
	"context"

	"github.com/sukryu/IV-auth-services/internal/domain"
	authv1 "github.com/sukryu/IV-auth-services/internal/proto/auth/v1"
	"github.com/sukryu/IV-auth-services/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PlatformService는 gRPC 기반 플랫폼 계정 관리 서비스 어댑터입니다.
// proto/auth/v1.PlatformService 인터페이스를 구현하며, 도메인 서비스를 호출합니다.
type PlatformService struct {
	authv1.UnimplementedPlatformServiceServer
	domainSvc *services.PlatformService
}

// NewPlatformService는 새로운 gRPC PlatformService 인스턴스를 생성하여 반환합니다.
// 도메인 PlatformService를 주입받아 초기화합니다.
func NewPlatformService(domainSvc *services.PlatformService) *PlatformService {
	return &PlatformService{domainSvc: domainSvc}
}

// ConnectPlatform은 사용자의 외부 플랫폼 계정을 연동합니다.
// proto/auth/v1.ConnectPlatform RPC를 구현합니다.
func (s *PlatformService) ConnectPlatform(ctx context.Context, req *authv1.ConnectPlatformRequest) (*authv1.PlatformAccountResponse, error) {
	if req.UserId == "" || req.Platform == "" || req.AuthCode == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, platform, and auth_code are required")
	}

	account, err := s.domainSvc.ConnectPlatform(ctx, req.UserId, domain.PlatformType(req.Platform), req.AuthCode)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case domain.ErrUserNotActive:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "failed to connect platform: %v", err)
		}
	}

	return &authv1.PlatformAccountResponse{
		Account: &authv1.PlatformAccount{
			Id:               account.ID,
			UserId:           account.UserID,
			Platform:         string(account.Platform),
			PlatformUserId:   account.PlatformUserID,
			PlatformUsername: account.PlatformUsername,
			AccessToken:      account.AccessToken,
			RefreshToken:     account.RefreshToken,
			TokenExpiresAt:   timestamppb.New(account.TokenExpiresAt),
			CreatedAt:        timestamppb.New(account.CreatedAt),
			UpdatedAt:        timestamppb.New(account.UpdatedAt),
		},
	}, nil
}

// RefreshPlatformToken은 플랫폼 계정의 토큰을 갱신합니다.
// proto/auth/v1.RefreshPlatformToken RPC를 구현합니다.
func (s *PlatformService) RefreshPlatformToken(ctx context.Context, req *authv1.RefreshPlatformTokenRequest) (*authv1.PlatformAccountResponse, error) {
	if req.PlatformAccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "platform_account_id is required")
	}

	account, err := s.domainSvc.RefreshPlatformToken(ctx, req.PlatformAccountId)
	if err != nil {
		switch err {
		case domain.ErrPlatformAccountNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case domain.ErrUserNotActive:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "failed to refresh platform token: %v", err)
		}
	}

	return &authv1.PlatformAccountResponse{
		Account: &authv1.PlatformAccount{
			Id:               account.ID,
			UserId:           account.UserID,
			Platform:         string(account.Platform),
			PlatformUserId:   account.PlatformUserID,
			PlatformUsername: account.PlatformUsername,
			AccessToken:      account.AccessToken,
			RefreshToken:     account.RefreshToken,
			TokenExpiresAt:   timestamppb.New(account.TokenExpiresAt),
			CreatedAt:        timestamppb.New(account.CreatedAt),
			UpdatedAt:        timestamppb.New(account.UpdatedAt),
		},
	}, nil
}

// DisconnectPlatform은 플랫폼 계정을 해제합니다.
// proto/auth/v1.DisconnectPlatform RPC를 구현합니다.
func (s *PlatformService) DisconnectPlatform(ctx context.Context, req *authv1.DisconnectPlatformRequest) (*authv1.DisconnectPlatformResponse, error) {
	if req.PlatformAccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "platform_account_id is required")
	}

	err := s.domainSvc.DisconnectPlatform(ctx, req.PlatformAccountId)
	if err != nil {
		switch err {
		case domain.ErrPlatformAccountNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case domain.ErrUserNotActive:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "failed to disconnect platform: %v", err)
		}
	}

	return &authv1.DisconnectPlatformResponse{Success: true}, nil
}
