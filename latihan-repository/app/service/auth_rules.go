package service

import (
	"strings"
	"unicode"

	"latihan-repository/app/model"
)

func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper bool
	var hasLower bool
	var hasNumber bool
	var hasSpecial bool

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		default:
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func ValidateRegister(req model.RegisterRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.Username) == "" {
		errs["username"] = "wajib diisi"
	}

	if strings.TrimSpace(req.Email) == "" {
		errs["email"] = "wajib diisi"
	}

	if strings.TrimSpace(req.Password) == "" {
		errs["password"] = "wajib diisi"
	} else if !IsStrongPassword(req.Password) {
		errs["password"] = "minimal 8 karakter, harus memiliki huruf besar, huruf kecil, angka, dan karakter khusus"
	}

	return errs
}

func ValidateLogin(req model.LoginRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.Username) == "" {
		errs["username"] = "wajib diisi"
	}

	if req.Password == "" {
		errs["password"] = "wajib diisi"
	}

	return errs
}