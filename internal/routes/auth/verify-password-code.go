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

type verifyPasswordCodeBody struct {
	Email   string `json:"email"`
	Code    string `json:"code"`
	Captcha string `json:"captcha_token"`
}

// This route verifies the confirmation code which is stored in the password_reset_verifications table.
// If the user sends a code that is identical to the hash saved in the database, the server responds with a JWT which can be redeemed at the /change-password route.
func VerifyPasswordCode(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := &verifyPasswordCodeBody{}
		err := c.BodyParser(body)
		if err != nil {
			return c.JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid JSON",
			})
		}
		if !utils.EmailRegex.MatchString(body.Email) || len(body.Code) != 6 {
			return c.JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}

		if !ctx.Instances().Recaptcha.ValidateCaptchaToken(body.Captcha, "verify_password_code") {
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
