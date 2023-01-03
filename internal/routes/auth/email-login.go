package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/global"
	"github.com/maskrapp/backend/internal/models"
	"github.com/maskrapp/backend/internal/utils"
	dbModels "github.com/maskrapp/common/models"
)

type emailLoginBody struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	CaptchaToken string `json:"captcha_token"`
}

// TODO: harden this
func EmailLogin(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		body := &emailLoginBody{}
		err := c.BodyParser(body)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		if body.Email == "" || body.Password == "" || body.CaptchaToken == "" {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		if !ctx.Instances().Recaptcha.ValidateCaptchaToken(body.CaptchaToken, "email_login") {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Captcha failed. Try again."})
		}

		user := &dbModels.User{}
		db := ctx.Instances().Gorm

		err = db.Raw("select users.* from providers inner join users on providers.user_id = users.id where provider_name = 'email' and users.email = ?", body.Email).Limit(1).Scan(user).Error
		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Incorrect login details",
			})
		}
		if user.ID == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Incorrect login details",
			})
		}
		validPassword := utils.ComparePassword(body.Password, *user.Password)
		if !validPassword {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Incorrect login details",
			})
		}

		pair, err := ctx.Instances().JWT.CreatePair(user.ID, user.TokenVersion, true)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		return c.JSON(pair)
	}
}
