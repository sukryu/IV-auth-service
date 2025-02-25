# Protocol Buffers 정의 가이드 (Auth Service)

본 문서는 **ImmersiVerse Authentication Service**에서 사용되는 **Protocol Buffers** 정의 파일(`.proto`)에 대한 작성 규칙과 구조를 안내합니다. gRPC 서비스의 메서드, 메시지 스키마, 버저닝 전략 등을 명확히 정의하여, 팀원 간 협업과 안정적인 API 유지보수를 지원합니다.

---

## 1. 개요

- **언어/버전**: `proto3` 문법을 사용.
- **Package Naming**: `auth.v1`  
  - 패키지 구조: `proto/auth/v1/` 디렉토리에 관련 `.proto` 파일 배치.
  - Go code generation 시 `go_package` 옵션에 `github.com/immersiverse/auth-service/proto/auth/v1` 형식을 매핑.
- **서비스 정의**: 주요 gRPC 서비스(`AuthService`, `UserService` 등)와 관련 메시지들을 이 디렉토리에 작성.  
- **의존성**: 공유 메시지나 공통 `.proto`는 상위 디렉토리(`proto/common`)나 다른 repo에 위치할 수 있음.

---

## 2. 디렉토리 구조

```
proto/
  └── auth/
      └── v1/
          ├── auth_service.proto
          ├── user_service.proto
          ├── platform_service.proto
          ├── error_codes.proto
          └── README.md  # 본 파일
```

1. **`auth_service.proto`**: 로그인, 로그아웃, 토큰 갱신 등 인증 관련 gRPC 메서드와 메시지  
2. **`user_service.proto`**: 사용자 생성, 조회, 업데이트 등 사용자 관리 관련 gRPC 메서드와 메시지  
3. **`platform_service.proto`**: 외부 스트리밍 플랫폼 연동(Connect/Disconnect 등) 관련 API  
4. **`error_codes.proto`**: 공통 에러 코드 정의 및 확장 (권장)  
5. **`README.md`**: 프로토콜 정의 가이드 (본 문서)

(파일 구분은 예시이며, 실제 필요에 따라 통합·분리 가능)

---

## 3. 코드 예시

### 3.1 AuthService - `auth_service.proto`

```proto
syntax = "proto3";

package auth.v1;

option go_package = "github.com/immersiverse/auth-service/proto/auth/v1;authv1";

import "google/protobuf/timestamp.proto"; // 예: 시간 필드가 필요할 때

service AuthService {
  // 사용자 로그인
  rpc Login(LoginRequest) returns (LoginResponse);

  // 사용자 로그아웃
  rpc Logout(LogoutRequest) returns (LogoutResponse);

  // 액세스 토큰 갱신
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);

  // 토큰 유효성 검증
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
}

// 메시지 정의
message LoginRequest {
  string username = 1;
  string password = 2;
}

message LoginResponse {
  string access_token = 1;
  string refresh_token = 2;
  google.protobuf.Timestamp expires_at = 3;
  // 에러 발생 시 별도 ErrorInfo 메시지를 사용하는 방안도 있음
}

message LogoutRequest {
  string access_token = 1;
}

message LogoutResponse {
  bool success = 1;
}

message RefreshTokenRequest {
  string refresh_token = 1;
}

message RefreshTokenResponse {
  string access_token = 1;
  string refresh_token = 2;
  google.protobuf.Timestamp expires_at = 3;
}

message ValidateTokenRequest {
  string access_token = 1;
}

message ValidateTokenResponse {
  bool valid = 1;
  string user_id = 2; // 유효 시 사용자 ID
  string role = 3;    // 사용자 역할 등
}
```

### 3.2 UserService - `user_service.proto`

```proto
syntax = "proto3";

package auth.v1;

option go_package = "github.com/immersiverse/auth-service/proto/auth/v1;authv1";

import "google/protobuf/timestamp.proto";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (UserResponse);
  rpc GetUser(GetUserRequest) returns (UserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
}

message CreateUserRequest {
  string username = 1;
  string email = 2;
  string password = 3;
}

message UserResponse {
  User user = 1;
}

message User {
  string id = 1;
  string username = 2;
  string email = 3;
  google.protobuf.Timestamp created_at = 4;
  google.protobuf.Timestamp updated_at = 5;
  // etc...
}

message GetUserRequest {
  string user_id = 1;
}

message UpdateUserRequest {
  string user_id = 1;
  string email = 2;
  string password = 3; // optional new password
}

message DeleteUserRequest {
  string user_id = 1;
}

message DeleteUserResponse {
  bool success = 1;
}
```

### 3.3 PlatformAccountService - `platform_service.proto`

```proto
syntax = "proto3";

package auth.v1;

option go_package = "github.com/immersiverse/auth-service/proto/auth/v1;authv1";

service PlatformAccountService {
  rpc ConnectPlatformAccount(ConnectPlatformAccountRequest) returns (PlatformAccountResponse);
  rpc DisconnectPlatformAccount(DisconnectPlatformAccountRequest) returns (PlatformAccountResponse);
  rpc GetPlatformAccounts(GetPlatformAccountsRequest) returns (GetPlatformAccountsResponse);
}

message ConnectPlatformAccountRequest {
  string user_id = 1;
  string platform = 2; // e.g. "TWITCH", "YOUTUBE", ...
  string auth_code = 3;
}

message PlatformAccountResponse {
  bool success = 1;
  PlatformAccount account = 2; // optional
}

message PlatformAccount {
  string id = 1;
  string user_id = 2;
  string platform = 3;
  string platform_username = 4;
  // etc...
}

message DisconnectPlatformAccountRequest {
  string user_id = 1;
  string platform_id = 2;
}

message GetPlatformAccountsRequest {
  string user_id = 1;
}

message GetPlatformAccountsResponse {
  repeated PlatformAccount accounts = 1;
}
```

---

## 4. 네이밍 컨벤션

1. **패키지명**: `auth.v1` 식으로 **도메인.버전** 구조를 사용.  
2. **서비스명**: `AuthService`, `UserService`, `PlatformAccountService` 등 **PascalCase**.  
3. **메시지/필드명**:
   - **PascalCase**: `LoginRequest`, `CreateUserRequest`
   - **snake_case**보다 Go 스타일과의 매핑을 고려해 **LowerCamelCase**(proto3 default) 사용하는 경우도 있음.  
4. **rpc 메서드명**: 동사+명사 형태(예: `Login`, `GetUser`, `RefreshToken`).

---

## 5. 필드 번호 규칙

- **필드 번호 1~15**: 가장 자주 사용하는 필드(한 자리 serialization)  
- **중요 필드**(ex. user_id, email 등)는 **낮은 번호** 우선 할당  
- 새 필드 추가 시 **기존 번호** 재사용 금지, 중복 금지  
- **필드 삭제 시**: proto 상 삭제 대신 deprecate 처리 가능; 호환성 이슈 주의

---

## 6. 버저닝 전략

- **proto3** 문법 사용  
- `v1`, `v2` 등 패키지 레벨 버저닝으로 호환성 관리  
- 메이저 변경(하위 호환 불가 시) → `auth.v2` 디렉토리로 분리  
- minor/patch 변경(필드 추가 등 하위 호환) → `auth.v1` 내에서 필드 추가 (번호 충돌 주의)

---

## 7. 코드 생성(Compile) 및 Makefile

### 7.1 protoc 명령 예시

```bash
protoc -I . \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/auth/v1/*.proto
```

### 7.2 Makefile 예시

```Makefile
PROTO_SRC = proto/auth/v1
PROTO_TARGET = $(PROTO_SRC)/*.proto

proto:
	@protoc -I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_TARGET)
	@echo "Protocol Buffers compiled successfully!"

.PHONY: proto
```

- 프로젝트 루트에서 `make proto` 실행 시 `.pb.go` 파일이 자동 생성됨.

---

## 8. 공통/에러 메시지 정의 (옵션)

- **`error_codes.proto`** 예시:
  ```proto
  syntax = "proto3";

  package auth.v1;

  option go_package = "github.com/immersiverse/auth-service/proto/auth/v1;authv1";

  message ErrorInfo {
    int32 code = 1; // custom error code
    string message = 2;
    map<string, string> details = 3; // extra metadata
  }
  ```
- 다른 `.proto`에서 `ErrorInfo`를 임포트해 반환 메시지에 포함할 수 있음.

---

## 9. 모범 사례 및 주의사항

- **역호환성**: 필드 삭제/재번호화 금지, `reserved` 키워드로 예약
- **Message 확장**: 새 필드는 **새 번호**로 추가, 기존 필드는 유지  
- **정적 검사**: Buf/Prototool 등 lint 도구로 문법/스타일 검사 권장  
- **문서화**: proto 내 `//` 주석으로 각 필드/메서드 설명  
- **테스트**: 자동화 빌드 파이프라인에서 protoc 컴파일 확인, Swagger-UI나 Postman + gRPC plugin으로 API 테스트도 가능

---

## 10. 결론

**`proto/auth/v1/`** 디렉토리는 Auth Service의 모든 gRPC 인터페이스와 메시지를 관리하는 핵심 위치입니다.  
- **폴더 구조**: 기능별 `.proto` 분할 (auth_service.proto, user_service.proto 등).  
- **버저닝**: `v1` 패키지를 사용, 추후 하위 호환 깨질 시 `v2`로 마이그레이션.  
- **코드 생성**: Makefile 사용으로 팀원 간 일관성 유지.  
- **Naming/필드 규칙**: 문서화된 컨벤션 준수.  