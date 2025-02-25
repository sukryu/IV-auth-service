# 토큰 관리 전략

본 문서는 **ImmersiVerse Authentication Service**가 사용하는 **JWT 토큰**(Access Token, Refresh Token)의 생성·검증·만료·블랙리스트 정책을 종합적으로 설명합니다.

---

## 1. 개요

- **토큰 형식**: JWT (JSON Web Token)
  - **서명 알고리즘**: RS256 (비대칭키 사용 권장)
  - **Payload 클레임**: `sub`(user ID), `exp`(만료 시각), `jti`(토큰 ID), `iat`(발급 시각) 등
- **토큰 종류**:
  1. **액세스 토큰(Access Token)**  
     - 수명 짧음 (15분~1시간 내)  
     - API 호출 시 `Authorization: Bearer <token>` 헤더로 전송  
  2. **리프레시 토큰(Refresh Token)**  
     - 수명 길음 (7일~14일 등)  
     - 만료 전이라도 무효화 가능(로그아웃 등)  
     - 보안 높은 저장소(예: HTTPOnly 쿠키, secure storage) 사용

**목표**:
1. **무상태** 인증 → 서버 간 세션 공유 필요 없음  
2. **단기 만료** → 유출 시 피해 축소  
3. **블랙리스트**로 무효화 가능

---

## 2. Access Token 발급

### 2.1 로그인 시점

1. 사용자 `username/password` 검증 (DB 확인, Argon2id 해싱 비교)  
2. 검증 성공 → Access Token, Refresh Token 생성  
3. Access Token:
   - `exp`: 현재 시각 + 15분 (예)  
   - `sub`: user ID  
   - `jti`: UUID 등 고유 ID  
   - `scope` or roles(옵션)  
4. RS256으로 서명 → 클라이언트에 반환

### 2.2 Claims 예시

```json
{
  "sub": "uuid-1234",
  "iat": 1687080000,
  "exp": 1687080900,
  "jti": "token-uuid-5678",
  "scope": "user",
  "roles": ["USER"]
}
```

---

## 3. Refresh Token 관리

### 3.1 발급 시점

- **Login** 시 함께 발급:
  - 만료: 7~14일 (예시)
  - 별도의 `jti` 
  - 가능하면 별도 RSA 키 or 동일

### 3.2 갱신 시나리오

- Access Token 만료/직전 → 클라이언트가 `refresh_token`으로 새 Access/Refresh 요청
- Auth Service:
  1. Refresh Token 유효성 (서명, `exp`, 블랙리스트)  
  2. 새 Access Token, Refresh Token 생성  
  3. 이전 refresh token을 블랙리스트 처리(선택적) or allow usage (정책 결정)

### 3.3 보안 권장 사항

- Refresh Token은 **HTTPOnly 쿠키** 등 안전한 저장
- 유출 시 장기 토큰 유효 → MFA(2FA)나 IP/디바이스 검사로 위험 완화

---

## 4. Token Validation (유효성 검증)

1. **서명 검증**: RS256, 서버는 public key로 signature verify  
2. **만료 검사**: `exp < now()`면 401 Unauthorized  
3. **블랙리스트 체크**:
   - `jti`가 `token_blacklist`(DB or Redis)에 존재하면 무효
4. **Role/Scope**(옵션): 요청 리소스 접근 권한 확인
5. **토큰 손상/파싱 실패** 시 → `UNAUTHENTICATED`(gRPC), `401 Unauthorized`(REST)

### 4.1 인터셉터/미들웨어

- gRPC: Auth Interceptor
- REST Gateway: JWT middleware or internal call to `AuthService.ValidateToken`

---

## 5. 토큰 무효화(Logout/Invalidate)

### 5.1 로그아웃

- 사용자 `POST /auth/logout` + Access Token
- Auth Service:
  1. 추출한 `jti` → `token_blacklist` 테이블(또는 Redis set) 등록
  2. Refresh Token도 동일한 `jti`(혹은 mapping)로 블랙리스트(옵션)
  3. `Logout` 이벤트 발행(Kafka) for security log

### 5.2 블랙리스트 구조

- **DB**: 
  - `token_blacklist(token_id, user_id, expires_at, reason, blacklisted_at)`
  - 만료시각(`expires_at`) 이후 자동 clean-up
- **Redis**(병행):
  - `SADD blacklisted_jtis jti`
  - 만료 시 TTL 만료 or cron job

### 5.3 트리거 사례

- 비정상 로그인 감지 → 서버 측에서 강제 무효화  
- 사용자 비번 변경 → 기존 토큰 전부 무효화

---

## 6. Keys & Rotation

### 6.1 RS256 키

- **Private Key**: PEM 파일, Vault/HSM에 안전 저장
- **Public Key**: 배포되어 서명 검증용
- **경로**: e.g. `config/jwt-private.pem`, `config/jwt-public.pem`

### 6.2 키 순환(Key Rotation)

- 정기(6~12개월)로 새 키 생성
- 발급 토큰은 새로운 kid(header) 사용, 구 키는 `kid=old`로 검증
- 과도기(최대 Access Token 만료시간) 후 구 키 폐기

---

## 7. Token 형식 구체화

### 7.1 JWT 헤더

```json
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "key1"
}
```
- `kid` → 키 식별자

### 7.2 JWT 페이로드

```json
{
  "sub": "uuid-1234",
  "exp": 1687080900,
  "iat": 1687080000,
  "jti": "abcd-efgh-1234-5678",
  "roles": ["USER","STREAMER"],
  ...
}
```

### 7.3 Example

- Access Token: `header.payload.signature` (Base64url)
- Refresh Token: 유사 구조, 장기 만료

---

## 8. 쿠키/헤더 사용

1. **Access Token** typically in `Authorization: Bearer <token>`  
2. **Refresh Token** in:
   - Secure HTTPOnly cookie → `Set-Cookie: refresh_token=...; HttpOnly; Secure;`  
   - or store in local storage(less secure)
3. CSRF 보호(REST apps):
   - Double Submit Cookie or CSRF token if using cookies for Access Token

---

## 9. 모니터링 & 로깅

- **Token generation count**: Prometheus metric, e.g. `auth_tokens_issued_total`
- **Invalid token error**: `auth_token_errors_total`
- **JWT decode failure** logs
- **JWT latencies**(if relevant)

---

## 10. 권장 정책 요약

1. **Access Token TTL**: 15분  
2. **Refresh Token TTL**: 7~14일  
3. **Key Rotation**: 6~12개월 주기  
4. **Logout**: Access + Refresh token jti를 블랙리스트  
5. **Validation**: signature + exp + jti blacklist check

---

## 11. 예시 로그인-토큰 흐름 요약

1. **Login**: `username/password` → Auth Service → DB check → success → Generate tokens  
2. **Use Access Token** for normal requests  
3. **Refresh** when Access Token near expiry → `POST /auth/refresh` with Refresh Token  
4. **Logout** → Add jti to blacklist  
5. **Check**: Each request verifies signature + expiry + not blacklisted

---

## 12. 결론

**Token 관리 전략**은 **Authentication Service**의 핵심:

1. **JWT**(RS256)로 무상태 인증, Access/Refresh 분리  
2. **단기 Access Token**, 장기 Refresh Token  
3. **블랙리스트**(로그아웃, 보안 이슈)  
4. **정기 키 롤링**으로 보안 향상  

이로써 높은 성능과 유연성을 제공하면서도, 유출·탈취 시나리오에 대응해 안전성을 유지할 수 있습니다. 필요 시 2FA/MFA, OAuth SSO 등과 결합해 보안 강화를 고려하세요.