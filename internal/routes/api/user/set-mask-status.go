package user

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/models"
	dbmodels "github.com/maskrapp/common/models"
	"gorm.io/gorm"
)

type setMaskBody struct {
	Mask  string `json:"mask"`
	Value bool   `json:"value"`
}

func SetMaskStatus(db *gorm.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		b := &setMaskBody{}
		err := json.Unmarshal(c.Body(), &b)
		if err != nil {
			return c.SendStatus(500)
		}
		if b.Mask == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid Body",
			})
		}
		userID := c.Locals("user_id").(string)
		values := map[string]interface{}{
			"enabled": b.Value,
		}

		err = db.Model(&dbmodels.Mask{}).Where("mask = ? and user_id = ?", b.Mask, userID).Updates(values).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(400).JSON(&models.APIResponse{
					Success: false,
					Message: "You don't own that mask",
				})
			}
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}

		return c.Status(200).JSON(&models.APIResponse{
			Success: true,
		})
	}
}
