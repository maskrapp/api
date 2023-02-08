package masks

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"github.com/maskrapp/api/internal/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type mask struct {
	Mask              string `json:"mask"`
	Email             string `json:"email"`
	Enabled           bool   `json:"enabled"`
	MessagesReceived  int    `json:"messages_received"`
	MessagesForwarded int    `json:"messages_forwarded"`
}

// Get is used for retrieving the user's masks.
// This route is accessible at: GET /masks
func Get(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		masks := []mask{}
		err := ctx.Instances().Gorm.Table("masks").Select("masks.mask, masks.enabled, masks.messages_forwarded, masks.messages_received, emails.email").Joins("inner join emails on emails.id = masks.forward_to").Where("emails.user_id = ?", userID).Order("masks.created_at DESC").Find(&masks).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.JSON(masks)
	}
}

// Add is used for creating a new mask.
// This route is accessible at: POST /masks/new
func Add(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var body struct {
			Name   string `json:"name"`
			Email  string `json:"email"`
			Domain string `json:"domain"`
		}
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
			})
		}
		if body.Email == "" || body.Domain == "" || body.Name == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		fullEmail := strings.ToLower(body.Name + "@" + body.Domain)
		if !utils.EmailRegex.MatchString(fullEmail) {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid email address",
			})
		}

		_, err = ctx.Instances().Domains.Get(body.Domain)
		//TODO: check if user can use the domain with their plan
		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid domain",
			})
		}

		userID := c.Locals("user_id").(string)

		var result struct {
			Found bool
		}

		db := ctx.Instances().Gorm

		db.Raw("SELECT EXISTS(SELECT 1 FROM masks WHERE mask = ?) AS found",
			fullEmail).Scan(&result)
		if result.Found {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "That mask already exists",
			})
		}
		emailRecord := &models.Email{}

		err = db.Find(emailRecord, "email = ? AND user_id = ?", body.Email, userID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "You don't own that email",
				})
			}
			logrus.Errorf("db error: %v", err)
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
		return c.JSON(fiber.Map{"mask": maskRecord.Mask})
	}
}

// Delete is used for deleting an existing mask.
// This route is accessible at: DELETE /masks/{id}
func Delete(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		mask := c.Params("mask")
		if mask == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Missing mask parameter",
			})
		}

		if !utils.EmailRegex.MatchString(mask) {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Please provide a valid mask value",
			})
		}

		userID := c.Locals("user_id").(string)
		err := ctx.Instances().Gorm.Delete(&models.Mask{}, "mask = ? AND user_id = ?", mask, userID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "Please provide a valid mask value",
				})
			}
			logrus.Errorf("db error: %v", err)
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.JSON(&models.APIResponse{
			Success: true,
			Message: "Mask deleted",
		})
	}
}
