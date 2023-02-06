package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"github.com/maskrapp/api/internal/utils"
	"gorm.io/gorm"
)

type createAccountBody struct {
	Email        string `json:"email"`
	Code         string `json:"code"`
	Password     string `json:"password"`
	CaptchaToken string `json:"captcha_token"`
}

func CreateAccount(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		body := &createAccountBody{}
		err := c.BodyParser(body)

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

		err = db.Delete(&models.AccountVerification{}, "email = ?", body.Email).Error

		//TODO: why are we returning here?
		if err != nil {
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
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		provider := &models.Provider{ID: uuid.NewString(), ProviderName: "email", UserID: user.ID}

		err = db.Create(provider).Error
		if err != nil {
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
