package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/global"
	"github.com/maskrapp/backend/internal/models"
)

func Emails(ctx global.Context) func(*fiber.Ctx) error {
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
