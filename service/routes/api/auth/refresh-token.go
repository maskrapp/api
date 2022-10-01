package auth

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/jwt"
	"github.com/maskrapp/backend/models"
)

func RefreshToken(jwtHandler *jwt.JWTHandler) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.SendStatus(500)
		}
		val, ok := body["refresh_token"]
		if !ok {
			return c.Status(400).SendString("Invalid Body")
		}
		refreshToken := val.(string)
		claims, err := jwtHandler.Validate(refreshToken, true)
		if err != nil {
			return c.Status(401).JSON(&models.APIResponse{
				Success: false,
				Message: "Unauthorized",
			})
		}
		jwt, err := jwtHandler.GenerateAccessToken(claims.UserId)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		return c.JSON(jwt)
	}
}
