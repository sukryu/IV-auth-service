package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockUserRepository is a mock implementation of UserRepository
type mockUserRepository struct {
	users map[string]*User
}

func (m *mockUserRepository) InsertUser(ctx context.Context, user *User) error {
	m.users[user.Username()] = user
	return nil
}

func (m *mockUserRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user, exists := m.users[username]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *mockUserRepository) UpdateUser(ctx context.Context, user *User) error {
	m.users[user.Username()] = user
	return nil
}

func (m *mockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	_, exists := m.users[username]
	return exists, nil
}

func TestUserDomainServiceCreateUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		password string
		wantErr  bool
		wantUser bool
	}{
		{
			name:     "Valid user creation",
			username: "testuser",
			email:    "test@example.com",
			password: "StrongP@ssw0rd!",
			wantErr:  false,
			wantUser: true,
		},
		{
			name:     "Duplicate username",
			username: "existinguser",
			email:    "new@example.com",
			password: "StrongP@ssw0rd!",
			wantErr:  true,
		},
		{
			name:     "Invalid password",
			username: "newuser",
			email:    "new@example.com",
			password: "short",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{users: map[string]*User{}}
			if tt.name == "Duplicate username" {
				repo.users["existinguser"] = &User{username: "existinguser"}
			}
			s := NewUserDomainService(repo)

			user, err := s.CreateUser(context.Background(), tt.username, tt.email, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.wantUser {
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username())
				assert.Equal(t, tt.email, user.Email())
				assert.Equal(t, "ACTIVE", user.Status())
			}
		})
	}
}

func TestUserDomainServiceChangeUserStatus(t *testing.T) {
	repo := &mockUserRepository{users: map[string]*User{}}
	s := NewUserDomainService(repo)

	// 초기 사용자 생성
	user, err := NewUser("testuser", "test@example.com", "hashedpassword")
	assert.NoError(t, err)
	repo.users["testuser"] = user

	err = s.ChangeUserStatus(context.Background(), "testuser", "SUSPENDED")
	assert.NoError(t, err)
	assert.Equal(t, "SUSPENDED", user.Status())

	// 존재하지 않는 사용자 테스트
	err = s.ChangeUserStatus(context.Background(), "nonexistent", "SUSPENDED")
	assert.Error(t, err)
	assert.Equal(t, errUserNotFound, err)
}

func TestUserDomainServiceRecordLogin(t *testing.T) {
	repo := &mockUserRepository{users: map[string]*User{}}
	s := NewUserDomainService(repo)

	// 초기 사용자 생성
	user, err := NewUser("testuser", "test@example.com", "hashedpassword")
	assert.NoError(t, err)
	repo.users["testuser"] = user

	err = s.RecordLogin(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.NotNil(t, user.LastLoginAt())

	// 존재하지 않는 사용자 테스트
	err = s.RecordLogin(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, errUserNotFound, err)
}
