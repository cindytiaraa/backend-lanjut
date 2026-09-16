package service

import "testing"

func TestIsStrongPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "strong password",
			password: "Admin123!",
			expected: true,
		},
		{
			name:     "too short",
			password: "Admin1!",
			expected: false,
		},
		{
			name:     "no uppercase",
			password: "admin123!",
			expected: false,
		},
		{
			name:     "no lowercase",
			password: "ADMIN123!",
			expected: false,
		},
		{
			name:     "no number",
			password: "AdminTest!",
			expected: false,
		},
		{
			name:     "no special character",
			password: "Admin1234",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsStrongPassword(tt.password)

			if got != tt.expected {
				t.Errorf(
					"IsStrongPassword(%q) = %v, expected %v",
					tt.password,
					got,
					tt.expected,
				)
			}
		})
	}
}
