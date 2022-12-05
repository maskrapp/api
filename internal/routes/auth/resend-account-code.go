package auth

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/mailer"
	"github.com/maskrapp/backend/internal/models"
	"github.com/maskrapp/backend/internal/recaptcha"
	"github.com/maskrapp/backend/internal/utils"
	dbmodels "github.com/maskrapp/common/models"
	"gorm.io/gorm"
)

type resendAccountCodeBody struct {
	Email        string `json:"email"`
	CaptchaToken string `json:"captcha_token"`
}

// TODO: rate limit this
func ResendAccountCode(db *gorm.DB, mailer *mailer.Mailer, recaptcha *recaptcha.Recaptcha) func(*fiber.Ctx) error {
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
		if !recaptcha.ValidateCaptchaToken(body.CaptchaToken, "resend_account_code") {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Captcha failed. Try again."})
		}
		verificationRecord := &dbmodels.AccountVerification{}
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
		err = db.Model(&dbmodels.AccountVerification{}).Where("email = ?", verificationRecord.Email).Updates(&dbmodels.AccountVerification{VerificationCode: verificationCode, ExpiresAt: time.Now().Add(5 * time.Minute).Unix()}).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		err = mailer.SendUserVerificationMail(body.Email, verificationCode)
		if err != nil {
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
