package signin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"github.com/maskrapp/api/internal/utils"
	"github.com/sirupsen/logrus"
)

// Email is used for authenticating users with the `email` provider.
// This endpoint is accessible at: POST /auth/signin/email
func Email(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var body struct {
			Email        string `json:"email"`
			Password     string `json:"password"`
			CaptchaToken string `json:"captcha_token"`
		}

		err := c.BodyParser(&body)
		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
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

		user := &models.User{}
		db := ctx.Instances().Gorm

		err = db.Raw("select users.* from providers inner join users on providers.user_id = users.id where provider_name = 'email' and users.email = ?", body.Email).Limit(1).Scan(user).Error
		if err != nil {
			logrus.Errorf("db error %v", err)
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
			})
		}
		// Check if record exists
		if user.ID == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Incorrect login details",
			})
		}
		// Check if password is correct
		if !utils.CompareHash(body.Password, user.Password) {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Incorrect login details",
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
