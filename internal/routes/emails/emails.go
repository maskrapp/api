package emails

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"github.com/maskrapp/api/internal/utils"
)

// Get is used for retrieving the user's emails.
// This route is accessible at: GET /emails
func Get(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		userId := c.Locals("user_id").(string)
		emails := make([]*models.Email, 0)
		err := ctx.Instances().Gorm.Where("user_id = ?", userId).Order("created_at DESC").Find(&emails).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.JSON(emails)
	}
}

// Add is used for creating a new email.
// This route is accessible at: POST /emails/new
func Add(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var body struct {
			Email string `json:"email"`
		}
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
			})
		}
		if body.Email == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		if !utils.EmailRegex.MatchString(body.Email) {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid email",
			})
		}

		domain := strings.Split(body.Email, "@")[1]
		if _, err := ctx.Instances().Domains.Get(domain); err == nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "You cannot use that email",
			})
		}

		userId := c.Locals("user_id").(string)

		var result struct {
			Found bool
		}

		db := ctx.Instances().Gorm

		db.Raw("SELECT EXISTS(SELECT 1 FROM emails WHERE user_id = ? AND email = ?) AS found",
			userId, body.Email).Scan(&result)

		if result.Found {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "That email is already registered to your account",
			})
		}

		emailRecord := &models.Email{
			UserID:     userId,
			Email:      body.Email,
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
		emailVerification := &models.EmailVerification{
			EmailID:          emailRecord.Id,
			VerificationCode: utils.GenerateCode(5),
			ExpiresAt:        time.Now().Add(30 * time.Minute).Unix(),
		}
		err = db.Create(emailVerification).Error

		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.JSON(&models.APIResponse{
			Success: true,
		})
	}
}

// Delete is used for deleting an existing email.
// This route is accessible at: DELETE /emails/{email}
func Delete(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		email := c.Params("email")
		if email == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid parameter",
			})
		}
		userID := c.Locals("user_id").(string)
		err := ctx.Instances().Gorm.Delete(&models.Email{}, "email = ? AND user_id = ?", email, userID).Error
		if err != nil {
			if strings.Contains(err.Error(), "(SQLSTATE 23503)") {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "There are still masks connected to that email.",
				})
			}
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.JSON(&models.APIResponse{
			Success: true,
			Message: "Email deleted",
		})
	}
}
