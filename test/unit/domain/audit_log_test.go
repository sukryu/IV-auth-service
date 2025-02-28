package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/IV-auth-services/internal/core/domain"
)

func TestNewAuditLog(t *testing.T) {
	userID := "user-123"
	entityID := "entity-456"
	ip := "192.168.1.1"
	ua := "Mozilla/5.0"
	oldValues := map[string]interface{}{"status": "ACTIVE"}
	newValues := map[string]interface{}{"status": "SUSPENDED"}

	tests := []struct {
		name       string
		id         string
		action     string
		entityType string
		userID     *string
		entityID   *string
		ipAddress  *string
		userAgent  *string
		oldValues  map[string]interface{}
		newValues  map[string]interface{}
		wantErr    bool
	}{
		{
			name:       "Valid audit log",
			id:         "log-123",
			action:     "CHANGE_STATUS",
			entityType: "USER",
			userID:     &userID,
			entityID:   &entityID,
			ipAddress:  &ip,
			userAgent:  &ua,
			oldValues:  oldValues,
			newValues:  newValues,
			wantErr:    false,
		},
		{
			name:       "Empty ID",
			id:         "",
			action:     "CHANGE_STATUS",
			entityType: "USER",
			userID:     &userID,
			entityID:   &entityID,
			ipAddress:  &ip,
			userAgent:  &ua,
			oldValues:  oldValues,
			newValues:  newValues,
			wantErr:    true,
		},
		{
			name:       "Empty action",
			id:         "log-123",
			action:     "",
			entityType: "USER",
			userID:     &userID,
			entityID:   &entityID,
			ipAddress:  &ip,
			userAgent:  &ua,
			oldValues:  oldValues,
			newValues:  newValues,
			wantErr:    true,
		},
		{
			name:       "Empty entityType",
			id:         "log-123",
			action:     "CHANGE_STATUS",
			entityType: "",
			userID:     &userID,
			entityID:   &entityID,
			ipAddress:  &ip,
			userAgent:  &ua,
			oldValues:  oldValues,
			newValues:  newValues,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := domain.NewAuditLog(tt.id, tt.action, tt.entityType, tt.userID, tt.entityID, tt.ipAddress, tt.userAgent, tt.oldValues, tt.newValues)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, log)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, log)
				assert.Equal(t, tt.id, log.ID())
				assert.Equal(t, tt.action, log.Action())
				assert.Equal(t, tt.entityType, log.EntityType())
				if tt.userID != nil {
					assert.Equal(t, *tt.userID, *log.UserID())
				}
				if tt.entityID != nil {
					assert.Equal(t, *tt.entityID, *log.EntityID())
				}
				if tt.ipAddress != nil {
					assert.Equal(t, *tt.ipAddress, *log.IPAddress())
				}
				if tt.userAgent != nil {
					assert.Equal(t, *tt.userAgent, *log.UserAgent())
				}
				assert.Equal(t, tt.oldValues, log.OldValues())
				assert.Equal(t, tt.newValues, log.NewValues())
				assert.WithinDuration(t, time.Now(), log.CreatedAt(), time.Second)
			}
		})
	}
}
