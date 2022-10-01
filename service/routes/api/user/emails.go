package user

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/models"
	"gorm.io/gorm"
)

func Emails(db *gorm.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		userId := c.Locals("user_id").(string)
		emails := make([]*models.Email, 0)
		fmt.Println(userId)
		err := db.Where("user_id = ?", userId).Find(&emails).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.JSON(emails)
	}
}
