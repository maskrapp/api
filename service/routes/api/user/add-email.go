package user

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/mailer"
	"github.com/maskrapp/backend/models"
	dbmodels "github.com/maskrapp/common/models"
	"gorm.io/gorm"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func AddEmail(db *gorm.DB, mailer *mailer.Mailer) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		body := make(map[string]interface{})
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.SendStatus(500)
		}
		val, ok := body["email"]
		if !ok {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		email := val.(string)
		if !emailRegex.MatchString(email) {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid email",
			})
		}

		if strings.Split(email, "@")[1] == "relay.maskr.app" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid email",
			})
		}

		userId := c.Locals("user_id").(string)

		var result struct {
			Found bool
		}

		db.Raw("SELECT EXISTS(SELECT 1 FROM emails WHERE user_id = ? AND email = ?) AS found",
			userId, email).Scan(&result)

		if result.Found {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "That email is already registered to your account",
			})
		}

		emailRecord := &dbmodels.Email{
			UserID:     userId,
			Email:      email,
			IsVerified: false,
			IsPrimary:  false,
		}

		err = db.Create(emailRecord).Error

		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		emailVerification := &dbmodels.EmailVerification{
			EmailID:          emailRecord.Id,
			VerificationCode: generateCode(),
			ExpiresAt:        time.Now().Add(30 * time.Minute).Unix(),
		}
		err = db.Create(emailVerification).Error

		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		err = mailer.SendVerifyMail(email, emailVerification.VerificationCode)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Could not send verification email, try again later.",
			})
		}
		return c.JSON(&models.APIResponse{
			Success: true,
		})
	}
}
