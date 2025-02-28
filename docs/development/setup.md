# 개발 환경 설정 가이드

본 문서는 **ImmersiVerse Authentication Service**를 로컬(개발용) 환경에서 실행하기 위한 준비 과정을 안내합니다. 의존성 설치, Docker Compose 사용, 프로토콜 버퍼 컴파일, 환경 변수 설정 등을 순서대로 설명하여 빠른 시작이 가능하도록 합니다.

---

## 1. 필수 요구사항

- **Go 1.23+**  
  - Go 모듈을 사용하므로 최소 1.23 이상 필요, 권장 1.22+
- **Docker & Docker Compose**  
  - 로컬 DB(PostgreSQL), Redis, Kafka 등 서비스 의존성을 빠르게 실행
- **Protocol Buffers 컴파일러(`protoc`)**  
  - 버전 3.15 이상 권장
- **Git**  
  - 소스 코드 관리
- **Make** (optional but recommended)
  - 프로젝트에서 Makefile 사용

---

## 2. 저장소 클론 및 코드 구성

1. **Git Clone**:
   ```bash
   git clone https://github.com/immersiverse/auth-service.git
   cd auth-service
   ```
2. **프로젝트 구조**(요약):
   ```
  auth-service/
  ├── cmd/                    # 애플리케이션 실행 진입점
  │   ├── auth-service/       # 인증 서비스 실행 파일
  │   │   ├── main.go        # gRPC 서버와 Kafka 초기화
  │   │   └── wire.go        # 의존성 주입 설정
  │   └── tools/             # 개발 도구 실행 파일
  │       └── migrate.go     # DB 마이그레이션 CLI
  ├── api/                    # 외부 인터페이스 정의 (gRPC)
  │   └── proto/             # gRPC 프로토콜 버퍼 정의
  │       ├── auth/          # 인증 관련 API 정의
  │       │   └── v1/        
  │       │       └── auth.proto       # 인증 서비스 정의 (생성 코드 제외)
  │       └── platform/      # 플랫폼 연동 관련 API 정의
  │           └── v1/
  │               └── platform.proto   # 플랫폼 계정 연결 API (생성 코드 제외)
  ├── internal/               # 내부 구현
  │   ├── adapters/          # 외부 시스템 연결
  │   │   ├── db/            # 데이터베이스 어댑터
  │   │   │   ├── postgres/  # PostgreSQL 구현
  │   │   │   │   ├── user_repository.go      # 사용자 데이터 접근
  │   │   │   │   ├── token_repository.go     # 토큰 데이터 접근
  │   │   │   │   └── platform_repository.go  # 플랫폼 계정 데이터
  │   │   │   └── mocks/     # 테스트용 목
  │   │   ├── kafka/         # Kafka 이벤트 발행
  │   │   │   ├── producer.go         # Kafka Producer
  │   │   │   └── mocks/              # 테스트용 목
  │   │   ├── redis/         # Redis 캐싱
  │   │   │   ├── cache.go           # 캐싱 로직
  │   │   │   └── mocks/             # 테스트용 목
  │   │   └── grpc/          # gRPC 어댑터
  │   │       ├── client/    # 내부 서비스 호출용 클라이언트
  │   │       │   └── client.go      # gRPC 클라이언트
  │   │       ├── server/    # gRPC 서버
  │   │       │   └── server.go      # gRPC 서버 로직
  │   │       └── interceptors/      # 인터셉터
  │   │           ├── auth.go        # 토큰 검증
  │   │           └── logging.go     # 로깅
  │   ├── core/              # 도메인 로직
  │   │   ├── domain/        # 도메인 모델
  │   │   │   ├── user.go            # 사용자 엔티티
  │   │   │   ├── token.go           # 토큰 엔티티
  │   │   │   ├── platform_account.go # 플랫폼 계정 엔티티
  │   │   │   └── events.go          # 도메인 이벤트
  │   │   ├── service/       # 애플리케이션 서비스
  │   │   │   ├── auth_service.go    # 인증 로직
  │   │   │   ├── user_service.go    # 사용자 관리 로직
  │   │   │   └── platform_service.go # 플랫폼 연동 로직
  │   │   └── ports/         # 도메인 인터페이스
  │   │       ├── repository.go      # 저장소 인터페이스
  │   │       ├── event_publisher.go # 이벤트 발행 인터페이스
  │   │       ├── cache.go           # 캐시 인터페이스
  │   │       └── grpc_client.go     # gRPC 클라이언트 인터페이스
  │   ├── config/            # 설정 관리
  │   │   ├── config.go             # 환경 변수 로드
  │   │   └── config.yaml           # 기본 설정
  │   ├── middleware/        # 공통 미들웨어
  │   │   ├── auth.go               # 토큰 검증
  │   │   ├── logging.go            # 로깅
  │   │   └── rate_limit.go         # 속도 제한
  │   └── generated/         # protoc으로 생성된 gRPC 코드
  │       ├── auth/          # 인증 관련 생성 코드
  │       │   └── v1/        
  │       │       ├── auth.pb.go       # 생성된 Go 코드
  │       │       └── auth_grpc.pb.go  # gRPC 서비스 인터페이스
  │       └── platform/      # 플랫폼 연동 관련 생성 코드
  │           └── v1/
  │               ├── platform.pb.go   # 생성된 Go 코드
  │               └── platform_grpc.pb.go # gRPC 서비스 인터페이스
  ├── pkg/                    # 공용 유틸리티
  │   ├── crypto/            # 암호화 유틸리티
  │   │   ├── hash.go               # 비밀번호 해싱
  │   │   └── jwt.go                # JWT 생성/검증
  │   ├── logger/            # 로깅 유틸리티
  │   │   └── logger.go             # 구조화된 로깅
  │   ├── metrics/           # 메트릭 수집
  │   │   └── prometheus.go         # Prometheus 메트릭
  │   └── validator/         # 입력 검증
  │       └── validator.go          # 요청 데이터 검증
  ├── db/                     # 데이터베이스 관련 파일
  │   ├── migrations/        # 스키마 마이그레이션
  │   │   ├── 001_init.up.sql       # 초기 테이블 생성
  │   │   └── 001_init.down.sql     # 롤백
  │   └── seeds/             # 초기 데이터
  │       └── users.sql             # 샘플 데이터
  ├── test/                   # 테스트 코드
  │   ├── unit/              # 단위 테스트
  │   │   ├── domain/               # 도메인 로직 테스트
  │   │   ├── service/              # 서비스 로직 테스트
  │   │   └── adapters/             # 어댑터 테스트
  │   ├── integration/       # 통합 테스트
  │   │   ├── auth_test.go          # 인증 워크플로우 테스트
  │   │   └── platform_test.go      # 플랫폼 연동 테스트
  │   └── performance/       # 성능 테스트
  │       └── load_test.go          # 부하 테스트
  ├── deployments/            # 배포 관련 파일
  │   ├── docker/            # Docker 설정
  │   │   ├── Dockerfile            # 서비스 빌드
  │   │   └── docker-compose.yml    # 개발용 컴포즈
  │   └── kubernetes/        # Kubernetes 매니페스트
  │       ├── deployment.yaml       # 배포 설정
  │       ├── service.yaml          # 서비스 설정
  │       └── configmap.yaml        # 환경 설정
  ├── docs/                   # 문서
  │   ├── architecture/      # 아키텍처 문서
  │   ├── api/               # gRPC API 문서
  │   ├── development/       # 개발 가이드
  │   └── ...                # 기타 문서
  ├── Makefile               # 빌드 및 개발 작업 자동화
  ├── go.mod                 # Go 모듈 정의
  ├── go.sum                 # 의존성 체크섬
  ├── .gitignore             # Git 무시 패턴
  └── README.md              # 프로젝트 개요
   ```

---

## 3. 의존성 설치

### 3.1 Go 의존성

- 루트 디렉토리에서:
  ```bash
  go mod download
  ```
- `go.sum`이 최신 상태인지 확인:
  ```bash
  go mod tidy
  ```

### 3.2 Protocol Buffers

- **protoc** 설치 확인:
  ```bash
  protoc --version
  ```
- **Go용 protoc 플러그인**:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```

### 3.3 Docker Compose

- `docker-compose --version`으로 확인  
- `docker-compose.yml` 파일을 이용해 로컬 의존성(DB, Redis, Kafka 등) 실행

---

## 4. 로컬 의존성 실행

프로젝트에서는 `docker-compose.yml`을 통해 PostgreSQL, Redis, Kafka 등을 한 번에 띄울 수 있습니다.

```bash
# 예시: deployments/docker
cd deployments/docker
docker-compose up -d
```

- **PostgreSQL**: `localhost:5432`, DB명 `auth_db`, 유저/패스 `auth_user/auth_password` (예시)
- **Redis**: `localhost:6379`
- **Kafka**: `localhost:9092` (예: `kafka:9092` alias)

컨테이너 상태 확인:
```bash
docker-compose ps
```

---

## 5. 데이터베이스 마이그레이션

1. **DB 접속 확인**:
   ```bash
   psql -h localhost -U auth_user -d auth_db
   ```
2. **마이그레이션 실행**(프로젝트 내 `Makefile` 기준):
   ```bash
   make migrate
   ```
   - 또는 별도 툴(e.g. `golang-migrate`) 사용:  
     ```bash
     migrate -path db/migrations -database "postgres://auth_user:auth_password@localhost:5432/auth_db?sslmode=disable" up
     ```
3. **시드 데이터**(optional):  
   ```bash
   make seed
   ```
   - db/seeds/ 디렉토리에 기본 Admin 계정 등 삽입 스크립트

---

## 6. Proto 컴파일

- **Makefile**에 `proto` 타겟이 정의돼 있다면:
  ```bash
  make proto
  ```
  - `proto/auth/v1/*.proto` → `.pb.go`, `.pb.gw.go` 등 생성
- 생성된 파일은 `internal/grpc`, `pkg/...` 등에서 import해 사용

---

## 7. 환경 변수 설정

Authentication Service는 다양한 환경 변수를 통해 DB 접속 정보, JWT 시크릿(또는 경로), Kafka 브로커 주소 등을 주입받습니다. 예시:

1. **.env.example**:
   ```
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=auth_user
   DB_PASS=auth_password
   DB_NAME=auth_db
   REDIS_ADDR=localhost:6379
   KAFKA_BROKER=localhost:9092
   JWT_PRIVATE_KEY_PATH=./certs/private.pem
   JWT_PUBLIC_KEY_PATH=./certs/public.pem
   ```
2. **로컬 개발 시**:
   ```bash
   cp .env.example .env
   # .env 수정
   source .env
   ```
3. **Secrets**(민감 정보) 관리:
   - 실제 운영에서는 Docker Secret, Vault, AWS Parameter Store 등 사용 권장

---

## 8. 서비스 실행

### 8.1 Makefile 사용

```bash
make run
```
- `go run ./cmd/server` 형태로 Auth Service 실행
- 로그에 "Listening on :50051" 등 출력 확인

### 8.2 수동 실행

```bash
go build -o bin/auth-service ./cmd/server
./bin/auth-service
```
- 혹은 `go run ./cmd/server`.

---

## 9. 테스트

### 9.1 단위 테스트

```bash
make test
```
- 또는:
  ```bash
  go test ./internal/... -v
  ```

### 9.2 통합 테스트

- Docker Compose로 DB 등 의존성을 띄운 뒤:
  ```bash
  make integration-test
  ```
- gRPC 호출 모의, DB에 실제 레코드 확인 등.

### 9.3 프로토 테스트

- `grpcurl` 또는 BloomRPC로 `localhost:50051` 접속해 `AuthService.Login` 등 호출.

---

## 10. 디버깅 & 로깅

- 기본적으로 **zap**(또는 logrus, etc.)로 로깅
- 로컬 환경에서 로그 레벨 `DEBUG`로 설정
- **Go Delve**를 사용하면 VSCode나 GoLand에서 디버깅 가능

---

## 11. 문제 해결 FAQ

1. **DB 연결 실패**: 
   - Docker Compose가 정상 실행됐는지 `docker-compose logs postgres` 확인
   - ENV 설정(DB_HOST, DB_PORT 등) 체크
2. **프로토 컴파일 에러**: 
   - protoc-gen-go, protoc-gen-go-grpc 버전 확인
   - `$GOPATH/bin`이 PATH에 포함됐는지 확인
3. **Kafka 연결 오류**:
   - `docker-compose logs kafka`에서 에러 확인
   - `KAFKA_BROKER` 주소/포트 맞게 설정

---

## 12. 결론

위 과정을 따르면 로컬 머신에서 **Authentication Service**를 완전한 형태로 실행·개발할 수 있습니다.  
- **요약**:
  1. Git Clone 후 `go mod download`
  2. `docker-compose up -d` → DB/Redis/Kafka 실행
  3. `make migrate` → DB 스키마 설정
  4. `make proto` → 프로토 컴파일
  5. `.env` 생성하고 환경 변수 세팅
  6. `make run` → 서버 실행
  7. `make test` → 단위/통합 테스트
  
문제가 발생하면 [문제 해결 가이드](../references/troubleshooting.md) 및 FAQ를 먼저 참고하고, 해결되지 않을 시 팀 내에 문의하세요.