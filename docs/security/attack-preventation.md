# 공격 방어 전략

본 문서는 **ImmersiVerse Authentication Service**가 직면할 수 있는 대표적인 보안 위협(공격 시나리오)과 이를 방어하기 위한 전략을 정리합니다. **SQL Injection, 무차별 대입(Brute Force), 세션 하이재킹, XSS** 등 다양하며, 각 위협에 대해 **예방 기법, 모니터링 방안, 대응 절차** 등을 제시합니다.

---

## 1. 개요

- **언어**: Go  
- **주요 엔드포인트**: 로그인, 회원가입, 토큰 검증, 플랫폼 계정 연동 등  
- **취약점 범위**: 입력 검증, 세션/토큰, DB 접근, 외부 API, 운영환경 구성 등  
- **목표**:
  1. 주요 공격 벡터 식별 및 사전 차단  
  2. 침해 시나리오 발생 시 피해 최소화  
  3. 지속적인 보안 모니터링 및 업데이트

---

## 2. 공격 유형별 대응

### 2.1 SQL Injection

- **위협**: 악의적 쿼리 삽입으로 DB 접근/데이터 누출 가능  
- **방어**:
  1. **ORM/Prepared Statement** 사용 (Go에서는 `database/sql`, `sqlx`, `gorm` 등)  
  2. **문자열 연결 쿼리 금지**  
  3. **정적 쿼리** + 바인딩(`WHERE username = $1`)  
  4. **DB 권한 최소화**: DB 유저는 필요한 DML 권한만 갖도록  
- **모니터링**:  
  - 에러 로그(에러율 급증), `pg_stat_statements` 쿼리 이상 탐지  
  - WAF(Web Application Firewall) 레벨 SQLi 필터

### 2.2 Brute Force / 무차별 대입

- **위협**: 대량의 ID/비밀번호 시도 → 계정 탈취  
- **방어**:
  1. 로그인 실패 횟수 제한 (ex: 5회→10분 잠금)  
  2. Rate limiting (IP별 초당 요청 수 제한)  
  3. CAPTCHA or 2FA (MFA)  
  4. Redis/DB를 통한 실패 카운트 저장  
- **모니터링**:
  - 로그인 실패 로그 급증 시 경보 → Slack/Email 알림  
  - 의심 IP(해외 VPN 등) 임시 차단

### 2.3 세션/토큰 하이재킹

- **위협**: JWT 탈취, Refresh Token 도용 → 계정 불법 접근  
- **방어**:
  1. **Access Token** 단기 만료(15분)  
  2. **Refresh Token** 안전 저장(HTTPOnly Secure Cookie, etc.)  
  3. **로그아웃 시** 블랙리스트 등록(token jti)  
  4. **TLS** 전구간 적용, token sniffing 방지  
- **모니터링**:
  - 비정상 지역(IP) 로그인 감지 → 알림/잠금  
  - 다중 로그인 시나리오 로깅

### 2.4 XSS, CSRF

- **XSS**: 클라이언트 측 문제(HTML/JS). Auth API 자체는 JSON 응답만.  
  - 만약 Admin UI, HTML Template 있으면 output encoding(Go html/template), sanitization  
- **CSRF**:
  1. Refresh Token이 쿠키에 있다면 **Double Submit Cookie** or CSRF 토큰  
  2. SSR or Single-Page App에서 CSRF header 검증  
- **모니터링**:
  - 웹 방화벽(WAF)에서 스크립트 패턴 차단  
  - 프론트엔드 UI + 백엔드 협업

### 2.5 OAuth 공격

- **위협**: 플랫폼 OAuth code 변조, redirect_uri 변경 등  
- **방어**:
  1. **PKCE**(Proof Key for Code Exchange): code_challenge/code_verifier  
  2. redirect URI 화이트리스트  
  3. 플랫폼 별 `state` 파라미터로 CSRF 방어  
- **모니터링**:
  - OAuth code 교환 실패 로그 → 의심스러운 패턴 시 경보

### 2.6 내부자 공격

- **위협**: 관리자나 시스템 접근자가 DB dump, 키(Private Key) 유출  
- **방어**:
  1. **키 관리**: Vault/HSM, 접근 통제(2FA)  
  2. DB 접근 로그 감시(audit_logs)  
  3. 최소 권한 원칙(Ops 팀원도 SELECT 제한)  
- **모니터링**:
  - DB queries(SELECT * FROM users …) 과도 시 알림  
  - Key vault 접근 이력 추적

---

## 3. 종합 보안 모범 사례

1. **TLS/mTLS**: 서버 간 및 클라이언트 ↔ 서버 모두 HTTPS  
2. **OIDC Integration**(선택) : 대규모 환경에서 표준화된 IDP  
3. **로깅 구조화**: `zap` or `logrus`, 민감 정보 제거  
4. **주기적 보안 점검**: penetration test, code review

---

## 4. 모니터링 및 알림

- **Prometheus**(메트릭), **Grafana**(대시보드)  
- **로그 분석**: ELK Stack(Elasticsearch, Logstash, Kibana)  
- **Alert**:
  1. 로그인 실패율 급증  
  2. DB error rate 상승  
  3. OAuth exchange 실패율  
- **Intrusion Detection**(옵션): Falco, OSSEC 등 호스트 기반 IDS

---

## 5. 침해 대응 시나리오

1. **토큰 유출**:
   - 블랙리스트 등록 + 키 롤링(Access/Refresh)  
   - 사용자에게 비밀번호 변경/재로그인 안내  
2. **Brute Force**:
   - 로그인 실패 급증 → 자동 IP 차단  
   - WAF rule 강화  
3. **DB dump 유출**:
   - 즉시 비밀번호/페퍼 교체  
   - 모든 토큰 무효화  
   - 침해 조사(AuditLog 분석), 당국 신고(필요 시)

---

## 6. 주기적 업데이트

- **Go Dependencies** / OpenSSL / TLS library 등 보안 이슈 발생 시 즉시 패치  
- **취약점 스캔**: Snyk, Dependabot, Trivy 등 정기 실행  
- **정기 모의 해킹**(Pentest)으로 실제 공격 시나리오 테스트

---

## 7. 결론

**Authentication Service**가 안전하게 운영되려면 아래를 준수하세요:

1. **입력 검증**과 **Prepared Statements**로 SQL Injection 방지  
2. **Rate Limiting** + **로그인 실패 횟수 제한**으로 brute force 대응  
3. **JWT 단기 만료**, **Refresh Token 안전 저장**, **블랙리스트**로 세션 하이재킹/유출 대응  
4. **OAuth**는 PKCE 등 최신 보안 패턴 준수  
5. 내부자/운영자 접근 제어 및 로그 감사(AuditLog)  
6. 정기적인 보안 모니터링, 취약점 스캔, 키/암호화 모듈 업데이트

이 가이드를 기반으로 보안 방안을 지속적으로 개선·점검해, 사용자와 서비스 모두를 안전하게 보호하십시오.