# 저장소 구현 설명

본 문서는 **ImmersiVerse Authentication Service**의 저장소(Repository) 계층 설계 및 구현에 대해 설명합니다. 저장소 계층은 도메인 모델과 인프라스트럭처(DB, 캐시 등) 사이에서 데이터 접근을 캡슐화하여, 비즈니스 로직과 데이터 저장소 간의 결합도를 낮추고 테스트 용이성을 높이는 역할을 수행합니다.

---

## 1. 저장소 계층의 목적

- **데이터 추상화**:  
  - DB 접근, 쿼리 실행, 트랜잭션 관리를 캡슐화하여 도메인 로직에서 직접 SQL 문을 다루지 않도록 합니다.
  
- **테스트 용이성**:  
  - 인터페이스 기반의 저장소 구현을 통해 모킹(mocking) 및 스텁(stub)을 사용, 단위 테스트와 통합 테스트에서 실제 DB 대신 가상의 저장소를 사용 가능합니다.
  
- **유지보수성 및 확장성**:  
  - 저장소 관련 로직을 한 곳에 집중시켜, 스키마 변경, 쿼리 최적화, 캐싱 정책 등의 수정 시 영향 범위를 최소화합니다.

---

## 2. 저장소 계층 구성

### 2.1 주요 인터페이스

- **UserRepository**:  
  - 사용자 관련 CRUD(Create, Read, Update, Delete) 작업을 담당합니다.
  - 주요 메서드 예시:  
    - `InsertUser(ctx, user *domain.User) error`
    - `GetUserByID(ctx, id string) (*domain.User, error)`
    - `UpdateUser(ctx, user *domain.User) error`
    - `DeleteUser(ctx, id string) error`
    - `ExistsByUsername(ctx, username string) (bool, error)`

- **PlatformAccountRepository**:  
  - 외부 플랫폼 계정 연동 데이터의 저장 및 조회를 담당합니다.
  - 주요 메서드 예시:
    - `InsertPlatformAccount(ctx, account *domain.PlatformAccount) error`
    - `GetPlatformAccountsByUserID(ctx, userID string) ([]*domain.PlatformAccount, error)`
    - `DeletePlatformAccount(ctx, id string) error`

- **TokenRepository**:  
  - JWT 토큰 관련 데이터(예: 토큰 블랙리스트)의 저장 및 조회를 담당합니다.
  - 주요 메서드 예시:
    - `InsertTokenBlacklist(ctx, tokenID string, expiresAt time.Time, reason string) error`
    - `IsTokenBlacklisted(ctx, tokenID string) (bool, error)`

- **AuditLogRepository**:  
  - 감사 로그 데이터를 기록하고 조회하는 기능을 제공합니다.
  - 주요 메서드 예시:
    - `InsertAuditLog(ctx, log *domain.AuditLog) error`
    - `GetAuditLogsByUserID(ctx, userID string) ([]*domain.AuditLog, error)`

### 2.2 구현 방식

- **SQL 템플릿**:  
  - SQL 문은 하드코딩보다는 템플릿(예: Go의 `text/template` 또는 외부 라이브러리)을 사용해 작성하여 유지보수성을 높입니다.
  
- **쿼리 최적화**:  
  - 자주 사용되는 쿼리, JOIN, 인덱스 활용 등을 고려하여 최적화하며, `EXPLAIN ANALYZE` 등 도구로 성능을 점검합니다.
  
- **트랜잭션 관리**:  
  - 복수의 DB 작업이 필요한 경우, 트랜잭션을 사용해 원자성 보장
  - 예: 회원가입 시 사용자 생성과 감사 로그 기록이 모두 성공해야 하는 경우

- **모듈화 및 패키지화**:  
  - 각 저장소 구현은 별도의 패키지로 구성하여, 도메인 계층 및 서비스 계층에서 쉽게 주입(Dependency Injection)할 수 있도록 설계합니다.

---

## 3. 예제 코드

### 3.1 UserRepository 인터페이스 정의 (Pseudo Go)

```go
package repository

import (
    "context"
    "time"

    "github.com/immersiverse/auth-service/internal/domain"
)

// UserRepository는 사용자 데이터에 대한 CRUD 작업을 정의합니다.
type UserRepository interface {
    InsertUser(ctx context.Context, user *domain.User) error
    GetUserByID(ctx context.Context, id string) (*domain.User, error)
    GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
    UpdateUser(ctx context.Context, user *domain.User) error
    DeleteUser(ctx context.Context, id string) error
    ExistsByUsername(ctx context.Context, username string) (bool, error)
}
```

### 3.2 PostgreSQL 기반 구현 예시 (Pseudo Go)

```go
package repository

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/immersiverse/auth-service/internal/domain"
)

type PostgresUserRepository struct {
    db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
    return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) InsertUser(ctx context.Context, user *domain.User) error {
    query := `
        INSERT INTO users (id, username, email, password_hash, status, subscription_tier, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
    _, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.Email, user.PasswordHash, user.Status, user.SubscriptionTier, user.CreatedAt, user.UpdatedAt)
    return err
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
    query := `
        SELECT id, username, email, password_hash, status, subscription_tier, created_at, updated_at, last_login_at
        FROM users WHERE id = $1
    `
    row := r.db.QueryRowContext(ctx, query, id)
    var user domain.User
    if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Status, &user.SubscriptionTier, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt); err != nil {
        return nil, err
    }
    return &user, nil
}

// 나머지 메서드들도 유사하게 구현...
```

---

## 4. 테스트 및 모킹

- **테스트 전략**:  
  - 각 저장소 인터페이스에 대해 단위 테스트를 작성하여, SQL 쿼리 실행, 트랜잭션 관리, 에러 처리 등을 검증합니다.
  - 실제 DB 대신 모의 객체(mock)를 활용해 테스트를 수행합니다.
  
- **테스트 도구**:  
  - `sqlmock` 라이브러리를 활용하여 DB 연동 없이 쿼리 결과를 시뮬레이션합니다.
  
예시:
```go
func TestInsertUser(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("sqlmock error: %v", err)
    }
    defer db.Close()

    repo := NewPostgresUserRepository(db)
    user := &domain.User{
        ID:               "uuid-1234",
        Username:         "testuser",
        Email:            "testuser@example.com",
        PasswordHash:     "hashedPwd",
        Status:           "ACTIVE",
        SubscriptionTier: "FREE",
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
    }

    mock.ExpectExec("INSERT INTO users").
        WithArgs(user.ID, user.Username, user.Email, user.PasswordHash, user.Status, user.SubscriptionTier, user.CreatedAt, user.UpdatedAt).
        WillReturnResult(sqlmock.NewResult(1, 1))

    err = repo.InsertUser(context.Background(), user)
    assert.NoError(t, err)
}
```

---

## 5. 결론

저장소(Repository) 계층은 **Authentication Service**의 데이터 접근을 캡슐화하며, 도메인 로직과 인프라스트럭처 간의 결합도를 낮추는 핵심 역할을 합니다.  
- **인터페이스 기반 설계**로 모킹 및 테스트 용이성을 확보하고,  
- **SQL 쿼리 최적화**와 **트랜잭션 관리**를 통해 데이터 무결성과 성능을 보장합니다.

이 문서를 기반으로 저장소 구현 및 테스트 전략을 지속적으로 개선하여, 서비스 확장과 운영 안정성을 높여 나가시기 바랍니다.

---