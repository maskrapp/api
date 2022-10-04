package email

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
		val, ok := body["code"]
		if !ok {
			return c.SendStatus(400)
		}
		code := val.(string)
		verificationModel := &dbmodels.EmailVerification{}
		err = db.Find(verificationModel, "verification_code = ?", code).Error
		if err != nil {
			return c.Status(404).JSON(&models.APIResponse{
				Success: false,
				Message: "Incorrect code",
			})
		}
		if time.Now().Unix() > verificationModel.ExpiresAt {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Code has expired",
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
