package domain

import (
	"errors"
	"time"
)

// AuditLog represents an audit record of a significant action in the system.
type AuditLog struct {
	id         string
	userID     *string // nullable
	action     string
	entityType string
	entityID   *string // nullable
	oldValues  map[string]interface{}
	newValues  map[string]interface{}
	ipAddress  *string // nullable
	userAgent  *string // nullable
	createdAt  time.Time
}

// NewAuditLog creates a new AuditLog instance.
func NewAuditLog(id, action, entityType string, userID, entityID, ipAddress, userAgent *string, oldValues, newValues map[string]interface{}) (*AuditLog, error) {
	if id == "" {
		return nil, errors.New("audit log id must not be empty")
	}
	if action == "" {
		return nil, errors.New("action must not be empty")
	}
	if entityType == "" {
		return nil, errors.New("entity type must not be empty")
	}

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
		createdAt:  time.Now(),
	}, nil
}

// ID returns the audit log's unique identifier.
func (a *AuditLog) ID() string {
	return a.id
}

// UserID returns the ID of the user who performed the action, or nil if not applicable.
func (a *AuditLog) UserID() *string {
	return a.userID
}

// Action returns the action performed.
func (a *AuditLog) Action() string {
	return a.action
}

// EntityType returns the type of entity affected.
func (a *AuditLog) EntityType() string {
	return a.entityType
}

// EntityID returns the ID of the affected entity, or nil if not applicable.
func (a *AuditLog) EntityID() *string {
	return a.entityID
}

// OldValues returns the previous state of the entity, if applicable.
func (a *AuditLog) OldValues() map[string]interface{} {
	return a.oldValues
}

// NewValues returns the new state of the entity, if applicable.
func (a *AuditLog) NewValues() map[string]interface{} {
	return a.newValues
}

// IPAddress returns the IP address of the requester, or nil if not recorded.
func (a *AuditLog) IPAddress() *string {
	return a.ipAddress
}

// UserAgent returns the user agent string of the requester, or nil if not recorded.
func (a *AuditLog) UserAgent() *string {
	return a.userAgent
}

// CreatedAt returns the time the audit log was created.
func (a *AuditLog) CreatedAt() time.Time {
	return a.createdAt
}
