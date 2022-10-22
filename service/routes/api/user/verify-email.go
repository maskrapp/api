package user

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/models"
	dbmodels "github.com/maskrapp/common/models"
	"gorm.io/gorm"
)

func VerifyEmail(db *gorm.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		codeVal, ok := body["code"]
		emailVal, ok2 := body["email"]

		if !ok || !ok2 {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}

		code := codeVal.(string)
		email := emailVal.(string)

		userID := c.Locals("user_id").(string)
		emailModel := &dbmodels.Email{}
		err = db.Find(emailModel, "user_id = ? AND email = ? AND is_verified = false", userID, email).Error
		if err != nil {
			return c.Status(404).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid Email",
			})
		}

		verificationModel := &dbmodels.EmailVerification{}

		err = db.Find(verificationModel, "email_id = ? AND verification_code = ?", emailModel.Id, code).Error

		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid code",
			})
		}

		if time.Now().Unix() > verificationModel.ExpiresAt {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid code",
			})
		}

		err = db.Delete(&dbmodels.EmailVerification{}, "id", verificationModel.Id).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		values := make(map[string]interface{})
		values["is_verified"] = true
		err = db.Model(&dbmodels.Email{}).Where("id = ?", verificationModel.EmailID).Updates(values).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		return c.JSON(&models.APIResponse{
			Success: true,
		})
	}
}
