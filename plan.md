# ImmersiVerse Authentication Service 개발 계획 (Production-Ready)

## 1. 개요
- **프로젝트 목표**: ImmersiVerse 플랫폼의 사용자 인증, 계정 관리, 외부 플랫폼 연동을 위한 안정적이고 확장 가능한 마이크로서비스 개발
- **기능 범위**:
  - 사용자 인증: 로그인, 로그아웃, 토큰 발급/검증/갱신
  - 사용자 관리: 생성(Create), 조회(Read), 수정(Update), 삭제(Delete) (CRUD)
  - 외부 플랫폼 연동: OAuth 2.0 기반 (Twitch, YouTube 등)
  - RBAC: 역할 기반 접근 제어 및 권한 관리
  - 감사 로깅: 사용자 활동 추적
  - 모니터링: 성능, 오류, 보안 이벤트 실시간 관찰
- **아키텍처**: DDD(Domain-Driven Design), 헥사고날 아키텍처, 이벤트 기반 마이크로서비스
- **기술 스택**: Go 1.21+, gRPC, PostgreSQL 15, Redis, Kafka, Kubernetes, Prometheus, Grafana
- **예상 전체 소요 시간**: 약 10~12주 (품질 보증 및 최적화 시간 포함)
- **개발 팀 구성**:
  - 백엔드 개발자 (3명): 도메인 로직, gRPC, 저장소, 유틸리티 구현
  - DevOps 엔지니어 (1~2명): CI/CD, Kubernetes, 모니터링, 배포 자동화
  - 보안 전문가 (1명): 보안 설계, 감사, 취약점 점검
  - QA 엔지니어 (1명, 옵션): 테스트 계획 수립 및 실행

---

## 2. 개발 단계별 계획

### 단계 1: 프로젝트 초기 설정 (1주)
- **목표**: Production-ready 수준의 개발 환경과 품질 관리 기반을 구축
- **작업 항목**:
  1. **프로젝트 디렉토리 구조 생성** (1일, 백엔드 개발자)
     - 디렉토리: `cmd/`, `api/proto/`, `internal/`, `pkg/`, `db/`, `test/`, `deployments/`
     - `go.mod` 초기화: `go mod init github.com/immersiverse/auth-service`
     - 의존성 추가: `grpc-go`, `pgx/v5`, `go-redis/v9`, `segmentio/kafka-go`, `jwt`, `zap`, `prometheus`, `viper`, `golang-migrate`
  2. **Makefile 작성** (1일, 백엔드 개발자)
     - 명령어: `make proto`, `make build`, `make run`, `make test`, `make lint`, `make migrate`
     - 예: `make proto`는 `api/proto/*.proto` → `internal/generated/`로 코드 생성
  3. **개발 환경 설정** (2일, DevOps 엔지니어)
     - `deployments/docker/Dockerfile`: 멀티스테이지 빌드 (빌드 → 실행 이미지 최소화)
     - `deployments/docker/docker-compose.yml`: PostgreSQL, Redis, Kafka, Zookeeper 구성
     - `.env.example`: `DB_URL`, `REDIS_ADDR`, `KAFKA_BROKERS`, `JWT_PRIVATE_KEY` 등 정의
  4. **CI/CD 파이프라인 설정** (2일, DevOps 엔지니어)
     - GitHub Actions:
       - `build.yml`: 빌드, 유닛 테스트, `golangci-lint` 실행
       - `deploy.yml`: Docker 이미지 빌드 및 푸시 (태그: `vX.Y.Z`)
     - ArgoCD: `deployments/kubernetes/` 디렉토리 초기화 (빈 매니페스트 파일 생성)
  5. **코드 품질 도구 설정** (1일, 백엔드 개발자)
     - `.golangci.yml`: 린트 규칙 정의 (예: `unused`, `errcheck`, `staticcheck` 활성화)
     - `go fmt`, `go vet` 기본 적용
  6. **문서화 및 팀 동기화** (1일, 모든 팀원)
     - `README.md`: 프로젝트 개요, 셋업 가이드, 기여 가이드 초안 작성
     - 팀 킥오프 미팅: 계획 검토, 역할 분담, Jira/타스크 보드 설정
- **산출물**: 프로젝트 스켈레톤, 빌드/테스트 환경, CI/CD 초안, 초기 문서
- **품질 보증**: 코드 품질 도구 설정 완료, CI에서 빌드 성공 확인

---

### 단계 2: 데이터베이스 스키마 설계 및 초기화 (1.5주)
- **목표**: 안정적이고 확장 가능한 데이터베이스 스키마 설계 및 적용
- **작업 항목**:
  1. **스키마 설계 및 리뷰** (2일, 백엔드 개발자)
     - 테이블: `users`, `roles`, `platform_accounts`, `token_blacklist`, `audit_logs`
     - 관계: `users → roles (M:N)`, `users → platform_accounts (1:N)`
     - 인덱스: `users(username)`, `token_blacklist(token_jti)`, `audit_logs(timestamp)`
     - 팀 리뷰: 정규화, 성능, 확장성 검토
  2. **마이그레이션 파일 작성** (3일, 백엔드 개발자)
     - `db/migrations/0001_create_users.up.sql`: 테이블 생성, 기본 인덱스 추가
     - `db/migrations/0002_create_roles_and_permissions.up.sql`: RBAC 지원
     - `db/migrations/0003_create_platform_accounts.up.sql`: 외부 플랫폼 연동
     - `.down.sql` 파일: 롤백 로직 작성
     - `make migrate`로 로컬 테스트
  3. **초기 데이터 생성** (1일, 백엔드 개발자)
     - `db/seeds/users.sql`: 테스트용 사용자 및 기본 역할 삽입
     - `db/seeds/platform_accounts.sql`: 샘플 플랫폼 계정 데이터
  4. **DB 연결 코드 작성** (2일, 백엔드 개발자)
     - `internal/adapters/db/postgres/pg.go`: `pgxpool` 연결 풀 설정
     - 기본 쿼리 (`SELECT 1`)로 연결 검증
- **산출물**: 스키마 설계 문서, 마이그레이션/시드 파일, DB 연결 코드
- **품질 보증**: 스키마 리뷰 완료, 마이그레이션 롤백 테스트 성공

---

### 단계 3: 도메인 모델 설계 및 구현 (2주)
- **목표**: DDD 기반의 견고한 도메인 계층 구축
- **작업 항목**:
  1. **엔티티 설계 및 구현** (3일, 백엔드 개발자)
     - `internal/core/domain/user.go`: `User{ID, Username, Email, RoleIDs, Status}`
     - `internal/core/domain/platform_account.go`: `PlatformAccount{UserID, Platform, AccessToken}`
     - `internal/core/domain/token.go`: `Token{AccessToken, RefreshToken, JTI, Expiry}`
     - `internal/core/domain/audit_log.go`: `AuditLog{UserID, Action, Timestamp}`
  2. **값 객체 설계 및 구현** (3일, 백엔드 개발자)
     - `internal/core/domain/email.go`: 유효성 검사, 정규화 로직
     - `internal/core/domain/password.go`: `Hash()`, `Verify()` 메서드 (Argon2id)
     - `internal/core/domain/role.go`: 역할 및 권한 정의
  3. **도메인 서비스 설계** (4일, 백엔드 개발자)
     - `internal/core/domain/auth.go`: `Authenticate()`, `GenerateTokenPair()`
     - `internal/core/domain/user_management.go`: `CreateUser()`, `UpdateUserRole()`
     - `internal/core/domain/platform.go`: `LinkAccount()`, `RevokeAccount()`
  4. **도메인 이벤트 정의** (2일, 백엔드 개발자)
     - `internal/core/domain/events.go`: `UserCreated`, `LoginFailed`, `PlatformLinked`
     - 이벤트 구조체: `type UserCreated struct { UserID string; Timestamp time.Time }`
  5. **단위 테스트 작성** (3일, 백엔드 개발자)
     - `internal/core/domain/*_test.go`: 각 메서드별 테스트 케이스 (예: `TestUser_Create_InvalidEmail`)
     - 커버리지 목표: 80% 이상
- **산출물**: 도메인 모델 코드, 단위 테스트, 이벤트 정의
- **품질 보증**: 코드 리뷰 (팀원 간 상호 검토), 테스트 커버리지 보고서

---

### 단계 4: 저장소 계층 구현 (2주)
- **목표**: 데이터 접근 로직을 인터페이스와 구현으로 분리하여 확장성과 테스트 용이성 확보
- **작업 항목**:
  1. **저장소 인터페이스 정의** (2일, 백엔드 개발자)
     - `internal/core/ports/repository.go`:
       - `UserRepository`: `SaveUser()`, `FindUserByUsername()`
       - `PlatformAccountRepository`: `SavePlatformAccount()`, `FindByUserID()`
       - `TokenRepository`: `BlacklistToken()`, `IsBlacklisted()`
       - `AuditLogRepository`: `LogAction()`, `GetLogs()`
  2. **PostgreSQL 구현** (5일, 백엔드 개발자)
     - `internal/adapters/db/postgres/user.go`: 인터페이스 구현
     - `internal/adapters/db/postgres/platform_account.go`: 트랜잭션 처리 추가
     - `internal/adapters/db/postgres/token.go`: Redis와 연계된 블랙리스트 로직
     - `internal/adapters/db/postgres/audit_log.go`: 비동기 로깅 최적화
  3. **Redis 캐싱 구현** (2일, 백엔드 개발자)
     - `internal/adapters/redis/cache.go`: `SetTokenBlacklist()`, `GetTokenStatus()`
     - TTL 설정: 액세스 토큰 만료 시간과 동기화
  4. **테스트 작성** (3일, 백엔드 개발자)
     - `internal/adapters/db/postgres/*_test.go`: `pgxmock` 사용
     - `internal/adapters/redis/*_test.go`: `miniredis` 사용
     - 에러 처리 및 경계 케이스 테스트
- **산출물**: 저장소 인터페이스, PostgreSQL/Redis 구현, 테스트 코드
- **품질 보증**: 트랜잭션 테스트 완료, 90% 이상 커버리지 달성

---

### 단계 5: 서비스 로직 및 gRPC 구현 (2.5주)
- **목표**: 비즈니스 로직을 서비스 계층에 통합하고 gRPC로 외부 노출
- **작업 항목**:
  1. **gRPC 인터페이스 정의** (3일, 백엔드 개발자)
     - `api/proto/auth/v1/auth.proto`: `Login`, `Logout`, `RefreshToken`
     - `api/proto/auth/v1/user.proto`: `CreateUser`, `GetUser`
     - `api/proto/auth/v1/platform.proto`: `ConnectPlatformAccount`
     - `protoc` 실행: 출력 경로 `internal/generated/`
  2. **서비스 로직 구현** (5일, 백엔드 개발자)
     - `internal/core/service/auth_service.go`: `Login()`, `RefreshToken()`
     - `internal/core/service/user_service.go`: `CreateUser()`, `UpdateUser()`
     - `internal/core/service/platform_service.go`: `Connect()`, `Disconnect()`
     - 의존성 주입: `wire.go`에 서비스 초기화 추가
  3. **gRPC 서버 구현** (4일, 백엔드 개발자)
     - `internal/adapters/grpc/server/auth.go`: gRPC 메서드 매핑
     - `internal/adapters/grpc/server/user.go`: RBAC 검증 로직 포함
     - `internal/adapters/grpc/server/platform.go`: OAuth 토큰 처리
  4. **인터셉터 구현** (3일, 백엔드 개발자)
     - `internal/adapters/grpc/interceptors/auth.go`: JWT 검증
     - `internal/adapters/grpc/interceptors/logging.go`: `zap` 기반 구조화 로깅
     - `internal/adapters/grpc/interceptors/metrics.go`: Prometheus 메트릭
- **산출물**: `.proto` 파일, 생성된 코드, 서비스 로직, gRPC 서버
- **품질 보증**: gRPC 호출 테스트 (`grpcurl`), 인터셉터 동작 확인

---

### 단계 6: 인증 유틸리티 및 토큰 관리 (1.5주)
- **목표**: 안전하고 효율적인 JWT 및 토큰 관리 기능 구현
- **작업 항목**:
  1. **JWT 유틸리티 구현** (3일, 백엔드 개발자)
     - `pkg/crypto/jwt.go`: `GeneratePair()`, `Validate()`, `Revoke()`
     - RS256 키 관리: 환경 변수 또는 Kubernetes Secret 사용
  2. **비밀번호 해싱 구현** (2일, 백엔드 개발자)
     - `pkg/crypto/password.go`: `Hash()`, `Verify()` (Argon2id, salt 포함)
  3. **토큰 관리 통합** (2일, 백엔드 개발자)
     - `internal/core/service/token.go`: 블랙리스트 관리, 리프레시 로직
     - Redis와 연계: `SetNX`로 중복 방지
- **산출물**: 인증 유틸리티, 토큰 관리 로직, 단위 테스트
- **품질 보증**: 보안 리뷰 (토큰 서명 검증), 성능 테스트 (토큰 생성 속도)

---

### 단계 7: 이벤트 발행 및 Kafka 통합 (1.5주)
- **목표**: 이벤트 기반 아키텍처 구현 및 Kafka 연동
- **작업 항목**:
  1. **이벤트 퍼블리셔 인터페이스 정의** (1일, 백엔드 개발자)
     - `internal/core/ports/event_publisher.go`: `Publish(event Event)`
  2. **Kafka 구현** (3일, 백엔드 개발자)
     - `internal/adapters/kafka/producer.go`: `segmentio/kafka-go` 사용
     - 토픽 설정: `auth.events.user.*`, `auth.events.security.*`
  3. **서비스 로직 통합** (3일, 백엔드 개발자)
     - `internal/core/service/*`: `UserCreated`, `LoginFailed` 등 발행
     - 비동기 발행으로 성능 최적화
- **산출물**: Kafka 퍼블리셔, 이벤트 통합 코드, 테스트
- **품질 보증**: 이벤트 발행 확인 (`kcat`으로 토픽 검증)

---

### 단계 8: 테스트 및 품질 검증 (2.5주)
- **목표**: 코드 안정성과 품질을 Production 수준으로 보장
- **작업 항목**:
  1. **단위 테스트 보강** (5일, 백엔드 개발자)
     - `internal/core/*_test.go`: 경계 케이스, 에러 처리 테스트
     - `internal/adapters/*_test.go`: 모킹 기반 테스트
     - 목표: 90% 커버리지
  2. **통합 테스트 작성** (5일, 백엔드 개발자/QA)
     - `test/integration/auth_test.go`: 로그인 → 토큰 검증 → 로그아웃 워크플로우
     - Docker Compose로 의존 서비스 실행
  3. **성능 테스트** (3일, 백엔드 개발자/DevOps)
     - `test/performance/load_test.go`: `k6`로 500 RPS 목표, 응답 시간 < 100ms
     - 결과 분석 및 최적화 (DB 쿼리 튜닝 등)
  4. **보안 테스트** (3일, 보안 전문가)
     - OWASP Top 10 점검: SQL Injection, JWT 조작 테스트
     - `go-sec`, `snyk`으로 정적 분석
- **산출물**: 테스트 코드, 성능 보고서, 보안 보고서
- **품질 보증**: 모든 테스트 통과, 보안 취약점 0건

---

### 단계 9: 배포 및 모니터링 설정 (1.5주)
- **목표**: Kubernetes 기반 배포 및 운영 환경 구축
- **작업 항목**:
  1. **Kubernetes 매니페스트 작성** (3일, DevOps 엔지니어)
     - `deployments/kubernetes/deployment.yaml`: 3 replicas, HPA (CPU 80%)
     - `deployments/kubernetes/service.yaml`: gRPC 포트 노출
     - `deployments/kubernetes/secret.yaml`: JWT 키, DB 비밀번호
  2. **CI/CD 파이프라인 완성** (3일, DevOps 엔지니어)
     - GitHub Actions: 빌드 → 테스트 → 이미지 푸시 → ArgoCD 동기화
     - 롤백 전략: `kubectl rollout undo` 지원
  3. **모니터링 설정** (2일, DevOps 엔지니어)
     - `pkg/metrics/prometheus.go`: 커스텀 메트릭 추가 (예: `login_latency`)
     - Grafana 대시보드: 요청율, 오류율, 지연 시간 시각화
- **산출물**: Kubernetes 설정, CI/CD 워크플로우, 모니터링 대시보드
- **품질 보증**: 배포 성공, 모니터링 데이터 수집 확인

---

### 단계 10: 문서화 및 최종 검토 (1주)
- **목표**: 완성된 서비스를 문서화하고 배포 준비 완료
- **작업 항목**:
  1. **문서화 완료** (3일, 모든 팀원)
     - `README.md`: 설치, 실행, 배포 가이드
     - `docs/api/grpc-services.md`: gRPC API 상세
     - `docs/operations/monitoring.md`: 모니터링 설정
  2. **최종 검토 및 QA** (2일, 모든 팀원)
     - 코드 리뷰: 모든 PR 병합 전 최종 확인
     - 배포 테스트: 스테이징 환경에서 검증
- **산출물**: 최종 문서, 배포된 서비스
- **품질 보증**: 팀 승인, 스테이징 테스트 통과

---

## 3. 전체 타임라인
| 단계                  | 기간       | 시작 날짜 (예시) | 종료 날짜 (예시) |
|-----------------------|------------|-------------------|-------------------|
| 1. 초기 설정          | 1주       | 3/1             | 3/7             |
| 2. DB 스키마          | 1.5주     | 3/8             | 3/18            |
| 3. 도메인 모델        | 2주       | 3/19            | 4/1             |
| 4. 저장소 계층        | 2주       | 4/2             | 4/15            |
| 5. 서비스/gRPC        | 2.5주     | 4/16            | 5/2             |
| 6. 인증 유틸리티      | 1.5주     | 5/3             | 5/13            |
| 7. 이벤트/Kafka       | 1.5주     | 5/14            | 5/24            |
| 8. 테스트/검증        | 2.5주     | 5/25            | 6/10            |
| 9. 배포/모니터링      | 1.5주     | 6/11            | 6/21            |
| 10. 문서화/검토       | 1주       | 6/22            | 6/28            |
| **총 기간**            | **10~12주** | **3/1**         | **6/28**        |

- **참고**: 병렬 작업 시 (예: DevOps와 백엔드 동시 진행) 10주 내 완료 가능

---

## 4. 품질 보증 전략
- **코드 품질**: `golangci-lint` 실행, 90% 이상 테스트 커버리지, 팀 간 코드 리뷰 필수
- **보안**: OWASP Top 10 준수, 정기 취약점 스캔, mTLS 인증
- **성능**: 로그인 TPS 500 이상, 응답 시간 < 100ms, DB 쿼리 최적화
- **안정성**: 트랜잭션 처리, 에러 핸들링, 롤백 테스트 완료

---

## 5. 결론
이 계획은 **Production-ready** 수준의 **ImmersiVerse Authentication Service**를 개발하기 위한 상세 로드맵입니다. 각 단계는 세분화된 작업과 품질 보증 활동을 포함하여 안정성, 보안, 성능을 극대화하도록 설계되었습니다. 팀은 이 계획을 기반으로 작업을 분배하고, 주간 점검을 통해 진척도를 관리해 주세요.