# 에러 코드 및 처리 방법

본 문서는 **ImmersiVerse Authentication Service**에서 발생하는 에러 코드와 에러 처리 방식을 정리합니다. 각 에러 상황에 대한 gRPC status code, 사용자 메시지(또는 로깅/추적용 메시지) 등을 정의하여, 클라이언트와 서버 간 통신의 일관성과 가독성을 높입니다.

---

## 1. 개요

- **Protocol Buffers + gRPC**를 사용하며, 서버는 `status.Status`(Go의 경우 `grpc/status`) 객체와 함께 에러를 반환함.
- 주요 에러 상황에 대해 **custom error code**(e.g. `USER_NOT_FOUND`, `INVALID_PASSWORD` 등)를 추가로 정의할 수 있으며, gRPC status code(예: `NOT_FOUND`, `UNAUTHENTICATED`)와 매핑.
- 클라이언트나 API Gateway는 이를 HTTP status code 등으로 변환하여 사용자에게 전달.

---

## 2. gRPC Status Code 매핑

| Status Code       | 의미                                              | 사용 예시                                           |
|-------------------|---------------------------------------------------|-----------------------------------------------------|
| **OK (0)**        | 요청 성공                                         | 정상 응답                                           |
| **INVALID_ARGUMENT (3)** | 요청 데이터가 잘못되었거나 필수 필드 누락         | 필드 유효성 에러, 잘못된 이메일 형식 등             |
| **NOT_FOUND (5)** | 요청한 리소스를 찾을 수 없음                      | 사용자 미존재, 플랫폼 계정 미존재                   |
| **ALREADY_EXISTS (6)** | 중복 리소스 존재                                | 이메일/아이디 중복, 유저명 중복                     |
| **PERMISSION_DENIED (7)** | 권한 없음                                   | 관리자 전용 API 호출, OAuth 실패                    |
| **FAILED_PRECONDITION (9)** | 특정 전제조건 불충족 (e.g. OAuth flow)   | 이미 연결된 플랫폼 계정 재연동 시도 등              |
| **UNAUTHENTICATED (16)** | 인증 실패, 로그인 불가, 토큰 무효             | 비밀번호 불일치, 토큰 만료/블랙리스트               |
| **INTERNAL (13)**  | 서버 내부 에러 (DB/Redis 장애, 예상치 못한 오류) | DB 연결 문제, Kafka 전송 실패, Panic 등             |

> 실제 사용 상황에 따라 적절히 선택. 중첩 혹은 세분화가 필요한 경우 `error_codes.proto`나 `ErrorInfo` 메시지를 병행.

---

## 3. 커스텀 에러 코드(ErrorInfo 메시지)

### 3.1 `ErrorInfo` 메시지 예시

```proto
// error_codes.proto
syntax = "proto3";

package auth.v1;

message ErrorInfo {
  int32 code = 1;           // e.g. 1001, 2001, ...
  string message = 2;       // human-readable summary
  map<string, string> details = 3; // additional info
}
```

- `code`: 내부적으로 관리하는 구체적인 에러 번호 (예: `1001` = INVALID_CREDENTIALS, `1002` = USER_ALREADY_EXISTS 등).
- `message`: 짧은 에러 설명.
- `details`: 확장 가능.

### 3.2 예시 매핑

| 내부 코드 | 의미                             | gRPC Status Code  | message 예시                      |
|-----------|----------------------------------|-------------------|-----------------------------------|
| **1001**  | 잘못된 사용자명/비밀번호          | UNAUTHENTICATED   | "Invalid credentials"             |
| **1002**  | 중복된 username/email            | ALREADY_EXISTS    | "User already exists"             |
| **1003**  | user_id 미존재                   | NOT_FOUND         | "User not found"                  |
| **1004**  | 필요 권한 없음 (Admin Only API)  | PERMISSION_DENIED | "No permission to access resource"|
| **2001**  | OAuth 인증 실패(플랫폼 응답 문제)| FAILED_PRECONDITION | "Failed to exchange OAuth code"  |
| **3001**  | DB 장애                          | INTERNAL          | "Database error occurred"         |

서버가 에러를 반환할 때 `status.Status` + `ErrorInfo`를 함께 담아 전송할 수 있습니다.

---

## 4. 에러 처리 흐름

### 4.1 서버 측

1. 비즈니스 로직에서 문제 발생 시, 적절한 gRPC status code를 선정.
2. 필요하면 `ErrorInfo` 메시지를 생성(내부 code, message, details) 후 `status.WithDetails()`로 포장.
3. gRPC로 반환 → API Gateway가 수신.

예: Go pseudo-code

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/anypb"
    authv1 "github.com/immersiverse/auth-service/proto/auth/v1"
)

func (s *AuthServiceServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
    // ... logic ...
    if invalidUserOrPass {
        eInfo := &authv1.ErrorInfo{
            Code: 1001,
            Message: "Invalid credentials",
        }
        st := status.New(codes.Unauthenticated, "Unauthorized")
        stWithDetails, _ := st.WithDetails(eInfo)
        return nil, stWithDetails.Err()
    }

    // success
    return &authv1.LoginResponse{ ... }, nil
}
```

### 4.2 API Gateway 측

- gRPC error가 도착하면, `status.Code(err)`를 확인 → HTTP status code 변환(`401`, `404`, `409`, `500` 등).
- 필요하면 `ErrorInfo`를 JSON 변환하여 Response Body에 담아 사용자에게 전달.
- 예: `UNAUTHENTICATED` → HTTP `401 Unauthorized`.
- 로그에 `ErrorInfo.code`, `ErrorInfo.message` 기록.

### 4.3 클라이언트 측

- 클라이언트(웹/모바일 등)는 HTTP status code, JSON body(에러 세부 정보)를 수신하여 사용자에게 적절히 메시지 표시( "비밀번호가 틀렸습니다", "이미 가입된 이메일입니다" 등).

---

## 5. 예시 시나리오별 에러

1. **로그인 - 비밀번호 틀림**  
   - gRPC: `UNAUTHENTICATED (16)`  
   - ErrorInfo: `code=1001, message="Invalid credentials"`  
   - API Gateway: HTTP `401`

2. **회원가입 - 이미 존재하는 username**  
   - gRPC: `ALREADY_EXISTS (6)`  
   - ErrorInfo: `code=1002, message="User already exists"`  
   - API Gateway: HTTP `409 Conflict`

3. **DB 장애**  
   - gRPC: `INTERNAL (13)`  
   - ErrorInfo: `code=3001, message="Database error occurred"`  
   - API Gateway: HTTP `500 Internal Server Error`

4. **OAuth 인증 실패**  
   - gRPC: `FAILED_PRECONDITION (9)`  
   - ErrorInfo: `code=2001, message="Failed to exchange OAuth code"`  
   - API Gateway: HTTP `400 Bad Request`

---

## 6. 확장 및 버전 관리

- 새 에러 유형이 추가될 때:
  - `ErrorInfo` `code` 목록에 새 번호 할당.  
  - `proto`에서 해당 부분(또는 Docs) 업데이트.  
- 기존 에러 삭제 시:
  - 호환성 고려해 `code`를 **reserved** 처리 or deprecated.

---

## 7. 참고사항

- gRPC 인터셉터(서버 측)에서 **공통 에러 처리** 가능 (e.g. `Panic` → `INTERNAL`).
- 로깅 시에는 `ErrorInfo.details` 필드를 활용해 IP, Request ID 등 부가 정보 기록.
- 클라이언트 측이 **다국어 지원**(Localize) 할 경우, `ErrorInfo.message` 대신 에러 코드를 키로 사용해 번역 메시지를 매핑할 수도 있음.

---

## 8. 결론

**에러 코드 및 처리**는 **Authentication Service**의 API 품질과 사용자 경험에 직결됩니다. 위 표준을 따름으로써,

- **서버**는 특정 상황에 맞는 gRPC status code + `ErrorInfo`를 반환,  
- **API Gateway**는 이를 HTTP status code로 매핑,  
- **클라이언트**는 일관성 있는 방식으로 에러를 해석,