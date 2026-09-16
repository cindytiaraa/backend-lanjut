package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	) == nil
}

func GenerateRandomToken(byteLength int) (string, error) {
	if byteLength < 16 {
		byteLength = 16
	}

	buf := make([]byte, byteLength)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func SHA256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

var dummyHash = []byte("$2a$12$okNPpmIEVYuprl4JFMzlI.93TJQkIVZjuhhuVd97lVzDUz7dEaDWu")

func VerifyDummyPassword(plain string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(plain))
}
