package auth

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/global"
	"github.com/maskrapp/backend/internal/models"
	"github.com/maskrapp/backend/internal/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type resendAccountCodeBody struct {
	Email        string `json:"email"`
	CaptchaToken string `json:"captcha_token"`
}

func ResendAccountCode(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		body := &resendAccountCodeBody{}
		err := c.BodyParser(body)

		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		if body.Email == "" || body.CaptchaToken == "" {
			fmt.Println(body.Email)
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}

		if !utils.EmailRegex.MatchString(body.Email) {
			return c.Status(400).JSON(&models.APIResponse{Success: false,
				Message: "Invalid email"})
		}
		if !ctx.Instances().Recaptcha.ValidateCaptchaToken(body.CaptchaToken, "resend_account_code") {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Captcha failed. Try again."})
		}
		verificationRecord := &models.AccountVerification{}
		db := ctx.Instances().Gorm
		err = db.First(verificationRecord, "email = ?", body.Email).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "Invalid email",
				})
			}
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})

		}

		verificationCode := utils.GenerateCode(5)
		err = db.Model(&models.AccountVerification{}).Where("email = ?", verificationRecord.Email).Updates(&models.AccountVerification{VerificationCode: verificationCode, ExpiresAt: time.Now().Add(5 * time.Minute).Unix()}).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		err = ctx.Instances().Mailer.SendUserVerificationMail(body.Email, verificationCode)
		if err != nil {
			logrus.Errorf("mailer error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		return c.JSON(&models.APIResponse{
			Success: true,
			Message: "Check your mailbox for a code",
		})
	}
}
