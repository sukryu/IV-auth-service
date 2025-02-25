# 마이그레이션 전략

본 문서는 **ImmersiVerse Authentication Service**의 **데이터베이스 마이그레이션** 방법과 원칙을 정의합니다. 스키마 변경(테이블 추가, 컬럼 수정 등)을 안전하게 적용하고, 하위 호환성을 유지하며, 롤백 시에도 데이터 일관성을 보장하기 위한 절차를 안내합니다.

---

## 1. 개요

- **DB 엔진**: PostgreSQL (버전 15.x 이상)
- **마이그레이션 툴**: 예) [golang-migrate/migrate](https://github.com/golang-migrate/migrate), Flyway, 또는 Liquibase
- **마이그레이션 파일 위치**: `db/migrations/` 디렉토리
- **파일 형식**: `NNNN_description.up.sql` / `NNNN_description.down.sql`
  - 예: `0001_create_users.up.sql` / `0001_create_users.down.sql`
- **전략 요약**: 
  1. DDL 변경 시 신규 .sql 파일 작성  
  2. CI/CD 파이프라인 또는 수동 명령으로 DB에 적용  
  3. 필요 시 다운 스크립트로 롤백

---

## 2. 디렉토리 구조

```
auth-service/
└── db/
    ├── migrations/
    │   ├── 0001_create_users.up.sql
    │   ├── 0001_create_users.down.sql
    │   ├── 0002_create_platform_accounts.up.sql
    │   ├── 0002_create_platform_accounts.down.sql
    │   ...
    └── seeds/
        └── example_data.sql
```

1. **migrations/**: 스키마 정의 변경. 각 쌍의 `up`/`down` 파일로 구성  
2. **seeds/**: 개발/테스트 용 초기 데이터 삽입 스크립트(선택 사항)

---

## 3. 버전 관리 방식

### 3.1 번호 & 파일명 규칙

- **고유 번호** `NNNN`은 일련 증가(0001, 0002, 0003, …).  
- **snake_case**로 간략한 설명: `0003_add_audit_logs`.
- **업/다운**:
  - `0003_add_audit_logs.up.sql`
  - `0003_add_audit_logs.down.sql`

### 3.2 Git에서 다중 개발 시 주의

- 충돌(동일 번호 중복) 방지 → 새로운 migration 추가 시, 직전 번호보다 +1 한 값 사용  
- PR 머지 순서가 바뀌지 않도록 협업 시 신중

### 3.3 하위 호환성

- Auth Service가 Zero downtime 배포를 지향할 경우, DB 변경은 backwards-compatible하게 적용:
  - 먼저 새 컬럼/테이블 추가 → 코드에서 사용  
  - 나중에 구 컬럼 제거 (메이저 릴리스 시)
- 빅 테이블에 컬럼 추가 시, `CONCURRENT` 등 PostgreSQL 옵션 활용(성능 이슈 고려)

---

## 4. 마이그레이션 툴 사용

### 4.1 golang-migrate 예시

1. **설치**:
   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```
2. **업그레이드**:
   ```bash
   migrate -path db/migrations -database "postgres://user:pass@localhost:5432/auth_db?sslmode=disable" up
   ```
3. **다운그레이드**:
   ```bash
   migrate -path db/migrations -database "..." down 1
   ```
4. **Makefile**:  
   ```makefile
   migrate:
       migrate -path db/migrations \
               -database "$(DB_DSN)?sslmode=disable" up
   rollback:
       migrate -path db/migrations \
               -database "$(DB_DSN)?sslmode=disable" down 1
   ```

### 4.2 Flyway / Liquibase (대안)

- **Flyway**: Java 기반, Spring 등과 호환성 높음  
- **Liquibase**: XML/YAML/SQL 등 다양한 format 지원

---

## 5. 개발 환경 적용

1. **Docker Compose**로 PostgreSQL 컨테이너 실행
2. `.env` 설정(DB_HOST=localhost, etc.)
3. `make migrate` (또는 manual `migrate` 명령) 실행
4. **seeds** (선택): `make seed` → 개발용 예시 데이터 삽입

---

## 6. 운영 환경 롤아웃

1. **배포 파이프라인**(CI/CD)에서:
   - 새 코드 배포 전 DB 마이그레이션 수행 → `migrate up`
   - 성공 시 애플리케이션 롤링 업데이트  
   - 실패 시 `down` 또는 수동 조치
2. **Zero Downtime**:
   - DB 변경 → 애플리케이션이 새 필드/테이블 참조
   - 구 코드도 문제 없이 동작하도록 설계(필드 삭제는 나중에)
3. **롤백 시나리오**:
   - `migrate down` N → 단, 데이터 손실 주의  
   - Backup/restore(스냅샷) 사용 시점 고려

---

## 7. 마이그레이션 작성 규칙

1. **Atomic**: 하나의 마이그레이션 파일에서 한 가지 목적(테이블 추가, 컬럼 추가 등)  
2. **Idempotent**: `.up.sql` 여러 번 실행되지 않게 DB state 점검(tool이 version table 관리)  
3. **DOWN 스크립트**: 역방향 완전 복구가 가능한지 여부 결정  
   - 예) 컬럼 삭제 시 데이터 손실 발생 → 주의  
   - 메이저 변경 시 down 스크립트 생략하기도 함(팀 정책)
4. **Index & Constraint**: 별도 마이그레이션으로 분리 or 함께 처리, 필요 시 CONCURRENTLY 옵션

---

## 8. 예시 마이그레이션

### 8.1 0001_create_users.up.sql

```sql
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    subscription_tier VARCHAR(20),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP
);
```

### 8.2 0001_create_users.down.sql

```sql
DROP TABLE IF EXISTS users;
```

---

## 9. 운영 고려사항

- **백업**: 중요한 마이그레이션 전에는 DB 스냅샷/백업 수행
- **쪼개진 릴리스**: 대규모 변경은 여러 단계로 나누어 배포  
- **모니터링**: 마이그레이션 시간, 트랜잭션 락, 지연 발생 여부
- **문서화**: 각 마이그레이션 파일에 목적·변경 요약 주석 기재

---

## 10. 결론

이 **마이그레이션 전략**을 따르면, Authentication Service의 **DB 스키마 변경**이 일관성 있고 안전하게 진행됩니다.

1. **고유 번호** + **UP/DOWN** SQL 파일  
2. **golang-migrate** 등 툴로 자동화  
3. **Zero Downtime** 원칙: 하위 호환 우선, 큰 변경은 단계적  
4. **CI/CD**와 연계해 배포 시점에 자동 적용/검증

팀원들은 스키마 수정 시 **새로운 마이그레이션 파일**을 생성하고, 리뷰와 테스트 후 병합·배포하는 절차를 준수해 주세요.