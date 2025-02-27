package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuditLog(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	entityID := "entity123"
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"
	oldValues := map[string]interface{}{"status": "ACTIVE"}
	newValues := map[string]interface{}{"status": "SUSPENDED"}

	tests := []struct {
		name           string
		userID         *string
		action         string
		entityType     string
		entityID       *string
		oldValues      map[string]interface{}
		newValues      map[string]interface{}
		ipAddress      *string
		userAgent      *string
		wantErr        bool
		wantIDNotEmpty bool
		wantCreatedAt  bool
	}{
		{
			name:           "Valid audit log creation",
			userID:         &userID,
			action:         "UPDATE_USER",
			entityType:     "USER",
			entityID:       &entityID,
			oldValues:      oldValues,
			newValues:      newValues,
			ipAddress:      &ipAddress,
			userAgent:      &userAgent,
			wantErr:        false,
			wantIDNotEmpty: true,
			wantCreatedAt:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := NewAuditLog(tt.userID, tt.action, tt.entityType, tt.entityID, tt.oldValues, tt.newValues, tt.ipAddress, tt.userAgent)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, log)

			// ID 검증
			if tt.wantIDNotEmpty {
				assert.NotEmpty(t, log.ID())
			}

			// 필드 값 검증
			assert.Equal(t, tt.userID, log.UserID())
			assert.Equal(t, tt.action, log.Action())
			assert.Equal(t, tt.entityType, log.EntityType())
			assert.Equal(t, tt.entityID, log.EntityID())
			assert.Equal(t, tt.oldValues, log.OldValues())
			assert.Equal(t, tt.newValues, log.NewValues())
			assert.Equal(t, tt.ipAddress, log.IPAddress())
			assert.Equal(t, tt.userAgent, log.UserAgent())

			// 시간 관련 검증
			if tt.wantCreatedAt {
				assert.False(t, log.CreatedAt().IsZero())
			}
		})
	}
}
