# 인증 유틸리티 사용 가이드

**pkg/auth** 패키지는 ImmersiVerse Authentication Service에서 JWT 토큰 생성, 검증, 파싱 및 기타 인증 관련 기능을 제공하는 유틸리티 모듈입니다. 이 문서를 통해 해당 패키지의 주요 기능과 사용 방법, 그리고 예제 코드를 확인할 수 있습니다.

---

## 1. 목적

- **JWT 생성 및 검증**:  
  - 사용자 인증 후 발급되는 Access Token과 Refresh Token을 안전하게 생성 및 검증
- **토큰 파싱**:  
  - 클라이언트 요청 시 전달된 JWT의 클레임을 파싱하여 사용자 정보 및 권한 확인
- **보안 강화**:  
  - RS256 알고리즘을 활용한 비대칭 서명 및 토큰 무결성 검증 지원
- **공통 인터페이스 제공**:  
  - 다른 서비스 및 인터셉터에서 인증 관련 기능을 쉽게 호출할 수 있도록 API를 추상화

---

## 2. 주요 기능

### 2.1 JWT 생성

- **CreateToken(userID string, roles []string, duration time.Duration)**  
  - 사용자 ID, 역할 정보, 토큰 유효 기간을 입력받아 JWT를 생성합니다.
  - 내부적으로 RS256 알고리즘을 사용하여 서명합니다.

### 2.2 JWT 검증

- **ValidateToken(tokenString string) (Claims, error)**  
  - JWT 문자열을 입력받아 서명 및 만료 여부, 블랙리스트 상태 등을 검증합니다.
  - 유효한 토큰의 경우, 토큰에 포함된 클레임(Claims)을 반환합니다.

### 2.3 토큰 파싱

- **ParseToken(tokenString string) (Claims, error)**  
  - JWT의 페이로드 부분을 파싱하여, 사용자 ID, 발급 시각, 만료 시각 등 주요 클레임 정보를 추출합니다.

### 2.4 키 관리

- **LoadKeys(privateKeyPath, publicKeyPath string)**  
  - JWT 생성 및 검증에 필요한 개인키 및 공개키를 파일에서 로드합니다.
  - 키는 안전한 위치(Vault/HSM 등)에서 관리할 수 있도록 권장합니다.

---

## 3. 사용 예제

아래는 **pkg/auth** 패키지의 주요 함수를 사용하는 간단한 예제 코드입니다.

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/immersiverse/auth-service/pkg/auth"
)

func main() {
	// 키 파일 경로 (환경 변수나 시크릿 관리 도구에서 주입)
	privateKeyPath := "./config/jwt-private.pem"
	publicKeyPath := "./config/jwt-public.pem"

	// 키 로드
	err := auth.LoadKeys(privateKeyPath, publicKeyPath)
	if err != nil {
		log.Fatalf("키 로드 실패: %v", err)
	}

	// 사용자 ID와 역할 정보를 기반으로 Access Token 생성
	userID := "uuid-1234"
	roles := []string{"USER", "STREAMER"}
	duration := 15 * time.Minute

	token, err := auth.CreateToken(userID, roles, duration)
	if err != nil {
		log.Fatalf("토큰 생성 실패: %v", err)
	}
	fmt.Printf("발급된 토큰: %s\n", token)

	// 토큰 검증
	claims, err := auth.ValidateToken(token)
	if err != nil {
		log.Fatalf("토큰 검증 실패: %v", err)
	}
	fmt.Printf("토큰 유효: 사용자 ID=%s, 역할=%v\n", claims.Sub, claims.Roles)
}
```

---

## 4. 베스트 프랙티스

- **환경 변수 관리**:  
  - 키 파일 경로, 토큰 유효 기간 등 민감 정보는 코드에 하드코딩하지 않고, 환경 변수 또는 안전한 시크릿 관리 도구를 통해 주입합니다.
  
- **에러 처리**:  
  - 토큰 생성 및 검증 과정에서 발생하는 에러는 로깅하고, 적절한 HTTP 상태 코드(예: 401 Unauthorized)를 반환하도록 처리합니다.
  
- **주기적 키 순환**:  
  - 보안을 강화하기 위해 주기적으로 키를 교체하고, 새로운 키가 적용된 토큰은 `kid`(키 식별자)를 포함하도록 구현합니다.

- **테스트**:  
  - 각 함수에 대해 단위 테스트와 통합 테스트를 작성하여, JWT 생성, 검증, 파싱 기능의 정확성을 보장합니다.

---

## 5. 결론

**pkg/auth** 패키지는 **Authentication Service**의 핵심 인증 기능을 담당하며, JWT 생성 및 검증, 토큰 파싱, 키 관리 등의 공통 유틸리티를 제공합니다. 이 가이드를 따라, 팀원들은 인증 관련 작업을 일관되게 구현하고, 보안 및 성능 측면에서도 높은 수준의 인증 시스템을 유지할 수 있습니다.

---