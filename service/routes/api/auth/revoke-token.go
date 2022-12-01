package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/jwt"
	"github.com/maskrapp/backend/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func RevokeToken(db *gorm.DB, jwtHandler *jwt.JWTHandler, redisClient *redis.Client) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]string)
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.SendStatus(401)
		}
		refreshToken, ok := body["refresh_token"]
		if !ok {
			return c.Status(401).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		claims, err := jwtHandler.Validate(refreshToken, true)
		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid token",
			})
		}
		ctx := context.Background()
		key := fmt.Sprintf("blacklist:%v", refreshToken)
		cmd := redisClient.Get(ctx, key)
		err = cmd.Err()
		if err != nil {
			if err != redis.Nil {
				logrus.Error("redis err(revoke-token): ", err)
				return c.Status(500).JSON(&models.APIResponse{
					Success: false,
					Message: "Something went wrong",
				})
			} else {
				expiresAt := time.Unix(claims.ExpiresAt, 0)
				res := redisClient.Set(ctx, key, 1, expiresAt.Sub(time.Now()))
				err = res.Err()
				if err != nil {
					logrus.Error("redis err(revoke-token2): ", err)
					// TODO: retry this.
				}
			}
		}
		return c.Status(200).JSON(&models.APIResponse{
			Success: true,
		})
	}
}
