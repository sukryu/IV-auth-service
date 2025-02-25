# gRPC 서비스 상세 명세

이 문서는 **ImmersiVerse Authentication Service**에 정의된 **gRPC 서비스**들을 자세히 설명합니다. 각 서비스별 RPC 메서드, 요청/응답 메시지, 에러 처리 방식, 예시 호출 등을 안내하여 클라이언트 개발 및 운영팀이 참고할 수 있도록 합니다.

---

## 1. 개요

- **proto3** 문법을 사용하며, `auth.v1` 패키지에 정의되어 있음.  
- 파일 구조:
  - `auth_service.proto`: 로그인/로그아웃/토큰 갱신 등
  - `user_service.proto`: 사용자 생성/조회/수정/삭제
  - `platform_service.proto`: 외부 플랫폼 계정 연동
- 코드 생성 위치: `github.com/immersiverse/auth-service/proto/auth/v1` (Go code)

---

## 2. AuthService

> 정의 파일: [`auth_service.proto`](../../proto/auth/v1/auth_service.proto)

### 2.1 Login

- **메서드 시그니처**:  
  ```proto
  rpc Login(LoginRequest) returns (LoginResponse);
  ```
- **설명**: 사용자명+비밀번호를 검증 후, 성공 시 액세스 토큰/리프레시 토큰을 발급
- **요청 메시지**: `LoginRequest`  
  - `string username` (필수)  
  - `string password` (필수)
- **응답 메시지**: `LoginResponse`
  - `string access_token`  
  - `string refresh_token`  
  - `google.protobuf.Timestamp expires_at` → 액세스 토큰 만료 시각
- **에러 처리**:
  - 비밀번호 불일치 → gRPC `Unauthenticated (16)` + 에러 메시지
  - 내부 DB 오류 → gRPC `Internal (13)`
- **예시 호출** (Go pseudo-code):
  ```go
  req := &authv1.LoginRequest{
      Username: "myUser",
      Password: "mySecret",
  }
  resp, err := client.Login(ctx, req)
  if err != nil {
      // handle error
  }
  fmt.Println("AccessToken:", resp.AccessToken)
  ```

### 2.2 Logout

- **메서드 시그니처**:  
  ```proto
  rpc Logout(LogoutRequest) returns (LogoutResponse);
  ```
- **설명**: 액세스 토큰을 블랙리스트에 등록하여 무효화(만료 전이라도 로그아웃 가능)
- **요청 메시지**: `LogoutRequest`
  - `string access_token`
- **응답 메시지**: `LogoutResponse`
  - `bool success`
- **에러 처리**:
  - 이미 블랙리스트 상태인 토큰 → 응답은 `success=true`로 단순 처리
  - 내부 오류 → `Internal (13)`
- **예시**:
  ```go
  resp, err := client.Logout(ctx, &authv1.LogoutRequest{
      AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9....",
  })
  ```

### 2.3 RefreshToken

- **메서드 시그니처**:
  ```proto
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  ```
- **설명**: 리프레시 토큰의 유효성을 확인한 뒤, 새로운 액세스 토큰/리프레시 토큰을 재발급
- **요청 메시지**: `RefreshTokenRequest`
  - `string refresh_token`
- **응답 메시지**: `RefreshTokenResponse`
  - `string access_token`
  - `string refresh_token`
  - `google.protobuf.Timestamp expires_at`
- **에러 처리**:
  - 리프레시 토큰이 만료/블랙리스트 → `Unauthenticated (16)`
  - 내부 DB/Redis 오류 → `Internal (13)`
- **예시**:
  ```go
  resp, err := client.RefreshToken(ctx, &authv1.RefreshTokenRequest{
      RefreshToken: "some-old-refresh-token",
  })
  // resp.AccessToken, resp.RefreshToken ...
  ```

### 2.4 ValidateToken

- **메서드 시그니처**:
  ```proto
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  ```
- **설명**: 액세스 토큰이 유효한지(서명, 만료, 블랙리스트 등) 검사
- **요청 메시지**: `ValidateTokenRequest`
  - `string access_token`
- **응답 메시지**: `ValidateTokenResponse`
  - `bool valid`
  - `string user_id`
  - `string role` 등
- **에러 처리**:
  - 토큰 무효 → `valid=false`, gRPC는 OK 리턴(상황에 따라 `Unauthenticated (16)`도 가능)
  - 내부 오류 → `Internal (13)`

---

## 3. UserService

> 정의 파일: [`user_service.proto`](../../proto/auth/v1/user_service.proto)

### 3.1 CreateUser

- **메서드 시그니처**:
  ```proto
  rpc CreateUser(CreateUserRequest) returns (UserResponse);
  ```
- **설명**: 새 사용자 등록(회원가입)
- **요청 메시지**: `CreateUserRequest`
  - `string username`
  - `string email`
  - `string password`
- **응답 메시지**: `UserResponse`
  - `User user`
    - `user.id`, `user.username` 등
- **에러 처리**:
  - 중복 username/email → `AlreadyExists (6)`
  - 내부 DB 에러 → `Internal (13)`

### 3.2 GetUser

- **메서드 시그니처**:
  ```proto
  rpc GetUser(GetUserRequest) returns (UserResponse);
  ```
- **설명**: 사용자 상세 정보 조회
- **요청 메시지**: `GetUserRequest`
  - `string user_id`
- **응답 메시지**: `UserResponse`
  - `User user`
- **에러 처리**:
  - 존재하지 않는 userId → `NotFound (5)`
  - DB 에러 → `Internal (13)`

### 3.3 UpdateUser

- **메서드 시그니처**:
  ```proto
  rpc UpdateUser(UpdateUserRequest) returns (UserResponse);
  ```
- **설명**: 사용자 정보 업데이트(이메일 변경, 비밀번호 변경 등)
- **요청 메시지**: `UpdateUserRequest`
  - `string user_id`  
  - `string email` (optional)  
  - `string password` (optional)
- **응답 메시지**: `UserResponse`
  - `User user`
- **에러 처리**:
  - user_id 미존재 → `NotFound (5)`
  - DB 문제 → `Internal (13)`

### 3.4 DeleteUser

- **메서드 시그니처**:
  ```proto
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
  ```
- **설명**: 사용자 정보 삭제(또는 논리적 삭제)
- **요청 메시지**: `DeleteUserRequest`
  - `string user_id`
- **응답 메시지**: `DeleteUserResponse`
  - `bool success`
- **에러 처리**:
  - user_id 미존재 → `NotFound (5)`
  - 내부 에러 → `Internal (13)`

---

## 4. PlatformAccountService

> 정의 파일: [`platform_service.proto`](../../proto/auth/v1/platform_service.proto)

### 4.1 ConnectPlatformAccount

- **메서드**:
  ```proto
  rpc ConnectPlatformAccount(ConnectPlatformAccountRequest) returns (PlatformAccountResponse);
  ```
- **설명**: 외부 스트리밍 플랫폼(Twitch, YouTube, 등) 계정을 OAuth 과정을 통해 연결
- **요청 메시지**: `ConnectPlatformAccountRequest`
  - `string user_id`
  - `string platform` (예: "TWITCH", "YOUTUBE")
  - `string auth_code` (OAuth Authorization Code)
- **응답 메시지**: `PlatformAccountResponse`
  - `bool success`
  - `PlatformAccount account` (연결된 계정 정보)
- **에러 처리**:
  - OAuth 실패 → `PermissionDenied (7)` or `FailedPrecondition (9)`
  - user_id 없음 → `NotFound (5)`

### 4.2 DisconnectPlatformAccount

- **메서드**:
  ```proto
  rpc DisconnectPlatformAccount(DisconnectPlatformAccountRequest) returns (PlatformAccountResponse);
  ```
- **설명**: 연결된 외부 플랫폼 계정 해제
- **요청 메시지**: `DisconnectPlatformAccountRequest`
  - `string user_id`
  - `string platform_id` (DB 내 `PlatformAccount.id`)
- **응답 메시지**: `PlatformAccountResponse`
  - `bool success`
  - `PlatformAccount account`(optional, 해제된 계정 info)
- **에러 처리**:
  - user_id/platform_id 불일치 → `NotFound (5)`

### 4.3 GetPlatformAccounts

- **메서드**:
  ```proto
  rpc GetPlatformAccounts(GetPlatformAccountsRequest) returns (GetPlatformAccountsResponse);
  ```
- **설명**: 특정 사용자가 연동한 플랫폼 계정 목록 조회
- **요청 메시지**: `GetPlatformAccountsRequest`
  - `string user_id`
- **응답 메시지**: `GetPlatformAccountsResponse`
  - `repeated PlatformAccount accounts`
- **에러 처리**:
  - user_id 미존재 시 빈 배열 혹은 `NotFound (5)` 처리

---

## 5. 에러 처리 방식

- **gRPC Status Codes**:
  - `OK (0)`: 성공
  - `InvalidArgument (3)`, `NotFound (5)`, `AlreadyExists (6)`, `PermissionDenied (7)`, `Unauthenticated (16)`, `Internal (13)` 등
- **메시지 레벨 오류**:
  - 부분적으로 `ErrorInfo` 메시지를 사용할 수도 있음 (optional)
  - 예: `LoginResponse` 내 `error_info` 필드

---

## 6. 인증 / 보안

- **통신 암호화**: mTLS(서버<->서버), API Gateway 레벨 TLS
- **JWT**: 클라이언트 또는 내부 서비스가 gRPC 호출 시 인터셉터에서 토큰 검증 가능
- **Role Checking**: `UserService`, `PlatformAccountService` 일부 메서드는 `ADMIN` 권한 필요

---

## 7. 예시 호출 방법

### 7.1 Go 클라이언트

```go
conn, err := grpc.Dial("auth-service:50051", grpc.WithInsecure()) // use TLS in production
if err != nil { ... }
defer conn.Close()

authClient := authv1.NewAuthServiceClient(conn)

resp, err := authClient.Login(context.Background(), &authv1.LoginRequest{
  Username: "myUser",
  Password: "myPass",
})
if err != nil {
  // handle gRPC error (check status.Code(err))
}
fmt.Println("AccessToken:", resp.AccessToken)
```

### 7.2 CLI / BloomRPC

- 설치된 [BloomRPC](https://github.com/uw-labs/bloomrpc) 등으로 `.proto` 가져와, `Login` 메서드 호출
- Request JSON 예시:
  ```json
  {
    "username": "myUser",
    "password": "myPass"
  }
  ```

---

## 8. 버저닝 / 확장

- **`auth.v1`** 패키지가 현재 버전
- Breaking change 발생 시, `auth.v2` 패키지로 새 디렉토리를 생성해 호환성 유지
- 필드 추가 시 **새 번호** 사용, 필드 삭제는 **reserved** 또는 deprecate 처리
- 릴리스 노트에 gRPC 변경 사항 기재

---

## 9. 참고 문서

- [Protocol Buffers 정의 가이드](../../proto/auth/v1/README.md)
- [API Gateway → gRPC 맵핑](../integration/api-gateway.md)
- [에러 코드 및 처리 방법](error-codes.md)  
- [테스트 가이드](../../docs/testing/integration-testing.md)

---

## 10. 결론

**gRPC 서비스**(AuthService, UserService, PlatformAccountService 등)는 Auth Service의 핵심 API를 제공하며, ImmersiVerse 플랫폼의 인증·계정 관리 로직을 효율적으로 노출합니다. 각 메서드는 Request/Response 메시지와 적절한 gRPC status code를 통해 결과를 반환하며, 확장 시에는 버전 관리 원칙을 준수해 호환성을 유지합니다.