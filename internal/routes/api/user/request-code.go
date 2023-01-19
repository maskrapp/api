package user

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/global"
	"github.com/maskrapp/backend/internal/models"
	"github.com/maskrapp/backend/internal/utils"
)

func RequestCode(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]string)
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		email, ok := body["email"]
		if !ok {
			return c.Status(400).SendString("Invalid Body")
		}

		userID := c.Locals("user_id").(string)

		emailRecord := &models.Email{}
		db := ctx.Instances().Gorm
		err = db.Find(emailRecord, "email = ? AND user_id = ? AND is_verified = false", email, userID).Error

		if err != nil {
			return c.Status(404).JSON(&models.APIResponse{
				Success: false,
				Message: "Could not find email",
			})
		}
		verification := &models.EmailVerification{
			EmailID:          emailRecord.Id,
			VerificationCode: utils.GenerateCode(5),
			ExpiresAt:        time.Now().Add(5 * time.Minute).Unix(),
		}
		if db.Model(&verification).Where("email_id = ?", emailRecord.Id).Updates(&verification).RowsAffected == 0 {
			err = db.Create(&verification).Error
		}
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		err = ctx.Instances().Mailer.SendVerifyMail(email, verification.VerificationCode)
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
