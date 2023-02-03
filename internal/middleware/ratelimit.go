package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"github.com/sirupsen/logrus"
)

//TODO: add ratelimit headers

func UserRateLimit(ctx global.Context, maxRequests int, cooldown time.Duration, next func(*fiber.Ctx) error) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		userId := c.Locals("user_id").(string)
		limited := ctx.Instances().RateLimiter.CheckRateLimit(ctx, userId, path, maxRequests, cooldown)
		if limited {
			return c.Status(429).JSON(&models.APIResponse{
				Success: false,
				Message: "You are being rate limited",
			})
		}
		return next(c)
	}
}

func EmailRateLimit(ctx global.Context, maxRequests int, cooldown time.Duration, next func(*fiber.Ctx) error) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		var body map[string]string
		if err := c.BodyParser(&body); err != nil {
			return c.SendStatus(400)
		}
		email, ok := body["email"]
		if !ok {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		limited := ctx.Instances().RateLimiter.CheckRateLimit(ctx, email, path, maxRequests, cooldown)
		if limited {
			return c.Status(429).JSON(&models.APIResponse{
				Success: false,
				Message: "You are being rate limited",
			})
		}
		return next(c)
	}
}

func IPRateLimit(ctx global.Context, maxRequests int, cooldown time.Duration, next func(*fiber.Ctx) error) func(*fiber.Ctx) error {
	if ctx.Config().Production {
		return func(c *fiber.Ctx) error {
			headers := c.GetReqHeaders()
			ip, ok := headers["X-Real-Ip"]
			if !ok {
				logrus.Error("header 'X-Real-Ip' is missing!")
				return c.Status(500).JSON(&models.APIResponse{
					Success: false,
					Message: "Something went wrong",
				})
			}
			limited := ctx.Instances().RateLimiter.CheckRateLimit(ctx, ip, c.Path(), maxRequests, cooldown)
			if limited {
				return c.Status(429).JSON(&models.APIResponse{
					Success: false,
					Message: "You are being rate limited",
				})
			}
			return next(c)
		}
	}
	return func(c *fiber.Ctx) error {
		ip := c.Context().RemoteAddr().String()
		limited := ctx.Instances().RateLimiter.CheckRateLimit(ctx, ip, c.Path(), maxRequests, cooldown)
		if limited {
			return c.Status(429).JSON(&models.APIResponse{
				Success: false,
				Message: "You are being rate limited",
			})
		}
		return next(c)
	}
}
