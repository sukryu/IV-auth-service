package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/IV-auth-services/configs" // import 경로 수정
	"github.com/sukryu/IV-auth-services/internal/domain"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func setupTestConfig(t *testing.T) func() {
	// 임시 파일 생성
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	// 테스트용 설정 작성
	configContent := `
db:
  host: "localhost"
  port: 5432
  user: "auth_user"
  password: "auth_password"
  name: "auth_db"
  dsn: "postgres://auth_user:auth_password@localhost:5432/auth_db?sslmode=disable"
`
	err = os.WriteFile(tmpFile.Name(), []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// configs 패키지의 ConfigPath를 임시 파일로 설정
	originalPath := configs.ConfigPath
	configs.ConfigPath = tmpFile.Name()

	// 테스트 종료 시 정리 및 원래 경로 복구
	return func() {
		os.Remove(tmpFile.Name())
		configs.ConfigPath = originalPath
	}
}

func TestPostgresUserRepositoryInsertUser(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	user, _ := domain.NewUser("testuser", "test@example.com", "hashedpassword")

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID(), user.Username(), user.Email(), user.PasswordHash(), user.Status(), user.SubscriptionTier(), user.CreatedAt(), user.UpdatedAt(), user.LastLoginAt()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.InsertUser(context.Background(), user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepositoryGetUserByUsername(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "status", "subscription_tier", "created_at", "updated_at", "last_login_at"}).
		AddRow("user123", "testuser", "test@example.com", "hashedpassword", "ACTIVE", "FREE", now, now, nil)
	mock.ExpectQuery("SELECT id, username, email").WithArgs("testuser").WillReturnRows(rows)

	user, err := repo.GetUserByUsername(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user123", user.ID())
	assert.Equal(t, "testuser", user.Username())
	assert.Equal(t, "test@example.com", user.Email())
	assert.Equal(t, "hashedpassword", user.PasswordHash())
	assert.Equal(t, "ACTIVE", user.Status())
	assert.Equal(t, "FREE", user.SubscriptionTier())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepositoryGetUserByUsernameNotFound(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)

	mock.ExpectQuery("SELECT id, username, email").WithArgs("nonexistent").WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByUsername(context.Background(), "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}
