package user

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/models"
	"gorm.io/gorm"
)

func AddEmail(db *gorm.DB, mailer *mailer.Mailer) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.SendStatus(500)
		}
		val, ok := body["email"]
		if !ok {
			return c.Status(400).SendString("Invalid Body")
		}
		email := val.(string)
		// TODO: validate email with regex

		userId := c.Locals("user_id").(string)

		var result struct {
			Found bool
		}

		db.Raw("SELECT EXISTS(SELECT 1 FROM emails WHERE user_id = ? AND email = ?) AS found",
			userId, email).Scan(&result)

		if result.Found {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "That email is already registered to your account",
			})
		}

		emailRecord := &models.Email{
			UserID:     userId,
			Email:      email,
			IsVerified: false,
			IsPrimary:  false,
		}

		err = db.Create(emailRecord).Error

		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		emailVerification := &models.EmailVerification{
			EmailID:          emailRecord.Id,
			VerificationCode: uuid.New().String(),
			ExpiresAt:        time.Now().Add(30 * time.Minute).Unix(),
		}
		err = db.Create(emailVerification).Error

		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		//TODO: include name in jwt so we can use it here
		var name = "unknown"
		err = mailer.SendVerifyMail(email, name, emailVerification.VerificationCode)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Could not send verification email, try again later.",
			})
		}
		return c.JSON(&models.APIResponse{
			Success: true,
		})
	}
}
