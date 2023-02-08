package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"github.com/maskrapp/api/internal/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// /auth/signup
// /auth/signup/verify
// /auth/signup/create

// Signup is the first step in the signup process
// This endpoint is accessible at: /auth/signup
func Signup(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var body struct {
			Email        string `json:"email"`
			CaptchaToken string `json:"captcha_token"`
		}
		err := c.BodyParser(&body)
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

		if !ctx.Instances().Recaptcha.ValidateCaptchaToken(body.CaptchaToken, "create_account_code") {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Captcha failed. Try again."})
		}

		data := make(map[string]interface{}, 0)

		db := ctx.Instances().Gorm

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

		verificationRecord := &models.AccountVerification{}

		err = db.First(verificationRecord, "email = ?", body.Email).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			logrus.Errorf("db error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		//  We create a new record if there isn't one. Otherwise we update the existing record.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			verificationCode := utils.GenerateCode(5)
			err = db.Create(&models.AccountVerification{Email: body.Email, VerificationCode: verificationCode, ExpiresAt: time.Now().Add(time.Minute * 5).Unix()}).Error
			if err != nil {
				logrus.Errorf("db error %v", err)
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
			})
		}

		// We don't need to update the record if it hasn't expired yet.
		if time.Now().Unix() < verificationRecord.ExpiresAt {
			return c.Status(200).JSON(&models.APIResponse{
				Success: true,
			})
		}

		verificationCode := utils.GenerateCode(5)
		updatedModel := &models.AccountVerification{VerificationCode: verificationCode, ExpiresAt: time.Now().Add(5 * time.Minute).Unix()}
		err = db.Model(&models.AccountVerification{}).Where("email = ?", verificationRecord.Email).Updates(updatedModel).Error
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
		})
	}
}

// Resend code will re-generate a verification code.
// This endpoint is accessible at: /auth/signup/resend-code
func ResendCode(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var body struct {
			Email        string `json:"email"`
			CaptchaToken string `json:"captcha_token"`
		}
		err := c.BodyParser(&body)

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

// VerifySignup is the second step in the signup process, but is optional. This route solely exists to give more feedback to the user on the frontend.
// This route is accessible at /auth/signup/verify
func VerifySignup(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var body struct {
			Email        string `json:"email"`
			Code         string `json:"code"`
			CaptchaToken string `json:"captcha_token"`
		}

		err := c.BodyParser(&body)

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
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "Invalid code",
				})
			}
			logrus.Errorf("db error: %v", err)
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

// Create is the final step in the signup process
// This endpoint is accesible at: /auth/signup/create
func Create(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var body struct {
			Email        string `json:"email"`
			Code         string `json:"code"`
			Password     string `json:"password"`
			CaptchaToken string `json:"captcha_token"`
		}

		err := c.BodyParser(&body)

		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}

		if body.Email == "" || body.Code == "" || body.Password == "" || body.CaptchaToken == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}

		if !utils.IsValidPassword(body.Password) {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Password does not meet requirements",
			})
		}

		if !ctx.Instances().Recaptcha.ValidateCaptchaToken(body.CaptchaToken, "create_account") {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Captcha failed. Try again.",
			})
		}

		verificationRecord := &models.AccountVerification{}
		db := ctx.Instances().Gorm
		err = db.First(verificationRecord, "email = ? AND verification_code = ? ", body.Email, body.Code).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "Invalid code",
				})
			}
			logrus.Errorf("db error :%v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		err = db.Delete(&models.AccountVerification{}, "email = ?", body.Email).Error

		if err != nil {
			logrus.Errorf("db error :%v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		hashedPassword, err := utils.HashPassword(body.Password)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		user := &models.User{ID: uuid.NewString(), Role: 0, Password: hashedPassword, Email: body.Email}
		err = db.Create(user).Error
		if err != nil {
			logrus.Errorf("db error :%v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		provider := &models.Provider{ID: uuid.NewString(), ProviderName: "email", UserID: user.ID}

		err = db.Create(provider).Error
		if err != nil {
			logrus.Errorf("db error :%v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		err = db.Create(&models.Email{
			UserID:     user.ID,
			IsPrimary:  true,
			IsVerified: true,
			Email:      body.Email,
		}).Error
		if err != nil {
			logrus.Errorf("db error :%v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}

		pair, err := ctx.Instances().JWT.CreatePair(user.ID, user.TokenVersion, "email")

		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		return c.JSON(pair)
	}
}
