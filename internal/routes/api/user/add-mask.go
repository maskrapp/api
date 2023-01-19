package user

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/global"
	"github.com/maskrapp/backend/internal/models"
	"github.com/maskrapp/backend/internal/utils"
	"gorm.io/gorm"
)

type addMaskBody struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Domain string `json:"domain"`
}

func AddMask(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		b := &addMaskBody{}
		err := json.Unmarshal(c.Body(), &b)
		if err != nil {
			return c.SendStatus(500)
		}
		if b.Email == "" || b.Domain == "" || b.Name == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		fullEmail := strings.ToLower(b.Name + "@" + b.Domain)
		if !utils.EmailRegex.MatchString(fullEmail) {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid masked email address",
			})
		}

		_, err = ctx.Instances().Domains.Get(b.Domain)
		db := ctx.Instances().Gorm

		//TODO: check if user can use the domain with their plan
		if err != nil {
			return c.Status(404).JSON(&models.APIResponse{
				Success: false,
				Message: "Domain not found",
			})
		}

		userID := c.Locals("user_id").(string)

		var result struct {
			Found bool
		}

		db.Raw("SELECT EXISTS(SELECT 1 FROM masks WHERE mask = ?) AS found",
			fullEmail).Scan(&result)
		if result.Found {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "That mask already exists",
			})
		}
		emailRecord := &models.Email{}

		err = db.Find(emailRecord, "email = ? AND user_id = ?", b.Email, userID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "You don't own that email",
				})

			}
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		if !emailRecord.IsVerified {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Email is not verified",
			})
		}
		maskRecord := &models.Mask{
			Mask:      fullEmail,
			Enabled:   true,
			ForwardTo: emailRecord.Id,
			UserID:    userID,
		}

		err = db.Create(&maskRecord).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		return c.Status(200).JSON(&models.APIResponse{
			Success: true,
			Message: "Created mask",
		})
	}
}
