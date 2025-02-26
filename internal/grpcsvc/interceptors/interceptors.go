package interceptors

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/sukryu/IV-auth-services/internal/domain"
	authv1 "github.com/sukryu/IV-auth-services/internal/proto/auth/v1"
	"github.com/sukryu/IV-auth-services/internal/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor는 gRPC 요청/응답에 대한 로깅을 처리하는 Unary 인터셉터입니다.
// 요청 메서드, 처리 시간, 상태 코드를 기록합니다.
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	log.Printf("Received request: %s", info.FullMethod)

	resp, err := handler(ctx, req)
	duration := time.Since(start)

	statusCode := codes.OK
	if err != nil {
		statusCode = status.Code(err)
	}
	log.Printf("Completed request: %s, Status: %s, Duration: %v", info.FullMethod, statusCode.String(), duration)
	return resp, err
}

// AuthInterceptor는 gRPC 요청에 대한 인증을 처리하는 Unary 인터셉터입니다.
// Authorization 헤더에서 토큰을 검증하며, 인증 실패 시 요청을 차단합니다.
func AuthInterceptor(authSvc *services.AuthService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip 인증이 필요 없는 메서드 (예: Login)
		if strings.HasSuffix(info.FullMethod, "/Login") {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders, ok := md["authorization"]
		if !ok || len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token := strings.TrimPrefix(authHeaders[0], "Bearer ")
		userID, roles, err := authSvc.ValidateToken(ctx, token)
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

		// 인증된 사용자 정보를 컨텍스트에 추가
		ctx = context.WithValue(ctx, "user_id", userID)
		ctx = context.WithValue(ctx, "roles", roles)
		return handler(ctx, req)
	}
}

// ValidationInterceptor는 gRPC 요청의 입력 유효성을 검증하는 Unary 인터셉터입니다.
// 현재는 기본적인 null 체크만 수행하며, 필요 시 확장 가능합니다.
func ValidationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 요청별 유효성 검사 (예시로 간단히 처리)
	switch r := req.(type) {
	case *authv1.LoginRequest:
		if r.Username == "" || r.Password == "" {
			return nil, status.Error(codes.InvalidArgument, "username and password are required")
		}
	case *authv1.CreateUserRequest:
		if r.Username == "" || r.Email == "" || r.Password == "" {
			return nil, status.Error(codes.InvalidArgument, "username, email, and password are required")
		}
		// 다른 요청 타입 추가 가능
	}

	return handler(ctx, req)
}

// ChainUnaryInterceptors는 여러 Unary 인터셉터를 체인으로 연결합니다.
// 로깅, 인증, 검증 순서로 적용됩니다.
func ChainUnaryInterceptors(authSvc *services.AuthService) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		LoggingInterceptor,
		AuthInterceptor(authSvc),
		ValidationInterceptor,
	)
}

// Stream 인터셉터도 필요 시 추가 가능 (현재는 Unary만 구현)
