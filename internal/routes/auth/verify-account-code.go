package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"gorm.io/gorm"
)

type verifyAccountCodeBody struct {
	Email        string `json:"email"`
	Code         string `json:"code"`
	CaptchaToken string `json:"captcha_token"`
}

func VerifyAccountCode(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := &verifyAccountCodeBody{}

		err := c.BodyParser(body)

		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}

		if body.Email == "" || body.Code == "" || body.CaptchaToken == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		if !ctx.Instances().Recaptcha.ValidateCaptchaToken(body.CaptchaToken, "verify_account_code") {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Captcha failed. Try again."})
		}

		db := ctx.Instances().Gorm

		verificationRecord := &models.AccountVerification{}
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
