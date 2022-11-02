package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/models"
	dbmodels "github.com/maskrapp/common/models"
	"gorm.io/gorm"
)

type verifyAccountCodeBody struct {
	Email string
	Code  string
}

func VerifyAccountCode(db *gorm.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := &verifyAccountCodeBody{}

		err := c.BodyParser(body)

		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}

		if body.Email == "" || body.Code == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		verificationRecord := &dbmodels.AccountVerification{}
		err = db.First(verificationRecord, "email = ? AND verification_code = ? ", body.Email, body.Code).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "Invalid code",
				})
			}
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		return c.JSON(&models.APIResponse{
			Success: true,
			Message: "Code is valid",
		})
	}
}
