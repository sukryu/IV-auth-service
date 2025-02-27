package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserCreated(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		username string
		email    string
		wantErr  bool
	}{
		{
			name:     "Valid UserCreated event",
			userID:   "550e8400-e29b-41d4-a716-446655440000",
			username: "testuser",
			email:    "test@example.com",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewUserCreated(tt.userID, tt.username, tt.email)
			assert.NotNil(t, event)
			assert.Equal(t, "UserCreated", event.Name())
			assert.False(t, event.OccurredAt().IsZero())
			assert.Equal(t, tt.userID, event.UserID())
			assert.Equal(t, tt.username, event.Username())
			assert.Equal(t, tt.email, event.Email())
		})
	}
}

func TestLoginSucceeded(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{
			name:    "Valid LoginSucceeded event",
			userID:  "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewLoginSucceeded(tt.userID)
			assert.NotNil(t, event)
			assert.Equal(t, "LoginSucceeded", event.Name())
			assert.False(t, event.OccurredAt().IsZero())
			assert.Equal(t, tt.userID, event.UserID())
		})
	}
}
