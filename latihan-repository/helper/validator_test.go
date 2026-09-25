package helper

import "testing"

func TestPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantOK   bool
	}{
		{"kuat", "Admin123!", true},
		{"terlalu pendek", "Admin1!", false},
		{"tanpa huruf besar", "admin123!", false},
		{"tanpa huruf kecil", "ADMIN123!", false},
		{"tanpa angka", "AdminTest!", false},
		{"tanpa karakter khusus", "Admin1234", false},
		{"password umum", "Password123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := passwordStrength(tt.password) == ""
			if got != tt.wantOK {
				t.Errorf("passwordStrength(%q) valid=%v, expected %v", tt.password, got, tt.wantOK)
			}
		})
	}
}

func TestIsValidNIM(t *testing.T) {
	tests := []struct {
		nim  string
		want bool
	}{
		{"20230001", true},     
		{"202300012345", true}, 
		{"abc12345", false},    
		{"2023001", false},     
		{"99990001", false},    
	}

	for _, tt := range tests {
		if got := isValidNIM(tt.nim); got != tt.want {
			t.Errorf("isValidNIM(%q) = %v, expected %v", tt.nim, got, tt.want)
		}
	}
}
