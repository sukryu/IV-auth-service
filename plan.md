# ImmersiVerse Authentication Service 개발 계획

## 1. 프로젝트 기초 설정
- 디렉토리 구조 설정 (완료)
- 설정 파일 구성 (완료)
- Makefile 정의 (완료)
- Git 저장소 초기화 및 .gitignore 설정 (완료)

## 2. 데이터베이스 구성
- DB 마이그레이션 파일 작성 (완료)
- 마이그레이션 실행 스크립트 설정 (완료)
- 데이터베이스 연결 설정 (완료)
- 저장소(Repository) 인터페이스 정의

## 3. 도메인 모델 구현
- 주요 엔티티 정의 (User, PlatformAccount, Token 등) (완료)
- 값 객체(Value Objects) 구현 (Password, Email 등) (완료)
- 도메인 서비스 구현 (UserDomainService, AuthenticationDomainService 등) (완료)
- 도메인 이벤트 정의 (UserCreated, LoginSucceeded 등) (완료)

## 4. 저장소(Repository) 구현
- 데이터베이스 접근 로직 구현
- PostgreSQL 저장소 구현
- 캐싱 레이어 추가 (Redis)
- 테스트용 Mock 저장소 구현

## 5. 인프라 구성
- Kafka 연결 및 이벤트 발행 설정
- Redis 설정 (토큰 블랙리스트, 캐싱)
- 외부 플랫폼 OAuth 통합 (Twitch, YouTube 등)
- 로깅 설정 (구조화된 로깅)

## 6. gRPC 서비스 정의
- Proto 파일 작성 (인증, 사용자 관리, 플랫폼 계정 관리)
- gRPC 코드 생성
- 인터셉터 구현 (로깅, 인증, 검증)

## 7. 서비스 구현
- 인증 서비스 (로그인, 로그아웃, 토큰 관리)
- 사용자 서비스 (CRUD 작업)
- 플랫폼 계정 서비스 (연결, 해제, 조회)
- 이벤트 발행 로직 통합

## 8. 보안 구현
- JWT 토큰 생성 및 검증
- 비밀번호 해싱 (Argon2id)
- 속도 제한 (Rate Limiting)
- 토큰 블랙리스트 관리

## 9. 테스트 구현
- 단위 테스트 작성
- 통합 테스트 작성
- 성능 테스트 구현
- 테스트 자동화 설정

## 10. 배포 설정
- Docker 컨테이너화
- Kubernetes 매니페스트 작성
- CI/CD 파이프라인 구성 (GitHub Actions)
- 모니터링 및 로깅 설정 (Prometheus, Grafana)

## 11. 문서화
- API 문서 작성
- 개발 가이드 작성
- 배포 및 운영 가이드 작성
- 보안 및 감사 로깅 정책 문서화