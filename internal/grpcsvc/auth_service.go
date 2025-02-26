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

// AuthService는 gRPC 기반 인증 서비스 어댑터입니다.
// proto/auth/v1.AuthService 인터페이스를 구현하며, 도메인 서비스를 호출합니다.
type AuthService struct {
	authv1.UnimplementedAuthServiceServer
	domainSvc *services.AuthService
}

// NewAuthService는 새로운 gRPC AuthService 인스턴스를 생성하여 반환합니다.
// 도메인 AuthService를 주입받아 초기화합니다.
func NewAuthService(domainSvc *services.AuthService) *AuthService {
	return &AuthService{domainSvc: domainSvc}
}

// Login은 사용자 자격 증명을 검증하고 토큰 쌍을 발행합니다.
// proto/auth/v1.Login RPC를 구현합니다.
func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	tokenPair, err := s.domainSvc.Login(ctx, req.Username, req.Password)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			return nil, status.Error(codes.Unauthenticated, err.Error())
		case domain.ErrUserNotActive:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "failed to login: %v", err)
		}
	}

	return &authv1.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    timestamppb.New(tokenPair.ExpiresAt),
	}, nil
}

// ValidateToken은 액세스 토큰의 유효성을 검증합니다.
func (s *AuthService) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access token is required")
	}

	userID, roles, err := s.domainSvc.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		switch err {
		case domain.ErrInvalidToken:
			return nil, status.Error(codes.Unauthenticated, err.Error())
		case domain.ErrTokenBlacklisted:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "failed to validate token: %v", err)
		}
	}

	return &authv1.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
		Roles:  roles,
	}, nil
}

// RefreshToken은 리프레시 토큰으로 새 토큰 쌍을 발행합니다.
func (s *AuthService) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	tokenPair, err := s.domainSvc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		switch err {
		case domain.ErrInvalidToken:
			return nil, status.Error(codes.Unauthenticated, err.Error())
		case domain.ErrTokenBlacklisted:
			return nil, status.Error(codes.PermissionDenied, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "failed to refresh token: %v", err)
		}
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    timestamppb.New(tokenPair.ExpiresAt),
	}, nil
}

// BlacklistToken은 토큰을 블랙리스트에 추가합니다.
func (s *AuthService) BlacklistToken(ctx context.Context, req *authv1.BlacklistTokenRequest) (*authv1.BlacklistTokenResponse, error) {
	if req.Token == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "token and user_id are required")
	}

	err := s.domainSvc.BlacklistToken(ctx, req.Token, req.UserId, req.Reason)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to blacklist token: %v", err)
	}

	return &authv1.BlacklistTokenResponse{Success: true}, nil
}
