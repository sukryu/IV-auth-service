package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry for tracking actions
type AuditLog struct {
	id         string
	userID     *string
	action     string
	entityType string
	entityID   *string
	oldValues  map[string]interface{}
	newValues  map[string]interface{}
	ipAddress  *string
	userAgent  *string
	createdAt  time.Time
}

// NewAuditLog creates a new AuditLog instance
func NewAuditLog(userID *string, action, entityType string, entityID *string, oldValues, newValues map[string]interface{}, ipAddress, userAgent *string) (*AuditLog, error) {
	id := uuid.New().String()
	now := time.Now()

	return &AuditLog{
		id:         id,
		userID:     userID,
		action:     action,
		entityType: entityType,
		entityID:   entityID,
		oldValues:  oldValues,
		newValues:  newValues,
		ipAddress:  ipAddress,
		userAgent:  userAgent,
		createdAt:  now,
	}, nil
}

// newAuditLogFromDB creates an AuditLog instance with a specified ID (for database retrieval)
func NewAuditLogFromDB(id string, userID *string, action, entityType string, entityID *string, oldValues, newValues map[string]interface{}, ipAddress, userAgent *string, createdAt time.Time) *AuditLog {
	return &AuditLog{
		id:         id,
		userID:     userID,
		action:     action,
		entityType: entityType,
		entityID:   entityID,
		oldValues:  oldValues,
		newValues:  newValues,
		ipAddress:  ipAddress,
		userAgent:  userAgent,
		createdAt:  createdAt,
	}
}

// Getters
func (a *AuditLog) ID() string                        { return a.id }
func (a *AuditLog) UserID() *string                   { return a.userID }
func (a *AuditLog) Action() string                    { return a.action }
func (a *AuditLog) EntityType() string                { return a.entityType }
func (a *AuditLog) EntityID() *string                 { return a.entityID }
func (a *AuditLog) OldValues() map[string]interface{} { return a.oldValues }
func (a *AuditLog) NewValues() map[string]interface{} { return a.newValues }
func (a *AuditLog) IPAddress() *string                { return a.ipAddress }
func (a *AuditLog) UserAgent() *string                { return a.userAgent }
func (a *AuditLog) CreatedAt() time.Time              { return a.createdAt }
