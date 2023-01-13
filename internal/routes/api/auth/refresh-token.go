package auth

import (
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/global"
	"github.com/maskrapp/backend/internal/models"
	dbmodels "github.com/maskrapp/common/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TODO: harden this
func RefreshToken(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]string)
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.SendStatus(401)
		}
		refreshToken, ok := body["refresh_token"]
		if !ok {
			return c.Status(401).SendString("Invalid Body")
		}
		claims, err := ctx.Instances().JWT.Validate(refreshToken, true)
		if err != nil {
			return c.Status(401).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid token",
			})
		}
		key := fmt.Sprintf("blacklist:%v", refreshToken)
		err = ctx.Instances().Redis.Get(ctx, key).Err()
		if err == nil {
			return c.Status(401).JSON(&models.APIResponse{
				Success: false,
				Message: "That token is blacklisted",
			})
		} else {
			if err != redis.Nil {
				logrus.Error("redis error(refresh-token): ", err)
			}
		}
		user := &dbmodels.User{}
		db := ctx.Instances().Gorm
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
		jwt, err := ctx.Instances().JWT.GenerateAccessToken(claims.UserId, tokenVersion, claims.Provider)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		return c.JSON(jwt)
	}
}
