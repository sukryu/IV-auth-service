package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestPostgresTokenRepositoryInsertTokenBlacklist(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTokenRepository(db)
	tokenID := "token123"
	userID := "user123"
	expiresAt := time.Now().Add(15 * time.Minute)
	reason := "logout"

	mock.ExpectExec("INSERT INTO token_blacklist").
		WithArgs(tokenID, userID, expiresAt, reason, sqlmock.AnyArg()). // blacklisted_at는 현재 시간으로 변동 가능
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.InsertTokenBlacklist(context.Background(), tokenID, userID, expiresAt, reason)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTokenRepositoryIsTokenBlacklistedTrue(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTokenRepository(db)

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("token123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.IsTokenBlacklisted(context.Background(), "token123")
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTokenRepositoryIsTokenBlacklistedFalse(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTokenRepository(db)

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("token456").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.IsTokenBlacklisted(context.Background(), "token456")
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}
