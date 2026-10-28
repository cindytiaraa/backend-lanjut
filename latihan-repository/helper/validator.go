package helper

import (
	"errors"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	_ = v.RegisterValidation("nim", func(fl validator.FieldLevel) bool {
		return isValidNIM(fl.Field().String())
	})

	// httpurl hanya menerima http/https. Tag bawaan "url" juga menerima skema
	// seperti javascript: atau file:, berbahaya bila tautan ditampilkan client.
	_ = v.RegisterValidation("httpurl", func(fl validator.FieldLevel) bool {
		u, err := url.Parse(fl.Field().String())
		return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
	})

	_ = v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		return passwordStrength(fl.Field().String()) == ""
	})

	return v
}

var nimPattern = regexp.MustCompile(`^[0-9]{8,15}$`)

func isValidNIM(nim string) bool {
	if !nimPattern.MatchString(nim) {
		return false
	}

	tahun := 0
	for _, r := range nim[:4] {
		tahun = tahun*10 + int(r-'0')
	}

	return tahun >= 2000 && tahun <= 2099
}

func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return map[string]string{"_": "objek yang divalidasi tidak sah"}
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return map[string]string{"_": "validasi gagal"}
	}

	result := make(map[string]string, len(fieldErrors))
	for _, fe := range fieldErrors {
		key := fieldKey(fe)
		if _, exists := result[key]; !exists {
			result[key] = messageFor(fe)
		}
	}

	return result
}

// fieldKey memberi nama field yang menyertakan posisinya pada data bersarang,
// misalnya "projects[1].title". Field() saja hanya menghasilkan "title",
// sehingga galat dua proyek berbeda saling menimpa.
func fieldKey(fe validator.FieldError) string {
	ns := fe.Namespace()
	if i := strings.Index(ns, "."); i >= 0 {
		ns = ns[i+1:]
	}
	if ns == "" {
		return fe.Field()
	}
	return ns
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		switch fe.Kind() {
		case reflect.String:
			return "minimal " + fe.Param() + " karakter"
		case reflect.Slice, reflect.Array, reflect.Map:
			return "minimal " + fe.Param() + " item"
		}
		return "nilai minimal " + fe.Param()
	case "max":
		switch fe.Kind() {
		case reflect.String:
			return "maksimal " + fe.Param() + " karakter"
		case reflect.Slice, reflect.Array, reflect.Map:
			return "maksimal " + fe.Param() + " item"
		}
		return "nilai maksimal " + fe.Param()
	case "datetime":
		return "format tanggal harus YYYY-MM-DD"
	case "httpurl":
		return "harus berupa tautan http:// atau https://"
	case "alphanum":
		return "hanya boleh berisi huruf dan angka"
	case "nim":
		return "NIM harus 8-15 digit angka dengan tahun angkatan yang wajar"
	case "strongpassword":
		if value, ok := fe.Value().(string); ok {
			return passwordStrength(value)
		}
		return "password tidak memenuhi syarat"
	case "oneof":
		return "harus salah satu dari: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	default:
		return "tidak memenuhi aturan " + fe.Tag()
	}
}

var commonPasswords = map[string]bool{
	"password":    true,
	"password123": true,
	"12345678":    true,
	"qwerty123":   true,
	"admin123":    true,
}

func passwordStrength(password string) string {
	if len(password) < 8 {
		return "minimal 8 karakter"
	}

	if commonPasswords[strings.ToLower(password)] {
		return "password terlalu umum"
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
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

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return "harus memiliki huruf besar, huruf kecil, angka, dan karakter khusus"
	}

	return ""
}
