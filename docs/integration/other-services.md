# 다른 마이크로서비스와의 통합 가이드

이 문서는 **ImmersiVerse Authentication Service**가 플랫폼의 **다른 마이크로서비스**(Chat, Subscription, etc.)와 어떻게 통합하여 협업하는지 정리합니다. 주로 **gRPC 인터페이스**, **토큰 검증 흐름**, **이벤트(카프카) 발행/구독** 등 핵심 통합 포인트를 다룹니다.

---

## 1. 개요

- **마이크로서비스 아키텍처**: 각 서비스(예: Chat, TTS, Subscription)는 독립 배포 가능
- **Auth Service** 역할:
  - 사용자 인증(로그인/로그아웃, JWT 발급)
  - 사용자 데이터(기본 프로필, 계정 상태) 관리
  - 이벤트 발행(로그인 성공/실패, 사용자 생성)
- **타 서비스**(예: Chat Service, Subscription Service)는 **Auth**에 의존(토큰 검증 등)

---

## 2. 통합 시나리오

### 2.1 토큰 검증(ValidateToken)

1. **Chat Service**(예)에서 유저 메시지 처리 전, 토큰 유효성 확인 필요
   - `ValidateToken(accessToken)` gRPC 호출 → Auth Service
   - Auth Service:
     - 서명 확인(우회해서 토큰을 직접 파싱 검증 가능하지만, 중앙 Auth가 블랙리스트 체크 포함)
     - 만료 여부, 블랙리스트 체크
     - 유효 시 user_id, roles 등 반환
   - Chat Service는 **user_id**를 Session/Context에 저장

2. **장점**:
   - 중앙 집중된 인증 로직, Blacklist 등 반영
   - 타 서비스 개발자가 토큰 parsing, key rotation 등에 신경 덜 써도 됨

3. **권장 방식**:
   - Chat Service gRPC interceptor:
     - For each request → call `AuthService.ValidateToken`
   - Or, subscribe to `token_blacklist` events(고급 시나리오) but typically direct call is simpler

### 2.2 이벤트 연동(Kafka)

**Auth Service** 발행 이벤트:
- `auth.events.user_created`, `auth.events.user_updated`
- `auth.events.login_succeeded`, `auth.events.login_failed`
- `auth.events.platform_connected` 등

**타 서비스**(예: Analytics Service, Recommendation Service) 구독:
- **Analytics**: 로그인 통계, user_created → 유저 활성 통계
- **Security**(옵션): `login_failed` 감지 → 공격 패턴? IP 차단?

**Config**:
- Kafka topic: `auth.events.*`
- Schema(Proto/JSON) 구조 문서화
- Consumer groups for each service

---

## 3. gRPC Interface 정리

**Auth Service** 내 핵심 gRPC:
1. `AuthService`:
   - `Login`, `Logout`, `RefreshToken`, `ValidateToken`
2. `UserService`:
   - `CreateUser`, `GetUser`, `UpdateUser`, `DeleteUser`
3. `PlatformAccountService`:
   - `ConnectPlatformAccount`, `DisconnectPlatformAccount`, `GetPlatformAccounts`

**타 서비스**(예: CharacterService, SubscriptionService) 어떤 호출 시나리오?

- **SubscriptionService**: 
  - 사용자 가입 시 `UserCreated` 이벤트 → SubscriptionService가 구독 플랜 set up?
  - 만약 구독 상태 변동 시, AuthService가 user record update(등급이 subscriptionTier로 반영).

- **ChatService**:
  - 로그인 토큰이 필요, ValidateToken or direct parse (but recommended ValidateToken).
  - Chat user profile → `GetUser(user_id)`로 닉네임, 상태 check?

---

## 4. 공통 DB 접근 여부

- **원칙**: **Auth Service**가 user DB 독점 관리(마이크로서비스 분리).  
- **타 서비스**는 DB table `users` 직접 읽지 않는다(강 결합, 보안 리스크).  
- **대안**:
  - gRPC `GetUser`, `ListUsers`(필요 시)  
  - Kafka events to keep local caches updated(이벤트 소싱)

---

## 5. 예시 통합 흐름

### 5.1 사용자가 채팅 서비스 이용

1. **클라이언트**: `POST /chat/room` with `Authorization: Bearer <accessToken>`
2. **Chat Service**:
   - Interceptor calls `AuthService.ValidateToken(accessToken)`
   - OK → user_id 반환
   - Proceed to create Chat room
3. **Chat Service** internally can call `GetUser(user_id)` to get username, status etc. if needed

### 5.2 사용자 가입 → 구독 서비스 알림

1. Auth Service: `CreateUser` → DB insert user row  
2. `user_created` 이벤트(`auth.events.user_created`) to Kafka  
3. **Subscription Service** consumes → `NewUserHandler`: sets default subscription plan?  
4. Logs or user state updated accordingly

### 5.3 Subscription Upgrade → Auth user record update

1. **Subscription Service**: payment success → plan upgrade  
2. Could call `UserService.UpdateUser( { user_id, subscriptionTier="PREMIUM" } )`  
3. Auth DB updated → event `user_updated` if needed

---

## 6. 보안 고려

1. **mTLS**: Service-to-Service gRPC + TLS client cert  
2. **ACL**: Only specific cluster IP or service accounts can call `AuthService`  
3. **Rate limiting**: Each service usage? Possibly handle at Service Mesh or Gateway  
4. **Minimize direct calls**: Some services can just parse JWT. But for up-to-date blacklist check, call `ValidateToken`.

---

## 7. 운영/배포

- **Service Discovery**: e.g. Kubernetes service name `auth-service:50051`
- **Versioning**: 
  - If Auth Service gRPC changes, other services must update stubs.  
  - Keep stable `v1` if possible.

- **CI/CD**:
  - Integration tests: e.g. Chat Service pipeline can spin up local Auth Service mock or real container.

---

## 8. Troubleshooting

- **401** / `UNAUTHENTICATED` in other services:
  - Possibly token expired or blacklisted.  
  - Check logs in Auth Service.  
- **Event not consumed**:
  - Kafka topic mismatch, or consumer group config error.  
- **Performance**:
  - If many requests always call `ValidateToken`, might cache or parse locally. But then won't get real-time blacklist updates unless calling or event-based approach.

---

## 9. 결론

**Auth Service**와 **다른 마이크로서비스** 간 통합은 크게 **gRPC 호출**(토큰 검증, user info)과 **Kafka 이벤트**(로그인, user updates)로 구성됩니다. 

- **gRPC**: 
  - `ValidateToken`, `GetUser` 등으로 인증/계정 상태 공유  
- **Kafka**: 
  - `auth.events.*`로 user lifecycle events 전달, 타 서비스가 구독
- **DB 공유 금지**: 원칙적으로 Auth DB는 Auth만 직접 접근
- **보안**: mTLS, ACL, Service Mesh 가능

이를 준수하면, 각 마이크로서비스가 Auth 기능을 안정적으로 활용하며, 인증·계정 상태 변화를 실시간 이벤트 기반으로 동기화하여 전체 ImmersiVerse 플랫폼의 유기적 작동을 보장할 수 있습니다.