package user

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/models"
	dbmodels "github.com/maskrapp/common/models"
	"gorm.io/gorm"
)

func RequestCode(db *gorm.DB, mailer *mailer.Mailer) func(*fiber.Ctx) error {
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

		emailRecord := &dbmodels.Email{}
		err = db.Find(emailRecord, "email = ? AND user_id = ? AND is_verified = false", email, userID).Error

		if err != nil {
			return c.Status(404).JSON(&models.APIResponse{
				Success: false,
				Message: "Could not find email",
			})
		}
		verification := &dbmodels.EmailVerification{
			EmailID:          emailRecord.Id,
			VerificationCode: generateCode(),
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
		err = mailer.SendVerifyMail(email, verification.VerificationCode)
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

var set = "1234567890"

func generateCode() string {
	sb := strings.Builder{}
	sb.Grow(5)
	for i := 0; i < 5; i++ {
		sb.WriteByte(set[rand.Intn(len(set))])
	}
	return sb.String()

}
