package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"github.com/maskrapp/api/internal/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Resetting your password is a three step process:
// 1. User requests a 6 digit code, which gets send to their email.
// 2. User sends this code back to the server. If this code is correct, the server will respond with a JWT that has a short expiry time.
// 3. User sends this JWT back to the server, along with their desired password.

// Reset is the first step in the process.
// This endpoint is accessible at: POST /auth/reset-password
func Reset(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var body struct {
			Email   string `json:"email"`
			Captcha string `json:"captcha_token"`
		}
		err := c.BodyParser(&body)
		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
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

		if !ctx.Instances().Recaptcha.ValidateCaptchaToken(body.Captcha, "reset_password") {
			return c.JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid captcha token",
			})
		}

		db := ctx.Instances().Gorm

		userRecord := &models.User{}
		err = db.Table("users").Select("users.id").Joins("INNER JOIN providers ON providers.user_id = users.id").Where("users.email = ? AND providers.provider_name = 'email'", body.Email).Find(userRecord).Error

		// For some reason, the query above doesn't trigger an ErrRecordNotFound error.
		if userRecord.ID == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "There's no account associated with this email address",
			})
		}

		if err != nil {
			logrus.Errorf("db error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
			})
		}

		passwordRecord := &models.PasswordResetVerification{}
		err = db.First(passwordRecord, "user_id = ?", userRecord.ID).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Errorf("db error:%v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
			})
		}

		// We create a new record if there isn't one, otherwise we just update the record.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			code := utils.GenerateCode(6)
			hash, err := utils.HashCode(code)
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

		// We update the record if it DOES exist

		code := utils.GenerateCode(6)
		hash, err := utils.HashCode(code)
		if err != nil {
			logrus.Errorf("hashing error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
			})
		}
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
			logrus.Errorf("mailer error: %v", err)
		}

		return c.JSON(&models.APIResponse{
			Success: true,
		})
	}
}

// VerifyPassword is the second step in the reset-password process.
// If the provided code is correct, then the server will respond with a JWT token which is redeemable in the final step.
// This endpoint is accessible at: POST /auth/reset-password/verify
func VerifyPassword(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var body struct {
			Email   string `json:"email"`
			Code    string `json:"code"`
			Captcha string `json:"captcha_token"`
		}

		err := c.BodyParser(&body)
		if err != nil {
			return c.JSON(&models.APIResponse{
				Success: false,
			})
		}
		if !utils.EmailRegex.MatchString(body.Email) || len(body.Code) != 6 {
			return c.JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}

		if !ctx.Instances().Recaptcha.ValidateCaptchaToken(body.Captcha, "verify_password") {
			return c.JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid captcha token",
			})
		}

		var record struct {
			ID               string
			UserID           string
			VerificationCode string
			TokenVersion     int
			ExpiresAt        int64
		}

		db := ctx.Instances().Gorm

		err = db.Table("password_reset_verifications").Select("password_reset_verifications.*, users.token_version").Joins("INNER JOIN users ON users.id = password_reset_verifications.user_id").Where("users.email = ?", body.Email).First(&record).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
				})
			}
			logrus.Errorf("db error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
			})
		}

		if !utils.CompareHash(body.Code, record.VerificationCode) {
			return c.Status(400).JSON(&models.APIResponse{
				Message: "Incorrect code",
				Success: false,
			})
		}

		if time.Now().Unix() > record.ExpiresAt {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Code has expired",
			})
		}

		token, err := ctx.Instances().JWT.GeneratePasswordResetToken(record.UserID, record.TokenVersion)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		err = db.Delete(&models.PasswordResetVerification{}, "id = ?", record.ID).Error
		if err != nil {
			logrus.Errorf("db error: %v", err)
		}
		return c.JSON(fiber.Map{"token": token})
	}
}

// Confirm is the final step in the process.
// This route is accessible at POST /auth/reset-password/confirm.
func Confirm(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var body struct {
			Token    string `json:"token"`
			Password string `json:"password"`
			Captcha  string `json:"captcha_token"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
			})
		}

		if body.Token == "" || body.Password == "" || body.Captcha == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}

		if !utils.IsValidPassword(body.Password) {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Password does not meet requirements",
			})
		}

		if !ctx.Instances().Recaptcha.ValidateCaptchaToken(body.Captcha, "confirm_password") {
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
				Message: "Invalid token",
				Success: false,
			})
		}

		db := ctx.Instances().Gorm
		userRecord := &models.User{}
		err = db.Table("users").Where("id = ?", claims.UserId).First(userRecord).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
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

		if utils.CompareHash(body.Password, userRecord.Password) {
			return c.Status(400).JSON(&models.APIResponse{
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
			Password:     passwordHash,
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
