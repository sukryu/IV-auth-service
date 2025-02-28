package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sukryu/IV-auth-services/internal/core/domain"
	"go.uber.org/zap"
)

// auditLogRepository implements domain.AuditLogRepository for PostgreSQL.
type auditLogRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewAuditLogRepository creates a new auditLogRepository instance.
func NewAuditLogRepository(db *pgxpool.Pool, logger *zap.Logger) domain.AuditLogRepository {
	return &auditLogRepository{
		db:     db,
		logger: logger.With(zap.String("component", "audit_log_repository")),
	}
}

// LogAction logs an action to the audit_logs table.
func (r *auditLogRepository) LogAction(ctx context.Context, log *domain.AuditLog) error {
	if log == nil {
		return errors.New("audit log must not be nil")
	}

	oldValuesJSON, _ := json.Marshal(log.OldValues()) // 오류 무시 (필수 아님)
	newValuesJSON, _ := json.Marshal(log.NewValues())

	query := `
        INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `
	_, err := r.db.Exec(ctx, query,
		log.ID(),
		log.UserID(),
		log.Action(),
		log.EntityType(),
		log.EntityID(),
		oldValuesJSON,
		newValuesJSON,
		log.IPAddress(),
		log.UserAgent(),
		log.CreatedAt(),
	)
	if err != nil {
		r.logger.Error("Failed to log audit action", zap.Error(err), zap.String("audit_id", log.ID()))
		return errors.New("failed to log audit action: " + err.Error())
	}
	r.logger.Debug("Audit action logged successfully", zap.String("audit_id", log.ID()))
	return nil
}

// GetLogs retrieves audit logs by user ID (optional) with pagination.
func (r *auditLogRepository) GetLogs(ctx context.Context, userID string, limit, offset int) ([]*domain.AuditLog, error) {
	query := `
        SELECT id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
        FROM audit_logs
        WHERE ($1::text IS NULL OR user_id = $1)
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.Query(ctx, query, nullableString(userID), limit, offset)
	if err != nil {
		r.logger.Error("Failed to retrieve audit logs", zap.Error(err), zap.String("user_id", userID))
		return nil, errors.New("failed to retrieve audit logs: " + err.Error())
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var (
			id         string
			userID     sql.NullString
			action     string
			entityType string
			entityID   sql.NullString
			oldValues  []byte
			newValues  []byte
			ipAddress  sql.NullString
			userAgent  sql.NullString
			createdAt  time.Time
		)
		if err := rows.Scan(&id, &userID, &action, &entityType, &entityID, &oldValues, &newValues, &ipAddress, &userAgent, &createdAt); err != nil {
			r.logger.Error("Failed to scan audit log row", zap.Error(err))
			return nil, errors.New("failed to scan audit log: " + err.Error())
		}

		var oldValuesMap, newValuesMap map[string]interface{}
		json.Unmarshal(oldValues, &oldValuesMap)
		json.Unmarshal(newValues, &newValuesMap)

		log, err := domain.NewAuditLog(id, action, entityType, nullableStringPtr(userID), nullableStringPtr(entityID), nullableStringPtr(ipAddress), nullableStringPtr(userAgent), oldValuesMap, newValuesMap)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("error iterating audit log rows: " + err.Error())
	}
	return logs, nil
}

// nullableString converts a string to a nullable pointer for SQL.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nullableStringPtr converts sql.NullString to a pointer.
func nullableStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
