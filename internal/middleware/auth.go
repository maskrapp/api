package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
)

func AuthMiddleware(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		auth := c.GetReqHeaders()["Authorization"]
		if auth == "" {
			return c.SendStatus(401)
		}
		split := strings.Split(auth, " ")
		if len(split) != 2 {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid authorization header",
			})
		}
		accessToken := split[1]
		claims, err := ctx.Instances().JWT.Validate(accessToken, false)
		if err != nil {
			if strings.Contains(err.Error(), "token mismatch") {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "Token mismatch",
				})
			}
			if strings.Contains(err.Error(), "token is expired by") {
				return c.Status(401).JSON(&models.APIResponse{
					Success: false,
					Message: "Token is expired",
				})
			}
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
			})

		}
		c.Locals("user_id", claims.UserId)
		return c.Next()
	}
}
