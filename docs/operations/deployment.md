# 배포 프로세스 및 환경 설정

본 문서는 **ImmersiVerse Authentication Service**의 배포 전략과 환경 설정에 대한 전반적인 가이드를 제공합니다. 이 가이드는 로컬 개발 환경부터 스테이징, 프로덕션 환경까지의 배포 파이프라인 구성, 인프라 설정, 환경 변수 관리 및 모니터링 설정 등을 포함합니다.

---

## 1. 개요

- **목표**:  
  - 안정적이고 일관된 배포 프로세스 확립  
  - 무중단 배포(Zero Downtime) 및 롤백 전략 구현  
  - 다양한 환경(개발, 스테이징, 프로덕션)에서의 설정 일관성 유지

- **주요 기술 스택**:  
  - **컨테이너화**: Docker  
  - **오케스트레이션**: Kubernetes  
  - **CI/CD**: GitHub Actions, ArgoCD  
  - **환경 변수 관리**: .env 파일, 시크릿 관리(Vault, Docker Secrets)

---

## 2. 배포 프로세스 개요

1. **코드 커밋 및 PR**:  
   - 모든 코드 변경은 GitHub Pull Request를 통해 검토  
   - CI/CD 파이프라인에서 자동 테스트, 린트, 코드 커버리지 확인

2. **CI 단계**:  
   - GitHub Actions를 사용하여 코드 빌드, 테스트, 패키징 수행  
   - 테스트 성공 시 Docker 이미지 생성 및 컨테이너 레지스트리(Harbor, Docker Hub 등) 업로드

3. **CD 단계**:  
   - ArgoCD를 통해 Kubernetes 클러스터에 배포  
   - 배포 전략: 카나리 배포, 블루/그린 배포, 롤백 자동화  
   - 배포 후 헬스체크, 모니터링, 로그 수집을 통한 안정성 확인

---

## 3. 환경 설정

### 3.1 환경별 구성 파일

- **개발 환경**:  
  - 로컬 Docker Compose 구성 파일(`docker-compose.yml`)을 통해 PostgreSQL, Redis, Kafka 등 의존 서비스 실행  
  - `.env.development` 파일을 통해 DB, JWT, Kafka, 기타 필요한 설정 주입

- **스테이징 환경**:  
  - Kubernetes 클러스터 내 별도 네임스페이스(`staging`) 사용  
  - ConfigMap과 Secret을 활용하여 환경 변수와 민감 정보 주입  
  - 스테이징 전용 Ingress, TLS 설정 적용

- **프로덕션 환경**:  
  - 다중 리전 Kubernetes 클러스터 구성  
  - 높은 가용성을 위해 ReplicaSet, Horizontal Pod Autoscaler(HPA) 적용  
  - 롤링 업데이트 및 Canary 배포 전략 적용  
  - 프로덕션 전용 ConfigMap/Secret 관리, TLS/mTLS 적용

### 3.2 환경 변수 및 시크릿 관리

- **환경 변수**:  
  - DB 접속 정보, Kafka 브로커 주소, JWT 키 파일 경로 등을 포함  
  - 예시(.env 파일):
    ```dotenv
    DB_HOST=postgres-auth-service
    DB_PORT=5432
    DB_USER=auth_user
    DB_PASS=auth_password
    DB_NAME=auth_db
    REDIS_ADDR=redis:6379
    KAFKA_BROKER=kafka:9092
    JWT_PRIVATE_KEY_PATH=/etc/secrets/jwt-private.pem
    JWT_PUBLIC_KEY_PATH=/etc/secrets/jwt-public.pem
    ```
- **시크릿 관리**:  
  - 프로덕션에서는 Docker Secrets, Kubernetes Secrets, 혹은 Vault와 같은 도구를 사용하여 민감 정보를 안전하게 관리

---

## 4. Kubernetes 배포 구성

### 4.1 Deployment & Service 매니페스트

- **Deployment**:  
  - Replicas: 최소 3개 이상으로 가용성 확보  
  - Readiness/Liveness Probe: 주기적으로 애플리케이션 헬스 체크
  - 리소스 요청 및 제한: CPU, 메모리 설정으로 오버스케일 방지

- **Service**:  
  - ClusterIP 또는 LoadBalancer 타입 사용  
  - 내부 서비스 통신은 ClusterIP, 외부 접근은 Ingress + LoadBalancer 사용

### 4.2 Ingress 및 TLS 설정

- **Ingress Controller**(예: NGINX, Envoy) 설정을 통해 API Gateway 역할 수행  
- **TLS**: Let's Encrypt 등 인증서를 통한 HTTPS 적용  
- 예시 Ingress 스니펫:
  ```yaml
  apiVersion: networking.k8s.io/v1
  kind: Ingress
  metadata:
    name: auth-ingress
    namespace: production
    annotations:
      kubernetes.io/ingress.class: "nginx"
      cert-manager.io/cluster-issuer: "letsencrypt-prod"
  spec:
    tls:
      - hosts:
          - auth.immersiverse.com
        secretName: auth-tls
    rules:
      - host: auth.immersiverse.com
        http:
          paths:
            - path: /auth
              pathType: Prefix
              backend:
                service:
                  name: auth-service
                  port:
                    number: 80
  ```

---

## 5. 배포 전략

### 5.1 카나리 배포

- **목적**: 새로운 버전의 배포 전 소규모 트래픽으로 안정성 검증  
- **절차**:
  1. 새 버전 이미지 생성 및 별도 Deployment(예: `auth-service-canary`)로 배포  
  2. 일정 기간 동안 트래픽 일부 할당하여 모니터링  
  3. 이상 없으면 전체 트래픽으로 전환, 문제 발생 시 롤백

### 5.2 블루/그린 배포

- **목적**: 완전한 무중단 배포  
- **절차**:
  1. 현재 버전(블루)과 새 버전(그린)을 동시에 배포  
  2. 그린 버전이 준비되면 Ingress/LoadBalancer를 통해 트래픽 전환  
  3. 이전 버전은 일정 기간 대기 후 제거

### 5.3 롤백 전략

- **자동 롤백**:  
  - 배포 후 헬스체크 실패나 모니터링 지표 이상 시 자동 롤백  
- **수동 롤백**:  
  - 문제가 발생하면 `kubectl rollout undo deployment/auth-service` 명령어로 이전 버전 복원

---

## 6. CI/CD 연동

- **GitHub Actions**:  
  - 커밋/PR 시 자동 빌드, 테스트, Docker 이미지 생성  
  - 테스트 성공 시 Docker Registry에 이미지 푸시
- **ArgoCD**:  
  - GitOps 방식으로 Git 리포지토리의 Kubernetes 매니페스트를 감시하고 자동 배포  
  - 배포 상태 모니터링 및 롤백 기능 제공

---

## 7. 모니터링 및 운영

- **Prometheus & Grafana**:  
  - 배포 상태, 리소스 사용량, 헬스체크 결과 등 모니터링  
- **Logging**:  
  - ELK Stack이나 클라우드 로깅 솔루션으로 배포 로그 및 애플리케이션 로그 집계  
- **알림**:  
  - CI/CD 실패, 헬스체크 실패, 배포 롤백 시 Slack, Email, PagerDuty 등으로 알림

---

## 8. 결론

**ImmersiVerse Authentication Service**의 배포 프로세스는 다음과 같은 핵심 요소를 포함합니다:

- **CI/CD 자동화**: GitHub Actions와 ArgoCD를 통해 코드 빌드, 테스트, 이미지 생성, 배포를 자동화  
- **Kubernetes 기반 배포**: Deployment, Service, Ingress, TLS 설정으로 안전하고 확장 가능한 운영 환경 구성  
- **무중단 배포 전략**: 카나리 및 블루/그린 배포, 자동 롤백으로 안정성 확보  
- **환경 변수 및 시크릿 관리**: .env, Kubernetes Secrets, Vault 등을 통한 민감 정보 보호  
- **모니터링 및 운영**: Prometheus, Grafana, ELK Stack 등을 활용하여 배포 후 상태를 실시간으로 모니터링

이 가이드를 통해 팀원들은 일관된 방식으로 배포를 진행할 수 있으며, 문제가 발생할 경우 신속하게 롤백하거나 수정할 수 있습니다.  
각 환경(개발, 스테이징, 프로덕션)에 맞춘 설정과 배포 전략을 지속적으로 업데이트하여 운영 안정성을 유지하시기 바랍니다.