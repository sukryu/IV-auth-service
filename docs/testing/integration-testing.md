# 통합 테스트 전략

본 문서는 **ImmersiVerse Authentication Service**의 통합 테스트(Integration Testing) 작성 및 실행 방법을 정리합니다. 통합 테스트는 개별 모듈(단위 테스트)뿐 아니라, 여러 구성 요소(데이터베이스, 메시징 시스템, 캐시 등) 간의 상호작용을 검증하여 실제 운영 환경에서의 기능 동작과 안정성을 확인하는 데 목적이 있습니다.

---

## 1. 통합 테스트의 목적

- **시스템 상호작용 검증**:  
  - 서비스 간 gRPC 호출, DB 연동, Kafka 이벤트 발행/구독 등의 실제 통합 동작 확인
- **환경 의존성 검증**:  
  - PostgreSQL, Redis, Kafka 등 외부 의존 서비스와의 연동 테스트
- **실제 시나리오 재현**:  
  - 회원가입, 로그인, 토큰 갱신, 플랫폼 계정 연결 등 주요 사용자 흐름에 대해 end-to-end 테스트 진행
- **회귀 방지 및 안정성 보장**:  
  - 주요 기능 변경 후 전체 통합 테스트를 통해 기존 기능이 정상적으로 동작하는지 확인

---

## 2. 테스트 환경 구성

### 2.1 외부 의존 서비스 실행

- **Docker Compose**를 활용하여 테스트 환경에서 다음과 같은 의존 서비스를 실행합니다:
  - **PostgreSQL**: 사용자 및 계정 데이터 저장
  - **Redis**: 토큰 캐싱 및 블랙리스트 관리
  - **Kafka**: 이벤트 발행 및 구독 테스트

예시:
```bash
cd deployments/docker
docker-compose up -d
```

### 2.2 환경 변수 및 설정

- 테스트 실행 전, `.env.testing` 파일을 생성하여 DB 접속 정보, Kafka 브로커 주소, Redis 주소 등 테스트용 환경 변수를 설정합니다.
- 예시 (.env.testing):
  ```dotenv
  DB_HOST=localhost
  DB_PORT=5432
  DB_USER=auth_user
  DB_PASS=auth_password
  DB_NAME=auth_db_test
  REDIS_ADDR=localhost:6379
  KAFKA_BROKER=localhost:9092
  JWT_PRIVATE_KEY_PATH=./certs/private.pem
  JWT_PUBLIC_KEY_PATH=./certs/public.pem
  ```

---

## 3. 통합 테스트 작성 가이드

### 3.1 테스트 시나리오

각 통합 테스트는 실제 서비스 흐름을 재현하며, 아래와 같은 주요 시나리오를 포함합니다:

- **회원가입**:  
  - `CreateUser` API를 호출하여 사용자가 정상적으로 생성되는지 확인하고, DB에 데이터가 저장되는지 검증합니다.
- **로그인 및 토큰 발급**:  
  - 올바른 자격 증명을 입력하여 JWT 토큰이 발급되고, Redis 및 Kafka에 이벤트가 발행되는지 확인합니다.
- **토큰 검증**:  
  - `ValidateToken` API를 호출하여 유효한 토큰과 만료/블랙리스트 처리된 토큰에 대해 올바른 응답을 반환하는지 검증합니다.
- **플랫폼 계정 연결**:  
  - 외부 OAuth 코드 교환을 모의하여, 플랫폼 계정이 DB에 정상적으로 저장되고 관련 이벤트가 발행되는지 확인합니다.
- **로그아웃**:  
  - 로그아웃 API 호출 후, 토큰 블랙리스트 등록 및 해당 토큰의 무효화가 올바르게 수행되는지 테스트합니다.

### 3.2 테스트 도구 및 프레임워크

- **Go의 testing 패키지**를 기본으로 사용하며, 통합 테스트 전용 패키지는 `test/integration` 폴더에 위치합니다.
- **HTTP 클라이언트**(예: `net/http`, `grpcurl` CLI)를 활용하여 실제 API 호출을 시뮬레이션합니다.
- 테스트 실행 시, Docker Compose로 구성된 의존 서비스에 연결하도록 설정합니다.

### 3.3 테스트 코드 예시

```go
package integration_test

import (
    "context"
    "database/sql"
    "net/http"
    "testing"
    "time"

    _ "github.com/lib/pq"
    "github.com/stretchr/testify/assert"
    authpb "github.com/immersiverse/auth-service/proto/auth/v1"
    "google.golang.org/grpc"
)

func TestUserSignupIntegration(t *testing.T) {
    // 1. DB 연결 테스트: PostgreSQL
    db, err := sql.Open("postgres", "postgres://auth_user:auth_password@localhost:5432/auth_db_test?sslmode=disable")
    if err != nil {
        t.Fatalf("DB connection failed: %v", err)
    }
    defer db.Close()

    // 2. gRPC 클라이언트 연결
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        t.Fatalf("gRPC connection failed: %v", err)
    }
    defer conn.Close()
    client := authpb.NewUserServiceClient(conn)

    // 3. CreateUser 요청 실행
    req := &authpb.CreateUserRequest{
        Username: "testuser",
        Email:    "testuser@example.com",
        Password: "StrongP@ssw0rd!",
    }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    resp, err := client.CreateUser(ctx, req)
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.User.Id)

    // 4. DB에서 사용자 데이터 검증
    var username string
    err = db.QueryRow("SELECT username FROM users WHERE id = $1", resp.User.Id).Scan(&username)
    assert.NoError(t, err)
    assert.Equal(t, "testuser", username)
}

func TestLoginIntegration(t *testing.T) {
    // 1. gRPC 클라이언트 연결 (AuthService)
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        t.Fatalf("gRPC connection failed: %v", err)
    }
    defer conn.Close()
    client := authpb.NewAuthServiceClient(conn)

    // 2. 로그인 요청 실행
    req := &authpb.LoginRequest{
        Username: "testuser",
        Password: "StrongP@ssw0rd!",
    }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    resp, err := client.Login(ctx, req)
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.AccessToken)
    assert.NotEmpty(t, resp.RefreshToken)

    // 3. 추가: 토큰 검증, Redis 캐시, Kafka 이벤트 발행 확인 (옵션)
    // 필요 시, 외부 시스템 모니터링을 통해 확인
}
```

### 3.4 테스트 실행

- **로컬 실행**:
  ```bash
  go test ./test/integration -v
  ```
- **Makefile** 내 통합 테스트 타겟 활용:
  ```makefile
  integration-test:
      go test ./test/integration -v
  ```

---

## 4. 환경 정리 및 초기화

- 각 통합 테스트는 독립적으로 실행될 수 있도록, 테스트 전후 DB 스키마 초기화 및 seed 데이터를 적용합니다.
- 테스트 완료 후, 의존 컨테이너(예: PostgreSQL, Redis, Kafka)의 상태를 정리(cleanup)합니다.
- CI/CD 파이프라인에서 통합 테스트를 실행하여, 모든 의존 서비스와의 연동이 정상 작동하는지 자동으로 검증합니다.

---

## 5. 모니터링 및 로깅 연동

- 통합 테스트 동안 **애플리케이션 로그**(구조화된 JSON 로그)와 **Kafka 이벤트 로그**를 모니터링하여, 정상 동작 및 문제 상황을 확인합니다.
- Prometheus와 Grafana를 활용하여 주요 메트릭(예: 응답 시간, 에러율 등)을 테스트 중에도 실시간으로 확인합니다.

---

## 6. 결론

통합 테스트는 **Authentication Service**가 단위 테스트에서 검증된 개별 모듈들을 실제 환경에서 올바르게 연동하여 동작하는지를 확인하는 핵심 단계입니다.

- **주요 테스트 시나리오**: 회원가입, 로그인, 토큰 검증, 플랫폼 계정 연동 등  
- **환경 구성**: Docker Compose를 통한 외부 의존 서비스 실행 및 환경 변수 설정  
- **실행**: gRPC 호출, DB 검증, 이벤트 발행/구독 확인  
- **자동화**: CI/CD 파이프라인에 통합 테스트 포함

이 가이드를 통해 팀원들은 통합 테스트를 작성하고 실행하여, 서비스 전체의 안정성과 운영 환경에서의 상호작용을 지속적으로 검증할 수 있습니다.

---