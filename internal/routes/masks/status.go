package masks

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/api/internal/global"
	"github.com/maskrapp/api/internal/models"
	"gorm.io/gorm"
)

// Status is used for upating the status of an existing mask.
// This route is accessible at: PUT /masks/{mask}/status
func Status(ctx global.Context) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var body struct {
			Value bool `json:"value"`
		}
		err := c.BodyParser(&body)
		if err != nil {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		mask := c.Params("mask")
		if mask == "" {
			return c.Status(400).JSON(&models.APIResponse{
				Success: false,
				Message: "Invalid body",
			})
		}
		userID := c.Locals("user_id").(string)
		values := map[string]interface{}{
			"enabled": body.Value,
		}

		err = ctx.Instances().Gorm.Model(&models.Mask{}).Where("mask = ? and user_id = ?", mask, userID).Updates(values).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
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
