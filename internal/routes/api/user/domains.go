package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maskrapp/backend/internal/models"
	dbmodels "github.com/maskrapp/common/models"
	"gorm.io/gorm"
)

func Domains(db *gorm.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var availableDomains []dbmodels.Domain
		err := db.Find(&availableDomains).Error
		if err != nil {
			return c.Status(500).JSON(&models.APIResponse{
				Success: false,
				Message: "Something went wrong",
			})
		}
		return c.JSON(availableDomains)
	}
}
