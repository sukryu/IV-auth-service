package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		wantErr   bool
		wantEmail string
	}{
		{
			name:      "Valid email",
			email:     "Test.User@Example.com",
			wantErr:   false,
			wantEmail: "test.user@example.com",
		},
		{
			name:      "Valid email with whitespace",
			email:     "  user@example.com  ",
			wantErr:   false,
			wantEmail: "user@example.com",
		},
		{
			name:    "Invalid email format",
			email:   "invalid-email",
			wantErr: true,
		},
		{
			name:    "Too long email",
			email:   "a" + strings.Repeat("b", 255) + "@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantEmail, email.String())
		})
	}
}
