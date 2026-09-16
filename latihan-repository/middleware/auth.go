package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"latihan-repository/app/model"
	"latihan-repository/helper"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		if authHeader == "" {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)

			return helper.Fail(
				c,
				fiber.StatusUnauthorized,
				"token wajib disertakan",
			)
		}

		parts := strings.Fields(authHeader)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)

			return helper.Fail(
				c,
				fiber.StatusUnauthorized,
				"format authorization tidak valid",
			)
		}

		claims, err := jwtManager.ParseAccess(parts[1])

		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)

			if errors.Is(err, jwt.ErrTokenExpired) {
				return helper.Fail(
					c,
					fiber.StatusUnauthorized,
					"access token kedaluwarsa",
				)
			}

			return helper.Fail(
				c,
				fiber.StatusUnauthorized,
				"access token tidak valid",
			)
		}

		authUser := model.AuthUser{
			UserID:   claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		}

		c.Locals(helper.LocalsAuthUser, authUser)

		return c.Next()
	}
}
