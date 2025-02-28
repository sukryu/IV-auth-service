package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// PostgresAuditLogRepository implements AuditLogRepository with PostgreSQL
type PostgresAuditLogRepository struct {
	db *sql.DB
}

// NewPostgresAuditLogRepository creates a new instance
func NewPostgresAuditLogRepository(db *sql.DB) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{db: db}
}

// InsertAuditLog adds a new audit log entry to the database
func (r *PostgresAuditLogRepository) InsertAuditLog(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	// JSONB로 저장하기 위해 map을 []byte로 변환
	oldValuesJSON, err := json.Marshal(log.OldValues())
	if err != nil {
		return fmt.Errorf("failed to marshal old_values: %w", err)
	}
	newValuesJSON, err := json.Marshal(log.NewValues())
	if err != nil {
		return fmt.Errorf("failed to marshal new_values: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		log.ID(), log.UserID(), log.Action(), log.EntityType(), log.EntityID(),
		oldValuesJSON, newValuesJSON, log.IPAddress(), log.UserAgent(), log.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}
	return nil
}

// GetAuditLogsByUserID retrieves audit logs for a specific user
func (r *PostgresAuditLogRepository) GetAuditLogsByUserID(ctx context.Context, userID string) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
		FROM audit_logs WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var id, action, entityType string
		var userIDVal, entityID, ipAddress, userAgent *string
		var oldValuesJSON, newValuesJSON []byte
		var createdAt time.Time

		err := rows.Scan(&id, &userIDVal, &action, &entityType, &entityID, &oldValuesJSON, &newValuesJSON, &ipAddress, &userAgent, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		var oldValues, newValues map[string]interface{}
		if oldValuesJSON != nil {
			if err := json.Unmarshal(oldValuesJSON, &oldValues); err != nil {
				return nil, fmt.Errorf("failed to unmarshal old_values: %w", err)
			}
		}
		if newValuesJSON != nil {
			if err := json.Unmarshal(newValuesJSON, &newValues); err != nil {
				return nil, fmt.Errorf("failed to unmarshal new_values: %w", err)
			}
		}

		log := domain.NewAuditLogFromDB(id, userIDVal, action, entityType, entityID, oldValues, newValues, ipAddress, userAgent, createdAt)
		logs = append(logs, log)
	}
	return logs, nil
}
