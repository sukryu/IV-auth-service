# 서비스 로직 구현 설명

본 문서는 **ImmersiVerse Authentication Service**의 서비스 로직 계층에 대한 개요와 구현 방식을 설명합니다. 서비스 로직은 도메인 모델의 규칙을 실제 비즈니스 요구사항에 맞게 적용하고, 외부 인터페이스(gRPC, Kafka, Redis 등)와 통합하는 핵심 계층입니다.

---

## 1. 서비스 로직 계층 개요

- **목표**:  
  - 사용자 인증, 계정 관리, 토큰 생성/검증 등의 비즈니스 규칙을 구현  
  - 도메인 모델과 인프라스트럭처(Repository, Messaging 등) 간의 중재 역할 수행  
  - 재사용 가능하고 테스트 용이한 모듈로 구성

- **구성 요소**:  
  1. **AuthService**: 사용자 로그인, 로그아웃, 토큰 갱신 등 인증 관련 핵심 기능  
  2. **UserService**: 사용자 생성, 조회, 업데이트, 삭제 등의 사용자 관리 기능  
  3. **PlatformAccountService**: 외부 플랫폼 계정 연동 및 관리  
  4. **TokenService**: JWT 발급, 검증, 블랙리스트 처리 등 토큰 관리 로직

---

## 2. 주요 기능 및 처리 흐름

### 2.1 사용자 인증 및 토큰 관리 (AuthService)

- **로그인 처리**:
  - 클라이언트로부터 `LoginRequest` 수신
  - **UserRepository**를 통해 사용자의 존재 여부 및 비밀번호 해싱된 값 조회
  - **Argon2id** 기반 비밀번호 비교 수행
  - 인증 성공 시, **JWT** 생성 (Access Token 및 Refresh Token)
  - 생성된 토큰을 포함한 `LoginResponse` 반환
  - 성공/실패 이벤트를 Kafka에 발행하여, 보안 및 모니터링 연동

- **토큰 검증**:
  - 클라이언트 또는 내부 서비스가 `ValidateToken` 요청 시,  
    - JWT 서명, 만료 여부, 블랙리스트 상태를 확인
  - 검증 결과를 기반으로 유효한 토큰이면 사용자 정보를 반환, 그렇지 않으면 인증 오류 처리

- **토큰 갱신**:
  - 만료된 Access Token에 대해 Refresh Token을 사용해 새로운 토큰 발급
  - Refresh Token의 유효성 및 블랙리스트 여부 확인 후 갱신 진행

### 2.2 사용자 관리 (UserService)

- **회원가입 (CreateUser)**:
  - 입력 데이터 검증 및 중복 체크 수행
  - 비밀번호 해싱(Argon2id) 후, 신규 사용자 데이터 생성
  - **UserRepository**를 통해 DB에 사용자 정보 저장
  - 성공 시, `UserCreated` 도메인 이벤트를 Kafka에 발행

- **사용자 정보 조회/수정/삭제**:
  - 사용자 ID를 기준으로 DB에서 데이터를 조회
  - 정보 변경 시, 변경 내역을 `UserUpdated` 이벤트로 기록(옵션)
  - 삭제 시, 연관된 감사 로그 및 플랫폼 계정 연동 데이터와 함께 정리

### 2.3 외부 플랫폼 계정 연동 (PlatformAccountService)

- **계정 연결**:
  - 외부 플랫폼의 OAuth 인증 코드를 받아, 해당 플랫폼 API를 통해 Access Token 및 Refresh Token 교환
  - 플랫폼 계정 데이터를 **PlatformAccountRepository**에 저장
  - `PlatformConnected` 이벤트 발행을 통해 다른 서비스와 연동

- **계정 해제**:
  - 사용자가 계정 연결 해제 요청 시, 해당 플랫폼 계정 정보를 DB에서 제거
  - 관련 이벤트(`PlatformDisconnected`)를 발행하여 후속 처리를 유도

### 2.4 토큰 관련 로직 (TokenService)

- **토큰 생성 및 검증**:
  - JWT 생성 시, 필요한 클레임(예: `sub`, `exp`, `jti` 등)을 설정하고 RS256 알고리즘으로 서명
  - 토큰 검증 과정에서 서명 확인, 만료 여부 체크 및 블랙리스트 확인 수행

- **토큰 무효화**:
  - 로그아웃 요청 시, 토큰의 `jti`를 **TokenBlacklist**에 등록하여 해당 토큰을 무효화
  - Refresh Token 갱신 시, 이전 토큰을 폐기하는 로직 포함

---

## 3. 서비스 로직 구현 방식

### 3.1 계층화된 아키텍처

- **도메인 계층**: 순수 비즈니스 로직(도메인 모델, 값 객체, 도메인 이벤트) 구현
- **애플리케이션 계층 (Service Layer)**:  
  - 도메인 계층의 기능을 활용하여 구체적인 유스케이스(로그인, 회원가입, 토큰 관리 등)를 구현  
  - 외부 시스템(Repository, Kafka, Redis)과의 통합 및 인터페이스 역할 수행
- **인프라스트럭처 계층**:  
  - 데이터 접근, 메시징, 캐싱 등 외부 의존성과의 통신을 담당  
  - 서비스 로직에 필요한 외부 시스템 모듈을 추상화하여 주입

### 3.2 인터페이스 기반 설계

- **UserRepository, TokenRepository 등** 인터페이스를 정의하여, 서비스 로직과 인프라스트럭처 간 결합도를 낮춤  
- 모킹(mocking)이나 스텁(stub)을 활용하여 단위 테스트 및 통합 테스트에서 쉽게 대체 가능

### 3.3 이벤트 발행

- 각 서비스 로직은 주요 상태 변경 시 **도메인 이벤트**를 발행하여, Kafka를 통해 다른 서비스(Analytics, Security 등)와 통합
- 이벤트 발행은 별도의 **EventPublisher** 인터페이스를 통해 추상화

---

## 4. 코드 예시

아래는 **UserService**에서 사용자 생성 로직의 간단한 예시입니다.

```go
package service

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/immersiverse/auth-service/internal/domain"
    "github.com/immersiverse/auth-service/internal/repository"
)

type UserService struct {
    userRepo    repository.UserRepository
    eventPub    EventPublisher
}

func NewUserService(userRepo repository.UserRepository, eventPub EventPublisher) *UserService {
    return &UserService{userRepo: userRepo, eventPub: eventPub}
}

func (s *UserService) CreateUser(ctx context.Context, username, email, password string) (*domain.User, error) {
    // 중복 체크 및 입력 검증 (생략)
    
    // 비밀번호 해싱
    hashedPwd, err := domain.HashPassword(password)
    if err != nil {
        return nil, err
    }
    
    // 사용자 엔티티 생성
    user := &domain.User{
        ID:               uuid.New().String(),
        Username:         username,
        Email:            email,
        PasswordHash:     hashedPwd,
        Status:           "ACTIVE",
        SubscriptionTier: "FREE",
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
    }
    
    // DB에 사용자 저장
    if err := s.userRepo.InsertUser(ctx, user); err != nil {
        return nil, err
    }
    
    // 도메인 이벤트 발행 (UserCreated)
    s.eventPub.Publish(ctx, domain.UserCreatedEvent{
        UserID:   user.ID,
        Username: user.Username,
        Email:    user.Email,
        Time:     user.CreatedAt,
    })
    
    return user, nil
}
```

---

## 5. 결론

**서비스 로직 구현**은 **Authentication Service**의 핵심 비즈니스 기능을 담당하며, 도메인 주도 설계와 헥사고날 아키텍처 원칙에 따라 깔끔하게 분리되어 있습니다.

- **주요 서비스**: AuthService, UserService, PlatformAccountService, TokenService
- **구현 방식**: 계층화된 아키텍처, 인터페이스 기반 설계, 이벤트 발행을 통한 다른 서비스와의 통합
- **유지보수**: 각 서비스의 독립적인 테스트, 모킹 및 인터페이스 추상화를 통해 코드 품질과 확장성을 보장

이 문서를 통해 팀원들은 Authentication Service의 서비스 로직 구조를 명확히 이해하고, 기능 추가나 수정 시 일관된 방식으로 작업할 수 있습니다.

---