# 디버깅 가이드

본 문서는 **ImmersiVerse Authentication Service**를 개발·운영하면서 발생하는 문제(에러, 성능 저하, 예외 상황 등)를 **효율적으로 디버깅**하기 위한 방법을 안내합니다.

---

## 1. 개요

- **언어/프레임워크**: Go (1.21+)
- **주요 디버깅 수단**:  
  1. 로깅 (zap/klog/logrus 중 택1)  
  2. Go Delve (CLI/IDE)  
  3. IDE 내장 디버거 (VS Code, GoLand 등)  
  4. 분산 트레이싱 (OpenTelemetry)  
  5. 메트릭/모니터링 (Prometheus, Grafana)  
- 목표: 빠른 원인 파악, 정확한 수정, 최소 다운타임 유지

---

## 2. 로깅 기반 디버깅

### 2.1 로그 레벨

- **DEBUG**: 상세한 정보 (개발환경)  
- **INFO**: 주요 흐름 정보  
- **WARN**: 잠재적 문제, 성능 저하 가능성  
- **ERROR**: 요청 실패, DB 에러 등  
- **FATAL**: 프로세스 종료급 치명 에러

### 2.2 동적 레벨 변경

- 개발환경에서는 `DEBUG`로 시작, 운영환경에서는 `INFO` 또는 `WARN`.  
- 일부 로깅 라이브러리는 SIGHUP 등으로 동적 변경 가능(프로젝트 설정에 따라).

### 2.3 로깅 시 주의사항

- 민감 정보(비밀번호, 토큰 등)는 **노출 금지**  
- 구조화된 로깅(JSON)으로 필드 구분: `ip`, `user_id`, `request_id` 등  
- 에러 스택 추적/원인 메시지 함께 기록

---

## 3. Go Delve로 디버깅 (CLI)

### 3.1 설치

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

### 3.2 실행

```bash
dlv debug ./cmd/server -- --config=dev.yaml
```
- `--` 이후 인자는 본 앱에 전달  
- Delve CLI에서 `break`, `continue`, `print` 명령어로 디버깅

### 3.3 예시 명령어

- `break internal/service/auth_service.go:45` → 특정 위치 브레이크  
- `continue` → 다음 브레이크까지 진행  
- `print req.username` → 변수 값 출력  
- `list` → 소스 미리보기

---

## 4. IDE 디버거 (VS Code / GoLand)

### 4.1 VS Code

- **Go Extension** 설치  
- `launch.json` 생성 예시:
  ```json
  {
    "version": "0.2.0",
    "configurations": [
      {
        "name": "Debug AuthService",
        "type": "go",
        "request": "launch",
        "mode": "debug",
        "program": "${workspaceFolder}/cmd/server/main.go",
        "args": ["--config=dev.yaml"],
        "env": {
          "DB_HOST": "localhost",
          "DB_USER": "auth_user"
        }
      }
    ]
  }
  ```
- **F5** 키로 디버깅 시작 → 브레이크포인트 설정, 변수 조회 가능

### 4.2 GoLand

- **Run/Debug Configurations** → Add Go Build  
- **Program arguments**로 `--config=dev.yaml` 등 설정  
- **Environment variables**에 `DB_HOST=localhost` 등

---

## 5. 분산 트레이싱 (OpenTelemetry)

### 5.1 Setup

- Auth Service가 gRPC 인터셉터 또는 HTTP 미들웨어로 **OpenTelemetry** 스팬 생성.  
- Jaeger, Zipkin 등 백엔드로 export.

### 5.2 디버깅 활용

- **Trace**: 요청 경로(로그인 → DB 쿼리 → Kafka 이벤트 발행) 시각화  
- **Span**: 각 단계의 duration, 오류 발생 지점 파악  
- 성능 병목, gRPC 대기시간 등을 식별

---

## 6. 모니터링/메트릭

### 6.1 Prometheus Metrics

- **Auth Service** 내장 메트릭 예:
  - `auth_login_requests_total`
  - `auth_login_errors_total`
  - `auth_token_verify_latency_seconds`
- **Grafana Dashboards**:
  - 로그인 성공/실패률 그래프  
  - 평균 DB 응답시간  
  - Kafka publish 에러 등

### 6.2 알림(Alerts)

- CPU/RAM 사용량 초과, 에러율 급증, etc. → Slack/Email/PagerDuty

---

## 7. 일반 디버깅 팁

1. **재현**: 문제 발생 상황(입력, 환경변수, 브랜치 등)을 최대한 재현  
2. **점진적 좁히기**: 로그 레벨 높이거나 Delve로 단계별 확인  
3. **Mock vs 실제 의존성**: 통합테스트(실제 DB/Kafka)로 환경 차이로 인한 버그인지 확인  
4. **Rollback**: 최근 커밋/릴리스에서 변경점을 살펴본 뒤 문제 발생 전후 비교

---

## 8. 운영환경 디버깅

- 운영환경은 보안·성능 문제로 Delve 직접 사용 제한  
- **원격 로깅**(ELK, Cloud Logging), 메트릭, 분산 트레이싱으로 원인 파악  
- 필요 시 스테이징 환경에서 동일 데이터/트래픽 시뮬레이션 후 Delve / IDE Debug

---

## 9. 예시 시나리오별 디버깅

### 9.1 로그인 시 에러 다발

1. Prometheus에서 `auth_login_errors_total`이 급증  
2. 로그( `level=ERROR` ) 또는 Jaeger Trace 점검:
   - DB 연결 실패?  
   - Password hashing 라이브러리 에러?  
3. 로컬 재현: Docker Compose + Delve로, `Login` 함수 브레이크 시점 확인  
4. 수정 후 `make test`, CI 통과, 재배포

### 9.2 Kafka 메시지 발행 실패

1. 로그에 `"failed to publish login event"` 문구, `err=kafka: broker not available`  
2. Check Docker Compose logs, broker address  
3. Redeploy Kafka or fix env variable  
4. Retest → event normal

---

## 10. 결론

**디버깅**은 **로깅**, **Delve/IDE**, **분산 트레이싱**, **모니터링**을 종합적으로 활용해야 합니다.  
- 로컬에선 Delve/IDE로 브레이크포인트 디버깅,  
- 운영환경에선 로그+트레이싱+메트릭 조합으로 원인 파악.  
- 문제 상황 별 시나리오(로그인 에러, DB 장애, Kafka 실패 등)에 맞춰 적절히 기법을 선택하세요.