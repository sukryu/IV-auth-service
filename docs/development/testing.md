# 테스트 작성 가이드

본 문서는 **ImmersiVerse Authentication Service**의 Go 코드를 테스트할 때 따라야 할 원칙과 방법을 제시합니다. 단위 테스트, 통합 테스트, 모킹 전략, 커버리지 기준 등을 정의하여 팀원들이 일관된 방식으로 테스트를 작성하고, 서비스 품질을 유지·개선할 수 있도록 돕습니다.

---

## 1. 개요

- **언어/프레임워크**: Go 언어 (1.21+), 표준 `testing` 패키지
- **테스트 계층**:
  1. **단위 테스트(Unit Test)**: 함수/메서드 단위, mocking으로 외부 의존성 제거  
  2. **통합 테스트(Integration Test)**: 실제 DB, Kafka, Redis 등과 연동  
  3. **엔드투엔드 테스트(E2E)**: 전체 시스템(마이크로서비스 + API Gateway) 레벨(별도 레포지토리나 상위 테스트 프로젝트에서 진행 가능)

- **목표**:
  - 최소 80% 코드 커버리지(혹은 팀 합의된 수치)  
  - 각 주요 로직(로그인, 회원가입, 토큰 검증 등)에 대한 충분한 시나리오 커버  
  - 빌드 파이프라인(CI/CD)에서 자동 테스트 수행  

---

## 2. 디렉토리 구조

```
auth-service/
├── internal/
│   ├── domain/
│   ├── repository/
│   ├── service/
│   └── ...
├── test/
│   ├── integration/
│   │   └── integration_test.go
│   ├── performance/
│   │   └── load_test.go (optional)
│   └── ...
└── ...
```

1. **단위 테스트**: 보통 소스 파일과 같은 디렉토리에 `_test.go` 파일로 작성 (e.g. `service/auth_service_test.go`)  
2. **통합 테스트**: `test/integration/`에 두고, Docker Compose나 실제 DB를 사용.  
3. **성능/부하 테스트**: `test/performance/` 참고.

---

## 3. 단위 테스트 (Unit Tests)

### 3.1 목적

- 개별 함수/메서드를 **mock** or **fake** 의존성으로 테스트
- 빠른 실행, 빠른 피드백

### 3.2 작성 규칙

1. **Table-driven tests**(Go community/K8s 권장):
   ```go
   func TestLogin(t *testing.T) {
       tests := []struct{
           name string
           input LoginRequest
           mockRepo func() UserRepository
           wantErr bool
       }{
           {"valid", LoginRequest{...}, validRepoMock, false},
           {"invalid pw", LoginRequest{...}, invalidRepoMock, true},
       }
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T){
               repo := tt.mockRepo()
               service := AuthService{repo: repo}
               _, err := service.Login(tt.input)
               if (err != nil) != tt.wantErr {
                   t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
               }
           })
       }
   }
   ```
2. **Mocking**:
   - **Interface** 기반 의존성 주입: e.g. `UserRepository`, `TokenManager` etc.  
   - Use custom stubs or a library (gomock, testify mock, or else).
3. **Naming**: `_test.go` suffix, `TestXxx` function.  
4. **Coverage**: 핵심 로직(비밀번호 해싱, 토큰 생성/검증) 모두 커버.

### 3.3 에러 케이스

- 잘못된 입력(Empty username, invalid email format, etc.)
- DB insert fail mock
- Password hash mismatch

---

## 4. 통합 테스트 (Integration Tests)

### 4.1 목적

- 실제 DB(PostgreSQL), Redis, Kafka 등과 연동하여 동작 여부 검증
- 서비스 레벨 시나리오 (로그인→DB select→JWT 발급→이벤트 발행 등)

### 4.2 환경 준비

1. **Docker Compose**: `deployments/docker-compose.yml`를 로컬 실행
   ```bash
   docker-compose up -d
   ```
2. **마이그레이션**:
   ```bash
   make migrate
   ```
3. **테스트 실행**:
   ```bash
   go test ./test/integration -v
   ```
   또는 `make integration-test`

### 4.3 테스트 예시

```go
func TestUserSignupIntegration(t *testing.T) {
    // Setup DB connection
    db, err := sql.Open("postgres", "postgres://auth_user:auth_password@localhost:5432/auth_db?sslmode=disable")
    if err != nil {
        t.Fatalf("failed to connect db: %v", err)
    }
    // Cleanup if needed
    defer db.Close()

    // Possibly set up gRPC client stub to local service
    // or call "start local server" code

    // 1. CreateUser request
    // 2. Check DB record
    // 3. Check Kafka message (optional)
}
```

### 4.4 시나리오 커버

- 회원가입 → DB insert → user row 생성
- 로그인 → DB user read → JWT 생성 → Redis cache → Kafka login event
- 플랫폼 연동 → 외부 OAuth mock or staging → DB insert  
- 에러 상황(DB unavailable, etc.)

---

## 5. 모의(Stub) 외부 서비스

- **플랫폼 OAuth** 통합 시, 테스트 환경에서 실제 Twitch/YouTube API를 호출하면 불안정.  
- **Mock server**(httptest.Server) or [WireMock](https://wiremock.org) 등으로 외부 플랫폼 응답 모사  
- Kafka → local broker in Docker Compose  
- Redis → local Docker container

---

## 6. 커버리지 & CI

- **목표**: 80% 이상(또는 팀 합의 수치)  
- **명령**:
  ```bash
  go test ./... -cover -coverprofile=coverage.out
  go tool cover -func=coverage.out
  go tool cover -html=coverage.out
  ```
- CI에서 `go test -cover` 결과를 수집, 실패 임계치(`-covermode=count -coverprofile=...`)

---

## 7. Test 분류 태그 (선택 사항)

- K8s 등에서는 `_test` suffix + build tags(예: `// +build integration`) 사용  
- e.g. `//go:build integration` 상단에 써서 `go test -tags=integration`

---

## 8. 성능 테스트 (간단 언급)

- 별도 `test/performance/` 디렉토리
- [k6](https://k6.io), locust, or `go test -bench`
- 목표: 로그인 TPS, DB 응답 지연 파악

---

## 9. Pull Request & Review

- **자동화**:
  1. 코드 변경 → Pull Request  
  2. CI가 lint + test(단위/통합) 실행  
  3. 커버리지 보고서, 테스트 결과  
- **리뷰어**는 test coverage, table-driven style, mock usage 등 확인

---

## 10. 결론

이 가이드는 **Authentication Service**의 **테스트 작성 방법**을 일관성 있게 유지하고, 빠른 피드백 루프를 통해 서비스 품질을 높이기 위한 표준을 제공합니다.

1. **단위 테스트**: small, fast, mock dependencies  
2. **통합 테스트**: 실제 DB/Kafka/Redis 환경, 시나리오별 end-to-end flow  
3. **자동화**: CI에서 `make test` → lint & coverage 체크  
4. **고수준 목표**: 코드 커버리지≥80%, critical path(로그인, 토큰, DB연동) 철저 검증  