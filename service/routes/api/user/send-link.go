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

func SendLink(db *gorm.DB, mailer *mailer.Mailer) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		val, ok := body["email"]
		if !ok {
			return c.Status(400).SendString("Invalid Body")
		}

		email := val.(string)
		userID := c.Locals("user_id").(string)

		emailRecord := &models.Email{}

		err = db.Find(emailRecord, "email = ? AND user_id = ?", email, userID).Error

		if err != nil {
			return c.Status(404).JSON(&models.APIResponse{
				Success: false,
				Message: "Could not find email",
			})
		}
		verification := &models.EmailVerification{
			EmailID:          emailRecord.Id,
			VerificationCode: uuid.New().String(),
			ExpiresAt:        time.Now().Add(30 * time.Minute).Unix(),
		}
		if db.Model(&verification).Where("email_id = ?", userID).Updates(&verification).RowsAffected == 0 {
			err = db.Create(&verification).Error
		}
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		var name = "unknown"
		err = mailer.SendVerifyMail(email, name, verification.VerificationCode)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.JSON(&models.APIResponse{
			Success: true,
			Message: "A verification code has been sent to your email",
		})
	}
}
