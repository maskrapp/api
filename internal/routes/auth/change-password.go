package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"github.com/maskrapp/api/internal/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type changePasswordBody struct {
	Token    string `json:"token"`
	Password string `json:"password"`
	Captcha  string `json:"captcha_token"`
}

func ChangePassword(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		body := &changePasswordBody{}
		if err := c.BodyParser(body); err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid JSON",
			})
		}

		if body.Token == "" || body.Password == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
			})
		}

		if !utils.IsValidPassword(body.Password) {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Password does not meet requirements",
			})
		}

		if !ctx.Instances().Recaptcha.ValidateCaptchaToken(body.Captcha, "change_password") {
			return c.JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid captcha token",
			})
		}

		claims, err := ctx.Instances().JWT.ValidatePasswordResetToken(body.Token)
		if err != nil {
			if strings.Contains(err.Error(), "is expired by") {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "Your session has expired",
				})
			}
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
			})
		}

		db := ctx.Instances().Gorm
		userRecord := &models.User{}
		err = db.Table("users").Where("id = ?", claims.UserId).First(userRecord).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				logrus.Errorf("unexpected database error: expected record with id %v to exist", claims.UserId)
			} else {
				logrus.Errorf("db error: %v", err)
			}
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
			})
		}

		if claims.Version != userRecord.TokenVersion {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Token version mismatch",
			})
		}

		if utils.CompareHash(body.Password, *userRecord.Password) {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "You cannot use your previous password",
			})
		}

		passwordHash, err := utils.HashPassword(body.Password)
		if err != nil {
			logrus.Errorf("hashing error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		updatedModel := &models.User{
			TokenVersion: userRecord.TokenVersion + 1, //Invalidates every active session
			Password:     &passwordHash,
		}

		err = db.Model(&models.User{}).Where("id = ?", userRecord.ID).Updates(updatedModel).Error
		if err != nil {
			logrus.Errorf("db error: %v")
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		return c.Status(200).JSON(&models.APIResponse{
			Success: true,
		})
	}
}
