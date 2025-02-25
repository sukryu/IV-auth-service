# 백업 및 복구 절차

본 문서는 **ImmersiVerse Authentication Service**의 데이터 백업 및 복구 절차를 설명합니다. 이 절차는 PostgreSQL 데이터베이스를 대상으로 하며, 백업 생성, 보관, 복구 테스트 및 장애 발생 시 신속한 복구를 보장하는 데 중점을 둡니다.

---

## 1. 개요

- **대상**:  
  - PostgreSQL 데이터베이스 (예: `users`, `platform_accounts`, `token_blacklist`, `audit_logs` 등)
- **목표**:  
  - 데이터 손실 최소화 (RPO: 5분 이내)
  - 빠른 서비스 복구 (RTO: 15분 이내)
- **백업 유형**:  
  - 전체 백업 (Full Backup)
  - 증분 백업 및 WAL(Write-Ahead Log) 아카이빙 (PITR: Point-In-Time Recovery)
- **자동화**:  
  - 정기 스케줄(Crontab, CI/CD)을 통한 백업 및 복구 테스트 자동화

---

## 2. 백업 생성 절차

### 2.1 전체 백업 (Full Backup)

- **방법**: `pg_basebackup` 또는 `pg_dump`를 사용  
- **예시 (pg_basebackup)**:
  ```bash
  pg_basebackup -D /backups/auth_db/base_$(date +%Y%m%d) \
    -F tar -z -X stream -c fast \
    -U replication_user -h postgres-auth-service -p 5432
  ```
  - 결과물: `/backups/auth_db/base_YYYYMMDD.tar.gz`
  - 옵션 설명:
    - `-X stream`: WAL 로그도 함께 스트리밍하여 백업
    - `-F tar -z`: tar 파일로 압축 저장

### 2.2 증분 백업 및 WAL 아카이빙

- **방법**: PostgreSQL 설정(`archive_mode = on`, `archive_command`) 활용  
- **설정 예시 (postgresql.conf)**:
  ```conf
  archive_mode = on
  archive_command = 'cp %p /backups/auth_db/wal/%f'
  ```
- **WAL 아카이빙**:  
  - WAL 파일을 안전한 스토리지(예: S3, Glacier)에 저장하여, PITR 시 복원 가능하도록 함

### 2.3 백업 자동화

- **Crontab 예시**:
  ```bash
  # 매일 새벽 3시에 전체 백업 실행
  0 3 * * * /opt/scripts/backup_auth_db.sh >> /var/log/db_backup.log 2>&1
  ```
- **스크립트 내용 (backup_auth_db.sh)**:
  ```bash
  #!/bin/bash
  DATE=$(date +%Y%m%d)
  BACKUP_DIR="/backups/auth_db/base_$DATE"
  mkdir -p "$BACKUP_DIR"
  pg_basebackup -D "$BACKUP_DIR" -F tar -z -X stream -c fast -U replication_user -h postgres-auth-service -p 5432
  ```

---

## 3. 복구 절차

### 3.1 복구 전 준비

1. **백업 파일 확인**:  
   - 백업 디렉토리와 WAL 로그 파일 확인  
2. **DB 서비스 중지**:  
   - 현재 운영 중인 PostgreSQL 인스턴스를 안전하게 중지
3. **데이터 디렉토리 백업**:  
   - 현재 데이터 디렉토리 백업(안전 용도로)

### 3.2 전체 복구 (Full Backup 복원)

- **pg_basebackup 복원 절차**:
  1. 백업 파일 압축 해제:
     ```bash
     tar -xzvf /backups/auth_db/base_YYYYMMDD.tar.gz -C /var/lib/postgresql/data/
     ```
  2. 복구 구성 파일 생성 (PostgreSQL 12 이상):
     - `standby.signal` 파일 생성 (필요한 경우)
     - `postgresql.conf` 내 복구 관련 옵션 추가:
       ```conf
       restore_command = 'cp /backups/auth_db/wal/%f %p'
       recovery_target_time = '2025-06-01 03:15:00'
       ```
  3. PostgreSQL 재시작:
     ```bash
     systemctl start postgresql
     ```
  4. 복구 진행 상황 모니터링:
     ```bash
     tail -f /var/log/postgresql/postgresql.log
     ```

### 3.3 Logical 복구 (pg_dump 복원)

- **pg_dump 복원 절차**:
  1. 새로운 DB 생성:
     ```bash
     createdb -U auth_user new_auth_db
     ```
  2. pg_restore 실행:
     ```bash
     pg_restore -U auth_user -d new_auth_db --clean --create /backups/auth_db_YYYYMMDD.dump
     ```

---

## 4. 복구 테스트 및 검증

1. **정기 테스트**:  
   - 매월 또는 분기별로 백업 파일을 사용해 테스트 환경에서 복구 시나리오를 실행
   - 복구된 데이터와 최신 DB 상태 비교
2. **PITR 테스트**:  
   - 특정 시점으로 복구(PITR)하여 데이터 손실(RPO) 및 복구 시간(RTO) 확인
3. **문서화**:  
   - 복구 테스트 결과를 기록하고, 문제 발생 시 절차 수정

---

## 5. 보안 및 보관

- **암호화**:  
  - 백업 파일은 생성 후 AES-256으로 암호화하거나 GPG 암호화 적용  
- **오프사이트 보관**:  
  - 중요한 백업 파일은 다른 지역 또는 클라우드 스토리지(S3, Glacier)로 복사
- **보관 정책**:  
  - 최근 7일~30일 온라인 보관, 장기 보관은 규제 요구사항에 따라 별도 아카이브

---

## 6. 운영 및 모니터링

1. **백업 모니터링**:
   - 백업 작업 성공/실패 로그 모니터링 (예: Slack 알림)
   - 백업 스토리지 용량 및 상태 점검
2. **복구 시나리오 모의**:
   - 정기적으로 재해 복구 훈련(DR Drill)을 실시하여 실제 복구 절차 점검
3. **문제 발생 시 대응**:
   - 백업 실패 시 즉시 알림, 자동 스크립트 재실행 또는 수동 개입

---

## 7. 결론

**백업 및 복구 절차**는 **Authentication Service**의 안정성과 데이터 무결성을 보장하기 위한 필수 프로세스입니다.

- **정기 백업**(전체 백업 + WAL 아카이빙)과 **자동화된 스케줄**로 데이터 보호
- **신속한 복구**: 전체 복원 및 논리적 복원 방법을 마련하여 RPO/RTO 목표 달성
- **정기 복구 테스트**를 통해 절차의 신뢰성을 검증하고, 필요 시 업데이트

이 가이드를 준수하여, 예상치 못한 장애나 데이터 손실 상황에서도 신속하고 안전하게 시스템을 복구할 수 있도록 하시기 바랍니다.

---