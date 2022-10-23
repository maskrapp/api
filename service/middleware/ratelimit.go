package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/models"
	"github.com/maskrapp/backend/ratelimit"
)

func UserRateLimit(rl *ratelimit.RateLimiter) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		userId := c.Locals("user_id").(string)
		limited := rl.CheckUserRateLimit(userId, path)
		if limited {
			return c.Status(429).JSON(&models.APIResponse{
				Success: false,
				Message: "You are being rate limited",
			})
		}
		return c.Next()
	}
}
