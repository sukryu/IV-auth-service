package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/IV-auth-services/internal/core/domain"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    domain.Email // 타입 명확히
		wantErr bool
	}{
		{"Valid email", "test@example.com", domain.Email("test@example.com"), false},
		{"Empty email", "", "", true},
		{"Invalid format", "not-an-email", "", true},
		{"Long email", strings.Repeat("a", 256) + "@example.com", "", true},
		{"Whitespace trimmed", "  test@example.com  ", domain.Email("test@example.com"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := domain.NewEmail(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, email)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, email) // 타입 일치로 비교 성공
				assert.True(t, email.IsValid())
				assert.Equal(t, "example.com", email.Domain())
			}
		})
	}
}
