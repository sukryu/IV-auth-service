# 백업 및 복구 전략

본 문서는 **ImmersiVerse Authentication Service**에서 사용되는 **PostgreSQL 데이터베이스**를 어떻게 백업하고, 재해 상황에서 어떻게 복구할지에 대한 지침을 제공합니다. 정기적이고 안정적인 백업/복구 방안을 통해 데이터 무결성과 서비스를 안정적으로 유지할 수 있습니다.

---

## 1. 개요

- **DB 엔진**: PostgreSQL 15.x
- **주요 데이터**: 
  - 사용자 정보(`users`), 외부 플랫폼 연결(`platform_accounts`), 감사 로그(`audit_logs`) 등
  - 민감한 인증 관련 데이터(비밀번호 해시, OAuth 토큰, 등)
- **목표**:
  - **RPO**(Recovery Point Objective): 5분 이내 (데이터 손실 최대 5분치 허용)
  - **RTO**(Recovery Time Objective): 15분 이내 복구

---

## 2. 백업 유형

### 2.1 전체 백업 (Full Backup)

- **기술**: 
  - `pg_basebackup` (온라인 베이스 백업)
  - 또는 `pg_dump` (논리 백업)
- **주기**: 주간 1회 이상 (프로젝트 요구 사항에 따라)
- **장점**: 
  - 단순 복구(전체 리스토어)
- **단점**: 
  - 백업 용량 큼, 더 긴 시간 소요

### 2.2 증분(WAL) 백업 / PITR

- **WAL**(Write-Ahead Log) 아카이빙:
  - PostgreSQL 설정(`archive_mode = on`, `archive_command`) 
  - WAL 파일을 안전한 스토리지에 전송
- **장점**:
  - 시점 복구(PITR, Point-In-Time Recovery) 가능
  - 백업 공간 효율적
- **단점**:
  - 설정 다소 복잡, WAL 관리 필요

### 2.3 스냅샷 기반 (클라우드)

- **클라우드 기능**: AWS RDS snapshot, GCP Cloud SQL backup 등
- **장점**: 관리형, UI/CLI로 쉽게 복구
- **단점**: 클라우드 종속, 비용 발생

---

## 3. 권장 백업 시나리오

1. **주간/월간**: `pg_basebackup` or `pg_dump` (Full Logical or Physical Backup)
2. **실시간 WAL 아카이빙**: 로그 파일 별도 스토리지 업로드
3. **보관 주기**: 
   - 최근 7일 ~ 30일(주간 풀백업 + WAL)
   - 장기 보관(3~6개월)은 비용/규제 따라 결정
4. **클라우드 환경**: AWS RDS snapshot 병행 고려

---

## 4. 백업 절차 (예시)

### 4.1 Physical Backup (pg_basebackup)

```bash
# 1. Ensure PostgreSQL is running with archive_mode=on
# 2. Create a base backup directory
pg_basebackup -D /backups/auth_db/base_$(date +%Y%m%d) \
  -F tar -z \
  -X stream \
  -c fast \
  -U replication_user \
  -h db.host.com \
  -p 5432
```

- 결과물: `/backups/auth_db/base_YYYYMMDD.tar.gz` 형식
- WAL 로그는 `-X stream` 옵션으로 같이 수집 (단, archive_command도 활성화)

### 4.2 Logical Backup (pg_dump)

```bash
pg_dump -Fc -U auth_user -h localhost auth_db > auth_db_$(date +%Y%m%d).dump
```
- **-Fc**: Custom format, 압축/병렬복구 유리
- 복구 시: `pg_restore -d auth_db --clean --create auth_db_YYYYMMDD.dump`

---

## 5. 복구 절차

### 5.1 Physical 복구(PITR)

1. 중단된 DB 인스턴스 정지  
2. 기존 데이터 디렉토리 백업(안전용)  
3. base backup 압축 해제 → PostgreSQL data directory 복사  
4. WAL 아카이빙 경로 준비  
5. `recovery.conf`(PostgreSQL 12+는 `postgresql.conf` + `standby.signal`) 설정, `restore_command='cp /backups/wal/%f %p'` 등  
6. 특정 시점으로 복구 (recovery target time)  
7. 서버 재시작 → PITR 진행

### 5.2 Logical 복구(pg_restore)

```bash
pg_restore -U auth_user -d new_auth_db --clean --create auth_db_YYYYMMDD.dump
```
- **주의**: 데이터가 완전히 덮어씌워짐
- 테이블/인덱스 스키마 재생성(DDL) + 데이터 Insert

---

## 6. 암호화 및 보안

- **백업 파일 암호화**: AES-256 등으로 압축 후 암호화  
  - 예: `gpg --symmetric --cipher-algo AES256`
- **오프사이트 보관**: 백업 파일을 다른 지역/클라우드 스토리지 보관  
- **민감 정보**: DB에 저장된 비밀번호 해시(Argon2id) 등은 해싱되어도 백업파일 유출 시 위험. 추가 보호 고려.

---

## 7. 자동화 & 스케줄

- **Crontab**: 매일 새벽 3시, `pg_dump` or `pg_basebackup` → 보관  
- **CI/CD**: Jenkins/GitHub Actions로 주간 작업 가능  
- **Monitoring**: 백업 성공/실패 알림(Slack, Email)

예시 Crontab:

```bash
0 3 * * * /opt/scripts/backup_auth_db.sh >> /var/log/db_backup.log 2>&1
```

---

## 8. 테스트 복구

- **정기 DR Drill**: 실제로 백업 파일에서 복원해 테스트 환경 구동  
- **Verify**: 유저 레코드, 토큰 데이터, 플랫폼 계정 일치 여부  
- **모의 시나리오**: "DB 손실" 가정 후 복구 시간 측정

---

## 9. 유지보수 고려

1. **공간 관리**: 오래된 백업(>30일) 자동 삭제 or 장기 보관 시스템 이동  
2. **WAL 로그 축적**: archive 폴더가 가득 차지 않도록 모니터링  
3. **DB 업그레이드**: 마이그레이션 방식이 달라질 수 있으므로 백업 절차 재검토

---

## 10. 결론

**백업 및 복구 전략**은 인증 서비스를 안정적으로 운영하기 위한 필수 요소입니다. 다음을 핵심 지침으로 삼으세요:

- **정기 풀백업 + WAL 증분**(PITR 지원)  
- **암호화·오프사이트 보관**으로 보안/재해 대비  
- **정기 복구 테스트**로 유효성 확인  
- **자동화**(스케줄, 모니터링, 알림)

이 가이드를 준수하여, 만일의 DB 장애나 데이터 손실 상황에서도 빠르게 대응하고 플랫폼 가용성을 유지하기 바랍니다.