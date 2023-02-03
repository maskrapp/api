package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"github.com/maskrapp/api/internal/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type createPasswordCodeBody struct {
	Email   string `json:"email"`
	Captcha string `json:"captcha_token"`
}

// Resetting your password is a three step process:
// 1. User requests a 6 digit code, which gets send to their email. (this route)
// 2. User sends this code back to the server. If this code is correct, the server will respond with a JWT that has a short expiry time.
// 3. User sends this JWT back to the server, along with their desired password.

func CreatePasswordCode(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := &createPasswordCodeBody{}
		err := c.BodyParser(body)
		if err != nil {
			return c.JSON(&models.APIResponse{
				Success: false,
			})
		}
		if body.Email == "" || body.Captcha == "" {
			return c.JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		if !utils.EmailRegex.MatchString(body.Email) {
			return c.JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid email",
			})
		}

		if !ctx.Instances().Recaptcha.ValidateCaptchaToken(body.Captcha, "create_password_code") {
			return c.JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid captcha token",
			})
		}

		db := ctx.Instances().Gorm

		userRecord := &models.User{}

		err = db.Table("users").Select("users.id").Joins("INNER JOIN providers ON providers.user_id = users.id").Where("users.email = ? AND providers.provider_name = 'email'", body.Email).Find(userRecord).Error

		// For some reason, the query above doesn't trigger a ErrRecordNotFound error.
		if userRecord.ID == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "There's no account associated with this email address",
			})
		}

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
				})
			}
			logrus.Errorf("db error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
			})
		}

		passwordRecord := &models.PasswordResetVerification{}
		err = db.First(passwordRecord, "user_id = ?", userRecord.ID).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
			})
		}

		// We create a new record if there isn't one, otherwise we just update the record.
		if err == gorm.ErrRecordNotFound {
			code := utils.GenerateCode(6)
			hash, err := utils.HashCode(code)
			if err != nil {
				logrus.Errorf("hashing error: %v", err)
				return c.Status(500).JSON(&models.APIResponse{
					Success: false,
				})
			}

			if err != nil {
				logrus.Errorf("hashing error: %v", err)
				return c.Status(500).JSON(&models.APIResponse{
					Success: false,
				})
			}

			err = db.Create(&models.PasswordResetVerification{
				UserID:           userRecord.ID,
				VerificationCode: hash,
				ExpiresAt:        time.Now().Add(5 * time.Minute).Unix(),
			}).Error

			if err != nil {
				logrus.Errorf("db error: %v", err)
				return c.Status(500).JSON(&models.APIResponse{
					Success: false,
				})
			}

			err = ctx.Instances().Mailer.SendPasswordCodeMail(body.Email, code)
			if err != nil {
				logrus.Errorf("mailer error: %v", err)
				return c.Status(500).JSON(&models.APIResponse{
					Success: false,
				})
			}
			return c.JSON(&models.APIResponse{
				Success: true,
			})
		}

		code := utils.GenerateCode(6)
		hash, err := utils.HashCode(code)
		if err != nil {
			logrus.Errorf("hashing error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
			})
		}
		// we update the record if it DOES exist.
		updatedModel := &models.AccountVerification{VerificationCode: hash, ExpiresAt: time.Now().Add(5 * time.Minute).Unix()}
		err = db.Table("password_reset_verifications").Where("user_id = ?", userRecord.ID).Updates(&updatedModel).Error
		if err != nil {
			logrus.Errorf("database error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
			})
		}

		err = ctx.Instances().Mailer.SendPasswordCodeMail(body.Email, code)
		if err != nil {
			ctx.Instances().Mailer.SendPasswordCodeMail(body.Email, code)
		}

		return c.JSON(&models.APIResponse{
			Success: true,
		})
	}
}
