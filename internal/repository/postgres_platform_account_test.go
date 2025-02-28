package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/IV-auth-services/internal/domain"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestPostgresPlatformAccountRepositoryInsertPlatformAccount(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPlatformAccountRepository(db)
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)
	account, _ := domain.NewPlatformAccount("user123", "TWITCH", "twitch123", "TwitchUser", "access_token", "refresh_token", &expiresAt)

	mock.ExpectExec("INSERT INTO platform_accounts").
		WithArgs(account.ID(), account.UserID(), account.Platform(), account.PlatformUserID(), account.PlatformUsername(),
			account.AccessToken(), account.RefreshToken(), account.TokenExpiresAt(), account.CreatedAt(), account.UpdatedAt()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.InsertPlatformAccount(context.Background(), account)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresPlatformAccountRepositoryGetPlatformAccountsByUserID(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPlatformAccountRepository(db)

	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)
	rows := sqlmock.NewRows([]string{"id", "user_id", "platform", "platform_user_id", "platform_username", "access_token", "refresh_token", "token_expires_at", "created_at", "updated_at"}).
		AddRow("acc123", "user123", "TWITCH", "twitch123", "TwitchUser", "access_token", "refresh_token", &expiresAt, now, now)
	mock.ExpectQuery("SELECT id, user_id, platform").WithArgs("user123").WillReturnRows(rows)

	accounts, err := repo.GetPlatformAccountsByUserID(context.Background(), "user123")
	assert.NoError(t, err)
	assert.Len(t, accounts, 1)
	assert.Equal(t, "acc123", accounts[0].ID())
	assert.Equal(t, "user123", accounts[0].UserID())
	assert.Equal(t, "TWITCH", accounts[0].Platform())
	assert.Equal(t, "twitch123", accounts[0].PlatformUserID())
	assert.Equal(t, "TwitchUser", accounts[0].PlatformUsername())
	assert.Equal(t, "access_token", accounts[0].AccessToken())
	assert.Equal(t, "refresh_token", accounts[0].RefreshToken())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresPlatformAccountRepositoryGetPlatformAccountsByUserIDEmpty(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPlatformAccountRepository(db)

	mock.ExpectQuery("SELECT id, user_id, platform").WithArgs("user456").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	accounts, err := repo.GetPlatformAccountsByUserID(context.Background(), "user456")
	assert.NoError(t, err)
	assert.Empty(t, accounts)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresPlatformAccountRepositoryUpdatePlatformAccount(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresPlatformAccountRepository(db)
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)
	account, _ := domain.NewPlatformAccount("user123", "TWITCH", "twitch123", "TwitchUser", "old_access", "old_refresh", &expiresAt)
	account.SetTokens("new_access", "new_refresh", &expiresAt)

	mock.ExpectExec("UPDATE platform_accounts SET").
		WithArgs(account.AccessToken(), account.RefreshToken(), account.TokenExpiresAt(), account.UpdatedAt(), account.ID()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdatePlatformAccount(context.Background(), account)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
