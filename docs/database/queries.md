# 주요 쿼리 최적화 가이드

본 문서는 **ImmersiVerse Authentication Service**에 사용되는 **주요 SQL 쿼리**와, 이를 **최적화**하기 위한 권장 방법을 다룹니다. PostgreSQL 환경에서의 인덱스, 쿼리 구조, 실행 계획(Explain) 활용 등을 개략적으로 설명하며, 성능 병목을 줄이고 응답 속도를 높이기 위한 지침을 제공합니다.

---

## 1. 개요

- **DB 엔진**: PostgreSQL 15.x
- **테이블**: `users`, `platform_accounts`, `token_blacklist`, `audit_logs` 등
- **중요 쿼리**: 
  - 로그인(유저 조회 + 비밀번호 검증)
  - 플랫폼 계정(조건 검색)
  - 블랙리스트 조회(토큰 ID, 만료 시각)
  - 감사 로그 검색(필터 조건)

---

## 2. 로그인 관련 쿼리

### 2.1 유저 조회 (username/email)

로그인 시:
```sql
SELECT id, username, email, password_hash, status, last_login_at
FROM users
WHERE username = $1
LIMIT 1;
```
- **최적화**:
  1. `username`에 **UNIQUE 인덱스** 이미 있음 → O(1) 또는 매우 빠른 검색
  2. 만약 `email`로도 로그인 지원한다면 `email`에도 **UNIQUE 인덱스**  
  3. `EXPLAIN ANALYZE`로 쿼리 실행 계획을 점검

### 2.2 사용자 상태 업데이트 (마지막 로그인 시각)

```sql
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;
```
- 빠른 PK 접근(조건 `id`)이므로 별도 인덱스 필요 X (`id`가 PRIMARY KEY)
- 주의: 너무 빈번하면 I/O 증가. 실시간 로그 기록은 AuditLog로 대체할 수도 있음

---

## 3. 플랫폼 계정 쿼리

### 3.1 플랫폼 계정 목록 조회

```sql
SELECT id, platform, platform_username
FROM platform_accounts
WHERE user_id = $1;
```
- **최적화**:
  - `platform_accounts`에 `user_id` FK + 인덱스 자동 생성 (확인 필요)
  - 대량 데이터 시 ORDER BY, LIMIT 고려

### 3.2 플랫폼 계정 검색 (external ID)

```sql
SELECT *
FROM platform_accounts
WHERE platform = $1
  AND platform_user_id = $2
LIMIT 1;
```
- **인덱스**: 복합 인덱스(`(platform, platform_user_id)`) 권장
- 만약 자주 조회되면 성능 크게 향상

---

## 4. 토큰 블랙리스트 쿼리

### 4.1 블랙리스트 확인

```sql
SELECT token_id, expires_at
FROM token_blacklist
WHERE token_id = $1
  AND expires_at > NOW()
LIMIT 1;
```

- **최적화**:
  - `PRIMARY KEY(token_id)` 존재 시 `WHERE token_id = $1` 효율적
  - 만료 검사(`expires_at > NOW()`)는 잘못된 인덱스 선택(Partial Index) 가능
    - e.g. `CREATE INDEX idx_not_expired ON token_blacklist (token_id) WHERE expires_at > now();`
      - 하지만 동적 비교 => partial index 유지 어려움
- **정기 청소**: `expires_at < NOW()`인 레코드 주기적 삭제

---

## 5. 감사 로그(AuditLog) 검색

### 5.1 기본 쿼리

```sql
SELECT id, user_id, action, entity_type, entity_id, created_at
FROM audit_logs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 50 OFFSET 0;
```

- **최적화**:
  - `user_id, created_at DESC` 접근 빈번 → 복합 인덱스(`(user_id, created_at DESC)`) 고려
  - 로그가 매우 많으면 파티셔닝(`created_at` 기준 월별 파티션 등) + 인덱스 + 아카이빙
- **JSONB**(old_values, new_values) 검색 시, GIN 인덱스 사용 가능 (복잡도 ↑)

### 5.2 고급 필터(액션, 기간)

```sql
SELECT *
FROM audit_logs
WHERE action = 'CREATE_USER'
  AND created_at BETWEEN $1 AND $2
ORDER BY created_at DESC;
```
- 인덱스: `(action, created_at)` 또는 파티셔닝.  
- **EXPLAIN** 후 Index Scan 유도.

---

## 6. 일반 인덱스 팁

1. **PK**: 각 테이블 `id`는 PK  
2. **UNIQUE**: username, email 등 고유  
3. **FK**: `ON DELETE CASCADE` or `RESTRICT`  
4. **정렬/검색 패턴**: 자주 사용하는 `ORDER BY created_at DESC`, `WHERE user_id=?` 등 → 복합 인덱스
5. **Partial Index**: 특정 조건(`status='ACTIVE'`) 시 성능 향상 가능, 단 조건이 고정/빈도 높아야

---

## 7. EXPLAIN / EXPLAIN ANALYZE

- **EXPLAIN**: PostgreSQL 실행 계획 확인
- **EXPLAIN ANALYZE**: 실제 쿼리 실행 + 시간/행추정오차
- 예:
  ```sql
  EXPLAIN ANALYZE
  SELECT * FROM users WHERE username = 'test';
  ```
- **Key Metric**: `Index Scan` vs `Seq Scan`, Execution Time, Rows Removed by Filter, etc.
- **튜닝**:
  - Seq Scan이 불필요할 경우 → 인덱스 생성  
  - Stats/Autovacuum 설정 점검

---

## 8. 캐싱 고려

- **Redis**: Access Token/Refresh Token, 블랙리스트 1차 캐싱  
- DB 접근 빈도가 높은(로그인) 경우, **User** row를 단기 캐싱(단위 60s etc.) 시도.  
- Cache invalidation 로직 필요(유저 프로필 변경 시 캐시 갱신).

---

## 9. 모니터링 & 성능 분석

1. **pg_stat_statements**: 쿼리 빈도, 평균 실행 시간  
2. **Prometheus** + **Postgres Exporter**: DB 커넥션 수, lock 상태, slow queries  
3. **Grafana** 대시보드: `top N queries`, `avg execution time`, `rows read`, etc.

---

## 10. 운영 시나리오

- **트래픽 증가**: DB CPU/IO 부담 → 인덱스 최적화, 캐싱. 필요 시 수평 분할(sharding) 고려  
- **로그 테이블 폭증**: `audit_logs` 파티션 or TTL(archiving old data)  
- **주기적인 VACUUM**: autovacuum 설정 점검  
- **쿼리 리팩토링**: 중복 subquery/OR 연산 등 복잡 조건 → 인덱스 효율 떨어질 수 있음

---

## 11. 예시: 비밀번호 검사 시 최적화

**Login** 흐름에서:
```sql
SELECT id, password_hash, status
FROM users
WHERE username = $1
LIMIT 1;
```
- **UNIQUE(username)** → 빠른 lookup
- 비밀번호 검증은 애플리케이션 레벨에서 Argon2id hashing compare  
- DB 관점 최적화: index scan.  
- CPU 관점(hash compare)은 DB가 아닌 앱 측에서 수행.

---

## 12. 결론

**Authentication Service**에서는 다음을 권장:

1. **핵심 필드**(username, email 등)에 **UNIQUE 인덱스**를 정확히 설정  
2. **FK** 필드는 인덱스 기반(예: `platform_accounts.user_id`)  
3. 대규모 로그/감사 테이블은 **파티셔닝** + 적절한 인덱스로 성능 관리  
4. `EXPLAIN ANALYZE`로 실행 계획을 주기적으로 확인하고, 필요 시 인덱스/쿼리 수정  
5. **Autovacuum**, **Vacuum Analyze**로 통계/디스크 효율 유지

이 문서를 참고하여 DB 스키마와 쿼리를 개선, 성능 문제를 사전에 방지하고 Auth Service가 높은 TPS에서도 안정적으로 동작하도록 유지해 주세요.