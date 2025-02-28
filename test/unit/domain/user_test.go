package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/IV-auth-services/internal/core/domain"
)

func TestNewUser(t *testing.T) {
	validEmail, _ := domain.NewEmail("test@example.com")
	validPassword, _ := domain.NewPassword("StrongP@ssw0rd!")

	tests := []struct {
		name             string
		id               string
		username         string
		email            domain.Email
		passwordHash     domain.Password
		roleIDs          []string
		subscriptionTier string
		wantErr          bool
	}{
		{
			name:             "Valid user",
			id:               "user-123",
			username:         "testuser",
			email:            validEmail,
			passwordHash:     validPassword,
			roleIDs:          []string{"role-1"},
			subscriptionTier: "FREE",
			wantErr:          false,
		},
		{
			name:             "Empty ID",
			id:               "",
			username:         "testuser",
			email:            validEmail,
			passwordHash:     validPassword,
			roleIDs:          []string{"role-1"},
			subscriptionTier: "FREE",
			wantErr:          true,
		},
		{
			name:             "Empty username",
			id:               "user-123",
			username:         "",
			email:            validEmail,
			passwordHash:     validPassword,
			roleIDs:          []string{"role-1"},
			subscriptionTier: "FREE",
			wantErr:          true,
		},
		{
			name:             "Invalid email",
			id:               "user-123",
			username:         "testuser",
			email:            "", // 유효하지 않은 Email
			passwordHash:     validPassword,
			roleIDs:          []string{"role-1"},
			subscriptionTier: "FREE",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.NewUser(tt.id, tt.username, tt.email, tt.passwordHash, tt.roleIDs, tt.subscriptionTier)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.id, user.ID())
				assert.Equal(t, tt.username, user.Username())
				assert.Equal(t, tt.email, user.Email())
				assert.Equal(t, tt.roleIDs, user.RoleIDs())
				assert.Equal(t, domain.UserStatusActive, user.Status())
				assert.Equal(t, tt.subscriptionTier, user.SubscriptionTier())
				assert.WithinDuration(t, time.Now(), user.CreatedAt(), time.Second)
			}
		})
	}
}

func TestUserSetStatus(t *testing.T) {
	email, _ := domain.NewEmail("test@example.com")
	password, _ := domain.NewPassword("StrongP@ssw0rd!")
	user, _ := domain.NewUser("user-123", "testuser", email, password, nil, "FREE")

	tests := []struct {
		name    string
		status  domain.UserStatus
		wantErr bool
	}{
		{"Set to Suspended", domain.UserStatusSuspended, false},
		{"Set to Deleted", domain.UserStatusDeleted, false},
		{"Invalid status", "INVALID", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.SetStatus(tt.status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.status, user.Status())
				assert.WithinDuration(t, time.Now(), user.UpdatedAt(), time.Second)
			}
		})
	}
}
