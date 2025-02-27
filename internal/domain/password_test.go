package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPassword(t *testing.T) {
	tests := []struct {
		name       string
		plainText  string
		wantErr    bool
		wantVerify bool
	}{
		{
			name:       "Valid password",
			plainText:  "StrongP@ssw0rd!",
			wantErr:    false,
			wantVerify: true,
		},
		{
			name:      "Too short password",
			plainText: "short",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pwd, err := NewPassword(tt.plainText)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, pwd.String())

			// 해시 검증
			valid, err := pwd.Verify(tt.plainText)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantVerify, valid)

			// 잘못된 비밀번호로 검증
			invalid, err := pwd.Verify("wrongpassword")
			assert.NoError(t, err)
			assert.False(t, invalid)
		})
	}
}

func TestPasswordVerifyInvalidFormat(t *testing.T) {
	// 잘못된 형식의 해시로 검증 시도
	pwd := Password("invalid_hash_format")
	valid, err := pwd.Verify("anytext")
	assert.Error(t, err)
	assert.False(t, valid)
	assert.Contains(t, err.Error(), "invalid password hash format")
}
