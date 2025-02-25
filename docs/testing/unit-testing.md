# 단위 테스트 전략

본 문서는 **ImmersiVerse Authentication Service**에서 단위 테스트(Unit Testing)를 효과적으로 작성하고 관리하기 위한 전략과 권장사항을 설명합니다. 단위 테스트는 개별 함수 및 모듈의 동작을 검증하여, 코드 변경 시 예상치 못한 사이드 이펙트를 최소화하고, 전체 애플리케이션의 안정성을 확보하는 데 중요한 역할을 합니다.

---

## 1. 단위 테스트의 목적

- **정확성 검증**: 각 함수와 메서드의 입력, 출력, 에러 처리 등을 독립적으로 검증합니다.
- **회귀 방지**: 코드 수정 및 리팩토링 시 기존 기능이 올바르게 동작하는지 보장합니다.
- **문서화**: 테스트 코드는 기능의 사용 예와 기대 동작을 문서화하는 역할도 수행합니다.
- **빠른 피드백**: CI/CD 파이프라인에 포함되어 변경 사항에 대한 빠른 피드백을 제공합니다.

---

## 2. 테스트 작성 원칙

### 2.1 Table-driven 테스트

- **표준 패턴**: Go 커뮤니티 및 Kubernetes 프로젝트에서도 권장하는 방식입니다.
- **예시**:
  ```go
  func TestLogin(t *testing.T) {
      tests := []struct {
          name     string
          req      LoginRequest
          setup    func() UserRepository  // 의존성 모킹 함수
          wantErr  bool
      }{
          {
              name: "Valid credentials",
              req: LoginRequest{Username: "validUser", Password: "ValidP@ss123"},
              setup: func() UserRepository {
                  // 성공 케이스를 위한 mock 구현
                  return NewMockUserRepository(func(username string) (*User, error) {
                      return &User{ID: "uuid-1234", Username: "validUser", PasswordHash: "hashedValue"}, nil
                  })
              },
              wantErr: false,
          },
          {
              name: "Invalid password",
              req: LoginRequest{Username: "validUser", Password: "WrongPass"},
              setup: func() UserRepository {
                  // 비밀번호 불일치 상황을 위한 mock 구현
                  return NewMockUserRepository(func(username string) (*User, error) {
                      return &User{ID: "uuid-1234", Username: "validUser", PasswordHash: "hashedValue"}, nil
                  })
              },
              wantErr: true,
          },
      }

      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              repo := tt.setup()
              service := NewAuthService(repo, /* 기타 의존성 */)
              _, err := service.Login(context.Background(), &tt.req)
              if (err != nil) != tt.wantErr {
                  t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
              }
          })
      }
  }
  ```

### 2.2 Mocking & Stub

- **인터페이스 기반**: 외부 의존성(DB, Redis, Kafka 등)은 인터페이스로 추상화하고, 테스트 시에는 모킹(mocking) 또는 스텁(stub)을 활용합니다.
- **라이브러리**: [gomock](https://github.com/golang/mock) 또는 [testify/mock](https://github.com/stretchr/testify) 같은 라이브러리를 사용해 모의 객체를 생성합니다.
- **목적**: 테스트 대상 코드의 동작을 독립적으로 검증하며, 외부 시스템과의 실제 연결 없이 빠른 테스트 실행을 보장합니다.

### 2.3 에러 및 경계 조건 테스트

- **정상 케이스 뿐만 아니라** 잘못된 입력, 에러 반환, 경계 조건에 대한 테스트를 반드시 포함합니다.
- 예를 들어, 비밀번호 검증에서 잘못된 포맷이나 빈 문자열, DB 연결 실패 등의 상황을 테스트합니다.

---

## 3. 테스트 도구 및 실행

### 3.1 표준 Go testing 패키지

- Go의 기본 `testing` 패키지를 사용하며, `go test ./...` 명령어로 전체 테스트를 실행합니다.
- **커버리지 측정**:
  ```bash
  go test ./... -cover -coverprofile=coverage.out
  go tool cover -html=coverage.out
  ```

### 3.2 CI/CD 연동

- GitHub Actions나 기타 CI 도구에 단위 테스트 실행을 포함하여, 모든 PR 및 배포 전 테스트가 통과하도록 설정합니다.
- 테스트 결과와 커버리지 리포트를 자동으로 확인하여, 기준(예: 80% 이상) 미달 시 빌드 실패 처리

---

## 4. 테스트 코드 구조

- **파일 네이밍**: 테스트 파일은 반드시 `_test.go` 접미사를 사용합니다.
- **패키지**: 테스트는 동일 패키지 내에서 작성하거나, 별도의 `*_test` 패키지를 활용하여 export되지 않은 함수까지 테스트할 수 있습니다.
- **테스트 함수**: `TestXxx(t *testing.T)` 형식으로 작성합니다.

---

## 5. 모범 사례

- **Table-driven tests**: 여러 입력과 기대 결과를 한 번에 테스트하여 중복 코드를 줄이고, 가독성을 높입니다.
- **모듈 단위 테스트**: 각 모듈(예: 인증 로직, 토큰 관리, 사용자 등록)의 경계를 명확히 하고, 독립적으로 테스트합니다.
- **코드 커버리지 목표**: 최소 80% 이상의 커버리지를 목표로 하며, 핵심 로직은 반드시 포함합니다.
- **테스트 격리**: 테스트 간 상태 공유를 피하고, 독립적으로 실행되도록 setup/teardown 로직을 구성합니다.
- **Mock 활용**: 외부 시스템과의 상호작용은 반드시 모의 객체를 사용하여 테스트 속도와 안정성을 높입니다.

---

## 6. 결론

단위 테스트는 **Authentication Service**의 코드 안정성과 신뢰성을 보장하는 핵심 요소입니다.  
- **Table-driven** 및 **모킹**을 통해 다양한 케이스를 신속하게 검증하고,  
- CI/CD 파이프라인에 포함하여 코드 변경 시 즉각적인 피드백을 받을 수 있도록 합니다.  
  
모든 팀원은 이 가이드를 준수하여 단위 테스트를 작성하고, 테스트 커버리지를 지속적으로 향상시켜 주세요.

---