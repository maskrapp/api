package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/models"
	"gorm.io/gorm"
)

type record struct {
	Mask    string `json:"mask"`
	Email   string `json:"email"`
	Enabled bool   `json:"enabled"`
}

func Masks(db *gorm.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		masks := []record{}
		err := db.Table("masks").Select("masks.mask, masks.enabled, emails.email").Joins("left join emails on emails.id = masks.forward_to").Where("emails.user_id = ?", userID).Find(&masks).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.JSON(masks)
	}

}
