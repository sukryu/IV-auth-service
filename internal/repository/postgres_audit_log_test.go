package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestPostgresAuditLogRepositoryGetAuditLogsByUserID(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresAuditLogRepository(db)

	now := time.Now()
	userID := "user123"
	oldValues := map[string]interface{}{"key": "old"}
	newValues := map[string]interface{}{"key": "new"}
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	rows := sqlmock.NewRows([]string{"id", "user_id", "action", "entity_type", "entity_id", "old_values", "new_values", "ip_address", "user_agent", "created_at"}).
		AddRow("log123", userID, "CREATE_USER", "USER", "entity123", oldValuesJSON, newValuesJSON, "192.168.1.1", "Mozilla/5.0", now)
	mock.ExpectQuery("SELECT id, user_id, action").WithArgs(userID).WillReturnRows(rows)

	logs, err := repo.GetAuditLogsByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "log123", logs[0].ID()) // 기대값 "log123"과 일치
	assert.Equal(t, &userID, logs[0].UserID())
	assert.Equal(t, "CREATE_USER", logs[0].Action())
	assert.Equal(t, "USER", logs[0].EntityType())
	assert.Equal(t, oldValues, logs[0].OldValues())
	assert.Equal(t, newValues, logs[0].NewValues())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresAuditLogRepositoryGetAuditLogsByUserIDEmpty(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresAuditLogRepository(db)

	mock.ExpectQuery("SELECT id, user_id, action").WithArgs("user456").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	logs, err := repo.GetAuditLogsByUserID(context.Background(), "user456")
	assert.NoError(t, err)
	assert.Empty(t, logs)
	assert.NoError(t, mock.ExpectationsWereMet())
}
