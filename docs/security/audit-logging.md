# 감사 로깅 정책

본 문서는 **ImmersiVerse Authentication Service**에서 발생하는 중요한 이벤트를 **감사 로그**에 기록하고, 이를 안전하게 보관·조회·분석하는 방안을 제시합니다. 이를 통해 **부정 행위**, **오류 진단**, **규제 준수** 등에 필요한 근거를 확보하고, 시스템 운영 전반의 투명성을 높일 수 있습니다.

---

## 1. 개요

- **감사 로깅(Audit Logging)**: 시스템 내 중요한 동작(사용자 생성·삭제, 역할 변경, 로그인 성공·실패, 플랫폼 계정 연결 등)을 별도 기록
- **주요 목표**:
  1. **보안**: 의심스러운 활동(무단 접근, 권한 오남용 등) 추적
  2. **규제 준수**: 감사 로그 의무 보관
  3. **문제 해결**: 장애나 버그 발생 시 원인 파악
- **로그 저장 위치**:  
  - DB 테이블 `audit_logs` (주요 액션 기록)  
  - 추가로 ELK Stack(Elasticsearch) 등 외부 로그 시스템 연동 가능

---

## 2. 감사 로그 범위

### 2.1 주요 이벤트

1. **사용자 관리**:
   - 사용자 생성, 정보 업데이트, 삭제
   - 비밀번호 변경, 계정 잠금/해제
2. **인증 관련**:
   - 로그인 성공/실패 (로그인 이벤트 테이블 또는 audit_logs)
   - 토큰 무효화(로그아웃), 토큰 갱신
3. **플랫폼 계정**:
   - 외부 플랫폼 연결(ConnectPlatformAccount)
   - 연결 해제(DisconnectPlatformAccount)
4. **권한/역할 변경**(RBAC):
   - 역할(Role) 생성/삭제, 사용자에 역할 부여/제거
5. **중요한 설정 변경**:
   - OAuth 클라이언트 secret 수정, JWT 키 교체 등

### 2.2 기록 항목

- **실행 주체**(user_id, 혹은 system/service)
- **액션 유형(action)**: e.g. `CREATE_USER`, `LOGIN_FAILED`, `CHANGE_PASSWORD`
- **대상 엔티티**(entity_type + entity_id), 예: `user`, `platform_account`
- **이전/새 값**(optional) - old_values/new_values (JSONB)
- **ip_address**, **user_agent**(optional) - 웹 접근 시
- **timestamp** - 이벤트 발생 시각

---

## 3. DB 스키마: audit_logs

```sql
CREATE TABLE audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(36),
    old_values JSONB,
    new_values JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

- **FK**(user_id → users.id) with ON DELETE SET NULL  
- 인덱스: `(action, created_at)`, `(entity_type, entity_id)` 등 사용 빈도에 따라 추가

---

## 4. 로깅 시점 및 방법

1. **트랜잭션 성공 후 기록**: DB Insert/Update 가 확정된 뒤, 감사 로그도 함께 Insert
2. **도메인 서비스에서**: e.g. `UserDomainService.CreateUser()` → 끝나면 `CreateAuditLog(action="CREATE_USER", ...)`
3. **Kafka 연동**(선택):
   - 이벤트를 Kafka로 발행, 별도 Audit Consumer가 DB Insert

### 4.1 예시 코드 (Go pseudo)

```go
func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    // 1. Create user in DB
    user, err := s.repo.InsertUser(req)
    if err != nil { return nil, err }

    // 2. Insert to audit_logs
    s.auditRepo.LogAudit(audit.AuditLog{
        ID: uuid.New().String(),
        UserID: currentAdminOrSystemUser,
        Action: "CREATE_USER",
        EntityType: "USER",
        EntityID: user.ID,
        NewValues: toJSON(user),
        IpAddress: req.IpAddress,
        // ...
    })
    return user, nil
}
```

---

## 5. 보안 및 접근 통제

1. **민감 정보 마스킹**: 
   - 비밀번호, 토큰 등의 원문을 old/new_values에 저장 금지
   - `"******"` 표시 or 제외
2. **권한 제한**: 
   - DB 측: `audit_logs` SELECT는 admin/ops만 가능
   - API 측: 전용 AuditLog 조회 API(ADMIN only)
3. **무결성 보장**:
   - 가능하면 WORM(Write Once Read Many) 스토리지 or append-only 로그
   - KMS로 서명/암호화(고급 환경)

---

## 6. 보관 주기 및 아카이빙

- **단기 보관**: 6~12개월 온라인(빠른 조회)
- **장기 보관**: 1~2년 or 규제 요건에 따라 별도 S3/Glacier 등
- **파티셔닝**(audit_logs_monthly) or 정기적 아카이브 → 성능 유지

---

## 7. 조회 및 분석

- **단순**: `SELECT * FROM audit_logs WHERE action='LOGIN_FAILED' ORDER BY created_at DESC LIMIT 50`
- **고급**: Kibana/Elasticsearch 인덱싱(로그스태시 파이프라인)
- **대시보드**: Grafana or Kibana → "Top actions", "Login fail count" 등

---

## 8. 샘플 시나리오

1. **사용자 회원가입**: 
   - `action="CREATE_USER"`, `entity_type="USER"`, `new_values`= {username, email}
2. **로그인 실패**:
   - `action="LOGIN_FAILED"`, `entity_type="USER"`, `entity_id=<userID or null>`, ip_address="..."
3. **플랫폼 계정 연결**:
   - `action="CONNECT_PLATFORM"`, `entity_type="PLATFORM_ACCOUNT"`, `entity_id=<platformId>`
4. **역할 부여**:
   - `action="ASSIGN_ROLE"`, old_values= {roles=[...old]}, new_values= {roles=[...new]}

---

## 9. 운영 모니터링

- **유의미 지표**:
  - 일/주간 CREATE_USER 수, DELETE_USER 수
  - LOGIN_FAILED 증가 추이(브루트포스 징후?)
  - ACTION=CHANGE_PASSWORD 빈도(보안 사고 후 증가?)
- **알림**: 관리자 권한 이상 API 호출 시 Slack/E-mail 알림

---

## 10. 결론

감사 로그는 **핵심 보안/운영 도구**로서:

1. **모든 핵심 동작**(계정 생성, 권한 변경, 로그인 시도 등)에 대해 AuditLog Insert
2. **민감 정보는 마스킹**
3. **보안성**: 접근 제한 + 무결성 보장
4. **보관 주기**: 운영 정책 따라 6개월~1년, 필요 시 장기 아카이브

이 원칙을 준수하면, 보안 사건·운영 문제 발생 시 정확한 추적이 가능하며, 규제 요구사항도 충족할 수 있습니다.