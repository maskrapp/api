package auth

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/jwt"
	"github.com/maskrapp/backend/models"
	dbmodels "github.com/maskrapp/common/models"
	"gorm.io/gorm"
)

//TODO: harden this
func RefreshToken(jwtHandler *jwt.JWTHandler, db *gorm.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.SendStatus(401)
		}
		val, ok := body["refresh_token"]
		if !ok {
			return c.Status(401).SendString("Invalid Body")
		}
		refreshToken := val.(string)
		claims, err := jwtHandler.Validate(refreshToken, true)
		if err != nil {
			return c.Status(401).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid token",
			})
		}
		user := &dbmodels.User{}
		err = db.First(user, "id = ?", claims.UserId).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(401).JSON(&models.APIResponse{
					Success: false,
					Message: "The user that is associated with your token no longer exists",
				})
			}
			return c.Status(401).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		tokenVersion := claims.Version
		if tokenVersion != user.TokenVersion {
			return c.Status(401).JSON(&models.APIResponse{
				Success: false,
				Message: "Token version mismatch",
			})
		}
		jwt, err := jwtHandler.GenerateAccessToken(claims.UserId, tokenVersion)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		return c.JSON(jwt)
	}
}
