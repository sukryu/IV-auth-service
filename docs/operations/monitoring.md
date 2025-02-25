# 모니터링 및 알림 설정

본 문서는 **ImmersiVerse Authentication Service**의 안정적인 운영을 위해 필요한 모니터링 지표, 도구 구성, 알림 설정 및 대시보드 구성 방안을 제시합니다. 이 가이드를 통해 클러스터 내 애플리케이션, 인프라 및 네트워크 상태를 실시간으로 감시하고, 이상 징후 발생 시 신속하게 대응할 수 있습니다.

---

## 1. 개요

- **목표**:
  - 서비스 가용성, 성능, 리소스 사용량, 에러율 등을 실시간으로 모니터링
  - 이상 징후 발생 시 자동 알림을 통해 신속한 대응
  - 운영 및 장애 대응 시간을 최소화하고, 안정적인 서비스를 제공

- **주요 도구**:
  - **Prometheus**: 메트릭 수집 및 저장
  - **Grafana**: 대시보드 시각화
  - **Alertmanager**: 알림 규칙 설정 및 알림 전송 (Slack, Email, PagerDuty 등)
  - **ELK Stack**(선택): 애플리케이션 로그 집계 및 분석

---

## 2. 주요 모니터링 지표

### 2.1 애플리케이션 지표

- **로그인 및 인증 관련**:
  - `auth_login_requests_total`: 로그인 시도 총 건수
  - `auth_login_success_total`: 로그인 성공 건수
  - `auth_login_failure_total`: 로그인 실패 건수
  - `auth_token_validation_latency_seconds`: 토큰 검증 지연 시간

- **토큰 관리**:
  - `auth_tokens_issued_total`: 발급된 토큰 수
  - `auth_token_errors_total`: 토큰 검증 실패/오류 건수

- **API 응답 시간**:
  - 각 gRPC 메서드(예: `Login`, `CreateUser`)의 평균 응답 시간 및 95백분위 응답 시간

### 2.2 인프라 및 리소스 지표

- **컨테이너 리소스 사용량**:
  - CPU, 메모리 사용량 (Deployment 단위)
  - 네트워크 I/O 및 디스크 사용량

- **데이터베이스**:
  - PostgreSQL 쿼리 응답 시간 및 트랜잭션 처리량
  - 연결 수 및 에러율

- **메시징**:
  - Kafka의 메시지 처리량, 소비 지연, 브로커 상태

---

## 3. Prometheus 구성

### 3.1 Prometheus 설정 예시

- **prometheus.yml** 예시:
  ```yaml
  global:
    scrape_interval: 15s
    evaluation_interval: 15s

  scrape_configs:
    - job_name: 'auth-service'
      static_configs:
        - targets: ['auth-service.production.svc.cluster.local:50051']
  
    - job_name: 'kubernetes-nodes'
      kubernetes_sd_configs:
        - role: node

    - job_name: 'kubernetes-pods'
      kubernetes_sd_configs:
        - role: pod
      relabel_configs:
        - source_labels: [__meta_kubernetes_pod_label_app]
          regex: auth-service
          action: keep
  ```

### 3.2 메트릭 수집

- **내장 메트릭**: 애플리케이션 내 Prometheus 클라이언트 라이브러리를 사용해 기본 메트릭 수집 (예: Go의 `prometheus/client_golang`)
- **사용자 정의 메트릭**: 각 주요 함수나 gRPC 인터셉터에서 추가 메트릭 수집
- **Pod 모니터링**: Kubernetes Exporter를 통해 클러스터 내 Pod 상태 모니터링

---

## 4. Grafana 대시보드

### 4.1 대시보드 구성 요소

- **서비스 헬스 대시보드**:  
  - 로그인 성공/실패 비율, 토큰 발급 및 검증 지연 시간  
  - gRPC 메서드별 응답 시간 분포  
- **리소스 사용량 대시보드**:  
  - CPU, 메모리, 네트워크 I/O  
  - 컨테이너/Pod별 리소스 활용 현황  
- **데이터베이스 모니터링**:  
  - PostgreSQL 쿼리 응답, 연결 수, 에러율  
- **Kafka 대시보드**:  
  - 메시지 처리량, 소비 지연, 브로커 상태

### 4.2 Grafana 설정 예시

- Grafana 데이터 소스로 Prometheus를 추가하고, 아래와 같은 패널을 생성:
  - **로그인 성공률**:
    ```promql
    rate(auth_login_success_total[1m])
    ```
  - **gRPC 응답 시간 95th percentile**:
    ```promql
    histogram_quantile(0.95, sum(rate(grpc_server_handling_seconds_bucket[1m])) by (le))
    ```
  - **CPU 사용량**:
    ```promql
    sum(rate(container_cpu_usage_seconds_total{namespace="production", pod=~"auth-service.*"}[1m])) by (pod)
    ```

---

## 5. Alertmanager 설정

### 5.1 알림 규칙 예시

- **로그인 실패율 경고**:
  ```yaml
  groups:
    - name: auth-service-alerts
      rules:
        - alert: HighLoginFailureRate
          expr: rate(auth_login_failure_total[5m]) > 5
          for: 2m
          labels:
            severity: warning
          annotations:
            summary: "로그인 실패율 급증"
            description: "최근 5분간 로그인 실패 건수가 5건을 초과했습니다. 확인 바랍니다."
  ```
- **gRPC 응답 지연**:
  ```yaml
        - alert: HighGRPCResponseTime
          expr: histogram_quantile(0.95, sum(rate(grpc_server_handling_seconds_bucket[1m])) by (le)) > 0.5
          for: 1m
          labels:
            severity: critical
          annotations:
            summary: "gRPC 응답 지연"
            description: "최근 1분 동안 gRPC 95th percentile 응답 시간이 0.5초를 초과했습니다."
  ```

### 5.2 알림 채널

- **Slack**, **Email**, **PagerDuty** 등과 연동하여 Alertmanager가 실시간 알림 전송
- 각 알림의 심각도에 따라 다른 채널로 전송 (예: Critical은 PagerDuty, Warning은 Slack)

---

## 6. 운영 모니터링 & 유지보수

- **정기 점검**: 
  - 대시보드 및 알림 규칙을 주기적으로 검토하여 이상 징후 수정
  - 모니터링 도구 업데이트 및 새로운 메트릭 추가  
- **자동화**:
  - CI/CD 파이프라인에 Prometheus, Grafana 설정 검증 포함
  - 알림 실패나 과도한 노이즈 시, 필터링/재설정 작업 수행

---

## 7. 결론

**ImmersiVerse Authentication Service**의 모니터링 및 알림 설정은 서비스의 안정성과 성능을 보장하는 핵심 요소입니다.  
- **Prometheus**를 통해 핵심 메트릭을 수집하고,
- **Grafana**로 시각화하며,
- **Alertmanager**로 실시간 알림을 설정함으로써,
  
서비스 이상 시 신속하게 대응할 수 있습니다. 이 가이드를 기반으로 환경에 맞게 설정을 조정하고, 지속적으로 모니터링 지표를 업데이트해 주세요.

---