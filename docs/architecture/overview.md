# ImmersiVerse Authentication Service: 서비스 아키텍처 개요

## 1. 소개

ImmersiVerse Authentication Service는 플랫폼의 중앙 인증 및 계정 관리를 담당하는 마이크로서비스입니다. 본 문서는 서비스의 아키텍처 설계, 주요 컴포넌트, 데이터 흐름, 그리고 다른 서비스와의 통합 방식에 대한 포괄적인 개요를 제공합니다.

### 1.1 목적 및 범위

Authentication Service는 다음과 같은 주요 기능을 제공합니다:

- 사용자 인증 및 권한 부여
- 방송인 계정 관리
- 플랫폼 계정 통합 및 토큰 관리
- 접근 제어 정책 적용
- 보안 이벤트 모니터링 및 알림
- 사용자 인증 이벤트 스트리밍

이 서비스는 모든 인증 관련 기능을 중앙화하여 플랫폼 전체의 보안 일관성을 보장하고, 다른 서비스가 비즈니스 로직에 집중할 수 있도록 합니다.

## 2. 아키텍처 개요

### 2.1 아키텍처 스타일

Authentication Service는 다음과 같은 아키텍처 패턴과 원칙을 따릅니다:

- **마이크로서비스 아키텍처**: 독립적으로 배포 가능한 서비스로 설계
- **도메인 주도 설계(DDD)**: 명확한 도메인 모델과 경계를 정의
- **헥사고날 아키텍처**: 비즈니스 로직과 외부 시스템 간의 명확한 분리
- **이벤트 기반 아키텍처**: 주요 상태 변경은 이벤트로 발행
- **CQRS 패턴**: 명령과 쿼리를 분리하여 성능과 확장성 최적화

### 2.2 시스템 아키텍처 다이어그램

```
┌───────────────────┐        ┌──────────────────────────────────────┐
│                   │        │         Authentication Service       │
│   API Gateway     │◄──────►│                                      │
│   (Envoy Proxy)   │  gRPC  │ ┌────────────┐     ┌───────────────┐ │
└───────────────────┘        │ │            │     │               │ │
                             │ │   Domain   │     │  Application  │ │
┌───────────────────┐        │ │   Layer    │◄───►│     Layer     │ │
│                   │        │ │            │     │               │ │
│ Other Services    │◄──────►│ └────────────┘     └───────────────┘ │
│ (gRPC Clients)    │  gRPC  │         ▲                  ▲         │
└───────────────────┘        │         │                  │         │
                             │ ┌───────┴──────┐  ┌────────┴───────┐ │
┌───────────────────┐        │ │              │  │                │ │
│                   │        │ │ Repository   │  │ Infrastructure │ │
│ Event Consumers   │◄───────┤ │ Layer        │  │ Layer          │ │
│                   │ Kafka  │ │              │  │                │ │
└───────────────────┘        │ └──────┬───────┘  └────────┬───────┘ │
                             │        │                   │         │
                             └────────┼───────────────────┼─────────┘
                                      │                   │
                             ┌────────▼───────┐ ┌─────────▼─────────┐
                             │                │ │                   │
                             │  PostgreSQL    │ │      Kafka        │
                             │                │ │                   │
                             └────────────────┘ └───────────────────┘
                                      ▲                   ▲
                                      │                   │
                             ┌────────┴───────┐ ┌─────────┴─────────┐
                             │                │ │                   │
                             │     Redis      │ │   OpenTelemetry   │
                             │                │ │                   │
                             └────────────────┘ └───────────────────┘
```

## 3. 핵심 아키텍처 컴포넌트

### 3.1 도메인 레이어 (Domain Layer)

도메인 레이어는 비즈니스 개체와 규칙을 정의하며, 인프라나 애플리케이션의 세부 사항으로부터 격리됩니다.

#### 3.1.1 주요 도메인 엔티티

- **User**: 방송인 정보와 상태 관리
- **PlatformAccount**: 외부 스트리밍 플랫폼 계정 연동
- **Token**: 인증 토큰 표현 및 검증 로직
- **Role**: 사용자 역할 및 권한 관리
- **LoginEvent**: 로그인 시도 및 결과 기록

#### 3.1.2 도메인 서비스

- **UserDomainService**: 사용자 생성 및 상태 변경 관련 비즈니스 규칙
- **PlatformIntegrationDomainService**: 플랫폼 계정 연결 및 토큰 관리 규칙
- **AuthenticationDomainService**: 인증 및 권한 부여 핵심 규칙

#### 3.1.3 값 객체 (Value Objects)

- **Password**: 비밀번호 해싱, 검증 및 정책 관리
- **Email**: 이메일 유효성 검증 및 정규화
- **TokenPair**: 액세스 및 리프레시 토큰 쌍 관리

#### 3.1.4 도메인 이벤트

- **UserCreated**: 사용자 생성 이벤트
- **UserUpdated**: 사용자 정보 업데이트 이벤트
- **LoginSucceeded/LoginFailed**: 로그인 결과 이벤트
- **PlatformConnected/PlatformDisconnected**: 플랫폼 연결 상태 이벤트

### 3.2 애플리케이션 레이어 (Application Layer)

애플리케이션 레이어는 사용 사례를 구현하고 도메인 서비스를 조정합니다.

#### 3.2.1 서비스

- **AuthService**: 인증 관련 사용 사례 구현 (로그인, 로그아웃, 토큰 갱신)
- **UserService**: 사용자 관리 사용 사례 (생성, 조회, 업데이트)
- **PlatformAccountService**: 플랫폼 계정 연동 관리
- **TokenService**: 토큰 생성, 검증 및 갱신 로직

#### 3.2.2 명령/쿼리 처리기 (CQRS)

- **Commands**: CreateUser, UpdateUser, ConnectPlatform 등
- **Queries**: GetUserById, GetUserByUsername, ValidateToken 등

#### 3.2.3 이벤트 발행자

- **EventPublisher**: 도메인 이벤트를 Kafka 메시지로 변환하여 발행

### 3.3 인프라스트럭처 레이어 (Infrastructure Layer)

인프라스트럭처 레이어는 외부 시스템과의 통합을 구현합니다.

#### 3.3.1 외부 플랫폼 통합

- **TwitchClient**: Twitch OAuth API 통합
- **YouTubeClient**: YouTube API 통합
- **FacebookClient**: Facebook API 통합
- **AfreecaClient**: Afreeca API 통합

#### 3.3.2 메시징 및 이벤트

- **KafkaProducer**: Kafka를 통한 이벤트 발행
- **EventSerializer**: 이벤트 직렬화/역직렬화 처리

#### 3.3.3 캐싱

- **RedisTokenStore**: Redis를 활용한 토큰 캐싱 및 블랙리스트 관리
- **RedisCacheClient**: 고성능 데이터 캐싱

#### 3.3.4 메트릭 및 모니터링

- **PrometheusMetrics**: 메트릭 수집 및 노출
- **OpenTelemetryTracer**: 분산 추적 구현

### 3.4 리포지토리 레이어 (Repository Layer)

리포지토리 레이어는 데이터 접근 로직을 캡슐화합니다.

#### 3.4.1 사용자 데이터 관리

- **UserRepository**: 사용자 데이터 CRUD 작업
- **PlatformAccountRepository**: 플랫폼 계정 데이터 CRUD 작업

#### 3.4.2 인증 데이터 관리

- **TokenRepository**: 토큰 저장 및 검색
- **BlacklistRepository**: 무효화된 토큰 관리

#### 3.4.3 감사 및 로깅

- **AuditLogRepository**: 중요 작업에 대한 감사 로그 저장

### 3.5 gRPC 서비스 레이어

gRPC 서비스 레이어는 외부 시스템과의 통신 인터페이스를 제공합니다.

#### 3.5.1 서비스 구현체

- **AuthServiceServer**: 인증 관련 gRPC 서비스 구현
- **UserServiceServer**: 사용자 관리 gRPC 서비스 구현
- **TokenServiceServer**: 토큰 검증 gRPC 서비스 구현

#### 3.5.2 인터셉터 및 미들웨어

- **AuthenticationInterceptor**: 요청 인증 검증
- **LoggingInterceptor**: 요청 및 응답 로깅
- **ValidationInterceptor**: 입력 데이터 검증
- **MetricsInterceptor**: 요청 메트릭 수집

## 4. 데이터 흐름 및 통신 패턴

### 4.1 인증 흐름

#### 4.1.1 사용자 로그인 흐름

1. 클라이언트가 API 게이트웨이를 통해 로그인 요청 전송
2. API 게이트웨이가 gRPC를 통해 Authentication Service의 Login 메서드 호출
3. AuthService가 사용자 인증 처리:
   - UserRepository에서 사용자 정보 조회
   - 비밀번호 검증
   - 토큰 생성
4. 로그인 성공 시 LoginSucceeded 이벤트 발행 (Kafka)
5. 응답으로 액세스 토큰 및 리프레시 토큰 반환

#### 4.1.2 토큰 검증 흐름

1. 클라이언트가 보호된 엔드포인트 접근 시 Authorization 헤더에 토큰 포함
2. API 게이트웨이가 gRPC를 통해 Authentication Service의 ValidateToken 메서드 호출
3. TokenService가 토큰 검증:
   - 서명 확인
   - 만료 여부 확인
   - 블랙리스트 확인
4. 검증 결과 및 사용자 정보 반환

### 4.2 사용자 관리 흐름

#### 4.2.1 사용자 생성 흐름

1. 관리자 또는 가입 프로세스에서 CreateUser 요청 전송
2. UserService가 요청 처리:
   - 입력 데이터 검증
   - 중복 확인
   - 비밀번호 해싱
   - User 객체 생성
3. UserRepository를 통해 데이터베이스에 사용자 저장
4. UserCreated 이벤트 발행 (Kafka)
5. 생성된 사용자 정보 반환

#### 4.2.2 플랫폼 계정 연결 흐름

1. 사용자가 플랫폼 계정 연결 요청
2. PlatformAccountService가 요청 처리:
   - 플랫폼 OAuth 완료
   - 플랫폼 API에서 사용자 정보 조회
   - PlatformAccount 객체 생성 및 연결
3. PlatformAccountRepository를 통해 데이터베이스에 계정 정보 저장
4. PlatformConnected 이벤트 발행 (Kafka)
5. 연결된 계정 정보 반환

### 4.3 이벤트 발행 패턴

Authentication Service는 다음 Kafka 토픽으로 이벤트를 발행합니다:

- `auth.events.user`: 사용자 생성, 업데이트, 상태 변경 이벤트
- `auth.events.login`: 로그인 성공 및 실패 이벤트
- `auth.events.platform`: 플랫폼 계정 연결 및 해제 이벤트
- `auth.events.security`: 의심스러운 활동 이벤트

### 4.4 gRPC 인터페이스

Services는 다음과 같은 주요 gRPC 서비스를 제공합니다:

1. **AuthService**: 인증 관련 작업
   - Login: 사용자 로그인
   - Logout: 사용자 로그아웃
   - RefreshToken: 액세스 토큰 갱신
   - ValidateToken: 토큰 유효성 검증

2. **UserService**: 사용자 관리
   - CreateUser: 사용자 생성
   - GetUser: 사용자 정보 조회
   - UpdateUser: 사용자 정보 업데이트
   - DeleteUser: 사용자 삭제

3. **PlatformAccountService**: 플랫폼 계정 관리
   - ConnectPlatformAccount: 플랫폼 계정 연결
   - DisconnectPlatformAccount: 플랫폼 계정 연결 해제
   - GetPlatformAccounts: 연결된 계정 목록 조회

## 5. 데이터 모델

### 5.1 핵심 데이터 모델

#### 5.1.1 User

```sql
CREATE TABLE streamers (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    subscription_tier VARCHAR(20)
);
```

#### 5.1.2 PlatformAccount

```sql
CREATE TABLE platform_accounts (
    id VARCHAR(36) PRIMARY KEY,
    streamer_id VARCHAR(36) NOT NULL,
    platform VARCHAR(20) NOT NULL,
    platform_user_id VARCHAR(100) NOT NULL,
    platform_username VARCHAR(100) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (streamer_id) REFERENCES streamers(id) ON DELETE CASCADE
);
```

#### 5.1.3 TokenBlacklist

```sql
CREATE TABLE token_blacklist (
    token_id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    reason VARCHAR(50),
    blacklisted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES streamers(id) ON DELETE CASCADE
);
```

#### 5.1.4 AuditLog

```sql
CREATE TABLE audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(36),
    old_values JSONB,
    new_values JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES streamers(id) ON DELETE SET NULL
);
```

### 5.2 데이터 저장소

#### 5.2.1 PostgreSQL

- 모든 영구 데이터 저장
- 트랜잭션 일관성 보장
- 사용자 및 계정 데이터의 중앙 저장소

#### 5.2.2 Redis

- 활성 토큰 캐싱
- 토큰 블랙리스트
- 속도 제한 카운터
- 세션 데이터 저장

## 6. 보안 아키텍처

### 6.1 인증 메커니즘

#### 6.1.1 비밀번호 인증

- Argon2id 알고리즘을 사용한 비밀번호 해싱
- 솔트와 페퍼를 적용한 보안 강화
- 복잡성 정책 적용:
  - 최소 12자 이상
  - 대소문자, 숫자, 특수문자 포함
  - 공통 비밀번호 및 사전 기반 공격 방지

#### 6.1.2 OAuth 2.0 통합

- 지원 플랫폼: Twitch, YouTube, Facebook, Afreeca
- Authorization Code 흐름 구현
- PKCE(Proof Key for Code Exchange) 지원

#### 6.1.3 JWT 토큰 관리

- RS256 알고리즘을 사용한 비대칭 서명
- 액세스 토큰 만료 시간: 15분
- 리프레시 토큰 만료 시간: 7일
- 토큰 기반 블랙리스트 관리

### 6.2 권한 부여

#### 6.2.1 역할 기반 접근 제어(RBAC)

- 사전 정의된 역할: Admin, Streamer, Moderator
- 리소스 기반 권한 부여
- 동적 권한 정책 적용

#### 6.2.2 OAuth 2.0 스코프

- 세분화된 권한 범위 지원
- 플랫폼별 최소 권한 요청

### 6.3 통신 보안

#### 6.3.1 TLS

- 모든 서비스 간 통신에 TLS 1.3 적용
- 상호 TLS(mTLS) 인증 구현
- 인증서 자동 갱신 시스템

#### 6.3.2 API 보안

- 입력 검증 및 샌드박싱
- 요청 속도 제한
- CSRF 방어 메커니즘

### 6.4 감사 및 모니터링

#### 6.4.1 보안 감사 로깅

- 모든 인증 및 권한 부여 이벤트 기록
- 민감한 데이터 액세스 감사
- 구조화된 로그 형식

#### 6.4.2 실시간 보안 모니터링

- 다중 로그인 실패 탐지
- 비정상적인 액세스 패턴 탐지
- 지리적 위치 이상 탐지

## 7. 확장성 및 고가용성

### 7.1 수평적 확장

- 상태 비저장(Stateless) 서비스 설계
- Kubernetes 기반 자동 확장
- Redis를 통한 분산 세션 관리

### 7.2 데이터베이스 확장성

- 읽기/쓰기 분리
- 데이터베이스 샤딩 전략
- 효율적인 인덱싱 및 쿼리 최적화

### 7.3 고가용성 설계

- 다중 가용 영역 배포
- 데이터베이스 복제 및 자동 장애 조치
- 서킷 브레이커 패턴 구현

## 8. 지표 및 모니터링

### 8.1 핵심 성능 지표(KPI)

- 서비스 응답 시간: 평균, 95백분위수, 99백분위수
- 요청 처리량: 초당 요청 수
- 오류율: 총 요청 대비 오류 비율
- 인증 성공/실패율
- 토큰 검증 처리량

### 8.2 모니터링 도구

- Prometheus: 메트릭 수집
- Grafana: 대시보드 및 시각화
- OpenTelemetry: 분산 추적
- ELK Stack: 로그 집계 및 분석

### 8.3 알림 전략

- 심각도 기반 알림
- PagerDuty 또는 OpsGenie 통합
- 자동 스케일링 이벤트 알림

## 9. 운영 고려사항

### 9.1 배포 전략

- 블루/그린 배포
- 카나리 배포로 위험 최소화
- 자동화된 롤백 메커니즘

### 9.2 장애 복구

- 데이터베이스 백업 및 복구 전략
- 장애 시나리오 정기 테스트
- 재해 복구 계획

### 9.3 운영 정책

- 정기적인 보안 검토
- 비밀 및 인증서 순환 정책
- 로그 보존 정책

## 10. 기술 선택 및 의존성

### 10.1 언어 및 프레임워크

- **Go**: 성능 및 동시성 장점
- **gRPC**: 효율적인 서비스 간 통신
- **Protocol Buffers**: 강력한 타입 정의 및 스키마 진화

### 10.2 주요 라이브러리

- **go-grpc**: gRPC 서버 및 클라이언트 구현
- **sqlx/pgx**: PostgreSQL 데이터베이스 액세스
- **go-redis**: Redis 클라이언트
- **golang-jwt**: JWT 처리
- **zap**: 구조화된 로깅
- **uber-go/fx**: 의존성 주입
- **prometheus/client_golang**: 메트릭 수집
- **segmentio/kafka-go**: Kafka 클라이언트

### 10.3 인프라 의존성

- **PostgreSQL**: 영구 데이터 저장
- **Redis**: 캐싱 및 분산 잠금
- **Kafka**: 이벤트 메시징
- **Kubernetes**: 컨테이너 오케스트레이션

## 11. 결론

ImmersiVerse Authentication Service는 확장성, 보안성, 성능을 고려한 계층화된 아키텍처로 설계되었습니다. 도메인 주도 설계 원칙을 적용하여 비즈니스 로직의 명확한 분리를 보장하고, 헥사고날 아키텍처 패턴을 통해 외부 시스템과의 통합을 효율적으로 관리합니다.

gRPC 기반 통신 및 이벤트 기반 아키텍처의 조합은 서비스 간 효율적인 통신과 느슨한 결합을 제공합니다. 이러한 설계는 시스템의 확장성, 유지보수성, 그리고 진화 능력을 크게 향상시킵니다.

보안 측면에서는 최신 인증 및 권한 부여 메커니즘을 구현하여 플랫폼 전체의 보안을 강화합니다. 또한, 포괄적인 모니터링 및 로깅 시스템을 통해 운영 투명성과 장애 대응 능력을 보장합니다.

이 아키텍처는 ImmersiVerse 플랫폼의 성장과 함께 진화할 수 있도록 설계되었으며, 향후 기능 확장 및 서비스 통합에 대비한 유연성을 제공합니다.