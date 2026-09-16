package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"latihan-repository/helper"
)

type loginAttempt struct {
	count       int
	windowStart time.Time
}

var (
	loginAttempts = make(map[string]loginAttempt)
	loginMu       sync.Mutex
)

const (
	maxLoginAttempts = 5
	loginWindow      = 5 * time.Minute
)

func LoginRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()

		loginMu.Lock()

		attempt, exists := loginAttempts[ip]

		if !exists || now.Sub(attempt.windowStart) >= loginWindow {
			attempt = loginAttempt{
				count:       0,
				windowStart: now,
			}
		}

		if attempt.count >= maxLoginAttempts {
			remaining := int(loginWindow - now.Sub(attempt.windowStart))

			if remaining < 1 {
				remaining = 1
			}

			c.Set("Retry-After", "300")

			loginMu.Unlock()

			return helper.Fail(
				c,
				fiber.StatusTooManyRequests,
				"terlalu banyak percobaan login",
			)
		}

		attempt.count++
		loginAttempts[ip] = attempt

		loginMu.Unlock()

		return c.Next()
	}
}
