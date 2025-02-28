# gRPC 서비스 구현 설명

본 문서는 **ImmersiVerse Authentication Service**의 gRPC 서비스 계층 구현에 대해 설명합니다. gRPC는 내부 마이크로서비스 간의 효율적인 통신을 위해 사용되며, 강력한 타입 안정성, 낮은 지연 시간, 스트리밍 지원 등의 장점을 제공합니다.

---

## 1. 개요

- **목적**:  
  - 사용자 인증, 계정 관리, 토큰 처리 등의 핵심 기능을 gRPC API로 노출하여, 내부 서비스 및 API 게이트웨이와 통신  
  - 높은 성능과 확장성을 제공하면서도, 보안 및 오류 처리를 강화

- **주요 구성요소**:  
  - **gRPC 서비스 인터페이스**: Authentication, User, PlatformAccount, Token 관련 RPC 메서드 정의  
  - **인터셉터/미들웨어**: 인증, 로깅, 메트릭 수집, 입력 검증 등 공통 관심사를 처리  
  - **서비스 구현체**: 도메인 및 애플리케이션 계층의 로직을 gRPC 인터페이스에 매핑

---

## 2. gRPC 서비스 인터페이스

### 2.1 주요 서비스 정의

- **AuthService**  
  - `Login(LoginRequest) returns (LoginResponse)`
  - `Logout(LogoutRequest) returns (LogoutResponse)`
  - `RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse)`
  - `ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse)`

- **UserService**  
  - `CreateUser(CreateUserRequest) returns (UserResponse)`
  - `GetUser(GetUserRequest) returns (UserResponse)`
  - `UpdateUser(UpdateUserRequest) returns (UserResponse)`
  - `DeleteUser(DeleteUserRequest) returns (DeleteUserResponse)`

- **PlatformAccountService**  
  - `ConnectPlatformAccount(ConnectPlatformAccountRequest) returns (PlatformAccountResponse)`
  - `DisconnectPlatformAccount(DisconnectPlatformAccountRequest) returns (DisconnectPlatformAccountResponse)`
  - `GetPlatformAccounts(GetPlatformAccountsRequest) returns (GetPlatformAccountsResponse)`

*참고*: gRPC 서비스 정의는 Protocol Buffers (`proto/auth/v1/*.proto`)에 기술되어 있으며, 서비스 간 명확한 계약을 보장합니다.

---

## 3. 인터셉터 및 미들웨어

### 3.1 서버 인터셉터

- **AuthenticationInterceptor**:  
  - 모든 gRPC 요청에 대해 JWT 토큰 검증, 블랙리스트 확인, 유효하지 않은 요청 차단  
- **LoggingInterceptor**:  
  - 각 요청의 메서드, 요청 ID, 사용자 정보, 처리 시간 등을 구조화된 로그로 기록
- **ValidationInterceptor**:  
  - 입력 데이터의 유효성을 검사하여, 잘못된 데이터로 인한 오류 발생 방지
- **MetricsInterceptor**:  
  - Prometheus와 연동해 각 RPC 호출의 처리 시간, 오류율 등 성능 지표를 수집

### 3.2 클라이언트 인터셉터

- **ClientAuthInterceptor**:  
  - 외부 호출 시 JWT 토큰을 요청 헤더에 자동 추가하여 인증 유지
- **ClientLoggingInterceptor 및 ClientMetricsInterceptor**:  
  - gRPC 클라이언트 호출에 대한 로깅 및 메트릭 수집

---

## 4. 서비스 구현 방식

### 4.1 계층 구조

- **도메인 계층**: 비즈니스 로직과 핵심 도메인 모델 (예: User, Token, PlatformAccount)
- **애플리케이션 계층**: Use Case별 서비스 (예: AuthService, UserService) 구현  
  - gRPC 서비스 구현체는 이 계층의 함수를 호출하여 비즈니스 로직을 수행
- **인프라스트럭처 계층**: 데이터 접근(Repository), 메시징(Kafka), 캐싱(Redis) 등 외부 시스템과의 통합

### 4.2 구현 예시

아래는 `AuthService`의 `Login` 메서드 구현 예시입니다.

```go
package grpc

import (
    "context"
    "time"

    "github.com/immersiverse/auth-service/internal/domain"
    "github.com/immersiverse/auth-service/internal/service"
    authpb "github.com/immersiverse/auth-service/proto/auth/v1"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type AuthServiceServer struct {
    authService *service.AuthService
    authpb.UnimplementedAuthServiceServer
}

func NewAuthServiceServer(authService *service.AuthService) *AuthServiceServer {
    return &AuthServiceServer{authService: authService}
}

func (s *AuthServiceServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
    // 비즈니스 로직 호출: 사용자 인증 및 토큰 생성
    user, tokens, err := s.authService.Login(ctx, req.Username, req.Password)
    if err != nil {
        // 실패 시 적절한 gRPC 에러 반환
        return nil, status.Error(codes.Unauthenticated, "Invalid credentials")
    }
    
    // 성공 시 LoginResponse 생성 및 반환
    return &authpb.LoginResponse{
        UserId:       user.ID,
        AccessToken:  tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
        ExpiresAt:    tokens.ExpiresAt.Unix(),
    }, nil
}
```

---

## 5. 통합 및 배포

- **빌드 및 코드 생성**:  
  - Protocol Buffers를 이용해 gRPC 코드를 자동 생성 (`protoc` 명령어 활용)
- **CI/CD 연동**:  
  - gRPC 서비스 테스트 및 성능 검증을 포함한 빌드 자동화
- **Kubernetes 배포**:  
  - gRPC 서비스는 Kubernetes Deployment로 배포되며, 서비스 디스커버리 및 로드 밸런싱을 위해 ClusterIP 및 Ingress 설정과 연동

---

## 6. 결론

**gRPC 서비스 구현**은 **Authentication Service**의 핵심 기능을 외부와 통신할 수 있도록 하는 중요한 계층입니다.

- **서비스 인터페이스**: Protocol Buffers로 명확한 계약을 정의  
- **인터셉터**: 인증, 로깅, 입력 검증, 메트릭 수집을 통해 공통 관심사를 처리  
- **계층화된 구조**: 도메인, 애플리케이션, 인프라스트럭처 간 명확한 분리로 유지보수성 및 테스트 용이성 확보

이 문서를 기반으로 팀원들은 gRPC 서비스 구현 시 일관된 패턴을 따르고, 향후 서비스 확장 및 유지보수에 도움이 되는 코드를 작성할 수 있습니다.

---