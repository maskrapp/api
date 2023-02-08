package token

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Refresh is used for refreshing access token, this is done by providing the refresh token.
// This endpoint is accessible at: POST /token/refresh
func Refresh(ctx global.Context) func(*fiber.Ctx) error {
	//TODO: needs hardening
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
		user := &models.User{}
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

// Revoke is used for revoking a refresh token, the token is temporarily stored in a redis db.
// This endpoint is accessible at: POST /token/revoke
func Revoke(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var body struct {
			Token string `json:"refresh_token"`
		}
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.SendStatus(401)
		}
		claims, err := ctx.Instances().JWT.Validate(body.Token, true)
		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid token",
			})
		}
		key := fmt.Sprintf("rt-blacklist:%v", body.Token)
		cmd := ctx.Instances().Redis.Get(c.Context(), key)
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
				res := ctx.Instances().Redis.Set(c.Context(), key, 1, expiresAt.Sub(time.Now()))
				err = res.Err()
				if err != nil {
					logrus.Error("redis err(revoke-token2): ", err)
				}
			}
		}
		return c.Status(200).JSON(&models.APIResponse{
			Success: true,
		})
	}
}
