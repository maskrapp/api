package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/jwt"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/models"
	"github.com/maskrapp/backend/recaptcha"
	"github.com/maskrapp/backend/utils"
	dbmodels "github.com/maskrapp/common/models"
	"gorm.io/gorm"
)

type createAccountCodeBody struct {
	Email        string `json:"email"`
	CaptchaToken string `json:"captcha_token"`
}

func CreateAccountCode(db *gorm.DB, jwtHandler *jwt.JWTHandler, mailer *mailer.Mailer, recaptcha *recaptcha.Recaptcha) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := &createAccountCodeBody{}
		err := c.BodyParser(body)
		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		if body.Email == "" || body.CaptchaToken == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}

		if !utils.EmailRegex.MatchString(body.Email) {
			return c.Status(400).JSON(&models.APIResponse{Success: false,
				Message: "Invalid email"})
		}

		if !recaptcha.ValidateCaptchaToken(body.CaptchaToken, "create_account_code") {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Captcha failed. Try again."})
		}

		data := make(map[string]interface{}, 0)

		err = db.Raw("select * from providers inner join users on providers.user_id = users.id where provider_name = 'email' and users.email = ?", body.Email).Limit(1).Scan(data).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		if len(data) > 0 {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Email is already in use",
			})
		}

		verificationRecord := &dbmodels.AccountVerification{}

		err = db.First(verificationRecord, "email = ?", body.Email).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		if err == gorm.ErrRecordNotFound {
			verificationCode := utils.GenerateCode(5)
			err = db.Create(&dbmodels.AccountVerification{Email: body.Email, VerificationCode: verificationCode, ExpiresAt: time.Now().Add(time.Minute * 5).Unix()}).Error
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
		} else {
			if time.Now().Unix() > verificationRecord.ExpiresAt {
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
			}
		}
		return c.JSON(&models.APIResponse{
			Success: true,
		})
	}
}
