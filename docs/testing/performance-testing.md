# 성능 테스트 가이드

본 문서는 **ImmersiVerse Authentication Service**의 성능을 평가하기 위한 테스트 전략과 실행 방법을 설명합니다. 성능 테스트는 시스템의 응답 시간, 처리량, 부하 한계 및 안정성을 검증하여, 운영 환경에서의 원활한 서비스 제공을 보장하기 위한 필수 절차입니다.

---

## 1. 성능 테스트 목표

- **응답 시간**:  
  - 실시간 인증 요청에 대한 평균 응답 시간이 100ms 이하
  - gRPC 호출의 95번째 백분위 응답 시간 200ms 이하
- **처리량**:  
  - 동시 사용자가 최대 60명(또는 예측 부하)인 상황에서 안정적으로 동작하는지 검증
- **시스템 안정성**:  
  - 장시간 부하 테스트 시, 메모리 누수, CPU 과부하 및 장애 발생 없이 지속 운영 가능
- **장애 복구**:  
  - 부하 과다 시 자동 확장 및 장애 복구(RTO/RPO) 목표 달성

---

## 2. 테스트 환경 구성

### 2.1 외부 의존 서비스

- **PostgreSQL**, **Redis**, **Kafka** 등은 Docker Compose를 사용해 로컬 또는 테스트 클러스터에서 실행합니다.
- 실제 운영 환경과 유사한 설정을 구성하여, 테스트 결과가 신뢰할 수 있도록 합니다.

### 2.2 테스트 환경 변수

- 별도의 `.env.performance` 파일을 사용해 성능 테스트에 적합한 환경 변수(DB, Kafka, Redis 주소 등)를 설정합니다.

```dotenv
DB_HOST=localhost
DB_PORT=5432
DB_USER=auth_user
DB_PASS=auth_password
DB_NAME=auth_db_perf
REDIS_ADDR=localhost:6379
KAFKA_BROKER=localhost:9092
```

---

## 3. 성능 테스트 도구

- **JMeter / Locust / k6**: 부하 및 스트레스 테스트 도구로 활용  
  - k6: 스크립트 기반 부하 테스트 도구로, HTTP 및 gRPC 테스트 지원
  - JMeter: GUI를 통한 테스트 시나리오 구성 및 실행
- **Go Benchmark**: 단위 성능 테스트(예: 로그인 처리 시간) 실행

---

## 4. 테스트 시나리오

### 4.1 부하 테스트 (Load Test)

- **목표**: 예상 동시 사용자(예: 50~60명) 하에서 시스템 응답 시간, 처리량, 에러율 등을 측정
- **시나리오**:
  - 사용자들이 로그인, 회원가입, 토큰 갱신 등의 API를 지속적으로 호출
  - 각 API의 응답 시간, CPU/메모리 사용량, DB 쿼리 지연 등을 기록
- **측정 지표**:
  - 평균/최대 응답 시간
  - 초당 처리 요청 수(TPS)
  - 에러 발생률

### 4.2 스트레스 테스트 (Stress Test)

- **목표**: 시스템의 최대 한계를 파악하고, 과부하 상황에서의 동작 및 복구 능력 확인
- **시나리오**:
  - 부하를 점진적으로 증가시키며, 시스템이 어느 시점에서 성능 저하나 장애를 겪는지 모니터링
  - 장애 발생 후 자동 확장(HPA) 및 복구 절차 테스트
- **측정 지표**:
  - 장애 임계치(최대 TPS, 최대 동시 사용자 수)
  - 장애 발생 시 응답 시간 및 에러율 변화
  - 복구 시간(RTO)

### 4.3 엔드투엔드 성능 테스트 (End-to-End Performance Test)

- **목표**: 전체 인증 흐름(회원가입 → 로그인 → 토큰 갱신 → 로그아웃)의 지연 시간과 시스템 부하 평가
- **시나리오**:
  - 사용자가 순차적으로 여러 API를 호출하며 전체 흐름을 측정
  - 각 단계별(로그인, 토큰 발급, 로그아웃 등) 지연 시간 및 이벤트 발행/처리 시간을 측정
- **측정 지표**:
  - 전체 흐름 지연 시간
  - 단계별 응답 시간

---

## 5. 테스트 실행 방법

### 5.1 로컬 및 CI 환경

- **로컬 테스트**: Docker Compose 환경에서 각 성능 테스트 도구를 실행하여 결과 확인
- **CI/CD 통합**: GitHub Actions에 성능 테스트 스크립트를 포함하여, 코드 변경 시 성능 지표가 기준치 이하인지 확인

### 5.2 k6 예제 스크립트

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '1m', target: 50 }, // ramp-up to 50 users
        { duration: '5m', target: 50 }, // sustain 50 users
        { duration: '1m', target: 0 }   // ramp-down
    ],
    thresholds: {
        http_req_duration: ['p(95)<200'], // 95% of requests should complete below 200ms
    },
};

export default function () {
    let res = http.post('https://auth.immersiverse.com/auth/login', JSON.stringify({
        username: 'testuser',
        password: 'StrongP@ssw0rd!'
    }), {
        headers: { 'Content-Type': 'application/json' },
    });
    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time is below 200ms': (r) => r.timings.duration < 200,
    });
    sleep(1);
}
```

---

## 6. 결과 분석 및 보고

- **메트릭 수집**:  
  - Prometheus에서 주요 메트릭 수집 (응답 시간, TPS, 에러율 등)
- **대시보드**:  
  - Grafana를 통해 실시간 및 히스토리 데이터 시각화
- **보고서 작성**:  
  - 테스트 결과 요약, 병목 현상, 개선 사항 도출
  - CI/CD 빌드 결과에 성능 테스트 통합

---

## 7. 최적화 및 지속적 개선

- 테스트 결과를 바탕으로 애플리케이션, DB 쿼리, 캐싱, 오토스케일링 설정 등을 지속적으로 최적화
- 정기적으로 성능 테스트를 재실행하여, 트래픽 증가나 시스템 변경에 따른 영향을 모니터링

---

## 8. 결론

**성능 테스트**는 **Authentication Service**가 높은 트래픽과 부하 상황에서도 안정적으로 동작할 수 있도록 보장하는 핵심 단계입니다.  
- **부하 테스트, 스트레스 테스트, 엔드투엔드 테스트**를 통해 다양한 시나리오를 검증하고, 주요 성능 지표(응답 시간, TPS, 에러율)를 모니터링합니다.  
- 이를 CI/CD 파이프라인에 통합하여 코드 변경 시 즉각적인 성능 회귀를 확인하고, 지속적인 최적화를 진행합니다.

이 가이드를 기반으로 팀원들은 성능 테스트 시나리오를 작성하고 실행하여, 서비스의 확장성과 안정성을 확보해 주세요.

---