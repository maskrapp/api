package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/models"
	"gorm.io/gorm"
)

type record struct {
	Mask              string `json:"mask"`
	Email             string `json:"email"`
	Enabled           bool   `json:"enabled"`
	MessagesReceived  int    `json:"messages_received"`
	MessagesForwarded int    `json:"messages_forwarded"`
}

func Masks(db *gorm.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		masks := []record{}
		err := db.Table("masks").Select("masks.mask, masks.enabled, masks.messages_forwarded, masks.messages_received, emails.email").Joins("inner join emails on emails.id = masks.forward_to").Where("emails.user_id = ?", userID).Order("masks.created_at DESC").Find(&masks).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong!",
			})
		}
		return c.JSON(masks)
	}

}
