package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/IV-auth-services/internal/core/domain"
)

func TestNewPassword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid password", "StrongP@ssw0rd!", false},
		{"Empty password", "", true},
		{"Short password", "short", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pwd, err := domain.NewPassword(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, pwd)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, pwd.Hash())
				assert.NotEmpty(t, pwd.Salt())
				assert.True(t, pwd.Verify(tt.input))
				assert.False(t, pwd.Verify("wrong"))
			}
		})
	}
}
