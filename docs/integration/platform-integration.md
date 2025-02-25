# 외부 플랫폼 통합 가이드

본 문서는 **ImmersiVerse Authentication Service**가 외부 스트리밍 플랫폼(예: **Twitch, YouTube, Facebook, Afreeca** 등)과 **OAuth 2.0** 연동을 통해 계정을 연결하는 과정을 상세히 설명합니다. 이를 통해 사용자들은 자신의 외부 플랫폼 계정을 ImmersiVerse와 간단히 연결하여, 채널/스트리밍 메타데이터를 통합하거나 추가 기능을 활용할 수 있습니다.

---

## 1. 개요

- **플랫폼 계정 연결**: Auth Service는 외부 플랫폼의 OAuth 인증 과정을 대행.  
- **주요 목적**: 
  1. 사용자에게 한 번의 인증으로 ImmersiVerse + 외부 플랫폼 기능을 동시 이용할 수 있도록.  
  2. Access/Refresh Token을 안전하게 저장하고, 필요 시 다시 플랫폼 API 호출.  
- **구현 흐름**: 
  - Client 측(웹/앱) → 외부 플랫폼 OAuth → Auth Service가 `ConnectPlatformAccount` 처리 → DB `platform_accounts` 테이블에 저장.

---

## 2. 지원 플랫폼 및 OAuth 버전

1. **Twitch**: OAuth 2.0 Authorization Code + PKCE (추천)  
2. **YouTube**(Google): OAuth 2.0 Authorization Code  
3. **Facebook**: OAuth 2.0 Authorization Code  
4. **Afreeca**: 전용 OAuth or API Key(프로젝트 상황에 따라)  

### 2.1 공통 프로토콜: OAuth 2.0

- **PKCE**(Proof Key for Code Exchange) → 권장: 모바일/SPA 앱에서 Authorization Code 취약점 방지  
- **Redirect URI**: Must be whitelisted in platform dev console  
- **Scope**: 최소 범위만 요청(`chat:read`, `channel:edit` 등)

---

## 3. OAuth 흐름 요약

### 3.1 시나리오: Twitch 계정 연결

1. **사용자**: UI에서 "Connect Twitch" 버튼 → Twitch OAuth 페이지로 리다이렉트  
2. **사용자**가 Twitch 로그인 및 권한 승인 → Authorization Code 획득  
3. **클라이언트** → `POST /auth/platform/connect` with `{ user_id, platform="TWITCH", auth_code=... }`  
4. **Auth Service**: 
   - Twitch OAuth Token Endpoint 호출 (code 교환)  
   - `access_token`, `refresh_token`, `expires_in` 획득  
   - DB `platform_accounts` Insert  
   - Kafka `platform_connected` 이벤트 발행(선택)
5. **결과**: 플랫폼 계정 연결이 완료되어, ImmersiVerse가 필요 시 Twitch API 호출할 수 있음.

---

## 4. API 호출: `ConnectPlatformAccount`

### 4.1 Request

```json
{
  "user_id": "uuid-1234",
  "platform": "TWITCH",
  "auth_code": "ABCDEF..."
}
```
- **platform**: "TWITCH", "YOUTUBE", "FACEBOOK", "AFREECA" 등  
- **auth_code**: 외부 OAuth code

### 4.2 내부 로직

1. **PlatformIntegrationDomainService**: 해당 `platform`의 Client ID, Secret, Redirect URI 등 조회  
2. **HTTP POST** to platform's token endpoint (PKCE or client_secret)  
3. Parse JSON response: `{ access_token, refresh_token, expires_in, scope... }`  
4. DB `INSERT INTO platform_accounts (...)`  
5. Return `PlatformAccountResponse` with `success=true`, connected account info

### 4.3 Response

```json
{
  "success": true,
  "account": {
    "id": "pacc-9876",
    "user_id": "uuid-1234",
    "platform": "TWITCH",
    "platform_username": "TwitchUser123",
    "access_token": "...",
    "refresh_token": "...",
    "token_expires_at": "2025-06-01T10:00:00Z"
  }
}
```

---

## 5. 플랫폼별 설정

### 5.1 Twitch Developer Settings

- **Client ID**, **Client Secret**: 저장 (Vault 등)  
- **Redirect URI**: `https://yourdomain.com/auth/twitch/callback` or Gateway route  
- **Scope**: e.g. `user:read:email`, `moderation:read` (필요 기능만)

### 5.2 YouTube (Google Cloud Console)

- **OAuth consent screen** + **Credentials**  
- **Client ID**, **Secret**  
- **Scope**: `'profile', 'email', 'youtube.readonly'` etc.

### 5.3 Facebook

- **App** → **Developer Portal**  
- **OAuth Redirect URI** 등록  
- Permissions(‘pages_manage_metadata’, etc.)

### 5.4 Afreeca

- 별도 API docs. Possibly OAuth or a custom SSO approach.

---

## 6. 갱신(Refresh) & 만료

- 각 플랫폼의 Access Token 만료(1~2시간 등) 시 **Auth Service**가 internally refresh token → new access token  
- 저장: `platform_accounts.refresh_token`, `token_expires_at`  
- 주기적 CRON job or on-demand refresh if 401 from platform.

---

## 7. 보안 이슈 & 대응

1. **Client Secret**: 절대 소스코드에 노출 금지, Vault or environment variables  
2. **auth_code**: 단발성 → 짧은 만료. PKCE로 intercept 방지  
3. **Stored tokens**(platform `access_token`/`refresh_token`): Encrypt or at least store in DB with caution, restrict DB access  
4. **Scope**: 최소 권한(Principle of Least Privilege)

---

## 8. 모니터링 & 로깅

- **메트릭**: `platform_connected_total`, `platform_refresh_total`, failure rates  
- **로그**: connection success/fail → `platform_integration.log` or DB `audit_logs`  
- **Alert**: refresh token failures escalated → Slack/E-mail

---

## 9. 예시 Disconnect Flow

`POST /auth/platform/disconnect` → Auth Service: remove DB entry from `platform_accounts`, optionally call platform API to revoke token. Kafka `platform_disconnected` event.  
- UI conveys “Account unlinked” success.

---

## 10. 결론

**외부 플랫폼 통합**은 ImmersiVerse 사용자가 Twitch, YouTube 등과 연동할 수 있게 해주며, **Auth Service**가 OAuth 코드 교환과 토큰 관리(저장/갱신)를 전담:

1. **ConnectPlatformAccount**: Auth Code 교환 → `platform_accounts` Insert  
2. **Refresh**: 자동/수동으로 Access Token 갱신  
3. **Disconnect**: DB Remove + optional platform revoke  
4. **보안**: Client Secret 보호, PKCE 사용, DB 암호화(민감 토큰)

이 가이드를 따라 구현하면, 여러 외부 플랫폼 계정도 일관된 방식으로 안전하게 연결할 수 있습니다.