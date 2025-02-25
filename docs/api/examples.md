# API 호출 예제

본 문서는 **ImmersiVerse Authentication Service**의 **API 호출 예시**를 제공합니다.  
- **REST 예시**: 외부 클라이언트(웹/앱)가 API Gateway를 통해 HTTP/JSON 형태로 요청.  
- **gRPC 예시**: 내부 마이크로서비스 또는 CLI 툴(예: `grpcurl`, BloomRPC, etc.)을 통한 직접 호출.  

이를 통해 개발자·QA가 실제 요청/응답 시나리오를 손쉽게 파악하고 테스트할 수 있습니다.

---

## 1. 개요

- **Auth Service** 핵심 API:
  - 회원가입 (CreateUser)
  - 로그인 (Login)
  - 로그아웃 (Logout)
  - 토큰 갱신 (RefreshToken)
  - 사용자 조회/수정/삭제 (GetUser, UpdateUser, DeleteUser)
  - 플랫폼 계정 연결 (ConnectPlatformAccount 등)

- **API Gateway**는 REST Endpoint를 제공하고 내부적으로 gRPC로 Auth Service를 호출.  
- **직접 gRPC**를 사용하려면 `.proto` 파일 기반으로 gRPC 툴을 사용하거나, Go/Python/Node 등에서 gRPC 클라이언트를 생성해 호출.

---

## 2. REST 예시 (through API Gateway)

### 2.1 회원가입 (Signup)

- **Endpoint**: `POST /auth/signup`
- **Request Body**:

```json
{
  "username": "myUser",
  "email": "myuser@example.com",
  "password": "mySecretP@ss"
}
```

- **Response 200 OK** (성공 시):

```json
{
  "user": {
    "id": "uuid-1234-5678",
    "username": "myUser",
    "email": "myuser@example.com",
    "created_at": "2025-01-01T12:34:56Z",
    "updated_at": "2025-01-01T12:34:56Z"
  }
}
```

- **실패 시** 예: 이메일 중복 → `409 Conflict`:

```json
{
  "error": {
    "code": 1002,
    "message": "User already exists"
  }
}
```

### 2.2 로그인 (Login)

- **Endpoint**: `POST /auth/login`
- **Request Body**:

```json
{
  "username": "myUser",
  "password": "mySecretP@ss"
}
```

- **Response 200 OK** (성공 시):

```json
{
  "access_token": "eyJhbGciOi...",
  "refresh_token": "eyJhbGciOi...",
  "expires_at": "2025-05-20T14:30:00Z"
}
```

- **401 Unauthorized** (비밀번호 불일치):

```json
{
  "error": {
    "code": 1001,
    "message": "Invalid credentials"
  }
}
```

### 2.3 로그아웃 (Logout)

- **Endpoint**: `POST /auth/logout`
- **Authorization Header**: `Bearer <access_token>`

```http
POST /auth/logout
Authorization: Bearer eyJhbGciOi...
```

- **Response** `200 OK`:

```json
{ "success": true }
```

### 2.4 토큰 갱신 (RefreshToken)

- **Endpoint**: `POST /auth/refresh`
- **Request Body**:

```json
{
  "refresh_token": "eyJhbGciOi..." 
}
```

- **Response 200 OK**:

```json
{
  "access_token": "newAccessTokenHere",
  "refresh_token": "newRefreshTokenHere",
  "expires_at": "2025-05-21T10:00:00Z"
}
```

- **401** (refresh 토큰이 무효):

```json
{
  "error": {
    "code": 1001,
    "message": "Invalid credentials"
  }
}
```

### 2.5 사용자 조회 (GetUser)

- **Endpoint**: `GET /auth/users/{userId}`
- **Request**:

```http
GET /auth/users/uuid-1234-5678
Authorization: Bearer <valid-access-token>
```

- **Response 200 OK**:

```json
{
  "user": {
    "id": "uuid-1234-5678",
    "username": "myUser",
    "email": "myuser@example.com",
    "created_at": "2025-01-01T12:34:56Z",
    ...
  }
}
```

- **404** (userId 존재 X):

```json
{
  "error": {
    "code": 1003,
    "message": "User not found"
  }
}
```

### 2.6 플랫폼 계정 연결 (ConnectPlatformAccount)

- **Endpoint**: `POST /auth/platform/connect`
- **Request Body**:

```json
{
  "user_id": "uuid-1234-5678",
  "platform": "TWITCH",
  "auth_code": "xxxxxxxx"
}
```

- **Response 200**:

```json
{
  "success": true,
  "account": {
    "id": "pacc-9876",
    "user_id": "uuid-1234-5678",
    "platform": "TWITCH",
    "platform_username": "TwitchUser123",
    ...
  }
}
```

- **실패 시**(OAuth code 만료 등) → `400 Bad Request`, `FailedPrecondition (9)` 변환:

```json
{
  "error": {
    "code": 2001,
    "message": "Failed to exchange OAuth code"
  }
}
```

---

## 3. gRPC 예시 (직접 호출)

이 섹션은 **직접 gRPC**를 사용할 때의 호출 방법 예시를 보여줍니다.  
- **CLI** 툴: `grpcurl`, BloomRPC, evans 등  
- **Generated Client**: Go, Node, Python, etc.

### 3.1 grpcurl 예시

#### 3.1.1 Login

```bash
grpcurl -plaintext \
  -d '{"username":"myUser","password":"mySecretP@ss"}' \
  localhost:50051 auth.v1.AuthService/Login
```

- **응답** 예시:

```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": "2025-05-20T14:30:00Z"
}
```

#### 3.1.2 CreateUser

```bash
grpcurl -plaintext \
  -d '{"username":"newGuy","email":"newguy@example.com","password":"abcd1234"}' \
  localhost:50051 auth.v1.UserService/CreateUser
```

- **응답**:

```json
{
  "user": {
    "id": "uuid-1234-5678",
    "username": "newGuy",
    "email": "newguy@example.com"
  }
}
```

### 3.2 Go Client 예시

```go
package main

import (
    "context"
    "fmt"
    "log"
    "google.golang.org/grpc"
    authv1 "github.com/immersiverse/auth-service/proto/auth/v1"
)

func main() {
    conn, err := grpc.Dial("auth-service:50051", grpc.WithInsecure())
    if err != nil { log.Fatalf("could not connect: %v", err) }
    defer conn.Close()

    client := authv1.NewAuthServiceClient(conn)
    req := &authv1.LoginRequest{
        Username: "myUser",
        Password: "mySecretP@ss",
    }
    resp, err := client.Login(context.Background(), req)
    if err != nil {
        log.Fatalf("Login failed: %v", err)
    }
    fmt.Println("AccessToken:", resp.AccessToken)
}
```

---

## 4. 에러 처리 예시

- **gRPC**: 서버에서 `status.Error(codes.Unauthenticated, "Invalid credentials")` 반환 → `grpcurl`은 JSON 형태로 status code/메시지 출력
- **REST**: Gateway에서 `401 Unauthorized` + `{"error":{"code":1001,"message":"Invalid credentials"}}`

---

## 5. 주의사항 & 모범 사례

1. **토큰 보안**: 
   - Access Token은 15분 등 짧은 만료, Refresh Token은 7일 등 설정.  
   - 보안 민감 데이터(비밀번호, 토큰 등)는 로컬에 평문 저장 금지.
2. **HTTP Header**: REST API 호출 시, `Authorization: Bearer <access_token>` 형식 준수.
3. **Rate Limiting**: 로그인/회원가입 등 민감 API에서 속도 제한 가능.  
4. **TLS/mTLS**: 실제 운영환경에선 gRPC는 mTLS, REST는 HTTPS 적용.  
5. **에러 핸들링**: 응답에서 `error.code`, `error.message`를 활용해 클라이언트 측 분기.

---

## 6. 결론

이 문서의 **API 호출 예제**를 통해,

- **REST(through API Gateway)** 기반 접근: cURL, Postman, etc.  
- **gRPC Direct** 접근: `grpcurl`, BloomRPC, 또는 언어별 generated client.