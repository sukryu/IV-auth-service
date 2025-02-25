# Kafka 이벤트 스트림 규격 및 통합 가이드

본 문서는 **ImmersiVerse Authentication Service**에서 발행하는 **Kafka 이벤트**의 스키마, 토픽 구조, 소비(Consume) 방식 등을 정의합니다. 다른 마이크로서비스(예: Analytics, Recommendation, UserProfile)에서 이 이벤트를 구독해 인증/사용자 관련 정보를 실시간으로 동기화하거나 분석할 수 있습니다.

---

## 1. 개요

- **메시지 브로커**: Apache Kafka (버전 2.x 이상)
- **Auth Service** → **Kafka**: 주요 상태 변경(회원가입, 로그인 성공/실패, 플랫폼 계정 연결 등)에 대해 이벤트 발행
- **토픽 네이밍**: `auth.events.*` (다중 토픽 or 단일 토픽 + eventType 필드)
- **포맷**: JSON(기본 예시) 또는 Protobuf(고급). 여기서는 JSON 예시 중심
- **컨슈머**: AnalyticsService, LoggingService, etc.

---

## 2. 토픽 구조

### 2.1 권장 토픽

1. `auth.events.user`  
   - 사용자 생성/수정/삭제 이벤트
2. `auth.events.login`  
   - 로그인 성공/실패, 로그아웃 등
3. `auth.events.platform`  
   - 플랫폼 계정 연결/해제
4. (선택) `auth.events.security`
   - 의심스러운 로그인, 무차별 대입 감지, 블랙리스트 추가 이벤트 등

각 토픽에는 이벤트 유형(eventType) 필드로 구분 가능.

---

## 3. 메시지 구조(예시: JSON)

```json
{
  "eventType": "UserCreated",
  "timestamp": "2025-06-01T10:15:30Z",
  "payload": {
    "userId": "uuid-1234",
    "username": "newuser",
    "email": "user@example.com",
    "createdAt": "2025-06-01T10:15:30Z"
  }
}
```

### 3.1 공통 필드

- **eventType**: 이벤트 명 (`UserCreated`, `LoginSucceeded` 등)  
- **timestamp**: ISO8601 시각 (UTC)  
- **payload**: 실제 데이터(유저 ID, 플랫폼 ID, etc.)

### 3.2 이벤트별 예시

1. **UserCreated**  
   - topic: `auth.events.user`  
   - payload: `{ userId, username, email, createdAt }`
2. **LoginSucceeded**  
   - topic: `auth.events.login`  
   - payload: `{ userId, ipAddress, userAgent, loginTime }`
3. **PlatformConnected**  
   - topic: `auth.events.platform`  
   - payload: `{ userId, platform, platformUserId, connectedAt }`

---

## 4. 발행 정책

1. **트랜잭션 성공 후**: Auth Service DB Insert/Update 완료 시 이벤트 발행  
2. **Kafka Producer**: 
   - Delivery guarantee: at-least-once (중복 가능성)  
   - 재시도 → idempotent consumer 설계 권장
3. **토픽 파티션**: 
   - userId 해시(`PartitionKey = userId`)로 파티셔닝  
   - 순서(로그인/로그아웃 등) 유지를 위함

### 4.1 이벤트 순서 보장

- Kafka 파티션 단위로 순서가 유지  
- 만약 multi-partition, userId-based partitioning으로 유저별 순서 유지

---

## 5. 소비(Consume) 예시

1. **AnalyticsService**:
   - Subscribes `auth.events.login`, `auth.events.user`  
   - 집계: 로그인 성공률, 일별 신규 유저 등
2. **SecurityService**(선택):
   - Subscribes `auth.events.login`(login_failed)  
   - 이상 징후(브루트포스) 발생 시 Slack 알림
3. **UserProfileService**:
   - Subscribes `auth.events.user` → 캐시/프로필 동기화

### 5.1 Consumer 그룹 config

```yaml
auth-consumer:
  bootstrap.servers: "kafka:9092"
  group.id: "auth-consumer-group"
  topics:
    - "auth.events.user"
    - "auth.events.login"
```

- Offset commit 자동/수동 결정(팀 표준에 따라)

---

## 6. 보안 & 인증

- **Kafka TLS**: enable server/client TLS encryption  
- **SASL**(plain, SCRAM, or GSSAPI) for auth  
- **ACL**: restrict produce to `auth.events.*` only from Auth Service  
- Consumer ACL to read specific topics

---

## 7. 스키마 관리

- **JSON**이 가볍지만, schema evolution(필드 추가)은 consumer가 handle.  
- 대안: **Protobuf** + Schema Registry → more robust.  
- 기존 JSON: eventType + payload, consumer must parse unknown fields gracefully

---

## 8. 에러/재시도 처리

- **Producer**: Auth Service retries if broker not available.  
- **Consumer**: handle duplicates or reorder if at-least-once. E.g. store last event version.  
- **Dead-letter queue**: consumer parse failure → send to `auth.events.dlq`.

---

## 9. 모니터링

- **Prometheus**:
  - `kafka_producer_requests_total`  
  - `kafka_consumer_lag`  
- **Grafana** dashboards: 
  - consumer lag, message throughput, error rates

---

## 10. 예시 사용자 등록 이벤트

```json
{
  "eventType": "UserCreated",
  "timestamp": "2025-07-01T09:00:00Z",
  "payload": {
    "userId": "uuid-1234",
    "username": "alice",
    "email": "alice@example.com",
    "createdAt": "2025-07-01T09:00:00Z"
  }
}
```
- **topic**: `auth.events.user`
- **Partition key**: "uuid-1234" 
- Consumer receives → parse JSON → update local stats or replicate user profile

---

## 11. 운영 가이드

1. **Topic retention**: 7일 or more if needed for replays  
2. **Partition count**: scale with user base  
3. **Schema evolution**: add fields in payload carefully (consumers ignore unknown)  
4. **Replay**: if a service was down, it can re-process from offset

---

## 12. 결론

**Auth Service**의 Kafka 이벤트 스트림은 **로그인·사용자·플랫폼 계정** 등의 실시간 상태 변화를 다른 마이크로서비스와 공유하는 핵심 메커니즘입니다.

- **주요 토픽**: `auth.events.user`, `auth.events.login`, `auth.events.platform`  
- **JSON 메시지 구조**: `eventType`, `timestamp`, `payload`  
- **Partition key**: userId  
- **Consumer**: Analytics, Security, UserProfile 등

이 구조를 준수하면, 확장 시 새로운 서비스가 손쉽게 Auth 이벤트를 구독·활용하여 플랫폼 기능을 풍부하게 만들 수 있습니다.