# ImmersiVerse Authentication Service 개발 계획

## 1. 개요
- **프로젝트 목표**: ImmersiVerse 플랫폼의 사용자 인증, 계정 관리, 외부 플랫폼 연동을 위한 마이크로서비스 개발
- **기능 범위**: 
  - 사용자 인증 (로그인, 로그아웃, 토큰 발급/검증)
  - 사용자 관리 (CRUD)
  - 외부 플랫폼 연동 (OAuth 2.0 기반)
  - RBAC (역할 기반 접근 제어)
  - 감사 로깅 및 모니터링
- **아키텍처**: DDD, 헥사고날 아키텍처, 이벤트 기반 마이크로서비스
- **기술 스택**: Go 1.21+, gRPC, PostgreSQL, Redis, Kafka, Kubernetes
- **예상 전체 소요 시간**: 약 8~10주 (팀 리소스 및 병렬 작업 여부에 따라 조정 가능)
- **개발 팀 구성**: 
  - 백엔드 개발자 (2~3명): 도메인 로직, gRPC, 저장소 구현
  - DevOps 엔지니어 (1명): CI/CD, Kubernetes 배포, 모니터링 설정
  - 보안 전문가 (1명, 옵션): 보안 테스트 및 규정 준수 검토

---

## 2. 개발 단계별 계획

### 단계 1: 프로젝트 초기 설정 (1주)
- **목표**: 기본 프로젝트 구조와 개발 환경을 설정하여 팀이 작업을 시작할 수 있도록 준비
- **작업 항목**:
  1. **프로젝트 디렉토리 구조 생성** (1일, 백엔드 개발자)
     - `cmd/`, `internal/`, `pkg/`, `proto/`, `db/`, `test/` 디렉토리 생성
     - `go.mod` 초기화 및 기본 의존성 추가 (`grpc-go`, `pgx`, `go-redis`, `kafka-go`, `jwt`, `zap`, `prometheus`, 등)
  2. **Makefile 작성** (1일, 백엔드 개발자)
     - `make proto`, `make migrate`, `make run`, `make test` 등 작업 자동화 스크립트 구현
  3. **Docker Compose 설정** (2일, DevOps 엔지니어)
     - `docker-compose.yml` 작성: PostgreSQL, Redis, Kafka 실행 환경 구성
     - `.env.example` 작성: 환경 변수 템플릿 제공
  4. **CI/CD 초기 설정** (2일, DevOps 엔지니어)
     - GitHub Actions 워크플로우 작성: 빌드, 테스트, 린트 실행
     - ArgoCD 초기 설정 준비 (매니페스트 디렉토리 생성: `deployments/kubernetes/`)
  5. **문서화 및 팀 동기화** (1일, 모든 팀원)
     - README.md 작성: 프로젝트 개요, 셋업 가이드, 기여 가이드
     - 팀 킥오프 미팅: 개발 계획 공유 및 역할 분담
- **산출물**: 프로젝트 스켈레톤, 빌드/실행 환경, CI/CD 파이프라인 초안
- **의존성**: 없음 (프로젝트 시작 단계)

---

### 단계 2: 데이터베이스 스키마 설계 및 마이그레이션 (1주)
- **목표**: PostgreSQL 스키마를 정의하고 초기 마이그레이션 파일을 작성하여 데이터 저장 기반 구축
- **작업 항목**:
  1. **스키마 설계 검토 및 최종화** (1일, 백엔드 개발자)
     - `users`, `platform_accounts`, `token_blacklist`, `audit_logs` 테이블 설계
     - 인덱스 및 제약 조건 정의 (예: `UNIQUE(username)`, `FK(user_id)`)
  2. **마이그레이션 파일 작성** (2일, 백엔드 개발자)
     - `db/migrations/0001_create_users.up.sql` 및 `.down.sql`
     - `db/migrations/0002_create_platform_accounts.up.sql` 및 `.down.sql` 등
     - `golang-migrate`로 마이그레이션 테스트
  3. **시드 데이터 생성** (1일, 백엔드 개발자)
     - `db/seeds/example_data.sql`: 테스트용 초기 사용자 데이터 삽입
  4. **DB 연결 및 초기 테스트** (1일, 백엔드 개발자)
     - `internal/repository/postgres.go`에 PostgreSQL 연결 코드 작성
     - 기본 쿼리 실행 및 로컬 환경에서 연결 검증
- **산출물**: DB 스키마 정의, 마이그레이션 파일, 시드 데이터, 초기 DB 연결 코드
- **의존성**: 단계 1 (환경 변수 및 Docker Compose 설정)

---

### 단계 3: 도메인 모델 설계 및 구현 (1.5주)
- **목표**: 핵심 도메인 엔티티, 값 객체, 도메인 서비스를 설계하고 구현
- **작업 항목**:
  1. **엔티티 정의** (2일, 백엔드 개발자)
     - `internal/domain/user.go`: `User` 구조체 정의 및 생성자 메서드
     - `internal/domain/platform_account.go`: `PlatformAccount` 정의
     - `internal/domain/token.go`: `Token` 정의
     - `internal/domain/audit_log.go`: `AuditLog` 정의
  2. **값 객체 구현** (2일, 백엔드 개발자)
     - `internal/domain/password.go`: `HashPassword`, `ValidatePassword`
     - `internal/domain/email.go`: 유효성 검증 및 정규화
     - `internal/domain/token_pair.go`: Access/Refresh 토큰 쌍 관리
  3. **도메인 서비스 구현** (3일, 백엔드 개발자)
     - `internal/domain/user_service.go`: `CreateUser`, `ChangeUserStatus`
     - `internal/domain/platform_integration.go`: `ConnectPlatformAccount`, `RefreshPlatformToken`
     - `internal/domain/authentication.go`: `Login`, `Logout`, `ValidateToken`
  4. **도메인 이벤트 정의** (1일, 백엔드 개발자)
     - `internal/domain/events.go`: `UserCreated`, `LoginSucceeded` 등 구조체 정의
- **산출물**: 도메인 모델 코드, 초기 단위 테스트 케이스
- **의존성**: 없음 (도메인 계층은 순수 비즈니스 로직)

---

### 단계 4: 저장소 계층 구현 (1.5주)
- **목표**: 데이터 접근 로직을 저장소 인터페이스와 PostgreSQL 구현으로 캡슐화
- **작업 항목**:
  1. **저장소 인터페이스 정의** (1일, 백엔드 개발자)
     - `internal/repository/user_repository.go`: `UserRepository` 인터페이스
     - `internal/repository/platform_account_repository.go`: `PlatformAccountRepository`
     - `internal/repository/token_repository.go`: `TokenRepository`
     - `internal/repository/audit_log_repository.go`: `AuditLogRepository`
  2. **PostgreSQL 저장소 구현** (3일, 백엔드 개발자)
     - `internal/repository/postgres_user.go`: `InsertUser`, `GetUserByID` 등 구현
     - `internal/repository/postgres_platform_account.go`: `InsertPlatformAccount`, `GetPlatformAccountsByUserID`
     - `internal/repository/postgres_token.go`: `InsertTokenBlacklist`, `IsTokenBlacklisted`
     - `internal/repository/postgres_audit_log.go`: `InsertAuditLog`, `GetAuditLogsByUserID`
  3. **단위 테스트 작성** (2일, 백엔드 개발자)
     - `sqlmock` 사용, 각 메서드에 대한 테스트 케이스 작성
- **산출물**: 저장소 인터페이스 및 PostgreSQL 구현, 단위 테스트
- **의존성**: 단계 2 (DB 스키마), 단계 3 (도메인 모델)

---

### 단계 5: 서비스 로직 및 gRPC 구현 (2주)
- **목표**: 비즈니스 로직을 서비스 계층에 통합하고 gRPC API로 노출
- **작업 항목**:
  1. **gRPC 인터페이스 정의** (2일, 백엔드 개발자)
     - `proto/auth/v1/auth_service.proto`: `Login`, `Logout`, `RefreshToken`, `ValidateToken`
     - `proto/auth/v1/user_service.proto`: `CreateUser`, `GetUser`, `UpdateUser`, `DeleteUser`
     - `proto/auth/v1/platform_account_service.proto`: `ConnectPlatformAccount`, `DisconnectPlatformAccount`
     - `make proto` 실행으로 코드 생성
  2. **서비스 로직 구현** (4일, 백엔드 개발자)
     - `internal/service/auth_service.go`: 로그인, 토큰 관리 로직
     - `internal/service/user_service.go`: 사용자 관리 로직
     - `internal/service/platform_account_service.go`: 플랫폼 연동 로직
  3. **gRPC 서비스 구현** (3일, 백엔드 개발자)
     - `internal/grpc/auth_service.go`: `AuthServiceServer` 구현
     - `internal/grpc/user_service.go`: `UserServiceServer`
     - `internal/grpc/platform_account_service.go`: `PlatformAccountServiceServer`
  4. **인터셉터 구현** (2일, 백엔드 개발자)
     - 인증, 로깅, 검증, 메트릭 인터셉터 추가
- **산출물**: gRPC `.proto` 파일, 서비스 로직, gRPC 서버 코드
- **의존성**: 단계 3 (도메인 모델), 단계 4 (저장소)

---

### 단계 6: 인증 유틸리티 및 토큰 관리 (1주)
- **목표**: JWT 생성/검증 유틸리티 및 토큰 관리 기능을 구현
- **작업 항목**:
  1. **인증 유틸리티 구현** (3일, 백엔드 개발자)
     - `pkg/auth/jwt.go`: `CreateToken`, `ValidateToken`, `ParseToken`, `LoadKeys`
     - RS256 키 로드 및 토큰 서명/검증 로직 작성
  2. **토큰 관리 로직 통합** (2일, 백엔드 개발자)
     - `internal/service/token_management.go`: 블랙리스트 관리, 토큰 갱신 로직
     - Redis 기반 캐싱 및 블랙리스트 추가
- **산출물**: `pkg/auth` 패키지, 토큰 관리 로직
- **의존성**: 단계 5 (서비스 로직)

---

### 단계 7: 이벤트 발행 및 Kafka 통합 (1주)
- **목표**: 도메인 이벤트를 Kafka로 발행하여 다른 서비스와 연동
- **작업 항목**:
  1. **이벤트 퍼블리셔 인터페이스 정의** (1일, 백엔드 개발자)
     - `internal/repository/event_publisher.go`: `Publish` 메서드 정의
  2. **Kafka 퍼블리셔 구현** (2일, 백엔드 개발자)
     - `internal/repository/kafka_publisher.go`: `segmentio/kafka-go` 사용
     - `auth.events.*` 토픽에 이벤트 발행 로직 작성
  3. **서비스 로직에 이벤트 통합** (2일, 백엔드 개발자)
     - `UserService`, `AuthService` 등에서 이벤트 발행 추가
- **산출물**: Kafka 퍼블리셔, 서비스 로직 내 이벤트 발행 코드
- **의존성**: 단계 5 (서비스 로직)

---

### 단계 8: 테스트 및 검증 (2주)
- **목표**: 단위, 통합, 성능, 보안 테스트를 통해 코드 품질 및 안정성 확보
- **작업 항목**:
  1. **단위 테스트 작성** (4일, 백엔드 개발자)
     - `internal/domain/*_test.go`, `internal/service/*_test.go`, `pkg/auth/*_test.go`
     - Table-driven 테스트, `gomock` 활용
  2. **통합 테스트 작성** (4일, 백엔드 개발자)
     - `test/integration/`에 로그인, 토큰 검증, 플랫폼 연동 등 테스트 구현
     - Docker Compose로 의존 서비스 실행
  3. **성능 테스트** (2일, 백엔드 개발자)
     - k6 스크립트 작성: 로그인 TPS 60 이상, 응답 시간 100ms 목표
  4. **보안 테스트** (2일, 보안 전문가)
     - SAST (GoSec), DAST (OWASP ZAP), 의존성 스캔 (Snyk)
- **산출물**: 테스트 코드, 테스트 보고서, 성능/보안 검증 결과
- **의존성**: 단계 1~7 (전체 코드)

---

### 단계 9: 배포 및 모니터링 설정 (1주)
- **목표**: Kubernetes 배포 및 운영 환경 구축
- **작업 항목**:
  1. **Kubernetes 매니페스트 작성** (2일, DevOps 엔지니어)
     - `deployments/kubernetes/deployment.yaml`: 3+ replicas, HPA
     - `deployments/kubernetes/service.yaml`: ClusterIP
     - `deployments/kubernetes/ingress.yaml`: TLS 설정
  2. **CI/CD 파이프라인 완성** (2일, DevOps 엔지니어)
     - GitHub Actions: 빌드 → 테스트 → 이미지 푸시
     - ArgoCD: Kubernetes 배포 자동화
  3. **모니터링 설정** (2일, DevOps 엔지니어)
     - Prometheus 메트릭 정의: `auth_login_requests_total`, `auth_token_validation_latency_seconds`
     - Grafana 대시보드 구성, Alertmanager 알림 설정
- **산출물**: Kubernetes 매니페스트, CI/CD 워크플로우, 모니터링 설정
- **의존성**: 단계 8 (테스트 완료)

---

### 단계 10: 문서화 및 최종 검토 (0.5주)
- **목표**: 프로젝트 문서를 완성하고 최종 검토 완료
- **작업 항목**:
  1. **개발 문서 업데이트** (2일, 모든 팀원)
     - README.md, API 문서, 문제 해결 가이드 등 최신화
  2. **최종 테스트 및 리뷰** (1일, 모든 팀원)
     - 전체 기능 점검, 코드 리뷰, 배포 테스트
- **산출물**: 최종 문서, 배포된 서비스
- **의존성**: 단계 1~9 (모든 작업 완료)

---

## 3. 전체 타임라인
| 단계                  | 기간       | 시작 날짜 (예시) | 종료 날짜 (예시) |
|-----------------------|------------|-------------------|-------------------|
| 1. 프로젝트 초기 설정   | 1주       | 3/1             | 3/7             |
| 2. DB 스키마 설계       | 1주       | 3/8             | 3/14            |
| 3. 도메인 모델 설계     | 1.5주     | 3/15            | 3/25            |
| 4. 저장소 계층 구현     | 1.5주     | 3/26            | 4/5             |
| 5. 서비스 로직/gRPC     | 2주       | 4/6             | 4/19            |
| 6. 인증 유틸리티        | 1주       | 4/20            | 4/26            |
| 7. 이벤트/Kafka 통합   | 1주       | 4/27            | 5/3             |
| 8. 테스트 및 검증       | 2주       | 5/4             | 5/17            |
| 9. 배포 및 모니터링     | 1주       | 5/18            | 5/24            |
| 10. 문서화 및 검토      | 0.5주     | 5/25            | 5/28            |
| **총 기간**            | **8~10주** | **3/1**         | **5/28**        |

- **참고**: 병렬 작업 가능 시 (예: DevOps와 백엔드 작업 동시 진행) 약 8주까지 단축 가능

---

## 4. 리소스 및 의존성 관리
- **필요 리소스**:
  - 개발 환경: Go 1.21+, Docker, Kubernetes 클러스터
  - 외부 서비스: PostgreSQL, Redis, Kafka
  - 도구: GitHub, ArgoCD, Prometheus, Grafana
- **의존성 관리**:
  - `go.mod`로 Go 모듈 관리
  - Docker 이미지 버전 관리 (예: `immersiverse/auth-service:v1.0.0`)
  - 외부 의존성 업데이트 주기 (최소 분기별 점검)

---

## 5. 위험 관리 및 대응 계획
- **위험 1: 개발 지연**
  - **대응**: 작업 분할 및 병렬 진행, 추가 리소스 투입
- **위험 2: 버그 및 안정성 문제**
  - **대응**: 철저한 단위/통합 테스트, 카나리 배포로 점진적 롤아웃
- **위험 3: 보안 취약점**
  - **대응**: 정기 보안 테스트, 최신 의존성 유지

---

## 6. 결론
이 개발 계획은 **ImmersiVerse Authentication Service**를 처음부터 완성까지 체계적으로 구현하기 위한 상세 로드맵입니다. 각 단계는 명확한 목표와 작업 항목을 포함하며, 팀 협업과 리소스 활용을 최적화하여 약 8~10주 내에 안정적인 서비스를 배포할 수 있도록 설계되었습니다. 팀원들은 이 계획을 기준으로 작업을 분배하고, 진행 상황을 주기적으로 점검하여 목표를 달성해 주세요.