# 도메인 모델 설계 및 구현 설명

본 문서는 **ImmersiVerse Authentication Service**의 도메인 계층 설계 및 구현에 대해 설명합니다. 도메인 모델은 서비스의 핵심 비즈니스 규칙을 캡슐화하며, 외부 인프라(데이터베이스, 메시징 등)와는 분리된 순수한 비즈니스 로직을 포함합니다.

---

## 1. 개요

- **목표**:  
  - 인증과 사용자 관리와 관련된 비즈니스 규칙을 명확히 표현  
  - 도메인 전문가와 개발자 간의 공통 언어(유비쿼터스 언어)를 확립  
  - 테스트 용이성 및 유지보수성을 높여, 서비스 확장에 유연하게 대응

- **설계 원칙**:
  - **도메인 주도 설계(DDD)**: 엔티티, 값 객체, 도메인 서비스, 도메인 이벤트를 활용하여 핵심 비즈니스 규칙을 모델링
  - **헥사고날 아키텍처**: 애플리케이션 코어(도메인 모델)와 인프라(데이터베이스, 외부 API 등)를 명확히 분리
  - **유비쿼터스 언어**: 도메인 전문가와의 소통을 위해 용어와 개념을 통일

---

## 2. 주요 도메인 엔티티

### 2.1 User (사용자)

- **설명**: 플랫폼 내 사용자(예: 스트리머, 일반 사용자)의 기본 정보를 관리하는 엔티티  
- **주요 속성**:
  - `ID`: UUID 형식의 고유 식별자
  - `Username`: 고유 사용자명
  - `Email`: 사용자 이메일 (unique)
  - `PasswordHash`: Argon2id 해싱된 비밀번호
  - `Status`: 계정 상태 (예: ACTIVE, SUSPENDED)
  - `SubscriptionTier`: 구독 등급 (예: FREE, PREMIUM)
  - `CreatedAt`, `UpdatedAt`, `LastLoginAt`: 생성, 수정, 마지막 로그인 시각

- **역할**:  
  - 사용자 인증, 권한 부여, 로그인/로그아웃 상태 관리 등 핵심 기능 제공

### 2.2 PlatformAccount (플랫폼 계정)

- **설명**: 외부 스트리밍 플랫폼(예: Twitch, YouTube 등)과 연동된 계정 정보를 저장  
- **주요 속성**:
  - `ID`: UUID
  - `UserID`: 연결된 User 엔티티의 참조
  - `Platform`: 플랫폼 종류 (TWITCH, YOUTUBE, 등)
  - `PlatformUserID`: 외부 플랫폼의 사용자 식별자
  - `PlatformUsername`: 외부 플랫폼에서의 닉네임
  - `AccessToken`, `RefreshToken`, `TokenExpiresAt`: OAuth 토큰 관련 정보

- **역할**:  
  - 외부 플랫폼과의 계정 연동 및 토큰 관리 기능 지원

### 2.3 Token (토큰)

- **설명**: 인증 토큰 관련 데이터 및 검증 로직을 캡슐화  
- **주요 속성**:
  - `AccessToken`: 단기 토큰 (JWT)
  - `RefreshToken`: 장기 토큰
  - `IssuedAt`, `ExpiresAt`: 발급 및 만료 시각
  - `JTI`: 토큰 고유 식별자

- **역할**:  
  - JWT 생성 및 검증, 토큰 무효화(블랙리스트) 등 관리

---

## 3. 값 객체 (Value Objects)

### 3.1 Password

- **설명**: 비밀번호 해싱, 검증 및 정책 관련 로직을 포함  
- **역할**:
  - 입력된 비밀번호를 Argon2id 알고리즘으로 해싱하고, 저장된 해시와 비교
  - 솔트와(필요 시) 페퍼를 적용하여 보안을 강화

### 3.2 Email

- **설명**: 이메일의 유효성 검증, 정규화 로직 포함  
- **역할**:  
  - 입력된 이메일의 형식 검증  
  - 소문자 변환 및 공백 제거 등 정규화 수행

### 3.3 TokenPair

- **설명**: 액세스 토큰과 리프레시 토큰 쌍을 관리  
- **역할**:  
  - JWT 발급 시 두 토큰을 함께 반환하고, 이후 갱신 절차에 활용

---

## 4. 도메인 서비스

### 4.1 UserDomainService

- **설명**: 사용자 생성, 업데이트, 상태 변경 등 사용자 관련 비즈니스 로직을 처리  
- **주요 기능**:
  - 신규 사용자 등록 시 중복 검사, 비밀번호 해싱 처리  
  - 계정 상태 변경, 로그인 성공/실패 처리 로직 포함  
  - 감사 이벤트 발행을 위한 로직 내재화

### 4.2 PlatformIntegrationDomainService

- **설명**: 외부 플랫폼 계정 연동 및 토큰 갱신 로직을 담당  
- **주요 기능**:
  - OAuth 인증 코드 교환 및 토큰 수신 처리  
  - 외부 API 호출 후 받은 토큰 정보의 저장과 갱신
  - 연동 계정의 상태 변경 및 이벤트 발행

### 4.3 AuthenticationDomainService

- **설명**: 사용자 인증 및 토큰 관리와 관련된 핵심 규칙을 구현  
- **주요 기능**:
  - 로그인 시 비밀번호 검증 및 JWT 토큰 생성  
  - 토큰 검증, 갱신 및 무효화 처리  
  - 실패/성공 이벤트에 따른 처리 로직 내재화

---

## 5. 도메인 이벤트

- **UserCreated**: 신규 사용자 생성 시 발생  
- **UserUpdated**: 사용자 정보 수정 시 발생  
- **LoginSucceeded/LoginFailed**: 로그인 성공 또는 실패 시 발생  
- **PlatformConnected/PlatformDisconnected**: 외부 플랫폼 계정 연결/해제 이벤트

각 이벤트는 Kafka를 통해 발행되어, 다른 마이크로서비스(예: Analytics, Security 등)와 연동됩니다.

---

## 6. 구현 예시

### 6.1 사용자 생성 로직 (Pseudo Go Code)

```go
package domain

import (
    "context"
    "time"
    "github.com/google/uuid"
)

// User 엔티티 정의
type User struct {
    ID              string
    Username        string
    Email           string
    PasswordHash    string
    Status          string
    SubscriptionTier string
    CreatedAt       time.Time
    UpdatedAt       time.Time
    LastLoginAt     time.Time
}

// UserDomainService: 사용자 생성 로직
type UserDomainService struct {
    repo UserRepository // 인터페이스로 정의된 User 데이터 접근 계층
    // 기타 의존성(예: 이벤트 발행기)
}

// CreateUser: 신규 사용자 등록 처리
func (s *UserDomainService) CreateUser(ctx context.Context, username, email, password string) (*User, error) {
    // 1. 중복 검사
    exists, err := s.repo.ExistsByUsername(ctx, username)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, ErrUserAlreadyExists
    }
    
    // 2. 비밀번호 해싱 (Password 값 객체 활용)
    hash, err := HashPassword(password) // HashPassword 함수는 Argon2id 구현
    if err != nil {
        return nil, err
    }
    
    // 3. 사용자 엔티티 생성
    user := &User{
        ID:              uuid.New().String(),
        Username:        username,
        Email:           email,
        PasswordHash:    hash,
        Status:          "ACTIVE",
        SubscriptionTier: "FREE",
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }
    
    // 4. DB에 저장
    if err := s.repo.InsertUser(ctx, user); err != nil {
        return nil, err
    }
    
    // 5. UserCreated 도메인 이벤트 발행 (옵션)
    // s.eventPublisher.Publish(UserCreated{...})
    
    return user, nil
}
```

---

## 7. 결론

**도메인 모델 설계 및 구현**은 **Authentication Service**의 비즈니스 로직을 명확히 표현하고, 유지보수 및 확장성을 높이기 위한 핵심 요소입니다.

- **핵심 엔티티**: User, PlatformAccount, Token  
- **값 객체**: Password, Email, TokenPair  
- **도메인 서비스**: 사용자 관리, 플랫폼 연동, 인증 관련 핵심 로직 구현  
- **도메인 이벤트**: 사용자 및 인증 관련 상태 변화 전달

이 문서를 바탕으로 도메인 모델을 계속 발전시키고, 도메인 규칙이 변경될 때마다 업데이트하여 팀 내 지식 공유와 시스템의 일관성을 유지하시기 바랍니다.

---