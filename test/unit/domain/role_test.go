package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/IV-auth-services/internal/core/domain"
)

func TestNewRole(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		roleName    string
		description string
		wantErr     bool
	}{
		{"Valid role", "role-1", "ADMIN", "Administrator role", false},
		{"Empty ID", "", "USER", "Regular user", true},
		{"Empty name", "role-1", "", "No name", true},
		{"Invalid name", "role-1", "INVALID", "Invalid role", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := domain.NewRole(tt.id, tt.roleName, tt.description)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, role)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.id, role.ID())
				assert.Equal(t, tt.roleName, role.Name())
				assert.Equal(t, tt.description, role.Description())
			}
		})
	}
}
