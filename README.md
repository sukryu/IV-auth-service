# ImmersiVerse: Authentication Service

## 개요

ImmersiVerse Authentication Service는 플랫폼 내 모든 사용자 인증 및 계정 관리를 담당하는 고성능 마이크로서비스입니다. 이 서비스는 Go 언어로 구현되었으며, 내부 마이크로서비스 간 통신에는 gRPC를, 비동기 이벤트 처리에는 메시지 큐(Kafka)를 사용합니다.

## 아키텍처

### 핵심 컴포넌트

- **gRPC 서버**: 인증 및 사용자 관리 API 제공
- **데이터베이스 계층**: PostgreSQL을 활용한 사용자 및 계정 데이터 관리
- **토큰 관리**: JWT 기반 토큰 생성, 검증 및 관리
- **이벤트 발행자**: Kafka를 통한 인증 이벤트 발행
- **플랫폼 통합**: 다양한 스트리밍 플랫폼 인증 연동 (Twitch, YouTube 등)

### 시스템 상호작용

```
[API 게이트웨이] ←→ [Authentication Service] ←→ [PostgreSQL]
                                ↓
                      [기타 마이크로서비스]
                                ↓
                         [Kafka Event Bus]
```

## 주요 기능

- **사용자 인증**: 사용자명/비밀번호 로그인, OAuth 기반 소셜 로그인
- **토큰 관리**: 액세스 및 리프레시 토큰 발급, 검증, 갱신
- **사용자 관리**: 사용자 등록, 수정, 조회, 비활성화
- **권한 관리**: 역할 기반 접근 제어(RBAC)
- **플랫폼 연동**: 다양한 스트리밍 플랫폼 계정 연결 및 관리
- **이벤트 발행**: 인증 및 사용자 관련 이벤트 발행
- **보안 기능**: 비밀번호 해싱, 속도 제한, 의심스러운 활동 탐지

## 기술 스택

- **언어**: Go 1.21+
- **프레임워크**: gRPC, Protocol Buffers
- **데이터베이스**: PostgreSQL 15
- **메시지 큐**: Apache Kafka
- **캐싱**: Redis
- **CI/CD**: GitHub Actions, ArgoCD
- **컨테이너화**: Docker, Kubernetes
- **모니터링**: Prometheus, Grafana

## 개발 환경 설정

### 필수 요구사항

- Go 1.21 이상
- Docker 및 Docker Compose
- Protocol Buffer 컴파일러(protoc) 및 Go 플러그인
- 개발용 PostgreSQL 및 Kafka 인스턴스

### 빠른 시작

```bash
# 저장소 클론
git clone https://github.com/immersiverse/auth-service.git
cd auth-service

# 의존성 설치
go mod download

# Protocol Buffers 컴파일
make proto

# 개발 환경 설정 (Docker Compose 사용)
docker-compose up -d postgres kafka

# 마이그레이션 실행
make migrate

# 서비스 실행
make run
```

## 프로젝트 구조

```
auth-service/
├── cmd/                  # 실행 진입점
│   └── server/
├── internal/             # 비공개 애플리케이션 코드
│   ├── config/           # 구성 관리
│   ├── domain/           # 도메인 모델 및 인터페이스
│   ├── repository/       # 데이터 접근 구현
│   ├── service/          # 비즈니스 로직
│   └── grpc/             # gRPC 서비스 구현
├── pkg/                  # 공개 라이브러리 코드
│   ├── auth/             # 인증 관련 유틸리티
│   ├── events/           # 이벤트 발행 유틸리티
│   ├── metrics/          # 메트릭 수집 유틸리티
│   └── validator/        # 입력 검증 유틸리티
├── proto/                # 프로토콜 버퍼 정의
│   └── auth/
│       └── v1/
├── db/                   # 데이터베이스 관련 파일
│   ├── migrations/       # 스키마 마이그레이션
│   └── seeds/            # 초기 데이터
├── test/                 # 테스트 코드
│   ├── integration/      # 통합 테스트
│   └── performance/      # 성능 테스트
├── deployments/          # 배포 관련 파일
│   ├── kubernetes/       # Kubernetes 매니페스트
│   └── docker/           # Docker 관련 파일
├── go.mod                # Go 모듈 정의
├── go.sum                # 의존성 체크섬
├── Makefile              # 빌드 및 개발 작업 자동화
├── .gitignore            # Git 무시 패턴
└── README.md             # 프로젝트 문서
```

## 프로젝트 문서

Authentication Service의 개발, 유지보수 및 운영을 위한 포괄적인 문서를 제공합니다.

### 설계 및 아키텍처 문서

- [서비스 아키텍처 개요](docs/architecture/overview.md)
- [주요 설계 결정 및 근거](docs/architecture/design-decisions.md)
- [도메인 모델 및 관계 설명](docs/architecture/domain-model.md)
- [데이터 흐름 다이어그램 및 설명](docs/architecture/data-flow.md)
- [서비스 간 통신 패턴](docs/architecture/communication-patterns.md)

### API 및 프로토콜 문서

- [Protocol Buffers 정의 가이드](proto/auth/v1/README.md)
- [gRPC 서비스 상세 명세](docs/api/grpc-services.md)
- [에러 코드 및 처리 방법](docs/api/error-codes.md)
- [API 호출 예제](docs/api/examples.md)
- [API 버전 관리 전략](docs/api/versioning.md)

### 개발 가이드

- [개발 환경 설정 가이드](docs/development/setup.md)
- [코드 스타일 가이드](docs/development/style-guide.md)
- [테스트 작성 가이드](docs/development/testing.md)
- [디버깅 가이드](docs/development/debugging.md)
- [기여 가이드라인](docs/development/contribution.md)

### 데이터베이스 문서

- [데이터베이스 스키마 설명](docs/database/schema.md)
- [마이그레이션 전략](docs/database/migrations.md)
- [주요 쿼리 최적화 가이드](docs/database/queries.md)
- [백업 및 복구 전략](docs/database/backups.md)

### 보안 문서

- [인증 흐름 상세 설명](docs/security/authentication-flow.md)
- [토큰 관리 전략](docs/security/token-management.md)
- [비밀번호 정책 및 해싱 전략](docs/security/password-policies.md)
- [공격 방어 전략](docs/security/attack-prevention.md)
- [감사 로깅 정책](docs/security/audit-logging.md)

### 통합 가이드

- [API 게이트웨이 통합 가이드](docs/integration/api-gateway.md)
- [다른 마이크로서비스와의 통합 가이드](docs/integration/other-services.md)
- [Kafka 이벤트 스트림 규격 및 통합 가이드](docs/integration/event-streams.md)
- [외부 플랫폼 통합 가이드](docs/integration/platform-integration.md)

### 운영 문서

- [배포 프로세스 및 환경 설정](docs/operations/deployment.md)
- [Kubernetes 구성 상세 가이드](docs/operations/kubernetes-config.md)
- [모니터링 및 알림 설정](docs/operations/monitoring.md)
- [확장성 및 성능 최적화 가이드](docs/operations/scaling.md)
- [백업 및 복구 절차](docs/operations/backup-recovery.md)
- [재해 복구 계획](docs/operations/disaster-recovery.md)

### 테스트 문서

- [단위 테스트 전략](docs/testing/unit-testing.md)
- [통합 테스트 전략](docs/testing/integration-testing.md)
- [성능 테스트 가이드](docs/testing/performance-testing.md)
- [보안 테스트 가이드](docs/testing/security-testing.md)

### 주요 구현 문서

- [도메인 모델 설계 및 구현 설명](internal/domain/README.md)
- [서비스 로직 구현 설명](internal/service/README.md)
- [저장소 구현 설명](internal/repository/README.md)
- [gRPC 서비스 구현 설명](internal/grpc/README.md)
- [인증 유틸리티 사용 가이드](pkg/auth/README.md)

### 참조 문서

- [용어 사전](docs/references/glossary.md)
- [외부 의존성 목록 및 버전](docs/references/external-dependencies.md)
- [규정 준수 문서](docs/references/compliance.md)
- [문제 해결 가이드](docs/references/troubleshooting.md)

## 인증 워크플로우

### 1. 사용자 등록 및 로그인

```mermaid
sequenceDiagram
    Client->>API Gateway: 로그인 요청
    API Gateway->>Auth Service: gRPC: Login
    Auth Service->>Database: 사용자 검증
    Database-->>Auth Service: 사용자 정보
    Auth Service->>Auth Service: 토큰 생성
    Auth Service-->>API Gateway: LoginResponse
    Auth Service->>Kafka: 로그인 이벤트 발행
    API Gateway-->>Client: 토큰 및 사용자 정보 반환
```

### 2. 토큰 갱신

```mermaid
sequenceDiagram
    Client->>API Gateway: 토큰 갱신 요청
    API Gateway->>Auth Service: gRPC: RefreshToken
    Auth Service->>Auth Service: 리프레시 토큰 검증
    Auth Service->>Auth Service: 새 액세스 토큰 생성
    Auth Service-->>API Gateway: RefreshTokenResponse
    API Gateway-->>Client: 새 액세스 토큰 반환
```

## gRPC API 명세

주요 gRPC 서비스 정의:

```protobuf
service AuthService {
  // 인증 관련 API
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc Logout(LogoutRequest) returns (LogoutResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  
  // 사용자 관리 API
  rpc CreateUser(CreateUserRequest) returns (UserResponse);
  rpc GetUser(GetUserRequest) returns (UserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
  
  // 플랫폼 계정 연동 API
  rpc ConnectPlatformAccount(ConnectPlatformAccountRequest) returns (PlatformAccountResponse);
  rpc DisconnectPlatformAccount(DisconnectPlatformAccountRequest) returns (DisconnectPlatformAccountResponse);
  rpc GetPlatformAccounts(GetPlatformAccountsRequest) returns (GetPlatformAccountsResponse);
}
```

## 이벤트 토픽

Authentication Service에서 발행하는 주요 Kafka 이벤트:

- `auth.events.user_created`: 사용자 생성 이벤트
- `auth.events.user_updated`: 사용자 정보 업데이트 이벤트
- `auth.events.login`: 사용자 로그인 이벤트
- `auth.events.logout`: 사용자 로그아웃 이벤트
- `auth.events.platform_connected`: 플랫폼 계정 연결 이벤트
- `auth.events.platform_disconnected`: 플랫폼 계정 연결 해제 이벤트
- `auth.events.security`: 보안 관련 이벤트 (의심스러운 로그인 시도 등)

## 보안 고려사항

- 모든 비밀번호는 Argon2id를 사용하여 해싱
- 액세스 토큰은 RS256 알고리즘으로 서명되며 짧은 만료 시간(15분) 설정
- 모든 서비스 간 통신은 mTLS를 통해 암호화
- 토큰 폐기 및 블랙리스트 관리를 위한 Redis 활용
- 인증 요청에 대한 속도 제한 구현
- IP 기반 의심스러운 활동 탐지

## 모니터링 및 로깅

- Prometheus 메트릭 수집:
  - 인증 성공/실패 횟수
  - 토큰 발급/검증 지연 시간
  - 요청 처리량
  - 오류율
- Grafana 대시보드 제공
- 구조화된 JSON 로그 출력
- OpenTelemetry를 활용한 분산 추적

## 테스트 전략

- **단위 테스트**: 각 패키지 및 모듈에 대한 단위 테스트
- **통합 테스트**: 실제 데이터베이스 및 Kafka와 통합 테스트
- **성능 테스트**: 부하 테스트 및 스트레스 테스트
- **API 테스트**: gRPC API 호출 테스트
- **보안 테스트**: 토큰 검증 및 해킹 시도 테스트

## 배포 가이드

### Kubernetes 배포

```bash
# 네임스페이스 생성
kubectl create namespace immersiverse

# 시크릿 생성
kubectl create secret generic auth-secrets \
  --from-literal=jwt-private-key="$(cat private.pem)" \
  --from-literal=jwt-public-key="$(cat public.pem)" \
  --from-literal=db-password="your-password" \
  --namespace immersiverse

# 구성 맵 생성
kubectl apply -f deployments/kubernetes/config.yaml

# 서비스 배포
kubectl apply -f deployments/kubernetes/deployment.yaml
```

## 기여 가이드라인

1. 이슈 트래커를 확인하여 기존 이슈를 찾거나 새 이슈 생성
2. 개발하기 전 최신 `main` 브랜치에서 기능 브랜치 생성 (`feature/기능명`)
3. 코드 작성 및 테스트 추가
4. 모든 테스트 통과 확인
5. 변경 사항 커밋 및 푸시
6. Pull Request 생성
7. 코드 리뷰 및 피드백 반영
8. 변경 사항 병합

## 라이센스

이 프로젝트는 Proprietary 라이센스로 배포되며, ImmersiVerse의 승인 없이 사용, 수정 또는 배포할 수 없습니다.

## 연락처

기술적 문의사항은 개발팀 이메일 [dev@immersiverse.com](mailto:dev@immersiverse.com)으로 연락 주시기 바랍니다.