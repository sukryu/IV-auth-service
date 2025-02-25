# 코드 스타일 가이드

본 문서는 **ImmersiVerse Authentication Service**의 **Go 코드 작성 스타일**을 정의합니다.  
**Kubernetes 프로젝트**(k8s.io)에서 사용하는 [Go 코드 스타일](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/go-code-review-comments.md)·패턴을 최대한 준수하여, 팀원 간 일관된 코딩 습관과 높은 유지보수성을 달성합니다.

---

## 1. 개요

- **언어**: Go 1.21 이상  
- **스타일 기준**: [Kubernetes Official Go Style](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/go-code-review-comments.md), [Effective Go](https://go.dev/doc/effective_go)  
- **자동화**: `gofmt`, `goimports`, `golangci-lint` 등 도구를 통해 스타일/린팅 체크  
- **목표**:  
  - 간결하고 읽기 쉬운 코드  
  - 함수, 변수 이름의 일관성  
  - 명확한 패키지 구조, public/private 구분  
  - 에러 처리, 동시성 패턴 등 K8s 스타일 준수

---

## 2. 파일/디렉토리 구조

1. **패키지 구조**  
   - Kubernetes 스타일은 `pkg/`, `cmd/`, `internal/` 디렉토리를 중심으로 역할별로 코드를 배치  
   - 예: 
     - `cmd/` - 메인 함수가 위치 (서버 실행 진입점)  
     - `internal/` - 도메인 로직, 비공개 패키지  
     - `pkg/` - 외부에서도 재사용 가능한 라이브러리 코드  

2. **파일명**:  
   - 소문자/스네이크케이스(`token_manager.go`, `auth_service.go`)  
   - 테스트 파일은 `_test.go` suffix

3. **Go Modules**:  
   - `go.mod`에서 **import path**를 깔끔하게 유지.  
   - Mismatched module paths 금지 (K8s도 모듈 명칭 정합 중요)

---

## 3. Imports & Formatting

1. **가장 기본**: `gofmt -s` 적용  
   - 코드 저장 시 자동 포맷 (VSCode, GoLand 등)  
2. **Imports 정렬**: `goimports` 사용  
   - 표준 라이브러리, 제3자, 로컬 패키지를 구분 블록  
3. **Unused Imports**: 허용 불가, CI에서 린터가 경고/오류 처리

예시:
```go
import (
    "fmt"
    "net/http"

    "github.com/myorg/somepkg"
    "k8s.io/klog/v2"

    "github.com/immersiverse/auth-service/internal/domain"
)
```

---

## 4. Naming & Commenting

### 4.1 일반 규칙

- **Go 표준**: `PascalCase`로 exported identifiers, `camelCase`로 non-exported  
- Kubernetes 레포 스타일처럼, 간결하고 충돌 없는 이름  
- 축약형 남발 지양 (`usr` 대신 `user`, `cfg` 대신 `config` → 명확성이 우선)  
- Acronyms(약어)는 대문자로(`HTTPServer`, `JSONData`)  
- 길이가 너무 긴 이름은 피하지만, 의미 명확성을 우선

### 4.2 함수/메서드 주석

- **exported 함수**: "FuncName ... " 형태로 시작하는 Godoc 주석 (K8s도 동일)  
- **비공개 함수**: 주석은 필요 시(복잡 로직)만 작성, 과도한 주석 X  
- 예시:
  ```go
  // Login authenticates user credentials and returns a token pair.
  func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
      // ...
  }
  ```

### 4.3 패키지 주석

- `doc.go`를 활용할 수도 있음. K8s도 일부 SIG에서 doc.go 사용  
- 패키지의 목적과 사용법 간략히 설명

---

## 5. 에러 처리

1. **에러 변수**: Kubernetes 스타일, `fmt.Errorf()` 또는 `errors.New()` / `%w`로 wrapping  
   - `return fmt.Errorf("failed to do X: %w", err)`  
2. **Error strings**: 소문자로 시작, 마침표 등 불필요. (K8s convention)  
   - e.g. `"failed to connect database: %w"`  
3. **단일 `if err != nil`** 블록에서 처리, K8s도 `utilerrors` 패턴 sometimes.  
4. 구체적 context wrapping:
   ```go
   if err := s.db.Insert(user); err != nil {
       return fmt.Errorf("insert user in db: %w", err)
   }
   ```

---

## 6. Logging & klog.v2

- **Kubernetes**는 klog(기본 `klog/v2`)를 사용  
- 레벨별 로깅(`klog.V(2).Infof("message", ...)`) 권장  
- 프로젝트가 zap/logrus 사용 시, klog와 혼합하지 않도록. (단, K8s 코드 스타일 따라 klog 쓰고 싶다면, 전체 통일)

---

## 7. Concurrency & Channels

- **Kubernetes**는 channel, goroutine 사용 시 매우 신중하게 설계  
- shared var → sync.Mutex, atomic 등 명확히.  
- prefer "Go idioms" for concurrency, e.g. "errgroup" for parallel tasks, or "context cancellation"  
- **K8s**는 context cancellation 패턴 (ctx.Done()) 중시

---

## 8. Lint & CI

1. **golangci-lint**:
   - `.golangci.yml` or `.golangci-lint.yaml` 설정  
   - Enable: `govet`, `golint`, `goconst`, `gocyclo`, `goimports`, `misspell` etc.
2. **CI/CD**:
   - PR 생성 시 lint+unit test 자동 실행  
   - K8s style PR bot flow(approve, lgtm) → optional

---

## 9. Testing

- **Test 파일**: `<filename>_test.go`
- **Table-driven tests**: K8s, Go community 권장. e.g.:
  ```go
  func TestLogin(t *testing.T) {
      tests := []struct{
          name string
          input ...
          wantErr bool
      }{
          {"valid", ..., false},
          {"invalid pw",..., true},
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T){
              ...
          })
      }
  }
  ```
- **Coverage**: 목표 80% 이상(또는 팀 정책)

---

## 10. Imports from Kubernetes libraries

- 만약 Kubernetes 공식 라이브러리(예: `k8s.io/apimachinery`, `k8s.io/utils`) 등을 사용한다면:
  - import alias 가능: `import utilrand "k8s.io/apimachinery/pkg/util/rand"`
  - vendor folder or go mod tidy ensure consistent versions

---

## 11. 기타 모범 사례

- **Getter**: Go style → `func (u *User) Name() string` vs. `GetName()`. K8s generally avoids `GetXxx` if not mandatory.  
- **Interface**: "accept interfaces, return structs" → K8s' typical approach.  
- **Package name**: short, all lower-case, no underscores.  
- Avoid stuttering: package `auth` → type `Service`, not `auth.AuthService`.

---

## 12. 결론

이 문서는 **Kubernetes 공식 소스 코드**와 **Go community** 스타일을 결합한 **Authentication Service**의 코드 컨벤션입니다.  
- 포매팅(gofmt, goimports),  
- 네이밍(Exported CamelCase, error strings in lower-case),  
- 주석(Godoc style),  
- 에러 처리(fmt.Errorf + wrap),  
- 로깅(K8s klog 또는 zap)  
등을 준수함으로써, 팀 내 높은 가독성과 유지보수성을 확보할 수 있습니다.