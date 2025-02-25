# Kubernetes 구성 상세 가이드

본 문서는 **ImmersiVerse Authentication Service**의 Kubernetes 환경 구성 및 배포 매니페스트 작성 방법, 각 컴포넌트의 역할과 설정 옵션, 그리고 운영 시 고려사항을 설명합니다. 이를 통해 클러스터 내에서 Auth Service가 안정적이고 효율적으로 운영될 수 있도록 합니다.

---

## 1. 개요

- **목표**:  
  - 안정적, 확장 가능, 무중단 배포를 위한 Kubernetes 환경 구성  
  - 애플리케이션 구성 요소(Deployment, Service, Ingress, ConfigMap, Secret 등)와 각 설정의 목적과 사용 방법 명확화  
- **대상**:  
  - 개발자, 운영팀, DevOps 담당자

---

## 2. 주요 구성 요소

### 2.1 Deployment

- **목적**: 애플리케이션 컨테이너의 배포, 스케일링, 롤링 업데이트 관리  
- **주요 설정 항목**:
  - **replicas**: 최소 3개 이상의 복제본으로 가용성 확보  
  - **컨테이너 이미지**: CI/CD를 통해 최신 이미지 사용 (예: `immersiverse/auth-service:v1.2.3`)  
  - **리소스 요청 및 제한**: CPU, 메모리 설정으로 오버스케일 및 자원 경쟁 최소화  
  - **Readiness/Liveness Probe**: 애플리케이션 헬스 체크 (예: `/healthz` 엔드포인트)
  - **환경 변수**: ConfigMap 또는 Secret을 통해 주입

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: production
  labels:
    app: auth-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
        - name: auth-service
          image: immersiverse/auth-service:v1.2.3
          imagePullPolicy: Always
          ports:
            - containerPort: 50051
          envFrom:
            - configMapRef:
                name: auth-service-config
            - secretRef:
                name: auth-service-secrets
          resources:
            requests:
              cpu: "500m"
              memory: "512Mi"
            limits:
              cpu: "1"
              memory: "1Gi"
          readinessProbe:
            httpGet:
              path: /healthz
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 5
          livenessProbe:
            httpGet:
              path: /healthz
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
```

### 2.2 Service

- **목적**: 내부 서비스 디스커버리와 외부 트래픽 라우팅 지원  
- **유형**:
  - **ClusterIP**: 내부 통신 (기본)  
  - **LoadBalancer** 또는 **NodePort**: 외부 접근 필요 시 (Ingress와 연동)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: production
  labels:
    app: auth-service
spec:
  type: ClusterIP
  ports:
    - port: 50051
      targetPort: 50051
      protocol: TCP
      name: grpc
  selector:
    app: auth-service
```

### 2.3 Ingress

- **목적**: 외부 클라이언트 요청(REST/HTTPS)을 API Gateway로 라우팅  
- **주요 설정 항목**:
  - **host**: 외부 접속 도메인 (예: `auth.immersiverse.com`)
  - **TLS**: HTTPS 보안을 위한 인증서 적용  
  - **경로 매핑**: `/auth` 경로를 내부 Auth Service로 전달

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

### 2.4 ConfigMap 및 Secret

- **ConfigMap**:  
  - 비민감 설정 (DB URL, Kafka 브로커, API endpoint 등)  
  - 예시 파일: `auth-service-config.yaml`
  
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: auth-service-config
  namespace: production
data:
  DB_HOST: "postgres-auth-service"
  DB_PORT: "5432"
  DB_NAME: "auth_db"
  KAFKA_BROKER: "kafka:9092"
```

- **Secret**:  
  - 민감 정보 (DB 비밀번호, JWT 키 등)  
  - 예시 파일: `auth-service-secrets.yaml`
  
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: auth-service-secrets
  namespace: production
type: Opaque
stringData:
  DB_USER: "auth_user"
  DB_PASS: "auth_password"
  JWT_PRIVATE_KEY: |-
    -----BEGIN RSA PRIVATE KEY-----
    ...
    -----END RSA PRIVATE KEY-----
  JWT_PUBLIC_KEY: |-
    -----BEGIN PUBLIC KEY-----
    ...
    -----END PUBLIC KEY-----
```

---

## 3. 고가용성 및 오토스케일링

### 3.1 Horizontal Pod Autoscaler (HPA)

- **목적**: 트래픽 변화에 따라 Pod 수 자동 조정  
- **설정**: CPU 사용률, 메모리 사용률 등의 메트릭 기반

```yaml
apiVersion: autoscaling/v2beta2
kind: HorizontalPodAutoscaler
metadata:
  name: auth-service-hpa
  namespace: production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: auth-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

---

## 4. 운영 모니터링 및 로깅

- **Prometheus & Grafana**:  
  - 배포 상태, 리소스 사용량, 헬스체크 결과 모니터링
- **ELK Stack** 또는 **Cloud Logging**:  
  - 애플리케이션 로그 집계 및 분석
- **Alerting**:  
  - Deployment 실패, 리소스 부족, HPA 이벤트 발생 시 알림 설정

---

## 5. 배포 전략 및 롤백

### 5.1 카나리 배포

- 새 버전을 소규모로 배포하여 모니터링 후 전체 트래픽 전환  
- 실패 시 자동 롤백 설정

### 5.2 블루/그린 배포

- 새로운 버전(그린)과 기존 버전(블루)을 병행 배포, 트래픽 전환 후 구 버전 제거

### 5.3 롤백 전략

- Kubernetes `rollout undo` 명령어를 활용한 수동 롤백  
- CI/CD에서 배포 실패 시 자동 롤백 정책 적용

---

## 6. 결론

**ImmersiVerse Authentication Service**의 Kubernetes 배포 구성은 다음과 같은 원칙을 따릅니다:

- **안정성**: 최소 3개의 복제본, Readiness/Liveness Probe, HPA 설정을 통한 고가용성 확보  
- **보안**: Ingress의 TLS, Secret을 통한 민감 정보 보호  
- **유연성**: ConfigMap 및 Secret 관리, Canary/Blue-Green 배포 전략 적용  
- **모니터링**: Prometheus, Grafana, ELK를 통한 실시간 모니터링 및 알림

이 가이드를 통해 팀원들은 Kubernetes 환경에서 Auth Service를 효과적으로 배포·운영할 수 있으며, 문제 발생 시 신속하게 대응할 수 있습니다. 필요한 경우 각 항목을 상황에 맞게 업데이트하고, CI/CD 파이프라인과 연계하여 자동화된 배포 환경을 유지해 주세요.