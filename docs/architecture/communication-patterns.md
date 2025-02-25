# Authentication Service: 서비스 간 통신 패턴

본 문서는 **ImmersiVerse Authentication Service**가 내부 플랫폼(다른 마이크로서비스) 및 외부 시스템(스트리밍 플랫폼 등)과 어떻게 통신하는지 정리합니다. 주 통신 방식(gRPC, 메시지 큐, API Gateway 연동)과 그에 따른 시나리오, 프로토콜, 보안 사항 등을 설명하여, 개발 및 운영 과정에서 참조할 수 있는 지침을 제공합니다.

---

## 1. 개요

ImmersiVerse 플랫폼은 **마이크로서비스 아키텍처**를 채택하여, **Authentication Service**(이하 'Auth Service')를 비롯한 여러 서비스가 각각 독립적으로 배포·확장됩니다. 각 서비스 간에는 주로 **gRPC**(동기 호출)와 **Kafka**(비동기 이벤트)를 통해 통신합니다. 또한, 외부 API(스트리밍 플랫폼 OAuth 등)와의 연동도 존재합니다.

전체 통신 흐름은 크게 다음과 같이 구분할 수 있습니다:

1. **클라이언트 ↔ API Gateway**: 외부 사용자 요청(REST over HTTP)
2. **API Gateway ↔ Auth Service**: 내부 gRPC 호출
3. **Auth Service ↔ DB/Redis**: 내부 데이터 액세스
4. **Auth Service ↔ Kafka**: 이벤트 발행(및 필요 시 구독)
5. **Auth Service ↔ 외부 플랫폼(API)**: OAuth 인증 및 토큰 교환

---

## 2. 내부 서비스 통신(gRPC)

### 2.1 API Gateway 연동

- **API Gateway**(예: Envoy Proxy)가 **REST**(외부) ↔ **gRPC**(내부)를 변환해 Auth Service와 통신.
- 클라이언트는 HTTP/JSON 형태로 `/auth/...` 엔드포인트를 호출하면, Gateway가 내부적으로 `AuthService.Login`, `AuthService.CreateUser` 등 gRPC 메서드를 호출.
- **장점**:
  - 타입 안전한 Protocol Buffers
  - 양방향 스트리밍, 낮은 지연
  - API Gateway 계층에서 인증/라우팅/모니터링 용이

### 2.2 다른 마이크로서비스와의 Direct gRPC

- 일부 상황에서, Auth Service가 다른 서비스와 직접 gRPC를 주고받을 수 있음 (예: `UserProfileService` 등).
- **주요 패턴**:
  - **Request-Response**: 인증 토큰 검증 요청(ValidateToken)이나 사용자 프로필 조회.
  - **Server Streaming**: 실시간 권한 변경 알림(현재는 Kafka 이벤트 기반으로 대체하는 경우가 많음).

> 실제로는 **API Gateway**를 통한 통합 경로가 일반적이지만, 내부 최적화 목적(Latency)에 따라 Direct gRPC 호출을 고려할 수 있음.

### 2.3 보안 설정

- **mTLS**(mutual TLS)를 통해 Auth Service ↔ 다른 서비스 간 통신을 암호화 및 상호 인증.
- gRPC 레벨에서 **인터셉터**를 사용해 토큰 검증, 로깅, 트레이싱 자동화.

---

## 3. 비동기 이벤트 통신(Kafka)

### 3.1 이벤트 발행(Publish)

- Auth Service는 인증/사용자 관련 상태 변경 시 **Kafka**에 이벤트를 발행:
  - `auth.events.user_created`
  - `auth.events.login_succeeded`
  - `auth.events.platform_connected`
  - 등등 (세부 목록은 [Event Topics] 문서 참조)
- **이벤트 포맷**: Proto 메시지 또는 JSON 직렬화. 일반적으로 Kafka에서 `application/json` / `application/x-protobuf` 방식을 채택.
- **발행 시점**:
  - **트랜잭션 완료 후**(DB Insert 성공 등) → 이벤트 발행
  - **Login** 함수 내에서 성공/실패 결과에 따라 이벤트 선택 발행
  - **ConnectPlatform** 함수 내에서 플랫폼 연동 완료 시 `PlatformConnected` 이벤트 발행

### 3.2 이벤트 구독(Consume)

- 기본적으로 Auth Service가 다른 서비스 이벤트를 구독하는 케이스는 적으나, 다음과 같은 시나리오가 가능:
  - `UserProfileService`나 `AnalyticsService`가 사용자 삭제/정지 등 이벤트를 발행 → Auth Service가 구독해 토큰 블랙리스트 처리를 할 수도 있음
  - 현재는 주로 Auth Service → 타 서비스 일방 발행이 주 시나리오.

> 구독 시에는 Kafka Consumer Group 이름, 오프셋 관리 전략, 재시도/데드레터 큐 등에 대한 정책이 필요함.

### 3.3 보안 설정 및 QoS

- **TLS 암호화**를 적용한 Kafka 클러스터 사용.
- **SASL** 인증, Access Control Lists(ACLs)를 통해 Auth Service 전용 토픽 접근을 제한.
- 재시도(Exponential backoff) 및 데드레터 큐(Dead Letter Queue) 구성으로 안정성 보장.

---

## 4. 외부 플랫폼 통합(OAuth 등)

### 4.1 스트리밍 플랫폼 OAuth

- **Twitch, YouTube, Facebook, Afreeca** 등 플랫폼 계정 연결 시, OAuth 2.0 Authorization Code 흐름을 사용.
- Auth Service 내 `PlatformIntegrationDomainService`가 각 플랫폼의 OAuth 엔드포인트와 통신(HTTP/HTTPS).
- **통신 패턴**:
  - Auth Service → 외부 API : HTTP(S) 요청
  - 외부 API → Auth Service : Access Token, Refresh Token 응답
  - Auth Service DB에 `PlatformAccount`를 저장하고, 필요 시 Kafka 이벤트(`PlatformConnected`) 발행
- **보안**:
  - PKCE(Proof Key for Code Exchange) 등 보안 강화
  - 토큰 갱신 로직(만료 시 재갱신)

### 4.2 알림/이벤트 연동(차후 확장)

- 일부 플랫폼에서는 Webhook/Callback 형태로 이벤트를 제공(스트리머 계정 상태 변경 등).
- 그 경우 Auth Service가 HTTP Endpoint를 개방해서 수신하거나, **새 마이크로서비스**(예: Webhook Service)로 라우팅 후 Auth Service와 gRPC 통신할 수 있음.

---

## 5. 대표 통신 시나리오 정리

### 5.1 사용자 가입

1. **Client** → (REST) → **API Gateway** → (gRPC) → **AuthService.CreateUser**  
2. AuthService가 DB에 User 레코드 생성 후 **Kafka** 이벤트 발행(`UserCreated`)

### 5.2 로그인 / 토큰 검증

1. 로그인 시나리오:
   - **Client** → **Gateway** → **AuthService.Login**  
   - DB 조회, 비밀번호 검증 → 성공 시 Token 발급, `LoginSucceeded` 이벤트 발행
2. 다른 서비스가 Auth 필요 시:
   - gRPC 인터셉터 or REST header에서 JWT → **AuthService.ValidateToken**(Direct or Gateway)
   - Redis를 통한 토큰 블랙리스트 확인

### 5.3 플랫폼 계정 연결

1. **Client**가 플랫폼(OAuth code) → **AuthService.ConnectPlatformAccount**  
2. Auth Service가 **외부 플랫폼** API로 토큰 교환 → DB Insert → `PlatformConnected` 이벤트 발행

---

## 6. 속도 제한 및 Rate Limiting

### 6.1 API Gateway 수준

- Gateway에서 **IP별 요청 제한**, **Burst Limit** 적용
- 429(Too Many Requests) 에러 처리

### 6.2 Auth Service 수준

- **Redis** 활용 Rate Limiter (ex: 로그인 시도 5회/분 초과 시 잠금)
- gRPC 인터셉터에서 호출 횟수 세션 추적 가능

---

## 7. 오류 처리를 위한 패턴

### 7.1 gRPC 리턴 코드

- **OK(0)**: 성공
- **InvalidArgument(3)**: 유효성 검증 실패
- **Unauthenticated(16)**: 비밀번호/토큰 검증 실패
- **Internal(13)**: DB 장애, Kafka 장애 등 내부 에러
- …  
이 코드를 API Gateway에서 HTTP 스테이터스(200, 400, 401, 500 등)로 매핑.

### 7.2 Kafka 장애 / 재시도

- **Producer Retries**: KafkaProducer가 ACK 실패 시 자동 재시도
- **DeadLetterQueue**: Consumer 쪽에서 처리 실패 시 DLQ에 저장(주로 다른 서비스에서 사용)

---

## 8. 분산 트레이싱(Distributed Tracing)

- **OpenTelemetry**를 통해 gRPC 호출 간 **Trace** / **Span**을 전파
- Auth Service 각 함수(로그인, 토큰 검증 등)에 스팬 생성 → Prometheus, Jaeger 등으로 시각화
- Kafka 이벤트 발행 시 헤더에 TraceContext 전파(옵션)

---

## 9. 보안 고려사항

### 9.1 TLS 및 mTLS

- **API Gateway ↔ Auth Service**: TLS / mTLS
- **Auth Service ↔ Kafka**: TLS 암호화, SASL 인증
- **Auth Service ↔ DB**: SSL/TLS 사용

### 9.2 접근 제어

- Auth Service 내 RBAC(역할기반) 적용
- 다른 서비스가 Auth Service 호출 시, 내부 서비스용 인증서(mTLS), 혹은 JWT 인증(서버-서버)이 필요할 수 있음

---

## 10. 결론

- Auth Service는 **동기 gRPC 호출**과 **비동기 이벤트(Kafka)**를 혼합 사용해 **다른 마이크로서비스** 및 **외부 플랫폼**과 통신합니다.
- 주요 지점:
  - **API Gateway**: 외부 REST → 내부 gRPC 변환 지점
  - **Kafka**: 사용자/로그인/플랫폼계정 등의 이벤트 발행(주로 Outbound)
  - **OAuth**: 외부 플랫폼 계정 연결 시 HTTP API 통신
- 이러한 통신 패턴은 **성능**, **확장성**, **보안성**을 만족시키며, 추후 마이크로서비스 증설/변경에도 유연히 대응 가능합니다.