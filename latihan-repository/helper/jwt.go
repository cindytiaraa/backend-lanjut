package helper

import (
	"errors"
	"time"

	"latihan-repository/app/model"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

type AccessClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTManager(
	secret string,
	issuer string,
	accessTTL time.Duration,
) *JWTManager {
	return &JWTManager{
		secret:    []byte(secret),
		issuer:    issuer,
		accessTTL: accessTTL,
	}
}

func (m *JWTManager) GenerateAccess(user model.User) (string, error) {
	now := time.Now()

	claims := AccessClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(m.secret)
}

func (m *JWTManager) ParseAccess(tokenString string) (*AccessClaims, error) {
	claims := &AccessClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("algoritma JWT tidak diizinkan")
			}

			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token tidak valid")
	}

	return claims, nil
}
