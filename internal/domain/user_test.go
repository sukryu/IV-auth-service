package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	tests := []struct {
		name            string
		username        string
		email           string
		passwordHash    string
		wantErr         bool
		wantIDNotEmpty  bool
		wantStatus      string
		wantTier        string
		wantCreatedAt   bool
		wantUpdatedAt   bool
		wantLastLoginAt *time.Time
	}{
		{
			name:            "Valid user creation",
			username:        "testuser",
			email:           "test@example.com",
			passwordHash:    "hashedpassword",
			wantErr:         false,
			wantIDNotEmpty:  true,
			wantStatus:      "ACTIVE",
			wantTier:        "FREE",
			wantCreatedAt:   true,
			wantUpdatedAt:   true,
			wantLastLoginAt: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.username, tt.email, tt.passwordHash)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, user)

			// ID 검증
			if tt.wantIDNotEmpty {
				assert.NotEmpty(t, user.ID())
			}

			// 필드 값 검증
			assert.Equal(t, tt.username, user.Username())
			assert.Equal(t, tt.email, user.Email())
			assert.Equal(t, tt.passwordHash, user.PasswordHash())
			assert.Equal(t, tt.wantStatus, user.Status())
			assert.Equal(t, tt.wantTier, user.SubscriptionTier())

			// 시간 관련 검증
			if tt.wantCreatedAt {
				assert.False(t, user.CreatedAt().IsZero())
			}
			if tt.wantUpdatedAt {
				assert.False(t, user.UpdatedAt().IsZero())
			}
			assert.Equal(t, tt.wantLastLoginAt, user.LastLoginAt())
		})
	}
}

func TestUserSetStatus(t *testing.T) {
	user, err := NewUser("testuser", "test@example.com", "hashedpassword")
	assert.NoError(t, err)

	originalUpdatedAt := user.UpdatedAt()
	time.Sleep(1 * time.Millisecond) // updatedAt 변경 확인을 위해 약간 대기

	user.SetStatus("SUSPENDED")
	assert.Equal(t, "SUSPENDED", user.Status())
	assert.True(t, user.UpdatedAt().After(originalUpdatedAt))
}

func TestUserSetSubscriptionTier(t *testing.T) {
	user, err := NewUser("testuser", "test@example.com", "hashedpassword")
	assert.NoError(t, err)

	originalUpdatedAt := user.UpdatedAt()
	time.Sleep(1 * time.Millisecond)

	user.SetSubscriptionTier("PREMIUM")
	assert.Equal(t, "PREMIUM", user.SubscriptionTier())
	assert.True(t, user.UpdatedAt().After(originalUpdatedAt))
}

func TestUserRecordLogin(t *testing.T) {
	user, err := NewUser("testuser", "test@example.com", "hashedpassword")
	assert.NoError(t, err)

	originalUpdatedAt := user.UpdatedAt()
	time.Sleep(1 * time.Millisecond)

	user.RecordLogin()
	assert.NotNil(t, user.LastLoginAt())
	assert.True(t, user.LastLoginAt().After(originalUpdatedAt))
	assert.Equal(t, user.LastLoginAt(), &user.updatedAt) // 포인터 값 동일성 확인
}
