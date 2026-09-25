package helper

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return 0, false
	}

	return id, true
}

func RequestID(c *fiber.Ctx) string {
	id, _ := c.Locals("requestid").(string)
	return id
}
