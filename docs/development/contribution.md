# 기여 가이드라인 (Contribution Guide)

이 문서는 **ImmersiVerse Authentication Service**에 기여하고자 하는 모든 팀원(개발자, QA, 문서 담당 등)을 위한 지침을 제공합니다.  
- **목표**: 협업 시 일관된 개발 프로세스, 코드 품질, 문서화 수준을 보장해 원활한 프로젝트 운영을 지원.

---

## 1. 개요

- **Repo**: [https://github.com/immersiverse/auth-service](https://github.com/immersiverse/auth-service) (예시)
- **언어**: Go 1.21+
- **메인 브랜치**: `main`
- **CI/CD**: GitHub Actions + ArgoCD
- **Issue Tracker**: GitHub Issues, Kanban board

---

## 2. 이슈 생성

1. **이슈 유형**:
   - Bug Report
   - Feature Request
   - Improvement / Refactoring
   - Documentation
2. **이슈 등록 규칙**:
   - 문제 상황(재현 방법), 기대 동작, 스크린샷/로그 등 상세 설명
   - 라벨(label) 지정: `bug`, `enhancement`, `question`, `documentation`
   - 중복 이슈 없는지 검색 후 생성

---

## 3. 브랜치 전략

- **메인 브랜치**(`main`): 항상 배포 가능한 안정 버전 유지
- **개발 브랜치**(옵션): 일부 팀은 `develop` 브랜치를 둘 수 있음
- **기능 브랜치**: `feature/<이슈번호>`, `bugfix/<이슈번호>` 형식
  - 예: `feature/123-add-logout-api`, `bugfix/234-fix-db-conn`
- **릴리스 브랜치**: `release/v1.2.0` 등 (버전에 따라 운영)
- **Hotfix 브랜치**: `hotfix/<이슈번호>`

---

## 4. 개발 프로세스

1. **이슈 선택**: GitHub Issues에서 작업할 이슈를 골라 **Assignee** 설정
2. **브랜치 생성**: `git checkout -b feature/123-new-oauth-flow`
3. **개발 & 커밋**: 규칙 준수(아래 커밋 메시지 가이드 참조)
4. **로컬 테스트**: `make test`, `make integration-test` 등
5. **PR 생성**: `main`(또는 `develop`)으로 Merge 요청, 이슈번호/설명 포함
6. **리뷰 & 수정**: 리뷰어 피드백 반영
7. **머지 후**: CI/CD가 자동 빌드·배포(스테이징 → 프로덕션)

---

## 5. 커밋 메시지 가이드

- **형식**: `<타입>: <간단한 요약>` + 본문(옵션)
  - **타입 예**: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`
- **예시**:
  ```
  feat: add refresh token API

  - implement refresh token method in AuthService
  - add integration tests for token refresh
  ```
- **K8s 스타일**: 원문(영문) 선호 시 영문으로 작성, 상세 본문은 필요 시 추가.

---

## 6. Pull Request 규칙

1. **PR 제목**: 이슈번호 + 요약 (ex: `#123 Add Refresh Token API`)
2. **Description**:
   - 변경 사항(무엇을, 왜)
   - 참고 이슈, 스크린샷 등
   - 테스트 방법 & 결과
3. **체크리스트**:
   - [ ] 단위 테스트 통과  
   - [ ] 통합 테스트(옵션)  
   - [ ] 린트/포맷 OK  
   - [ ] 문서(필요 시) 업데이트
4. **CI 자동화**:
   - PR 생성 시 `make lint`, `make test`, 커버리지 검증
5. **리뷰어**: 1~2명 배정, LGTM(**Looks Good To Me**) 후 머지

---

## 7. 리뷰 & 머지

- **Code Review**: 2인 이상 승인이 권장
- **K8s 스타일**: "LGTM", "Approved" 라벨로 머지 결정
- 작은 피드백은 추가 커밋 or `git --fixup` 등 사용
- Merge 시점에서 **Rebase**(squash) or 병합 결정 (팀 합의에 따라)

---

## 8. 테스트 작성

- **단위 테스트**: table-driven, `_test.go` suffix
- **통합 테스트**: Docker Compose 기반, DB/Redis/Kafka 실제 사용
- **커버리지 목표**: 80% 이상
- PR 시 Test 결과가 CI에서 모두 Green 이어야.

(자세한 내용은 [테스트 작성 가이드](./testing.md) 참고)

---

## 9. 문서 기여

- 문서 변경(PR):
  - Markdown 파일(.md) 수정 시 [docs/](../) 디렉토리 내
  - 예) `docs/development/setup.md` 업데이트, `docs/architecture/overview.md` 추가
- 번역/개정 시 문서 유지보수에 주의.  
- README 등 주요 문서에 큰 변경은 리뷰어 최소 1인 이상 검토 필요

---

## 10. 릴리스 프로세스

1. **버전 태그**: `v1.2.0` 등 SemVer
2. **Change Log** 작성(이슈/PR 자동 링크 or 수동 정리)
3. **Draft Release** → QA 테스트 완료 후 **Publish Release**
4. ArgoCD or CI/CD로 프로덕션 배포

(자세한 내용은 [API 버전 관리 전략](../api/versioning.md) 참고)

---

## 11. 코드 오브 컨덕트 (Code of Conduct)

- 팀원 간 협업 시 상호 존중, 친절한 의사소통
- 차별 발언/성희롱/인신공격 금지
- K8s Community 가이드에 준하는 기본 윤리준수

---

## 12. 결론

이 **기여 가이드**를 통해, **Authentication Service** 발전에 참여하는 모든 팀원은 일관된 협업 절차를 준수하고, 고품질 코드를 기여할 수 있습니다.

1. **Issue** 생성/할당 → **Branch** → **Commit** → **PR** → **Review** → **Merge**  
2. 충분한 **테스트**와 **문서 업데이트**는 필수  
3. 코드 품질과 협업 효율을 위해 **lint**, **CI**, **review** 절차를 자동화